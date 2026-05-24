-- +goose Up
-- +goose StatementBegin
INSERT INTO achievements (id, title, description, badge_icon, category) VALUES
('bot_slayer', 'Tueur de Géants', 'Battre Scrabby (l''ordinateur) lors d''une partie en 1vs1', '🤖', 'special')
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM achievements WHERE id = 'bot_slayer';
-- +goose StatementEnd
