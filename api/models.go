package api

import "time"

// A Message represents a persisted message.
type Message struct {
	ID        string
	Text      string
	UserID    string
	CreatedAt time.Time
}

// A Reaction represents a reaction to a message such as a like.
type Reaction struct {
	ID        string
	MessageID string
	Type      string
	Score     int
	UserID    string
	CreatedAt time.Time
}
