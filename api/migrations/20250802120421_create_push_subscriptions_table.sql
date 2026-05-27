-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS push_subscriptions (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    subscription JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS push_subscriptions;
-- +goose StatementEnd
