package trello_services

import (
	"anemone_notes/internal/model/trello_model"
	"anemone_notes/internal/repository/trello_repository"
	"context"
)

type ColumnService struct {
	Repo *trello_repository.ColumnRepo
}

func NewColumnService(r *trello_repository.ColumnRepo) *ColumnService {
	return &ColumnService{Repo: r}
}

func (s *ColumnService) CreateColumn(ctx context.Context, boardID, columnTitle string) (*trello_model.Column, error) {
	return s.Repo.CreateColumn(ctx, boardID, columnTitle)
}

func (s *ColumnService) DeleteColumn(ctx context.Context, boardID, columnID string) error {
	return s.Repo.DeleteColumn(ctx, boardID, columnID)
}

func (s *ColumnService) RenameColumn(ctx context.Context, boardID, columnID, newName string) (*trello_model.Column, error) {
	return s.Repo.RenameColumn(ctx, boardID, columnID, newName)
}
