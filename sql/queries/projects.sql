-- name: CreateProject :one
INSERT INTO projects (
    name,
    repo_url
)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: GetProjects :many
SELECT *
FROM projects
ORDER BY created_at DESC;


-- name: GetProject :one
SELECT *
FROM projects
WHERE id = $1;