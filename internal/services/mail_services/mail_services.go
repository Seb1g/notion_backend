package mail_services

import (
	"anemone_notes/internal/model/mail_model"
	"anemone_notes/internal/repository/mail_repository"
)

type MailService struct {
	repo   *mail_repository.MailRepository
	domain string
}

func New(repo *mail_repository.MailRepository, domain string) *MailService {
	return &MailService{
		repo:   repo,
		domain: domain,
	}
}

type GeneratedAddressResponse struct {
	Address string `json:"address"`
}

func (s *MailService) GenerateAddress(userID int) (*mail_model.TempAddress, error) {
	addr, err := s.repo.CreateTempAddress(s.domain, userID)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (s *MailService) GetInboxForAddress(addressID int) ([]mail_model.Email, error) {
	return s.repo.GetEmailsForAddress(addressID)
}

func (s *MailService) ListAddresses(userID int) ([]mail_model.TempAddress, error) {
	return s.repo.GetAddressesForUser(userID)
}

func (s *MailService) DeleteAddress(addressID int, userID int) error {
	return s.repo.DeleteAddress(addressID, userID)
}
