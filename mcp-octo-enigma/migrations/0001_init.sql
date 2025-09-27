-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS documents (
  id TEXT PRIMARY KEY,
  content TEXT NOT NULL,
  embedding vector(1536) NOT NULL,
  metadata JSONB DEFAULT {}::jsonb,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_documents_embedding 
ON documents USING ivfflat (embedding vector_l2_ops) WITH (lists = 100);

-- +goose Down
DROP TABLE IF EXISTS documents;
