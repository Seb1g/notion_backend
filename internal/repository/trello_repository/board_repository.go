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

	// 1. Создание доски
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

func (r *BoardRepo) GetOneUserBoard(ctx context.Context, boardID, userID string) (*trello_model.BoardWithColumns, error) {
	var board trello_model.Board
	err := r.DB.GetContext(ctx, &board, "SELECT * FROM boards WHERE id = $1 AND user_id = $2", boardID, userID)
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

func (r *BoardRepo) GetAllUserBoards(ctx context.Context, userID string) ([]*trello_model.Board, error) {
	var boards []*trello_model.Board
	q := `SELECT * FROM boards WHERE user_id = $1;`
	err := r.DB.SelectContext(ctx, &boards, q, userID)
	if err != nil {
		return nil, err
	}
	return boards, nil
}

func (r *BoardRepo) DeleteBoard(ctx context.Context, boardID string, userID int) error {
	q := `DELETE FROM boards WHERE id = $1 AND user_id = $2 RETURNING id;`
	result, err := r.DB.ExecContext(ctx, q, boardID, userID)
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

func (r *BoardRepo) RenameBoard(ctx context.Context, boardID string, userID int, newName string) (*trello_model.Board, error) {
	q := `UPDATE boards SET title = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3 RETURNING *;`
	var board trello_model.Board
	err := r.DB.QueryRowxContext(ctx, q, newName, boardID, userID).StructScan(&board)
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
	// Гарантируем откат
	defer func() {
		if r := recover(); r != nil || err != nil {
			_ = tx.Rollback()
			if r != nil {
				panic(r)
			}
		}
	}()

	// 1. Проверка владения
	// ... (код проверки) ...

	// --- 🔑 ФИНАЛЬНАЯ СТРАТЕГИЯ: Полная очистка и вставка ---

	// 2. Очистка старых данных (DELETE)
	// Удаляем все карточки, привязанные к этой доске
	_, err = tx.ExecContext(ctx, `
        DELETE FROM cards 
        WHERE column_id IN (SELECT id FROM columns WHERE board_id = $1);
    `, boardID)
	if err != nil {
		return fmt.Errorf("%w: failed to delete old cards: %v", ErrBoardUpdateFailed, err)
	}

	// Удаляем все колонки, привязанные к этой доске
	_, err = tx.ExecContext(ctx, "DELETE FROM columns WHERE board_id = $1", boardID)
	if err != nil {
		return fmt.Errorf("%w: failed to delete old columns: %v", ErrBoardUpdateFailed, err)
	}

	// 3. Вставка новых данных (INSERT)
	for i, col := range boardData {
		// 🔑 Вставка колонки
		var columnID string
		// ⚠️ Используем ID из пейлоада для сохранения ссылочной целостности и UUID
		err = tx.GetContext(ctx, &columnID, `
            INSERT INTO columns (id, board_id, column_title, position) 
            VALUES ($1, $2, $3, $4) 
            RETURNING id;
        `, col.ID, boardID, col.Title, i+1)

		if err != nil {
			return fmt.Errorf("%w: failed to insert column %s: %v", ErrBoardUpdateFailed, col.ID, err)
		}

		// 🔑 Вставка карточек
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

	// 4. Коммит
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}
