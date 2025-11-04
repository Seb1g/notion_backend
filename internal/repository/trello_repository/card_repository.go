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
	ErrCardNotFound          = errors.New("card not found")
	ErrCardMoveFailed        = errors.New("card move failed")
	ErrColumnNotFoundForCard = errors.New("column not found for card operation")
)

type CardRepo struct {
	DB *sqlx.DB
}

func NewCardRepo(db *sqlx.DB) *CardRepo {
	return &CardRepo{DB: db}
}

func (r *CardRepo) CreateCard(ctx context.Context, columnID, cardTitle string) (*trello_model.Card, error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	cardID := uuid.New().String()

	var newPosition int
	qPos := `SELECT COALESCE(MAX(position), 0) + 1 FROM cards WHERE column_id = $1`
	err = tx.GetContext(ctx, &newPosition, qPos, columnID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get max position: %w", err)
	}

	card := &trello_model.Card{}
	qInsert := `INSERT INTO cards (id, content, column_id, position) VALUES ($1, $2, $3, $4) RETURNING *;`
	err = tx.QueryRowxContext(ctx, qInsert, cardID, cardTitle, columnID, newPosition).StructScan(card)
	if err != nil {
		return nil, fmt.Errorf("failed to insert card: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return card, nil
}

func (r *CardRepo) DeleteCard(ctx context.Context, columnID, cardID string) error {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	var deletedPosition int
	qFindPos := `SELECT position FROM cards WHERE id = $1 AND column_id = $2 FOR UPDATE;`
	err = tx.GetContext(ctx, &deletedPosition, qFindPos, cardID, columnID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCardNotFound
		}
		return fmt.Errorf("failed to get card position: %w", err)
	}

	qDelete := `DELETE FROM cards WHERE id = $1 AND column_id = $2;`
	result, err := tx.ExecContext(ctx, qDelete, cardID, columnID)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrCardNotFound
	}

	qUpdatePos := `UPDATE cards SET position = position - 1 WHERE column_id = $1 AND position > $2;`
	_, err = tx.ExecContext(ctx, qUpdatePos, columnID, deletedPosition)
	if err != nil {
		return fmt.Errorf("failed to update positions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (r *CardRepo) RenameCard(ctx context.Context, columnID, cardID, newName string) (*trello_model.Card, error) {
	q := `UPDATE cards SET content = $1 WHERE id = $2 AND column_id = $3 RETURNING *;`
	var card trello_model.Card
	err := r.DB.QueryRowxContext(ctx, q, newName, cardID, columnID).StructScan(&card)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCardNotFound
		}
		return nil, err
	}
	return &card, nil
}