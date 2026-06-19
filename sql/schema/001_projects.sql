-- +goose Up
CREATE TABLE projects(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    repo_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE projects;