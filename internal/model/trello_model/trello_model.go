package trello_model

import (
	"time"
)

type Card struct {
	ID       string `db:"id" json:"id"`
	Content  string `db:"content" json:"content"`
	ColumnID string `db:"column_id" json:"column_id"`
	Position int    `db:"position" json:"position"`
}

type Column struct {
	ID       string  `db:"id" json:"id"`
	Title    string  `db:"column_title" json:"title"`
	BoardID  string  `db:"board_id" json:"-"`
	Position int     `db:"position" json:"position"`
	Cards    []*Card `db:"-" json:"cards"`
}

type Board struct {
	ID        string     `db:"id" json:"id"`
	Title     string     `db:"title" json:"title"`
	UserID    int     `db:"user_id" json:"user_id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type BoardWithColumns struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Columns []*Column `json:"columns"`
}

type DefaultColumnData struct {
	Title string
	Cards []string
}

type ContextKey string

const UserIDKey ContextKey = "userID"