-- +goose Up
INSERT INTO users (email, password_hash)
VALUES ('demo@gmail.com', 'login-disabled')
ON CONFLICT (email) DO NOTHING;

UPDATE projects
SET user_id = (SELECT id FROM users WHERE email = 'demo@gmail.com');

ALTER TABLE projects
ALTER COLUMN user_id SET NOT NULL;

CREATE OR REPLACE FUNCTION enforce_global_deployment_limit()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(861042);

    IF (
        SELECT COUNT(*)
        FROM deployments
        WHERE status IN ('queued', 'running', 'success')
    ) >= 10 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            CONSTRAINT = 'deployments_global_limit',
            MESSAGE = 'global deployment limit reached';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER deployments_global_limit
BEFORE INSERT ON deployments
FOR EACH ROW
EXECUTE FUNCTION enforce_global_deployment_limit();

-- +goose Down
DROP TRIGGER IF EXISTS deployments_global_limit ON deployments;
DROP FUNCTION IF EXISTS enforce_global_deployment_limit();

ALTER TABLE projects
ALTER COLUMN user_id DROP NOT NULL;
