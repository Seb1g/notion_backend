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

type Response struct {
	Status string `json:"status"`
}

type PageHandler struct {
	Service     *notes_services.PageService
	AuthService *auth_services.AuthService
}

func NewPageHandler(s *notes_services.PageService, a *auth_services.AuthService) *PageHandler {
	return &PageHandler{Service: s, AuthService: a}
}

func (h *PageHandler) PagesRoutes(r *mux.Router) {
	// Create note - Status: WORK
	r.Handle("/api/v1/notes/create_note",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.createPage)),
	).Methods("POST")
	// Get one note by id - Status: WORK
	r.Handle("/api/v1/notes/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.getPage)),
	).Methods("GET")
	// Get all user notes - Status: WORK
	r.Handle("/api/v1/notes",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.getAllPages)),
	).Methods("GET")
	// Update title note by id - Status: WORK
	r.Handle("/api/v1/notes/update_title",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.updateTitle)),
	).Methods("PUT")
	// Update content note by id - Status: WORK
	r.Handle("/api/v1/notes/update_content",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.updateContent)),
	).Methods("PUT")
	// Get all notes from folder - Status: Unknown
	// TODO: Conduct tests
	r.Handle("/api/v1/folder/get_notes/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.getAllNotesFromFolder)),
	)
	// Add note by id to folder - Status: WORK
	r.Handle("/api/v1/notes/add_to_folder",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.addNoteToFolder)),
	).Methods("POST")
	// Cancel note from folder - Status: WORK but need bugfix
	// FIXME: The value that should be written to the database should be null, not zero.
	r.Handle("/api/v1/notes/cancel_from_folder/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.cencelingNoteFromFolder)),
	).Methods("POST")
	// Soft delete one note - Status: WORK
	r.Handle("/api/v1/notes/{id}/soft-delete",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.markDeletedNote)),
	).Methods("PUT")
	// Unmark note soft delete - Status: WORK
	r.Handle("/api/v1/notes/{id}/soft-undelete",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.unmarkDeletedNote)),
	).Methods("PUT")
	// Soft delete more notes - Status: WORK
	r.Handle("/api/v1/notes/bulk-delete",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.markDeletedMoreNotes)),
	).Methods("PUT")
	// Unmark notes soft delete - Status: WORK
	r.Handle("/api/v1/notes/bulk-undelete",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.unmarkDeletedMoreNotes)),
	).Methods("PUT")
	// Soft Delete all notes - Status: WORK
	r.Handle("/api/v1/notes/all_items_delete/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.markDeletedAllNotes)),
	).Methods("PUT")
	// Unmark all notes soft delete - Status: WORK
	r.Handle("/api/v1/notes/all_items_undelete/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.unmarkDeletedAllNotes)),
	).Methods("PUT")
	// Clear trash bin user by id - Status: WORK
	r.Handle("/api/v1/notes/trash/clear/{id}",
		middlewares.AuthMiddleware(h.AuthService, http.HandlerFunc(h.deleteAllMarkNotes)),
	).Methods("DELETE")
}

func (h *PageHandler) createPage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  int    `json:"user_id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	p, err := h.Service.CreatePage(r.Context(), req.UserID, req.Title, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) getPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := h.Service.GetPage(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) getAllPages(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	p, err := h.Service.GetAllPages(r.Context(), userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) updateTitle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int    `json:"id"`
		NewTitle string `json:"new_title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	p, err := h.Service.UpdateTitle(r.Context(), req.ID, req.NewTitle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) updateContent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         int    `json:"id"`
		NewContent string `json:"new_content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	p, err := h.Service.UpdateContent(r.Context(), req.ID, req.NewContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) getAllNotesFromFolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p, err := h.Service.GetAllNotesFromFolder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) addNoteToFolder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NoteID   int `json:"note_id"`
		FolderID int `json:"folder_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	p, err := h.Service.AddNoteToFolder(r.Context(), req.NoteID, req.FolderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) cencelingNoteFromFolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p, err := h.Service.CencelingNoteFromFolder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PageHandler) markDeletedNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.Service.MarkDeletedNote(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}

func (h *PageHandler) unmarkDeletedNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.Service.UnmarkDeletedNote(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}

func (h *PageHandler) markDeletedMoreNotes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []int `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	err := h.Service.MarkDeletedMoreNotes(r.Context(), req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}

func (h *PageHandler) unmarkDeletedMoreNotes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []int `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}

	err := h.Service.UnmarkDeletedMoreNotes(r.Context(), req.Items)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}

func (h *PageHandler) markDeletedAllNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.Service.MarkDeletedAllNotes(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}

func (h *PageHandler) unmarkDeletedAllNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.Service.UnmarkDeletedAllNotes(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}

func (h *PageHandler) deleteAllMarkNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.Service.DeleteAllMarkNotes(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := Response{Status: "Success"}
	json.NewEncoder(w).Encode(response)
}
