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
ORDER BY deployments.created_at DESC;


-- name: UpdateDeploymentStatus :one
UPDATE deployments
SET status = $2
WHERE id = $1
RETURNING *;

-- name: ClaimDeployment :one
UPDATE deployments
SET status = 'running'
WHERE id = $1
AND status = 'queued'
RETURNING *;

-- name: UpdateDeploymentStatusIfCurrent :one
UPDATE deployments
SET status = sqlc.arg(new_status)
WHERE id = sqlc.arg(id)
AND status = sqlc.arg(current_status)
RETURNING *;


-- name: GetDeployment :one
SELECT *
FROM deployments
WHERE id = $1;


-- name: UpdateDeploymentPort :exec
UPDATE deployments
SET port = $2
WHERE id = $1;

-- name: UpdateDeploymentType :exec
UPDATE deployments
SET deployment_type = $2
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



-- name: GetActiveDeploymentByProject :one
SELECT *
FROM deployments
WHERE project_id = $1
AND status IN ('queued', 'running')
ORDER BY created_at DESC
LIMIT 1;

-- name: GetDeploymentsForCleanup :many
SELECT *
FROM deployments
WHERE (status = 'success' AND created_at < NOW() - INTERVAL '1 hour')
OR (status IN ('queued', 'running') AND created_at < NOW() - INTERVAL '30 minutes')
ORDER BY created_at;
