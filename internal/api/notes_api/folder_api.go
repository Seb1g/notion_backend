package notes_api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"anemone_notes/internal/api/middlewares"
	"anemone_notes/internal/services/auth_services"
	"anemone_notes/internal/services/notes_services"

	"github.com/gorilla/mux"
)

type FolderHandler struct {
	FolderService *notes_services.FolderService
	AuthService   *auth_services.AuthService
}

func NewFolderHandler(fs *notes_services.FolderService, a *auth_services.AuthService) *FolderHandler {
	return &FolderHandler{FolderService: fs, AuthService: a}
}

func (h *FolderHandler) FolderRoutes(r *mux.Router) {
	// Create folder - Status: WORK
	r.Handle("/api/v1/folder/create",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.createFolder)),
	).Methods("POST")
	// Get all folders by user id - Status: WORK
	r.Handle("/api/v1/folder/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.getAllFolders)),
	).Methods("GET")
	// Update title folder by id - Status: WORK
	r.Handle("/api/v1/folder/update",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.updateTitleFolder)),
	).Methods("PUT")
	// Delete folder by id - Status: WORK
	r.Handle("/api/v1/folder/delete/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.deleteFolder)),
	).Methods("DELETE")
}

func (h *FolderHandler) createFolder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int    `json:"user_id"`
		Title  string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	p, err := h.FolderService.CreateFolder(r.Context(), req.UserID, req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *FolderHandler) getAllFolders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := h.FolderService.GetAllFolders(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *FolderHandler) updateTitleFolder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int    `json:"id"`
		NewTitle string `json:"new_title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	p, err := h.FolderService.UpdateTitleFolder(r.Context(), req.ID, req.NewTitle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *FolderHandler) deleteFolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.FolderService.DeleteFolders(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}