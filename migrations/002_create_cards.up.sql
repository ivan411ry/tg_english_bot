CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    word TEXT NOT NULL CHECK (btrim(word) <> ''),
    translation TEXT NOT NULL CHECK (btrim(translation) <> ''),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, word, translation)
);
