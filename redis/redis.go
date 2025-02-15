package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/redis/go-redis/v9"
)

// Redis provides caching in Redis.
type Redis struct {
	cli *redis.Client
}

// Connect connects to the Redis server and pings the server to ensure the
// connection is working.
func Connect(ctx context.Context, addr string) (*Redis, error) {
	cli := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Redis{
		cli: cli,
	}, nil
}

const (
	messagePrefix = "messages"
	maxSize       = 10
)

// ListMessages returns a list of message from Redis. The messages are sorted
// by the timestamp in descending order.
func (r *Redis) ListMessages(ctx context.Context) ([]api.Message, error) {
	vals, err := r.cli.ZRevRangeByScore(ctx, messagePrefix, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", time.Now().UnixNano()),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("zrange: %w", err)
	}

	out := make([]api.Message, len(vals))
	for i, key := range vals {
		var msg message
		err = r.cli.HGetAll(ctx, key).Scan(&msg)
		if err != nil {
			return nil, fmt.Errorf("hgetall: %w", err)
		}

		out[i], err = msg.APIMessage()
		if err != nil {
			return nil, fmt.Errorf("hgetall: %w", err)
		}
	}

	return out, nil
}

// InsertMessage adds the message to Redis with the message:MESSAGE_ID as the key and adds the key to a sorted set.
func (r *Redis) InsertMessage(ctx context.Context, msg api.Message) error {
	m, err := r.toRedisMessage(msg)
	if err != nil {
		return err
	}
	err = r.cli.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			key := fmt.Sprintf("%s:%s", messagePrefix, m.ID)
			pipe.HSet(ctx, key, m)
			pipe.ZAdd(ctx, messagePrefix, redis.Z{
				Score:  float64(msg.CreatedAt.UnixNano()),
				Member: key,
			})

			return nil
		})
		return err
	}, m.ID)

	if err != nil {
		return fmt.Errorf("redis insert message: %w", err)
	}

	// Simulate an eviction strategy by removing the oldest key in case the max cache size is exceeded.
	err = r.evictOldest(ctx)
	if err != nil {
		return fmt.Errorf("evict oldest: %w", err)
	}
	return nil
}

// GetMessage retrieves a message from Redis by its ID.
func (r *Redis) GetMessage(ctx context.Context, messageID string) (*api.Message, error) {
	key := fmt.Sprintf("%s:%s", messagePrefix, messageID)

	var m message

	err := r.cli.HGetAll(ctx, key).Scan(&m)
	if err != nil {
		return nil, fmt.Errorf("redis get message: %w", err)
	}

	if m.ID == "" {
		return nil, api.ErrMessageNotFoundInCache
	}

	apiMsg, err := m.ToAPIMessage()
	if err != nil {
		return nil, fmt.Errorf("redis get message: %w", err)
	}
	return apiMsg, nil
}

// DeleteMessage removes a message from Redis by its ID.
func (r *Redis) DeleteMessage(ctx context.Context, messageID string) error {
	key := fmt.Sprintf("%s:%s", messagePrefix, messageID)

	err := r.cli.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)
			pipe.ZRem(ctx, messagePrefix, key) // Remove from sorted set

			return nil
		})
		return err
	}, messageID)

	if err != nil {
		return fmt.Errorf("redis delete message: %w", err)
	}

	return nil
}

// ToAPIMessage convert redis message to api message
func (m message) ToAPIMessage() (*api.Message, error) {
	am := &api.Message{}

	am.ID = m.ID
	am.Text = m.Text
	am.UserID = m.UserID
	am.CreatedAt = m.CreatedAt
	if m.MessageReactionCounts != "" {
		err := json.Unmarshal([]byte(m.MessageReactionCounts), &am.MessageReactionCounts)
		if err != nil {
			return nil, err
		}
	}
	return am, nil
}

func (r *Redis) toRedisMessage(apiMsg api.Message) (message, error) {
	m := message{
		ID:        apiMsg.ID,
		Text:      apiMsg.Text,
		UserID:    apiMsg.UserID,
		CreatedAt: apiMsg.CreatedAt,
	}
	if apiMsg.MessageReactionCounts != nil {
		reactionCountsJSON, err := json.Marshal(apiMsg.MessageReactionCounts)
		if err != nil {
			return message{}, fmt.Errorf("failed to marshal reactions: %w", err)
		}
		m.MessageReactionCounts = string(reactionCountsJSON)
	}
	return m, nil
}

func (r *Redis) evictOldest(ctx context.Context) error {
	vals, err := r.cli.ZRange(ctx, messagePrefix, 0, int64(-maxSize-1)).Result()
	if err != nil {
		return fmt.Errorf("zrevrange: %w", err)
	}

	for _, key := range vals {
		_ = r.cli.ZRem(ctx, messagePrefix, key).Err()
		_ = r.cli.Del(ctx, key).Err()
	}

	return nil
}
