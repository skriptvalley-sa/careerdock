-- Add name and updated_at columns to curated_lists for management UX.
ALTER TABLE curated_lists ADD COLUMN name VARCHAR(255);
ALTER TABLE curated_lists ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Backfill name with a sensible default.
UPDATE curated_lists SET name = 'Curated List', updated_at = created_at;
