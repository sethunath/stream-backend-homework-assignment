package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/uptrace/bun/driver/pgdriver"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// A DB provides a storage layer that persists messages.
type DB interface {
	ListMessages(ctx context.Context, excludeMsgIDs ...string) ([]Message, error)
	InsertMessage(ctx context.Context, msg Message) (Message, error)
	InsertReaction(ctx context.Context, reaction Reaction) (Reaction, error)
}

// A Cache provides a storage layer that caches messages.
type Cache interface {
	ListMessages(ctx context.Context) ([]Message, error)
	InsertMessage(ctx context.Context, msg Message) error
}

// API provides the REST endpoints for the application.
type API struct {
	Logger *slog.Logger
	DB     DB
	Cache  Cache

	once sync.Once
	mux  *http.ServeMux
}

func (a *API) setupRoutes() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /messages", a.listMessages)
	mux.HandleFunc("POST /messages", a.createMessage)
	mux.HandleFunc("POST /messages/{messageID}/reactions", a.createReaction)

	a.mux = mux
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.once.Do(a.setupRoutes)
	a.Logger.Info("Request received", "method", r.Method, "path", r.URL.Path)
	a.mux.ServeHTTP(w, r)
}

func (a *API) respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		a.Logger.Error("Could not encode JSON body", "error", err.Error())
	}
}

func (a *API) respondError(w http.ResponseWriter, status int, err error, msg string) {
	type response struct {
		Error string `json:"error"`
	}
	a.Logger.Error("Error", "error", err.Error())
	a.respond(w, status, response{Error: msg})
}

func (a *API) listMessages(w http.ResponseWriter, r *http.Request) {
	type message struct {
		ID        string `json:"id"`
		Text      string `json:"text"`
		UserID    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
	}
	type response struct {
		Messages []message `json:"messages"`
	}

	// Get messages from cache
	msgs, err := a.Cache.ListMessages(r.Context())
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
		return
	}

	a.Logger.Info("Got messages from cache", "count", len(msgs))

	// Get any remaining messages from DB
	msgIDs := make([]string, len(msgs))
	for i, msg := range msgs {
		msgIDs[i] = msg.ID
	}
	dbMsgs, err := a.DB.ListMessages(r.Context(), msgIDs...)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err, "Could not list messages")
		return
	}
	a.Logger.Info("Got remaining messages from DB", "count", len(dbMsgs))
	msgs = append(msgs, dbMsgs...)

	out := make([]message, len(msgs))
	for i, msg := range msgs {
		out[i] = message{
			ID:        msg.ID,
			Text:      msg.Text,
			UserID:    msg.UserID,
			CreatedAt: msg.CreatedAt.Format(time.RFC1123),
		}
	}
	res := response{
		Messages: out,
	}
	a.respond(w, http.StatusOK, res)
}

func (a *API) createMessage(w http.ResponseWriter, r *http.Request) {
	type (
		request struct {
			Text   string `json:"text"`
			UserID string `json:"user_id"`
		}
		response struct {
			ID        string `json:"id"`
			Text      string `json:"text"`
			UserID    string `json:"user_id"`
			CreatedAt string `json:"created_at"`
		}
	)

	var body request
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.respondError(w, http.StatusBadRequest, err, "Could not decode request body")
		return
	}
	r.Body.Close()

	msg, err := a.DB.InsertMessage(r.Context(), Message{
		Text:      body.Text,
		UserID:    body.UserID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, err, "Could not insert message")
		return
	}

	if err := a.Cache.InsertMessage(r.Context(), msg); err != nil {
		a.Logger.Error("Could not cache message", "error", err.Error())
	}

	res := response{
		ID:        msg.ID,
		Text:      msg.Text,
		UserID:    msg.UserID,
		CreatedAt: msg.CreatedAt.Format(time.RFC1123),
	}
	a.respond(w, http.StatusCreated, res)
}

func (a *API) createReaction(w http.ResponseWriter, r *http.Request) {
	type (
		request struct {
			Type   string `json:"type"`
			Score  int    `json:"score"`
			UserID string `json:"user_id"`
		}
		response struct {
			ID        string `json:"id"`         // reaction ID
			MessageID string `json:"message_id"` // message ID
			Type      string `json:"type"`       // reaction type, for example 'like', 'laugh', 'wow', 'thumbs_up'
			Score     int    `json:"score"`      // reaction score should default to 1 if not specified, but can be any positive integer. Think of claps on Medium.com
			UserID    string `json:"user_id"`    // the user ID submitting the reaction
			CreatedAt string `json:"created_at"` // the date/time the reaction was created
		}
	)

	messageID := r.PathValue("messageID")
	var body request
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.respondError(w, http.StatusBadRequest, err, "Could not decode request body")
		return
	}
	r.Body.Close()

	reaction, err := a.DB.InsertReaction(r.Context(), Reaction{
		MessageID: messageID,
		UserID:    body.UserID,
		Type:      body.Type,
		Score:     body.Score,
		CreatedAt: time.Now(),
	})
	if err != nil {
		var pgErr pgdriver.Error
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.IntegrityViolation() {
				a.Logger.Error("Duplicate reaction", "error", pgErr.Error())
				a.respondError(w, http.StatusConflict, err, "Already reacted to this message")
				return
			}
		}
		a.respondError(w, http.StatusInternalServerError, err, "Could not insert reaction")
		return
	}

	res := response{
		ID:        reaction.ID,
		MessageID: reaction.MessageID,
		Type:      reaction.Type,
		Score:     reaction.Score,
		UserID:    reaction.UserID,
		CreatedAt: reaction.CreatedAt.Format(time.RFC1123),
	}
	a.respond(w, http.StatusCreated, res)
}
