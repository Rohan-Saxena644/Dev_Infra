-- +goose Up
CREATE TABLE deployments(
    id SERIAL PRIMARY KEY,
    project_id INT NOT NULL REFERENCES projects(id),
    status TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE deployments;