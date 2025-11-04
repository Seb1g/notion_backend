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

type ColumnHandler struct {
	Service     *trello_services.ColumnService
	AuthService *auth_services.AuthService
}

func NewColumnHandler(s *trello_services.ColumnService, a *auth_services.AuthService) *ColumnHandler {
	return &ColumnHandler{Service: s, AuthService: a}
}

func (h *ColumnHandler) ColumnRoutes(r *mux.Router) {
	// Create Column in Board: Status: WORK
	r.Handle("/api/v1/trello/create_column",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.createColumn)),
	).Methods("POST")
	// Delete Column from Board: Status: WORK
	r.Handle("/api/v1/trello/delete_column",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.deleteColumn)),
	).Methods("DELETE")
	// Rename Column in Board: Status: WORK
	r.Handle("/api/v1/trello/rename_column",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.renameColumn)),
	).Methods("PUT")
}

func (h *ColumnHandler) createColumn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BoardID     string `json:"board_id"`
		ColumnTitle string `json:"column_title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	columnData, err := h.Service.CreateColumn(r.Context(), req.BoardID, req.ColumnTitle)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(columnData)
}

func (h *ColumnHandler) deleteColumn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BoardID  string `json:"board_id"`
		ColumnID string `json:"column_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	err := h.Service.DeleteColumn(r.Context(), req.BoardID, req.ColumnID)
	if err != nil {
		if errors.Is(err, trello_repository.ErrColumnNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Column not found"})
			return
		}
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Column deleted successfully"})
}

func (h *ColumnHandler) renameColumn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BoardID  string `json:"board_id"`
		ColumnID string `json:"column_id"`
		NewName  string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	column, err := h.Service.RenameColumn(r.Context(), req.BoardID, req.ColumnID, req.NewName)
	if err != nil {
		if errors.Is(err, trello_repository.ErrColumnNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Column not found"})
			return
		}
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(column)
}