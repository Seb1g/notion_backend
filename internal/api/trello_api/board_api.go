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

func (h *BoardHandler) getBoardRepoInterface() middlewares.BoardRepoInterface {
	return h.Service.Repo
}

func NewBoardHandler(s *trello_services.BoardService, a *auth_services.AuthService) *BoardHandler {
	return &BoardHandler{Service: s, AuthService: a}
}

func (h *BoardHandler) BoardRoutes(r *mux.Router) {
	boardRepo := h.getBoardRepoInterface()
	// Work
	r.Handle("/api/v1/trello/create_board",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.createBoard)),
	).Methods("POST")
	// Work
	r.Handle("/api/v1/trello/get_all_user_boards",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.getAllUserBoards)),
	).Methods("GET")

	boardRouter := r.PathPrefix("/api/v1/trello/board/{boardID}").Subrouter()
	// Work
	boardRouter.Handle("",
		middlewares.AuthMiddleware(h.AuthService,
			middlewares.IsBoardOwner_Path(boardRepo, http.HandlerFunc(h.getOneUserBoard))),
	).Methods("GET")
	// Work
	boardRouter.Handle("",
		middlewares.AuthMiddleware(h.AuthService,
			middlewares.IsBoardOwner_Path(boardRepo, http.HandlerFunc(h.deleteBoard))),
	).Methods("DELETE")
	// Work
	boardRouter.Handle("",
		middlewares.AuthMiddleware(h.AuthService,
			middlewares.IsBoardOwner_Path(boardRepo, http.HandlerFunc(h.renameBoard))),
	).Methods("PUT")
	// WORK
	boardRouter.Handle("",
		middlewares.AuthMiddleware(h.AuthService,
			middlewares.IsBoardOwner_Path(boardRepo, http.HandlerFunc(h.updateBoard))),
	).Methods("POST")
}

func (h *BoardHandler) createBoard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User authentication data missing", http.StatusInternalServerError)
		return
	}

	boardData, err := h.Service.CreateBoard(r.Context(), req.Title, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boardData)
}

func (h *BoardHandler) getOneUserBoard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID := vars["boardID"]

	oneUserBoard, err := h.Service.GetOneUserBoard(r.Context(), boardID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oneUserBoard)
}

func (h *BoardHandler) getAllUserBoards(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User authentication data missing", http.StatusInternalServerError)
		return
	}

	allUserBoards, err := h.Service.GetAllUserBoards(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allUserBoards)
}

func (h *BoardHandler) deleteBoard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardID := vars["boardID"]

	err := h.Service.DeleteBoard(r.Context(), boardID)
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
		NewName string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	boardID := vars["boardID"]

	board, err := h.Service.RenameBoard(r.Context(), boardID, req.NewName)
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
		BoardData []*trello_model.Column `json:"board_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, err)
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	boardID := vars["boardID"]

	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User authentication data missing", http.StatusInternalServerError)
		return
	}

	err := h.Service.UpdateBoard(r.Context(), boardID, userID, req.BoardData)
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
