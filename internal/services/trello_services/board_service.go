package trello_services

import (
	"anemone_notes/internal/model/trello_model"
	"anemone_notes/internal/repository/trello_repository"
	"context"
)

type BoardService struct {
	Repo *trello_repository.BoardRepo
}

func NewBoardService(r *trello_repository.BoardRepo) *BoardService {
	return &BoardService{Repo: r}
}

func (s *BoardService) CreateBoard(ctx context.Context, title string, userID int) (*trello_model.Board, error) {
	return s.Repo.CreateBoard(ctx, title, userID)
}

func (s *BoardService) GetOneUserBoard(ctx context.Context, boardID, userID string) (*trello_model.BoardWithColumns, error) {
	return s.Repo.GetOneUserBoard(ctx, boardID, userID)
}

func (s *BoardService) GetAllUserBoards(ctx context.Context, userID string) ([]*trello_model.Board, error) {
	return s.Repo.GetAllUserBoards(ctx, userID)
}

func (s *BoardService) DeleteBoard(ctx context.Context, boardID string, userID int) error {
	return s.Repo.DeleteBoard(ctx, boardID, userID)
}

func (s *BoardService) RenameBoard(ctx context.Context, boardID string, userID int, newName string) (*trello_model.Board, error) {
	return s.Repo.RenameBoard(ctx, boardID, userID, newName)
}

func (s *BoardService) UpdateBoard(ctx context.Context, boardID string, userID int, boardData []*trello_model.Column) error {
	return s.Repo.UpdateBoard(ctx, boardID, userID, boardData)
}
