package postgres

import (
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
)

// A message represents a message in the database.
type message struct {
	ID          string    `bun:",pk,type:uuid,default:uuid_generate_v4()"`
	MessageText string    `bun:"message_text,notnull"`
	UserID      string    `bun:",notnull"`
	CreatedAt   time.Time `bun:",nullzero,default:now()"`
}

func (m message) APIMessage() api.Message {
	return api.Message{
		ID:        m.ID,
		Text:      m.MessageText,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}
