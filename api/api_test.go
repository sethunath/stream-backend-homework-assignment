package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/neilotoole/slogt"
)

func TestAPI_listMessages(t *testing.T) {
	tests := []struct {
		name       string
		db         *testdb
		cache      *testcache
		wantStatus int
		wantBody   string
	}{
		{
			name: "DBError",
			cache: &testcache{
				listMessages: func(t *testing.T) ([]Message, error) {
					return nil, nil
				},
			},
			db: &testdb{
				listMessages: func(t *testing.T, excludeMsgIDs ...string) ([]Message, error) {
					return nil, errors.New("something went wrong")
				},
			},
			wantStatus: 500,
			wantBody: `{
				"error": "Could not list messages"
			}`,
		},
		{
			name: "CacheError",
			cache: &testcache{
				listMessages: func(t *testing.T) ([]Message, error) {
					return nil, errors.New("something went wrong")
				},
			},
			db: &testdb{
				listMessages: func(t *testing.T, excludeMsgIDs ...string) ([]Message, error) {
					return nil, nil
				},
			},
			wantStatus: 200,
			wantBody: `{
				"messages": []
			}`,
		},
		{
			name: "Empty",
			cache: &testcache{
				listMessages: func(t *testing.T) ([]Message, error) {
					return nil, nil
				},
			},
			db: &testdb{
				listMessages: func(t *testing.T, excludeMsgIDs ...string) ([]Message, error) {
					return nil, nil
				},
			},
			wantStatus: 200,
			wantBody: `{
				"messages": []
			}`,
		},
		{
			name: "Cache",
			cache: &testcache{
				listMessages: func(t *testing.T) ([]Message, error) {
					return []Message{
						{
							ID:        "1",
							Text:      "Hello",
							UserID:    "testuser",
							CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					}, nil
				},
			},
			db: &testdb{
				listMessages: func(t *testing.T, excludeMsgIDs ...string) ([]Message, error) {
					// Nothing in DB.
					return nil, nil
				},
			},
			wantStatus: 200,
			wantBody: `{
				"messages": [
					{
						"id": "1",
						"text": "Hello",
						"user_id": "testuser",
						"created_at": "Mon, 01 Jan 2024 00:00:00 UTC",
						"message_reactions": []
					}
				]
			}`,
		},
		{
			name: "DB",
			cache: &testcache{
				listMessages: func(t *testing.T) ([]Message, error) {
					// Nothing in cache.
					return nil, nil
				},
			},
			db: &testdb{
				listMessages: func(t *testing.T, excludeMsgIDs ...string) ([]Message, error) {
					return []Message{
						{
							ID:        "1",
							Text:      "Hello",
							UserID:    "testuser",
							CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					}, nil
				},
			},
			wantStatus: 200,
			wantBody: `{
				"messages": [
					{
						"id": "1",
						"text": "Hello",
						"user_id": "testuser",
						"created_at": "Mon, 01 Jan 2024 00:00:00 UTC",
						"message_reactions": []
					}
				]
			}`,
		},
		{
			name: "Mixed",
			cache: &testcache{
				listMessages: func(t *testing.T) ([]Message, error) {
					return []Message{
						{
							ID:        "1",
							Text:      "Hello",
							UserID:    "testuser",
							CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					}, nil
				},
			},
			db: &testdb{
				listMessages: func(t *testing.T, excludeMsgIDs ...string) ([]Message, error) {
					return []Message{
						{
							ID:        "2",
							Text:      "World",
							UserID:    "testuser",
							CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						},
					}, nil
				},
			},
			wantStatus: 200,
			wantBody: `{
				"messages": [
					{
						"id": "1",
						"text": "Hello",
						"user_id": "testuser",
						"created_at": "Mon, 01 Jan 2024 00:00:00 UTC",
						"message_reactions": []
					},
					{
						"id": "2",
						"text": "World",
						"user_id": "testuser",
						"created_at": "Tue, 02 Jan 2024 00:00:00 UTC",
 						"message_reactions": []
					}
				]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.db != nil {
				tt.db.T = t
			}
			if tt.cache != nil {
				tt.cache.T = t
			}
			api := &API{
				DB:     tt.db,
				Cache:  tt.cache,
				Logger: slogt.New(t),
				Validate: &MockValidator{
					ShouldFail: false,
				},
			}

			srv := httptest.NewServer(api)
			defer srv.Close()

			req, _ := http.NewRequest("GET", srv.URL+"/messages", nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			checkStatus(t, resp.StatusCode, tt.wantStatus)
			checkBody(t, resp, tt.wantBody)
		})
	}
}

func TestAPI_createMessage(t *testing.T) {
	tests := []struct {
		name        string
		cache       *testcache
		db          *testdb
		req         string
		wantStatus  int
		wantBody    string
		containsLog string
	}{
		{
			name:       "InvalidJSON",
			req:        `not json`,
			wantStatus: 400,
			wantBody: `{
				"error": "Could not decode request body"
			}`,
		},
		{
			name: "DBError",
			req: `{
				"text": "hello",
				"user_id": "test"
			}`,
			db: &testdb{
				insertMessage: func(t *testing.T, msg Message) (Message, error) {
					return Message{}, errors.New("something went wrong")
				},
			},
			wantStatus: 500,
			wantBody: `{
				"error": "Could not insert message"
			}`,
		},
		{
			name: "CacheError",
			req: `{
				"text": "hello",
				"user_id": "test"
			}`,
			cache: &testcache{
				insertMessage: func(t *testing.T, msg Message) error {
					return errors.New("something went wrong")
				},
			},
			db: &testdb{
				insertMessage: func(t *testing.T, msg Message) (Message, error) {
					return Message{
						ID:        "1",
						Text:      msg.Text,
						UserID:    msg.UserID,
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil
				},
			},
			wantStatus: 201,
			wantBody: `{
				"id": "1",
				"text": "hello",
				"user_id": "test",
				"created_at": "Mon, 01 Jan 2024 00:00:00 UTC"
			}`,
			containsLog: "Could not cache message",
		},
		{
			name: "OK",
			req: `{
				"text": "hello",
				"user_id": "test"
			}`,
			db: &testdb{
				insertMessage: func(t *testing.T, msg Message) (Message, error) {
					if msg.UserID != "test" {
						t.Errorf("Got UserID %q, want test", msg.UserID)
					}
					if msg.Text != "hello" {
						t.Errorf("Got Text %q, want test", msg.Text)
					}
					return Message{
						ID:        "1",
						Text:      msg.Text,
						UserID:    msg.UserID,
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil
				},
			},
			cache: &testcache{
				insertMessage: func(t *testing.T, msg Message) error {
					if msg.UserID != "test" {
						t.Errorf("Got UserID %q, want test", msg.UserID)
					}
					if msg.Text != "hello" {
						t.Errorf("Got Text %q, want test", msg.Text)
					}
					return nil
				},
			},
			wantStatus: 201,
			wantBody: `{
				"id": "1",
				"text": "hello",
				"user_id": "test",
				"created_at": "Mon, 01 Jan 2024 00:00:00 UTC"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			if tt.db != nil {
				tt.db.T = t
			}
			if tt.cache != nil {
				tt.cache.T = t
			}
			api := &API{
				DB:    tt.db,
				Cache: tt.cache,
				Validate: &MockValidator{
					ShouldFail: false,
				},
				Logger: slog.New(slog.NewTextHandler(buf, nil)),
			}

			srv := httptest.NewServer(api)
			defer srv.Close()

			req, _ := http.NewRequest("POST", srv.URL+"/messages", strings.NewReader(tt.req))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			checkStatus(t, resp.StatusCode, tt.wantStatus)
			checkBody(t, resp, tt.wantBody)
			checkLog(t, buf, tt.containsLog)
		})
	}
}

func TestAPI_createReaction(t *testing.T) {
	tests := []struct {
		name       string
		db         *testdb
		cache      *testcache
		messageID  string
		req        string
		wantStatus int
		wantBody   string
	}{
		{
			name: "OK",
			req: `{
				"type": "like",
				"user_id": "test"
			}`,
			messageID: "12345",
			db: &testdb{
				insertReaction: func(t *testing.T, reaction Reaction) (Reaction, error) {
					if reaction.UserID != "test" {
						t.Errorf("Got UserID %q, want test", reaction.UserID)
					}
					if reaction.Type != "like" {
						t.Errorf("Got Text %q, want test", reaction.Type)
					}
					return Reaction{
						ID:        "1",
						MessageID: "12345",
						Score:     1,
						Type:      reaction.Type,
						UserID:    reaction.UserID,
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil
				},
			},
			cache: &testcache{
				deleteMessage: func(t *testing.T, id string) error {
					return nil
				},
				getMessage: func(t *testing.T, id string) (*Message, error) {
					return nil, nil
				},
			},
			wantStatus: 201,
			wantBody: `{
				"id": "1",
				"message_id": "12345",
				"type": "like",
				"score": 1,
				"user_id": "test",
				"created_at": "Mon, 01 Jan 2024 00:00:00 UTC"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.db == nil {
				tt.db = &testdb{}
			}
			if tt.cache == nil {
				tt.cache = &testcache{}
			}
			tt.db.T = t
			tt.cache.T = t
			api := &API{
				DB:    tt.db,
				Cache: tt.cache,
				Validate: &MockValidator{
					ShouldFail: false,
				},
				Logger: slogt.New(t),
			}

			srv := httptest.NewServer(api)
			defer srv.Close()

			req, _ := http.NewRequest("POST", srv.URL+"/messages/"+tt.messageID+"/reactions", strings.NewReader(tt.req))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			checkStatus(t, resp.StatusCode, tt.wantStatus)
			checkBody(t, resp, tt.wantBody)
		})
	}
}

type testdb struct {
	T              *testing.T
	listMessages   func(t *testing.T, excludeMsgIDs ...string) ([]Message, error)
	insertMessage  func(t *testing.T, msg Message) (Message, error)
	insertReaction func(t *testing.T, reaction Reaction) (Reaction, error)
}

func (db *testdb) ListMessages(ctx context.Context, limit int, offset int, excludeMsgIDs ...string) ([]Message, error) {
	return db.listMessages(db.T, excludeMsgIDs...)
}

func (db *testdb) InsertMessage(_ context.Context, msg Message) (Message, error) {
	return db.insertMessage(db.T, msg)
}

func (db *testdb) InsertReaction(_ context.Context, reaction Reaction) (Reaction, error) {
	return db.insertReaction(db.T, reaction)
}

type testcache struct {
	T             *testing.T
	listMessages  func(t *testing.T) ([]Message, error)
	insertMessage func(t *testing.T, msg Message) error
	getMessage    func(t *testing.T, id string) (*Message, error)
	deleteMessage func(t *testing.T, id string) error
}

func (c *testcache) GetMessage(_ context.Context, messageID string) (*Message, error) {
	return c.getMessage(c.T, messageID)
}

func (c *testcache) DeleteMessage(_ context.Context, messageID string) error {
	return c.deleteMessage(c.T, messageID)
}

func (c *testcache) ListMessages(_ context.Context) ([]Message, error) {
	return c.listMessages(c.T)
}

func (c *testcache) InsertMessage(_ context.Context, msg Message) error {
	return c.insertMessage(c.T, msg)
}

type MockValidator struct {
	ShouldFail bool
	Err        error
}

func (m *MockValidator) Struct(interface{}) error {
	if m.ShouldFail {
		return m.Err
	}
	return nil
}

func checkStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("Got HTTP status %d, want %d", got, want)
	}
}

func checkBody(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	gotBody := normalizeJSON(t, resp.Body)
	wantBody := normalizeJSON(t, bytes.NewReader([]byte(want)))
	if gotBody != wantBody {
		t.Errorf("Body does not match\nGot\n  %s\n\nWant\n  %s", gotBody, wantBody)
	}
}

func checkLog(t *testing.T, buffer *bytes.Buffer, want string) {
	t.Helper()

	if s := buffer.String(); want != "" && !strings.Contains(s, want) {
		t.Errorf("Log does not contain  %s\n", want)
	}
}

func normalizeJSON(t *testing.T, r io.Reader) string {
	t.Helper()
	var buf bytes.Buffer
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Could not read JSON: %v", err)
	}
	if err := json.Indent(&buf, b, "  ", "  "); err != nil {
		t.Fatalf("Could not indent JSON: %v", err)
	}
	return strings.TrimSpace(buf.String())
}
