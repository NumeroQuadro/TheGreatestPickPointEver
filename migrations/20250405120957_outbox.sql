-- +goose Up
SELECT 'up SQL query';
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox (
    task_id BIGSERIAL PRIMARY KEY,
    task_status varchar(255) NOT NULL DEFAULT 'CREATED',
    task_type varchar(255) NOT NULL,
    entry_id BIGSERIAL NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    attempts_count INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

INSERT INTO outbox (task_type, entry_id, created_at, updated_at)
SELECT
    'AUDIT_LOG'    AS task_type,
    entry_id,
    created_at,
    created_at
FROM audit_logs;

INSERT INTO outbox (task_type, entry_id, created_at, updated_at)
SELECT
    'ORDER_STATUS_LOG' AS task_type,
    entry_id,
    created_at,
    created_at
FROM order_status_audit;

-- +goose StatementEnd
SELECT 'down SQL query';
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox;
-- +goose StatementEnd
