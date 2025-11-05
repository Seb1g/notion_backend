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
  BoardRepo   middlewares.BoardRepoInterface 
}

func NewCardHandler(s *trello_services.CardService, a *auth_services.AuthService, br middlewares.BoardRepoInterface) *CardHandler {
  return &CardHandler{Service: s, AuthService: a, BoardRepo: br}
}

func (h *CardHandler) CardRoutes(r *mux.Router) {
  boardRepo := h.BoardRepo

  r.Handle("/api/v1/trello/column/{columnID}/card",
    middlewares.AuthMiddleware(h.AuthService,
      middlewares.IsBoardOwner_ColumnPath(boardRepo, http.HandlerFunc(h.createCard))),
  ).Methods("POST")
  
  cardRouter := r.PathPrefix("/api/v1/trello/column/{columnID}/card/{cardID}").Subrouter()
  
  cardRouter.Handle("",
    middlewares.AuthMiddleware(h.AuthService,
      middlewares.IsBoardOwner_ColumnPath(boardRepo, http.HandlerFunc(h.deleteCard))),
  ).Methods("DELETE")
  
  cardRouter.Handle("",
    middlewares.AuthMiddleware(h.AuthService,
      middlewares.IsBoardOwner_ColumnPath(boardRepo, http.HandlerFunc(h.renameCard))),
  ).Methods("PUT")
}

func (h *CardHandler) createCard(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  columnID := vars["columnID"]

  var req struct {
    CardTitle string `json:"card_title"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    handleError(w, err)
    return
  }
  defer r.Body.Close()

  cardData, err := h.Service.CreateCard(r.Context(), columnID, req.CardTitle)
  if err != nil {
    handleError(w, err)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(cardData)
}

func (h *CardHandler) deleteCard(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  columnID := vars["columnID"]
  cardID := vars["cardID"]
  
  err := h.Service.DeleteCard(r.Context(), columnID, cardID)
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
  vars := mux.Vars(r)
  columnID := vars["columnID"]
  cardID := vars["cardID"]

  var req struct {
    NewName string `json:"new_name"`
  }

  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    handleError(w, err)
    return
  }
  defer r.Body.Close()

  card, err := h.Service.RenameCard(r.Context(), columnID, cardID, req.NewName)
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