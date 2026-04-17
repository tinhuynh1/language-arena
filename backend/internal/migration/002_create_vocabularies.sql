CREATE TABLE IF NOT EXISTS vocabularies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(100) NOT NULL,
    meaning VARCHAR(255) NOT NULL,
    language VARCHAR(5) NOT NULL,
    difficulty INT DEFAULT 1 CHECK (difficulty BETWEEN 1 AND 3),
    category VARCHAR(50)
);

CREATE INDEX idx_vocabularies_language ON vocabularies(language);
CREATE INDEX idx_vocabularies_difficulty ON vocabularies(language, difficulty);
