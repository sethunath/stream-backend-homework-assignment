
-- Messages
CREATE TABLE IF NOT EXISTS messages (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  message_text TEXT NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_user_id ON messages (user_id);
CREATE INDEX idx_messages_created_at ON messages (created_at DESC);

-- Message Reaction Type
CREATE TYPE reaction_type AS ENUM ('like', 'love', 'laugh', 'sad', 'clap', 'wow');
-- Message Reactions
CREATE TABLE IF NOT EXISTS message_reactions (
     id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
     message_id UUID NOT NULL,
     user_id VARCHAR(255) NOT NULL,
     type reaction_type NOT NULL,
     score INT DEFAULT 1,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     CONSTRAINT fk_message FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
     CONSTRAINT unique_reaction UNIQUE (message_id, user_id)
);

-- Indexes
CREATE INDEX idx_message_reactions_message_id ON message_reactions (message_id);
CREATE INDEX idx_message_reactions_user_id ON message_reactions (user_id);
CREATE INDEX idx_message_reactions_created_at ON message_reactions (created_at DESC);
