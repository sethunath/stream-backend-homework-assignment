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

type messageWithReactions struct {
	ID            string    `bun:",pk,type:uuid"`
	MessageText   string    `bun:"message_text,notnull"`
	UserID        string    `bun:",notnull"`
	CreatedAt     time.Time `bun:",nullzero,default:now()"`
	ReactionType  *string   `bun:"column:type"`
	ReactionCount *int      `bun:"reaction_count"`
}

func (m messageWithReactions) APIMessage() api.Message {
	return api.Message{
		ID:        m.ID,
		Text:      m.MessageText,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
}

type messageReaction struct {
	ID        string    `bun:",pk,type:uuid,default:uuid_generate_v4()"`
	MessageID string    `bun:"message_id,notnull"`
	UserID    string    `bun:"user_id,notnull"`
	Type      string    `bun:"type,notnull"`
	Score     int       `bun:"score,default:1"`
	CreatedAt time.Time `bun:",nullzero,default:now()"`
}

func (m messageReaction) APIMessageReaction() api.Reaction {
	return api.Reaction{
		ID:        m.ID,
		UserID:    m.UserID,
		MessageID: m.MessageID,
		Type:      m.Type,
		Score:     m.Score,
		CreatedAt: m.CreatedAt,
	}
}
