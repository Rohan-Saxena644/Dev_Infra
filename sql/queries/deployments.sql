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

-- name: GetDeploymentsByUser :many
SELECT deployments.*
FROM deployments
JOIN projects ON projects.id = deployments.project_id
WHERE projects.user_id = $1
ORDER BY deployments.created_at DESC;


-- name: UpdateDeploymentStatus :one
UPDATE deployments
SET status = $2
WHERE id = $1
RETURNING *;


-- name: GetDeployment :one
SELECT *
FROM deployments
WHERE id = $1;


-- name: UpdateDeploymentPort :exec
UPDATE deployments
SET port = $2
WHERE id = $1;

-- name: GetDeploymentsByProject :many
SELECT *
FROM deployments
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: DeleteDeployment :exec
DELETE FROM deployments
WHERE id = $1;

-- name: DeleteDeploymentsByProject :exec
DELETE FROM deployments
WHERE project_id = $1;
