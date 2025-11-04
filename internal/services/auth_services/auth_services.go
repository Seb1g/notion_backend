package auth_services

import (
	"anemone_notes/internal/config"
	"anemone_notes/internal/model/auth_model"
	"anemone_notes/internal/repository/auth_repository"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Здесь должна быть ваша реальная загрузка конфигурации
var cfg = config.Load() 

type AuthService struct {
	Users   *auth_repository.UserRepo
	Refresh *auth_repository.RefreshRepo
}

func NewAuthService(u *auth_repository.UserRepo, r *auth_repository.RefreshRepo) *AuthService {
	return &AuthService{Users: u, Refresh: r}
}

// ─── Регистрация ──────────────────────────────────────────────
func (s *AuthService) Register(ctx context.Context, email, password string) (string, string, *auth_model.User, error) {
	email = strings.TrimSpace(email) // Очистка email
	password = strings.TrimSpace(password) // Очистка пароля

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", nil, errors.New("failed to hash password")
	}

	u := &auth_model.User{Email: email, Password: string(hash)}
	if err := s.Users.Create(ctx, u); err != nil {
		return "", "", nil, err
	}
	accessToken, refreshToken, err := s.generateTokens(ctx, u)
	if err != nil {
		return "", "", nil, err
	}
	return accessToken, refreshToken, u, nil
}

// ─── Генерация токенов ─────────────────────────────────────────
func (s *AuthService) generateTokens(ctx context.Context, u *auth_model.User) (string, string, error) {
	// ... (без изменений)
	// access (60 минут - лучше 15-30)
	accessClaims := jwt.MapClaims{
		"user_id": u.ID,
		"exp":     time.Now().Add(60 * time.Minute).Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := at.SignedString([]byte(cfg.AccessSecret))
	if err != nil {
		log.Printf("ERROR signing access token: %v", err)
		return "", "", err
	}

	// refresh (7 дней)
	refreshExp := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"user_id": u.ID,
		"exp":     refreshExp.Unix(),
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := rt.SignedString([]byte(cfg.RefreshSecret))
	if err != nil {
		log.Printf("ERROR signing refresh token: %v", err)
		return "", "", err
	}

	// сохранить refresh в БД
	if err := s.Refresh.Store(ctx, u.ID, refreshToken, refreshExp); err != nil {
		log.Printf("ERROR storing refresh token: %v", err)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ─── Логин ─────────────────────────────────────────────────────
func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, *auth_model.User, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return "", "", nil, errors.New("invalid credentials") // 🚫 Не раскрываем, существует ли пользователь
	}

	// Сравнение хеша (уже очищенный в репозитории)
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return "", "", nil, errors.New("invalid credentials")
	}
	access, refresh, err := s.generateTokens(ctx, u)
	return access, refresh, u, nil
}

// 🆕 НОВАЯ ФУНКЦИЯ: Logout
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// Парсинг токена для получения userID
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		// Должны удалить токен, даже если он "испорчен", если он есть в БД
		// В этой простой реализации просто вернём ошибку парсинга
		return errors.New("invalid refresh token format")
	}

	userID := int(claims["user_id"].(float64))

	// Удаление токена из БД
	if err := s.Refresh.Delete(ctx, userID, refreshToken); err != nil {
		log.Printf("ERROR deleting refresh token: %v", err)
		return errors.New("failed to logout")
	}
	return nil
}


// ─── Смена пароля ──────────────────────────────────────────────
// 🔑 Улучшено: теперь требует старый пароль для смены.
func (s *AuthService) ChangePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	email = strings.TrimSpace(email)
	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)

	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	// 1. Проверяем старый пароль
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPassword)) != nil {
		return errors.New("invalid old password")
	}

	// 2. Хешируем новый пароль
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// 3. Обновляем пароль в БД (используя ID, а не email)
	return s.Users.UpdatePassword(ctx, u.ID, string(newHash))
}

// ─── Сброс пароля (для сценариев "Забыл пароль" с токеном сброса) ────────
// Сохраняем как отдельную функцию для "Забыл пароль" (если у вас есть такой механизм).
// Если нет, эту функцию можно удалить.
func (s *AuthService) ResetPassword(ctx context.Context, email, newPassword string) error {
	email = strings.TrimSpace(email)
	newPassword = strings.TrimSpace(newPassword)
	
	// Здесь должна быть дополнительная проверка токена сброса из email.
	// В текущем виде эта функция небезопасна, если вызывается напрямую!
	
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	return s.Users.UpdatePassword(ctx, u.ID, string(hash))
}


// ─── Обновление access по refresh ─────────────────────────────
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, *auth_model.User, error) {
	// ... (без изменений, кроме использования ParseWithClaims)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return "", nil, errors.New("invalid refresh token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return "", nil, errors.New("invalid user_id in token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", nil, errors.New("invalid exp in token")
	}

	// Проверяем наличие токена в БД и его срок годности
	ok, err = s.Refresh.Check(ctx, int(userID), refreshToken, time.Unix(int64(exp), 0))
	if err != nil || !ok {
		return "", nil, errors.New("refresh token not found or expired")
	}

	// Генерация нового access токена (сокращаем время жизни, 15 минут)
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	at, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(cfg.AccessSecret))
	if err != nil {
		return "", nil, errors.New("error")
	}

	ud, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return "", nil, errors.New("user data not found")
	}

	return at, ud, nil
}

func (s *AuthService) ParseAccessToken(tokenStr string) (int, error) {
	// ... (без изменений)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	userID := int(claims["user_id"].(float64))
	return userID, nil
}