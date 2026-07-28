-- +goose Up
CREATE TABLE deployment_usage (
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX deployment_usage_created_at_idx
ON deployment_usage (created_at);

CREATE FUNCTION enforce_daily_deployment_limit()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(861042);

    DELETE FROM deployment_usage
    WHERE created_at < NOW() - INTERVAL '24 hours';

    IF (SELECT COUNT(*) FROM deployment_usage) >= 10 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            CONSTRAINT = 'deployments_daily_limit',
            MESSAGE = 'daily deployment limit reached';
    END IF;

    INSERT INTO deployment_usage DEFAULT VALUES;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER deployments_daily_limit
BEFORE INSERT ON deployments
FOR EACH ROW
EXECUTE FUNCTION enforce_daily_deployment_limit();

-- +goose Down
DROP TRIGGER deployments_daily_limit ON deployments;
DROP FUNCTION enforce_daily_deployment_limit();
DROP TABLE deployment_usage;
