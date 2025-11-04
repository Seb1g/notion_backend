package trello_api

import (
	"anemone_notes/internal/api/middlewares"
	"anemone_notes/internal/model/trello_model"
	"anemone_notes/internal/repository/trello_repository"
	"anemone_notes/internal/services/auth_services"
	"anemone_notes/internal/services/trello_services"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func handleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	if errors.Is(err, trello_repository.ErrBoardNotFound) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Board not found"})
		return
	}

	var reqErr *json.SyntaxError
	if errors.As(err, &reqErr) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request payload"})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
}

type BoardHandler struct {
	Service     *trello_services.BoardService
	AuthService *auth_services.AuthService
}

func NewBoardHandler(s *trello_services.BoardService, a *auth_services.AuthService) *BoardHandler {
	return &BoardHandler{Service: s, AuthService: a}
}

func (h *BoardHandler) BoardRoutes(r *mux.Router) {
	// Create Kanban Board: Status: WORK
	r.Handle("/api/v1/trello/create_board",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.createBoard)),
	).Methods("POST")
	// Get one Kanban Board: Status: WORK
	r.Handle("/api/v1/trello/get_one_user_board",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.getOneUserBoard)),
	).Methods("GET")
	// Get all User Boards: Status: WORK
	r.Handle("/api/v1/trello/get_all_user_boards",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.getAllUserBoards)),
	).Methods("GET")
	// Delete Kanban Board: Status: WORK
	r.Handle("/api/v1/trello/delete_board",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.deleteBoard)),
	).Methods("DELETE")
	// Rename Kanban Board: Status: 
	r.Handle("/api/v1/trello/rename_board",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.renameBoard)),
	).Methods("PUT")
	// Update KanBan Board: Status: UNKNOWN
	r.Handle("/api/v1/trello/update_board",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.updateBoard)),
	).Methods("POST")
}

func (h *BoardHandler) createBoard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title  string `json:"title"`
		UserID int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	boardData, err := h.Service.CreateBoard(r.Context(), req.Title, req.UserID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boardData)
}

func (h *BoardHandler) getOneUserBoard(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	boardID := query.Get("boardId")
	userID := query.Get("userId")

	oneUserBoard, err := h.Service.GetOneUserBoard(r.Context(), boardID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oneUserBoard)
}

func (h *BoardHandler) getAllUserBoards(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")

	allUserBoards, err := h.Service.GetAllUserBoards(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allUserBoards)
}

func (h *BoardHandler) deleteBoard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BoardID string `json:"board_id"`
		UserID  int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	err := h.Service.DeleteBoard(r.Context(), req.BoardID, req.UserID)
	if err != nil {
		if errors.Is(err, trello_repository.ErrBoardNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Board not found"})
			return
		}
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Board deleted successfully"})
}

func (h *BoardHandler) renameBoard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BoardID string `json:"board_id"`
		UserID  int `json:"user_id"`
		NewName string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	board, err := h.Service.RenameBoard(r.Context(), req.BoardID, req.UserID, req.NewName)
	if err != nil {
		if errors.Is(err, trello_repository.ErrBoardNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Board not found"})
			return
		}
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}

func (h *BoardHandler) updateBoard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BoardID   string                 `json:"board_id"`
		BoardData []*trello_model.Column `json:"board_data"`
		UserID    int                 `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	err := h.Service.UpdateBoard(r.Context(), req.BoardID, req.UserID, req.BoardData)
	if err != nil {
		if errors.Is(err, trello_repository.ErrBoardNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Board not found"})
			return
		}
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"message": "Board updated successfully", "success": true})
}
