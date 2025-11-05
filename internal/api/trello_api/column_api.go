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
	BoardRepo   middlewares.BoardRepoInterface
}

func NewColumnHandler(s *trello_services.ColumnService, a *auth_services.AuthService, br middlewares.BoardRepoInterface) *ColumnHandler {
	return &ColumnHandler{Service: s, AuthService: a, BoardRepo: br}
}

func (h *ColumnHandler) getBoardRepoInterface() middlewares.BoardRepoInterface {
	return h.BoardRepo
}

func (h *ColumnHandler) ColumnRoutes(r *mux.Router) {
	boardRepo := h.getBoardRepoInterface()

	// WORK
	r.Handle("/api/v1/trello/board/{boardID}/column",
		middlewares.AuthMiddleware(h.AuthService,
			middlewares.IsBoardOwner_Path(boardRepo, http.HandlerFunc(h.createColumn))),
	).Methods("POST")

	columnRouter := r.PathPrefix("/api/v1/trello/board/{boardID}/column/{columnID}").Subrouter()
	// WORK
	columnRouter.Handle("",
		middlewares.AuthMiddleware(h.AuthService,
			middlewares.IsBoardOwner_Path(boardRepo, http.HandlerFunc(h.deleteColumn))),
	).Methods("DELETE")
	// WORK
	columnRouter.Handle("",
		middlewares.AuthMiddleware(h.AuthService,
			middlewares.IsBoardOwner_Path(boardRepo, http.HandlerFunc(h.renameColumn))),
	).Methods("PUT")
}

func (h *ColumnHandler) createColumn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID := vars["boardID"]

	var req struct {
		ColumnTitle string `json:"column_title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	columnData, err := h.Service.CreateColumn(r.Context(), boardID, req.ColumnTitle)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(columnData)
}

func (h *ColumnHandler) deleteColumn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID := vars["boardID"]
	columnID := vars["columnID"]

	err := h.Service.DeleteColumn(r.Context(), boardID, columnID)
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
	vars := mux.Vars(r)
	boardID := vars["boardID"]
	columnID := vars["columnID"]

	var req struct {
		NewName string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	column, err := h.Service.RenameColumn(r.Context(), boardID, columnID, req.NewName)
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
