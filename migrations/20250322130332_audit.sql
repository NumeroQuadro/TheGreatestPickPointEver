-- +goose Up
SELECT 'up SQL query';
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS audit_logs (
    entry_id BIGSERIAL PRIMARY KEY,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    request_header JSONB,
    request_body JSONB,
    query_params JSONB,
    status_code INTEGER NOT NULL,
    response_body JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query';
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd
