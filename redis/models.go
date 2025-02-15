package redis

import (
	"encoding/json"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
)

// A message represents a message in the database.
type message struct {
	ID                    string    `redis:"id" json:"id"`
	Text                  string    `redis:"text" json:"text"`
	UserID                string    `redis:"user_id" json:"user_id"`
	CreatedAt             time.Time `redis:"created_at" json:"created_at"`
	MessageReactionCounts string    `redis:"message_reaction_counts" json:"message_reaction_counts"`
}

func (m message) APIMessage() (api.Message, error) {
	am := api.Message{
		ID:        m.ID,
		Text:      m.Text,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
	}
	if m.MessageReactionCounts != "" {
		err := json.Unmarshal([]byte(m.MessageReactionCounts), &am.MessageReactionCounts)
		if err != nil {
			return api.Message{}, err
		}
	}
	return am, nil
}
