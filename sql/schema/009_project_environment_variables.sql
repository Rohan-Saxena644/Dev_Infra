-- +goose Up
CREATE TABLE project_env_vars (
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    encrypted_value BYTEA NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, name),
    CHECK (name ~ '^[A-Za-z_][A-Za-z0-9_]*$'),
    CHECK (char_length(name) <= 128)
);

-- +goose Down
DROP TABLE project_env_vars;
