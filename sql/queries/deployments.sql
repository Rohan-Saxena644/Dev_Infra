-- name: CreateDeployment :one
INSERT INTO deployments (
    project_id,
    status
)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: GetDeployments :many
SELECT *
FROM deployments
ORDER BY created_at DESC;