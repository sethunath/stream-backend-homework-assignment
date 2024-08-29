-- Add your db schema creation SQL for reactions here

-- Messages
CREATE TABLE IF NOT EXISTS messages (
  id uuid DEFAULT gen_random_uuid(),
  message_text TEXT NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
