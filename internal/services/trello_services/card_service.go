package trello_services

import (
	"anemone_notes/internal/model/trello_model"
	"anemone_notes/internal/repository/trello_repository"
	"context"
)

type CardService struct {
	Repo *trello_repository.CardRepo
}

func NewCardService(r *trello_repository.CardRepo) *CardService {
	return &CardService{Repo: r}
}

func (s *CardService) CreateCard(ctx context.Context, columnID, cardTitle string) (*trello_model.Card, error) {
	return s.Repo.CreateCard(ctx, columnID, cardTitle)
}

func (s *CardService) DeleteCard(ctx context.Context, columnID, cardID string) error {
	return s.Repo.DeleteCard(ctx, columnID, cardID)
}

func (s *CardService) RenameCard(ctx context.Context, columnID, cardID, newName string) (*trello_model.Card, error) {
	return s.Repo.RenameCard(ctx, columnID, cardID, newName)
}
