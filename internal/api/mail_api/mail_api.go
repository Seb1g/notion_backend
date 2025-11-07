package mail_api

import (
	"anemone_notes/internal/api/middlewares"
	"anemone_notes/internal/model/mail_model"
	"anemone_notes/internal/repository/mail_repository"
	"anemone_notes/internal/services/auth_services"
	"anemone_notes/internal/services/mail_services"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type MailHandler struct {
	Service     *mail_services.MailService
	AuthService *auth_services.AuthService
	Repo        *mail_repository.MailRepository
}

func NewMailHandler(
	service *mail_services.MailService,
	authService *auth_services.AuthService,
	repo *mail_repository.MailRepository,
) *MailHandler {
	return &MailHandler{
		Service:     service,
		AuthService: authService,
		Repo:        repo,
	}
}

func (h *MailHandler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/api/v1/mail").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return middlewares.AuthMiddleware(h.AuthService, next)
	})

	api.HandleFunc("/addresses", h.generateAddress).Methods("POST")
	api.HandleFunc("/addresses", h.listAddresses).Methods("GET")

	ownerRoutes := api.PathPrefix("").Subrouter()
	ownerRoutes.Use(middlewares.CheckAddressOwnerMiddleware(h.Repo))

	ownerRoutes.HandleFunc("/inbox/{id:[0-9]+}", h.getInbox).Methods("GET")
	ownerRoutes.HandleFunc("/addresses/{id:[0-9]+}", h.deleteAddress).Methods("DELETE")
}

func (h *MailHandler) generateAddress(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
		return
	}

	addr, err := h.Service.GenerateAddress(userID)
	if err != nil {
		log.Printf("ERROR: could not generate address for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := mail_services.GeneratedAddressResponse{
		Address: addr.Address,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *MailHandler) getInbox(w http.ResponseWriter, r *http.Request) {
	addressIDVal := r.Context().Value(middlewares.AddressIDContextKey)
	addressID, ok := addressIDVal.(int)
	if !ok {
		http.Error(w, "Could not retrieve address ID from context", http.StatusInternalServerError)
		return
	}

	emails, err := h.Service.GetInboxForAddress(addressID)
	if err != nil {
		log.Printf("ERROR: could not get inbox for address %d: %v", addressID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if emails == nil {
		emails = []mail_model.Email{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(emails)
}

func (h *MailHandler) listAddresses(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
		return
	}

	addresses, err := h.Service.ListAddresses(userID)
	if err != nil {
		log.Printf("ERROR: could not list addresses for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if addresses == nil {
		addresses = []mail_model.TempAddress{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(addresses)
}

func (h *MailHandler) deleteAddress(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	addressIDVal := r.Context().Value(middlewares.AddressIDContextKey)
	addressID, _ := addressIDVal.(int)

	err := h.Service.DeleteAddress(addressID, userID)
	if err != nil {
		log.Printf("ERROR: could not delete address %d for user %d: %v", addressID, userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}