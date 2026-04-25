-- 008_user_mistakes.sql
-- Create table to track user mistake history for targeted quizzes
CREATE TABLE IF NOT EXISTS user_mistakes (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    vocab_id UUID REFERENCES vocabularies(id) ON DELETE CASCADE,
    quiz_type VARCHAR(50) NOT NULL,
    incorrect_count INT DEFAULT 1,
    correct_count INT DEFAULT 0,
    last_mistake_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, vocab_id, quiz_type)
);

CREATE INDEX IF NOT EXISTS idx_user_mistakes_user_quiz ON user_mistakes(user_id, quiz_type);
