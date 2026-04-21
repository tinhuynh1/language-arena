-- Migration: Replace score-based ranking with avg reaction time
-- Adds avg_reaction_ms and total_correct columns to users table

ALTER TABLE users ADD COLUMN IF NOT EXISTS avg_reaction_ms INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_correct INT DEFAULT 0;

-- New index for leaderboard: sort by avg reaction time ascending (lower = better)
CREATE INDEX IF NOT EXISTS idx_users_avg_reaction ON users(avg_reaction_ms ASC) WHERE games_played > 0;

-- Drop old score-based index (no longer used for ranking)
DROP INDEX IF EXISTS idx_users_total_score;
