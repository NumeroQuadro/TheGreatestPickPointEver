-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE IF NOT EXISTS orders (
    order_id BIGINT UNIQUE PRIMARY KEY NOT NULL,
    user_id BIGINT NOT NULL,
    expiration_date TIMESTAMP not null,
    status varchar(255) NOT NULL DEFAULT 'confirmed',
    weight integer NOT NULL,
    cost integer NOT NULL,
    last_changed_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW () NOT NULL
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE orders;