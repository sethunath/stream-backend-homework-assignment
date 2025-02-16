//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/GetStream/stream-backend-homework-assignment/api"
	"github.com/google/go-cmp/cmp"
)

func TestPostgres_ListMessages(t *testing.T) {
	tests := []struct {
		name  string
		setup func(pg *Postgres) error
		want  []api.Message
	}{
		{
			name: "Empty",
			want: []api.Message{},
		},
		{
			name: "One",
			setup: func(pg *Postgres) error {
				msgs := []message{
					{
						ID:          "388d74ea-cc39-4566-860f-0df6068f3330",
						MessageText: "hello",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}
				_, err := pg.bun.NewInsert().Model(&msgs).Exec(context.Background())
				return err
			},
			want: []api.Message{
				{
					ID:                    "388d74ea-cc39-4566-860f-0df6068f3330",
					Text:                  "hello",
					UserID:                "test",
					CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					MessageReactionCounts: make([]api.MessageReactionCount, 0),
				},
			},
		},
		{
			name: "Two",
			setup: func(pg *Postgres) error {
				msgs := []message{
					{
						ID:          "4562fe69-42b3-46e5-b990-11581182f57c",
						MessageText: "hello",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "7c6d956b-58d6-4ac3-9984-f341346edc37",
						MessageText: "world",
						UserID:      "test",
						CreatedAt:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}
				_, err := pg.bun.NewInsert().Model(&msgs).Exec(context.Background())
				return err
			},
			want: []api.Message{
				{ // First because of DESC sorting on the created_at column.
					ID:                    "7c6d956b-58d6-4ac3-9984-f341346edc37",
					Text:                  "world",
					UserID:                "test",
					CreatedAt:             time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					MessageReactionCounts: []api.MessageReactionCount{},
				},
				{
					ID:                    "4562fe69-42b3-46e5-b990-11581182f57c",
					Text:                  "hello",
					UserID:                "test",
					CreatedAt:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					MessageReactionCounts: []api.MessageReactionCount{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			pg := connect(t)
			if tt.setup != nil {
				if err := tt.setup(pg); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			got, err := pg.ListMessages(ctx, 0, 0)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Diff (-got +want)\n%s", diff)
			}
		})
	}
}

func TestPostgres_InsertMessage(t *testing.T) {
	tests := []struct {
		name  string
		msg   api.Message
		check func(t *testing.T, pg *Postgres)
	}{
		{
			name: "OK",
			msg: api.Message{
				Text:   "Hello",
				UserID: "testuser",
			},
			check: func(t *testing.T, pg *Postgres) {
				var got message
				if err := pg.bun.NewSelect().Model(&got).Scan(context.Background()); err != nil {
					t.Fatal(err)
				}

				if got.MessageText != "Hello" {
					t.Errorf("Stored message text does not match; got %q, want %q", got.MessageText, "Hello")
				}
				if got.UserID != "testuser" {
					t.Errorf("Stored message user id does not match; got %q, want %q", got.UserID, "testuser")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			pg := connect(t)
			got, err := pg.InsertMessage(ctx, tt.msg)
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, pg)

			if got.ID == "" {
				t.Error("Returned message has empty ID")
			}
			if got.CreatedAt.IsZero() {
				t.Error("Returned message does not have a CreatedAt field")
			}
		})
	}
}

func connect(t *testing.T) *Postgres {
	t.Helper()
	connStr := "postgres://message-api:message-api@localhost:5432/message-api?sslmode=disable"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pg, err := Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("Could not connect to PostgreSQL: %v", err)
	}

	// Truncate the table before each test.
	if _, err := pg.bun.NewTruncateTable().Model((*message)(nil)).Cascade().Exec(ctx); err != nil {
		t.Fatalf("Could not truncate table: %v", err)
	}

	return pg
}
