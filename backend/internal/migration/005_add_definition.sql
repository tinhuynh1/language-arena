-- Add definition column for English definition-based quiz
ALTER TABLE vocabularies ADD COLUMN IF NOT EXISTS definition VARCHAR(500);
