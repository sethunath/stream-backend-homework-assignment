package redis

import (
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
)

// A message represents a message in the database.
type message struct {
	ID        string    `redis:"id" json:"id"`
	Text      string    `redis:"text" json:"text"`
	UserID    string    `redis:"user_id" json:"user_id"`
	CreatedAt time.Time `redis:"created_at" json:"created_at"`
}

func (m message) APIMessage() api.Message {
	return api.Message{
		ID:        m.ID,
		Text:      m.Text,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}
