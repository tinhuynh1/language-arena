-- Backfill avg_reaction_ms from historical game_session_players data
-- This runs after 002_reaction_scoring.sql which added the columns

UPDATE users u SET
  avg_reaction_ms = COALESCE(sub.avg_ms, 0),
  total_correct = COALESCE(sub.total_correct, 0)
FROM (
  SELECT
    gsp.user_id,
    CASE WHEN SUM(CASE WHEN gsp.avg_reaction_ms > 0 THEN 1 ELSE 0 END) > 0
      THEN SUM(CASE WHEN gsp.avg_reaction_ms > 0 THEN gsp.avg_reaction_ms ELSE 0 END)
           / SUM(CASE WHEN gsp.avg_reaction_ms > 0 THEN 1 ELSE 0 END)
      ELSE 0
    END AS avg_ms,
    SUM(gsp.score) AS total_correct
  FROM game_session_players gsp
  GROUP BY gsp.user_id
) sub
WHERE u.id = sub.user_id AND u.games_played > 0;
