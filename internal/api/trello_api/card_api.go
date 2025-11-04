package trello_api

import (
	"anemone_notes/internal/api/middlewares"
	"anemone_notes/internal/repository/trello_repository"
	"anemone_notes/internal/services/auth_services"
	"anemone_notes/internal/services/trello_services"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type CardHandler struct {
	Service     *trello_services.CardService
	AuthService *auth_services.AuthService
}

func NewCardHandler(s *trello_services.CardService, a *auth_services.AuthService) *CardHandler {
	return &CardHandler{Service: s, AuthService: a}
}

func (h *CardHandler) CardRoutes(r *mux.Router) {
	// Create card in column: Status: WORK
	r.Handle("/api/v1/trello/create_card",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.createCard)),
	).Methods("POST")
	// Delete card in column: Status: WORK
	r.Handle("/api/v1/trello/delete_card",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.deleteCard)),
	).Methods("DELETE")
	// Update card in column: Status: WORK
	r.Handle("/api/v1/trello/rename_card",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.renameCard)),
	).Methods("PUT")
}

func (h *CardHandler) createCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ColumnID  string `json:"column_id"`
		CardTitle string `json:"card_title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	cardData, err := h.Service.CreateCard(r.Context(), req.ColumnID, req.CardTitle)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cardData)
}

func (h *CardHandler) deleteCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ColumnID string `json:"column_id"`
		CardID   string `json:"card_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	err := h.Service.DeleteCard(r.Context(), req.ColumnID, req.CardID)
	if err != nil {
		if errors.Is(err, trello_repository.ErrCardNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Card not found"})
			return
		}
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Card deleted successfully"})
}

func (h *CardHandler) renameCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ColumnID string `json:"column_id"`
		CardID   string `json:"card_id"`
		NewName  string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	card, err := h.Service.RenameCard(r.Context(), req.ColumnID, req.CardID, req.NewName)
	if err != nil {
		if errors.Is(err, trello_repository.ErrCardNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Card not found"})
			return
		}
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"message": "Card success renamed", "card": card})
}
