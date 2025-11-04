package auth_api

import (
	"anemone_notes/internal/services/auth_services"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type AuthHandler struct {
	Service *auth_services.AuthService
}

func NewAuthHandler(s *auth_services.AuthService) *AuthHandler {
	return &AuthHandler{Service: s}
}

func (h *AuthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/auth/register", h.register).Methods("POST")
	r.HandleFunc("/api/v1/auth/login", h.login).Methods("POST")
	r.HandleFunc("/api/v1/auth/change-password", h.changePassword).Methods("POST")
	r.HandleFunc("/api/v1/auth/refresh", h.refresh).Methods("POST")
	r.HandleFunc("/api/v1/auth/logout", h.logout).Methods("POST")
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	// ... (почти без изменений, но Service теперь сам триммит)
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, u, err := h.Service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user_data":     u,
	})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	// ... (почти без изменений, но Service теперь сам триммит)
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	access, refresh, user_data, err := h.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
		"user_data": user_data,
	})
}

// 🆕 НОВЫЙ МАРШРУТ: Смена пароля
func (h *AuthHandler) changePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.Service.ChangePassword(r.Context(), req.Email, req.OldPassword, req.NewPassword); err != nil {
		// Стараемся не раскрывать слишком много информации об ошибке
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "invalid old password") || strings.Contains(err.Error(), "user not found") {
			status = http.StatusUnauthorized // Лучше использовать 401 для "неправильный старый пароль"
			http.Error(w, "Invalid email or old password", status)
			return
		}
		http.Error(w, "Password change failed", status)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newAccess, user_data, err := h.Service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"access_token": newAccess, "user_data": user_data})
}

// 🆕 НОВЫЙ МАРШРУТ: Logout
func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Удаление refresh-токена
	if err := h.Service.Logout(r.Context(), req.RefreshToken); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}