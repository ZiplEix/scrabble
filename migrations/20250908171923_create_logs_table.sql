-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS logs (
  id           bigserial PRIMARY KEY,
  received_at  timestamptz NOT NULL DEFAULT now(),
  raw          jsonb       NOT NULL,
  req_id       text
);

CREATE INDEX IF NOT EXISTS idx_logs_received_brin ON logs USING brin (received_at);
CREATE INDEX IF NOT EXISTS idx_logs_raw_gin      ON logs USING gin  (raw);
CREATE INDEX IF NOT EXISTS idx_logs_req_id ON logs (req_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs;
-- +goose StatementEnd
