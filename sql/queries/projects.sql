-- name: CreateProject :one
INSERT INTO projects (
    name,
    repo_url,
    user_id
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetProjectsByUser :many
SELECT *
FROM projects
WHERE user_id = $1
ORDER BY created_at DESC;


-- name: GetProjectByUser :one
SELECT *
FROM projects
WHERE id = $1
AND user_id = $2;

-- name: GetProject :one
SELECT *
FROM projects
WHERE id = $1;

-- name: DeleteProjectByUser :exec
DELETE FROM projects
WHERE id = $1
AND user_id = $2;
