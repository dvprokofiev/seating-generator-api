-- +goose Up
CREATE TABLE seating_plans (
    id SERIAL PRIMARY KEY,
    teacher_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    share_id UUID DEFAULT gen_random_uuid () UNIQUE,
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска по JSON, если рассадок будет много
CREATE INDEX idx_seating_plans_data ON seating_plans USING GIN (data);

-- +goose Down
DROP TABLE seating_plans;