-- Improve indexes to match actual query patterns

-- Leaderboard queries filter: WHERE games_played > 0 AND avg_reaction_ms > 0
-- ORDER BY avg_reaction_ms ASC
-- Old partial index only had WHERE games_played > 0, forcing an extra filter pass.
-- New index matches the exact predicate so PG can use a full index-only scan.
DROP INDEX IF EXISTS idx_users_avg_reaction;
CREATE INDEX IF NOT EXISTS idx_users_leaderboard
    ON users(avg_reaction_ms ASC)
    WHERE games_played > 0 AND avg_reaction_ms > 0;

-- Game history query joins game_session_players to game_sessions:
--   WHERE gsp.user_id = $1 ... JOIN gs ON gs.id = gsp.session_id
-- Old idx_gsp_user(user_id) required a heap fetch to retrieve session_id.
-- Covering index (user_id, session_id) satisfies both the WHERE and the JOIN
-- without touching the heap, and still serves FK lookups on user_id.
DROP INDEX IF EXISTS idx_gsp_user;
CREATE INDEX IF NOT EXISTS idx_gsp_user_session
    ON game_session_players(user_id, session_id);
