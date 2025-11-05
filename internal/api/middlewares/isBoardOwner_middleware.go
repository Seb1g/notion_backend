package middlewares

import (
	"anemone_notes/internal/repository/trello_repository"
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type BoardRepoInterface interface {
	GetBoardOwnerID(ctx context.Context, boardID string) (int, error)
	GetBoardOwnerIDByColumnID(ctx context.Context, columnID string) (int, error)
}

func IsBoardOwner_Query(boardRepo BoardRepoInterface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		boardID := r.URL.Query().Get("boardId")
		if boardID == "" {
			http.Error(w, "Board ID is missing in query parameters", http.StatusBadRequest)
			return
		}

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "User authentication data missing", http.StatusInternalServerError)
			return
		}

		ownerID, err := boardRepo.GetBoardOwnerID(r.Context(), boardID)
		if err != nil {
			if errors.Is(err, trello_repository.ErrBoardNotFound) {
				http.Error(w, "Board not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error checking board ownership", http.StatusInternalServerError)
			return
		}

		if ownerID != userID {
			http.Error(w, "Access Forbidden: Not the board owner", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func IsBoardOwner_Path(boardRepo BoardRepoInterface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		boardID, ok := vars["boardID"]
		if !ok || boardID == "" {
			http.Error(w, "Board ID is missing in URL path", http.StatusBadRequest)
			return
		}

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "User authentication data missing", http.StatusInternalServerError)
			return
		}

		ownerID, err := boardRepo.GetBoardOwnerID(r.Context(), boardID)
		if err != nil {
			if errors.Is(err, trello_repository.ErrBoardNotFound) {
				http.Error(w, "Board not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error checking board ownership", http.StatusInternalServerError)
			return
		}

		if ownerID != userID {
			http.Error(w, "Access Forbidden: Not the board owner", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func IsBoardOwner_ColumnPath(boardRepo BoardRepoInterface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		columnID, ok := vars["columnID"]
		if !ok || columnID == "" {
			http.Error(w, "Column ID is missing in URL path", http.StatusBadRequest)
			return
		}

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok || userID == 0 {
			http.Error(w, "Authentication context error", http.StatusUnauthorized)
			return
		}

		ownerID, err := boardRepo.GetBoardOwnerIDByColumnID(r.Context(), columnID)
		if err != nil {
			if errors.Is(err, trello_repository.ErrBoardNotFound) {
				http.Error(w, "Column or Board not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error checking board ownership", http.StatusInternalServerError)
			return
		}

		if ownerID != userID {
			http.Error(w, "Access Forbidden: Not the board owner", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
