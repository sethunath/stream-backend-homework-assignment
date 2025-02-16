package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Postgres provides storage in PostgreSQL.
type Postgres struct {
	bun *bun.DB
}

// Connect connects to the database and ping the DB to ensure the connection is
// working.
func Connect(ctx context.Context, connStr string) (*Postgres, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connStr)))
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	db := bun.NewDB(sqlDB, pgdialect.New())
	return &Postgres{
		bun: db,
	}, nil
}

// ListMessages returns all messages in the database.
func (pg *Postgres) ListMessages(ctx context.Context, limit int, offset int, excludeMsgIDs ...string) ([]api.Message, error) {
	var msgs []message
	q := pg.bun.NewSelect().
		Model(&msgs).
		Offset(offset).
		Limit(limit).
		Order("message.created_at DESC").
		Column("message.*")

	q = q.Join("LEFT JOIN message_reactions AS r ON r.message_id = message.id").
		ColumnExpr("r.type AS reaction_type, COUNT(r.id) AS reaction_count").
		Group("message.id", "r.type")

	if len(excludeMsgIDs) > 0 {
		q = q.Where("message.id NOT IN (?)", bun.In(excludeMsgIDs))
	}

	var messagesWithReactions []messageWithReactions
	if err := q.Scan(ctx, &messagesWithReactions); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return convertToMessages(messagesWithReactions), nil
}

// InsertMessage inserts a message into the database. The returned message
// holds auto generated fields, such as the message id.
func (pg *Postgres) InsertMessage(ctx context.Context, msg api.Message) (api.Message, error) {
	m := &message{
		MessageText: msg.Text,
		UserID:      msg.UserID,
	}
	if _, err := pg.bun.NewInsert().Model(m).Exec(ctx); err != nil {
		return api.Message{}, fmt.Errorf("insert: %w", err)
	}
	return m.APIMessage(), nil
}

// InsertReaction inserts a reaction into the database. The returned reaction
// holds auto-generated fields, such as the reaction id.
func (pg *Postgres) InsertReaction(ctx context.Context, reaction api.Reaction) (api.Reaction, error) {
	r := &messageReaction{
		MessageID: reaction.MessageID,
		UserID:    reaction.UserID,
		Type:      reaction.Type,
		Score:     reaction.Score,
	}

	// Insert the reaction into the database
	if _, err := pg.bun.NewInsert().Model(r).Exec(ctx); err != nil {
		return api.Reaction{}, fmt.Errorf("insert: %w", err)
	}

	// Return the inserted reaction, assuming there's a method to convert to API format
	return r.APIMessageReaction(), nil
}

// convertToMessages converts the database response object messageWithReactions to api.Message object
func convertToMessages(messagesWithReactions []messageWithReactions) []api.Message {
	messageMap := make(map[string]*api.Message)

	for _, mwr := range messagesWithReactions {
		if _, exists := messageMap[mwr.ID]; !exists {
			messageMap[mwr.ID] = &api.Message{
				ID:                    mwr.ID,
				Text:                  mwr.MessageText,
				UserID:                mwr.UserID,
				CreatedAt:             mwr.CreatedAt,
				MessageReactionCounts: make([]api.MessageReactionCount, 0),
			}
		}

		if mwr.ReactionType != nil && mwr.ReactionCount != nil {
			reaction := api.MessageReactionCount{
				Type:  *mwr.ReactionType,
				Count: *mwr.ReactionCount,
			}
			messageMap[mwr.ID].MessageReactionCounts = append(messageMap[mwr.ID].MessageReactionCounts, reaction)
		}
	}

	messages := make([]api.Message, 0, len(messageMap))
	for _, msg := range messageMap {
		messages = append(messages, *msg)
	}

	return messages
}
