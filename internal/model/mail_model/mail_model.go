package mail_model

import (
	"time"
	"github.com/lib/pq" 
)

type TempAddress struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `json:"-" db:"user_id"`
	Address   string    `db:"address" json:"address"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Email struct {
	ID         int       `db:"id" json:"id"`
	AddressID  int       `db:"address_id" json:"-"`
	Sender     string    `db:"sender" json:"sender"`
	Recipients pq.StringArray  `db:"recipients" json:"recipients"`
	Subject    string    `db:"subject" json:"subject"`
	Body       string    `db:"body" json:"body"`
	ReceivedAt time.Time `db:"received_at" json:"received_at"`
}
