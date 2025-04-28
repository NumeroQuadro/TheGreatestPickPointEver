-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS order_status_audit (
    entry_id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    previous_status varchar(255),
    current_status varchar(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_status_audit;
-- +goose StatementEnd
