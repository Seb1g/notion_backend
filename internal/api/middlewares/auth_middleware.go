package middlewares

import (
  "anemone_notes/internal/services/auth_services"
  "context"
  "net/http"
  "strings"
)

const userIDKey contextKey = "user_id"

func GetUserIDFromContext(ctx context.Context) (int, bool) {
    userIDVal := ctx.Value(userIDKey)
    userID, ok := userIDVal.(int)
    return userID, ok
}

func AuthMiddleware(auth *auth_services.AuthService, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
      http.Error(w, "missing token", http.StatusUnauthorized)
      return
    }

    tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
    userID, err := auth.ParseAccessToken(tokenStr)
    if err != nil {
      http.Error(w, "invalid token", http.StatusUnauthorized)
      return
    }

    ctx := context.WithValue(r.Context(), userIDKey, userID) 
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}