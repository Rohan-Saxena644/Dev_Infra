-- name: UpsertProjectEnvVar :exec
INSERT INTO project_env_vars (project_id, name, encrypted_value)
VALUES ($1, $2, $3)
ON CONFLICT (project_id, name)
DO UPDATE SET encrypted_value = EXCLUDED.encrypted_value, updated_at = NOW();

-- name: ListProjectEnvVars :many
SELECT name, encrypted_value
FROM project_env_vars
WHERE project_id = $1
ORDER BY name;

-- name: DeleteProjectEnvVar :exec
DELETE FROM project_env_vars
WHERE project_id = $1 AND name = $2;
