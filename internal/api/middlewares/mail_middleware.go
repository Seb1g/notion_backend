package middlewares

import (
	"anemone_notes/internal/repository/mail_repository"
	"context"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type contextKey string

const AddressIDContextKey contextKey = "address_id"

func CheckAddressOwnerMiddleware(repo *mail_repository.MailRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
				return
			}

			vars := mux.Vars(r)
			idStr, ok := vars["id"]
			if !ok {
				http.Error(w, "Missing address ID in URL path", http.StatusBadRequest)
				return
			}

			addressID, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid address ID", http.StatusBadRequest)
				return
			}

			isOwner, err := repo.CheckAddressOwner(addressID, userID)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !isOwner {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), AddressIDContextKey, addressID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
