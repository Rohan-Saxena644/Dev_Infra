-- +goose Up
ALTER TABLE deployments
ADD COLUMN deployment_type TEXT NOT NULL DEFAULT 'dockerfile'
CHECK (deployment_type IN ('dockerfile', 'compose'));

-- +goose Down
ALTER TABLE deployments
DROP COLUMN deployment_type;
