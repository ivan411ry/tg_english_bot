-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    word TEXT NOT NULL CHECK (btrim(word) <> ''),
    translation TEXT NOT NULL CHECK (btrim(translation) <> ''),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, word, translation)
);

-- +goose Down
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS users;