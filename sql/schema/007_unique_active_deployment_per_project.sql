-- +goose Up
CREATE UNIQUE INDEX deployments_one_active_per_project
ON deployments (project_id)
WHERE status IN ('queued', 'running');

-- +goose Down
DROP INDEX IF EXISTS deployments_one_active_per_project;
