-- +goose Up
ALTER TABLE deployments
ALTER COLUMN port DROP NOT NULL;

-- +goose Down
ALTER TABLE deployments
ALTER COLUMN port SET NOT NULL;