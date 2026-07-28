-- name: CreateProject :one
INSERT INTO projects (
    name,
    repo_url,
    user_id
)
VALUES (
    $1,
    $2,
    (SELECT id FROM users WHERE email = 'demo@gmail.com')
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

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;
