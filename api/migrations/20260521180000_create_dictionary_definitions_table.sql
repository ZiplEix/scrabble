-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dictionary_definitions (
    word TEXT PRIMARY KEY,
    definitions JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dictionary_definitions_word ON dictionary_definitions(word);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dictionary_definitions;
-- +goose StatementEnd
