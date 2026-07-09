-- +goose Up
ALTER TABLE projects
ADD COLUMN user_id INT REFERENCES users(id);

-- +goose Down
ALTER TABLE projects
DROP COLUMN user_id;