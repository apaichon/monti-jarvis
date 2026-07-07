ALTER TABLE km_embeddings ADD COLUMN IF NOT EXISTS created_at DateTime DEFAULT now();
ALTER TABLE km_embeddings ADD COLUMN IF NOT EXISTS created_by String DEFAULT 'system';
ALTER TABLE km_embeddings ADD COLUMN IF NOT EXISTS updated_by String DEFAULT 'system';
ALTER TABLE qa_events ADD COLUMN IF NOT EXISTS updated_at DateTime DEFAULT now();
ALTER TABLE qa_events ADD COLUMN IF NOT EXISTS created_by String DEFAULT 'system';
ALTER TABLE qa_events ADD COLUMN IF NOT EXISTS updated_by String DEFAULT 'system';