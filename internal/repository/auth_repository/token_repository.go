package auth_repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type RefreshRepo struct {
	DB *sqlx.DB
}

func NewRefreshRepo(db *sqlx.DB) *RefreshRepo {
	return &RefreshRepo{DB: db}
}

func (r *RefreshRepo) Store(ctx context.Context, userID int, token string, exp time.Time) error {
	q := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)`
	_, err := r.DB.ExecContext(ctx, q, userID, token, exp)
	return err
}

func (r *RefreshRepo) Check(ctx context.Context, userID int, token string, exp time.Time) (bool, error) {
	var count int
	q := `SELECT COUNT(*) FROM refresh_tokens WHERE user_id=$1 AND token=$2 AND expires_at>$3`
	err := r.DB.QueryRowContext(ctx, q, userID, token, time.Now()).Scan(&count)
	return count > 0, err
}

func (r *RefreshRepo) Delete(ctx context.Context, userID int, token string) error {
	q := `DELETE FROM refresh_tokens WHERE user_id=$1 AND token=$2`
	_, err := r.DB.ExecContext(ctx, q, userID, token)
	return err
}