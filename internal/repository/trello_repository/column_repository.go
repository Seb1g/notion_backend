package trello_repository

import (
	"anemone_notes/internal/model/trello_model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrColumnNotFound    = errors.New("column not found")
	ErrColumnMoveFailed  = errors.New("column move failed")
	ErrBoardAccessDenied = errors.New("board not found or access denied")
	ErrNoColumnsFound    = errors.New("no columns found")
)

type ColumnRepo struct {
	DB *sqlx.DB
}

func NewColumnRepo(db *sqlx.DB) *ColumnRepo {
	return &ColumnRepo{DB: db}
}

func (r *ColumnRepo) CreateColumn(ctx context.Context, boardID, columnTitle string) (*trello_model.Column, error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	columnID := uuid.New().String()

	var newPosition int
	qPos := `SELECT COALESCE(MAX(position), 0) + 1 AS new_position FROM columns WHERE board_id = $1`
	err = tx.GetContext(ctx, &newPosition, qPos, boardID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get max position: %w", err)
	}

	column := &trello_model.Column{}
	qInsert := `INSERT INTO columns (id, column_title, board_id, position) VALUES ($1, $2, $3, $4) RETURNING *;`
	err = tx.QueryRowxContext(ctx, qInsert, columnID, columnTitle, boardID, newPosition).StructScan(column)
	if err != nil {
		return nil, fmt.Errorf("failed to insert column: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return column, nil
}

func (r *ColumnRepo) DeleteColumn(ctx context.Context, boardID, columnID string) error {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	var deletedPosition int
	qFindPos := `SELECT position FROM columns WHERE id = $1 AND board_id = $2 FOR UPDATE;` // FOR UPDATE для транзакции
	err = tx.GetContext(ctx, &deletedPosition, qFindPos, columnID, boardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrColumnNotFound
		}
		return fmt.Errorf("failed to get column position: %w", err)
	}

	qDelete := `DELETE FROM columns WHERE id = $1 AND board_id = $2;`
	result, err := tx.ExecContext(ctx, qDelete, columnID, boardID)
	if err != nil {
		return fmt.Errorf("failed to delete column: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrColumnNotFound
	}

	qUpdatePos := `UPDATE columns SET position = position - 1 WHERE board_id = $1 AND position > $2;`
	_, err = tx.ExecContext(ctx, qUpdatePos, boardID, deletedPosition)
	if err != nil {
		return fmt.Errorf("failed to update positions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (r *ColumnRepo) RenameColumn(ctx context.Context, boardID, columnID, newName string) (*trello_model.Column, error) {
	q := `UPDATE columns SET column_title = $1 WHERE id = $2 AND board_id = $3 RETURNING *;`
	var column trello_model.Column
	err := r.DB.QueryRowxContext(ctx, q, newName, columnID, boardID).StructScan(&column)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrColumnNotFound
		}
		return nil, err
	}
	return &column, nil
}
