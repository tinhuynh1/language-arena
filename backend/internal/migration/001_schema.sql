-- Schema: all tables and indexes (idempotent — uses IF NOT EXISTS)

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    total_score BIGINT DEFAULT 0,
    games_played INT DEFAULT 0,
    best_reaction_ms INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_total_score ON users(total_score DESC);

CREATE TABLE IF NOT EXISTS vocabularies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(100) NOT NULL,
    meaning VARCHAR(255) NOT NULL,
    language VARCHAR(5) NOT NULL,
    level VARCHAR(10) NOT NULL DEFAULT 'A1',
    difficulty INT DEFAULT 1 CHECK (difficulty BETWEEN 1 AND 3),
    category VARCHAR(50),
    ipa VARCHAR(150),
    pinyin VARCHAR(150)
);

CREATE INDEX IF NOT EXISTS idx_vocabularies_language ON vocabularies(language);
CREATE INDEX IF NOT EXISTS idx_vocabularies_level ON vocabularies(language, level);

CREATE TABLE IF NOT EXISTS game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mode VARCHAR(10) NOT NULL CHECK (mode IN ('solo', 'duel', 'battle')),
    language VARCHAR(5) NOT NULL,
    winner_id UUID REFERENCES users(id),
    rounds INT DEFAULT 10,
    avg_reaction_ms INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS game_session_players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    score INT DEFAULT 0,
    avg_reaction_ms INT DEFAULT 0,
    best_reaction_ms INT DEFAULT 0,
    rank INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gsp_session ON game_session_players(session_id);
CREATE INDEX IF NOT EXISTS idx_gsp_user ON game_session_players(user_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_created ON game_sessions(created_at DESC);
