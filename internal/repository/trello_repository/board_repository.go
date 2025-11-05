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
	ErrBoardNotFound      = errors.New("board not found")
	ErrBoardUpdateFailed  = errors.New("board update failed")
	ErrBoardDeleteFailed  = errors.New("board delete failed")
	ErrColumnCreateFailed = errors.New("column creation failed")
	ErrCardCreateFailed   = errors.New("card creation failed")
)

type BoardRepo struct {
	DB *sqlx.DB
}

func NewBoardRepo(db *sqlx.DB) *BoardRepo {
	return &BoardRepo{DB: db}
}

func (r *BoardRepo) CreateBoard(ctx context.Context, title string, userID int) (*trello_model.Board, error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	boardID := uuid.New().String()
	board := &trello_model.Board{ID: boardID, Title: title, UserID: userID}

	qBoard := `INSERT INTO boards (id, title, user_id) VALUES ($1, $2, $3) RETURNING *;`
	err = tx.QueryRowxContext(ctx, qBoard, boardID, title, userID).StructScan(board)
	if err != nil {
		return nil, fmt.Errorf("failed to create board: %w", err)
	}

	defaultColumns := []trello_model.DefaultColumnData{
		{Title: "Need to do", Cards: []string{"Task 1", "Task 2", "Task 3"}},
		{Title: "In progress", Cards: []string{"Task A", "Task B", "Task C"}},
		{Title: "Ready", Cards: []string{"Task X", "Task Y", "Task Z"}},
	}

	for i, colData := range defaultColumns {
		columnID := uuid.New().String()
		position := i + 1

		qColumn := `INSERT INTO columns (id, column_title, board_id, position) VALUES ($1, $2, $3, $4);`
		_, err = tx.ExecContext(ctx, qColumn, columnID, colData.Title, boardID, position)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create column: %v", ErrColumnCreateFailed, err)
		}

		for j, cardContent := range colData.Cards {
			cardID := uuid.New().String()
			cardPosition := j + 1
			qCard := `INSERT INTO cards (id, content, column_id, position) VALUES ($1, $2, $3, $4);`
			_, err = tx.ExecContext(ctx, qCard, cardID, cardContent, columnID, cardPosition)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to create card: %v", ErrCardCreateFailed, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return board, nil
}

func (r *BoardRepo) GetOneUserBoard(ctx context.Context, boardID string) (*trello_model.BoardWithColumns, error) {
	var board trello_model.Board
	err := r.DB.GetContext(ctx, &board, "SELECT * FROM boards WHERE id = $1", boardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}

	var columns []*trello_model.Column
	err = r.DB.SelectContext(ctx, &columns, "SELECT id, column_title, position FROM columns WHERE board_id = $1 ORDER BY position", boardID)
	if err != nil {
		return nil, err
	}

	if len(columns) > 0 {
		columnIDs := make([]string, len(columns))
		columnMap := make(map[string]*trello_model.Column)
		for i, col := range columns {
			columnIDs[i] = col.ID
			columnMap[col.ID] = col
		}

		query, args, err := sqlx.In("SELECT id, content, column_id, position FROM cards WHERE column_id IN (?) ORDER BY column_id, position", columnIDs)
		if err != nil {
			return nil, err
		}
		query = r.DB.Rebind(query)

		var cards []*trello_model.Card
		err = r.DB.SelectContext(ctx, &cards, query, args...)
		if err != nil {
			return nil, err
		}

		for _, card := range cards {
			if col, ok := columnMap[card.ColumnID]; ok {
				col.Cards = append(col.Cards, card)
			}
		}
	}

	return &trello_model.BoardWithColumns{
		ID:      board.ID,
		Title:   board.Title,
		Columns: columns,
	}, nil
}

func (r *BoardRepo) GetAllUserBoards(ctx context.Context, userID int) ([]*trello_model.Board, error) {
	var boards []*trello_model.Board

	q := `SELECT * FROM boards WHERE user_id = $1;`
	err := r.DB.SelectContext(ctx, &boards, q, userID)
	if err != nil {
		return nil, err
	}
	return boards, nil
}

func (r *BoardRepo) DeleteBoard(ctx context.Context, boardID string) error {
	q := `DELETE FROM boards WHERE id = $1 RETURNING id;`
	result, err := r.DB.ExecContext(ctx, q, boardID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBoardNotFound
	}
	return nil
}

func (r *BoardRepo) RenameBoard(ctx context.Context, boardID string, newName string) (*trello_model.Board, error) {
	q := `UPDATE boards SET title = $1, updated_at = NOW() WHERE id = $2 RETURNING *;`
	var board trello_model.Board

	err := r.DB.QueryRowxContext(ctx, q, newName, boardID).StructScan(&board)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}
	return &board, nil
}

func (r *BoardRepo) UpdateBoard(ctx context.Context, boardID string, userID int, boardData []*trello_model.Column) (err error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
        DELETE FROM cards 
        WHERE column_id IN (SELECT id FROM columns WHERE board_id = $1);
    `, boardID)
	if err != nil {
		return fmt.Errorf("%w: failed to delete old cards: %v", ErrBoardUpdateFailed, err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM columns WHERE board_id = $1", boardID)
	if err != nil {
		return fmt.Errorf("%w: failed to delete old columns: %v", ErrBoardUpdateFailed, err)
	}

	for i, col := range boardData {
		var columnID string
		err = tx.GetContext(ctx, &columnID, `
            INSERT INTO columns (id, board_id, column_title, position) 
            VALUES ($1, $2, $3, $4) 
            RETURNING id;
        `, col.ID, boardID, col.Title, i+1)

		if err != nil {
			return fmt.Errorf("%w: failed to insert column %s: %v", ErrBoardUpdateFailed, col.ID, err)
		}

		for j, card := range col.Cards {
			_, err = tx.ExecContext(ctx, `
                INSERT INTO cards (id, column_id, content, position) 
                VALUES ($1, $2, $3, $4);
            `, card.ID, columnID, card.Content, j+1)

			if err != nil {
				return fmt.Errorf("%w: failed to insert card %s: %v", ErrBoardUpdateFailed, card.ID, err)
			}
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("transaction commit failed: %w", commitErr)
	}

	return nil
}

func (r *BoardRepo) GetBoardOwnerID(ctx context.Context, boardID string) (int, error) {
	var ownerID int
	query := `SELECT user_id FROM boards WHERE id = $1`

	err := r.DB.GetContext(ctx, &ownerID, query, boardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrBoardNotFound
		}
		return 0, fmt.Errorf("failed to get board owner ID: %w", err)
	}
	return ownerID, nil
}

func (r *BoardRepo) GetBoardOwnerIDByColumnID(ctx context.Context, columnID string) (int, error) {
	var ownerID int
	query := `
        SELECT b.user_id 
        FROM boards b 
        JOIN columns c ON b.id = c.board_id 
        WHERE c.id = $1;
    `
		
	err := r.DB.GetContext(ctx, &ownerID, query, columnID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrBoardNotFound
		}
		return 0, fmt.Errorf("failed to get board owner ID by column ID: %w", err)
	}
	return ownerID, nil
}
