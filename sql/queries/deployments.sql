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


-- name: UpdateDeploymentStatus :one
UPDATE deployments
SET status = $2
WHERE id = $1
RETURNING *;


-- name: GetDeployment :one
SELECT *
FROM deployments
WHERE id = $1;