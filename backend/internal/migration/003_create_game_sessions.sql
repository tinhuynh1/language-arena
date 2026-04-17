CREATE TABLE IF NOT EXISTS game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mode VARCHAR(10) NOT NULL CHECK (mode IN ('solo', 'duel')),
    language VARCHAR(5) NOT NULL,
    player1_id UUID NOT NULL REFERENCES users(id),
    player2_id UUID REFERENCES users(id),
    player1_score INT DEFAULT 0,
    player2_score INT DEFAULT 0,
    winner_id UUID REFERENCES users(id),
    rounds INT DEFAULT 10,
    avg_reaction_ms INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_game_sessions_player1 ON game_sessions(player1_id);
CREATE INDEX idx_game_sessions_player2 ON game_sessions(player2_id);
CREATE INDEX idx_game_sessions_created ON game_sessions(created_at DESC);
