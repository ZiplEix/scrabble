-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS achievements (
    id VARCHAR(50) PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    badge_icon VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_achievements (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id VARCHAR(50) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id)
);

CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);

INSERT INTO achievements (id, title, description, badge_icon, category) VALUES
('first_blood', 'Premier Sang', 'Remporter votre toute première victoire', '🏆', 'milestone'),
('bingo', 'Bingo !', 'Poser un mot de 7 lettres d''un coup (bonus de 50 points)', '⚡', 'gameplay'),
('scrabble_master', 'Grand Maître', 'Marquer plus de 400 points dans une seule partie', '👑', 'milestone'),
('high_scorer', 'Coup de Génie', 'Marquer plus de 75 points en un seul coup', '🎯', 'gameplay'),
('puzzle_solver', 'As des Puzzles', 'Résoudre correctement un puzzle quotidien', '🧩', 'special'),
('marathoner', 'Marathonien', 'Jouer 10 parties complètes', '🏃', 'milestone'),
('word_smith', 'Amoureux des Mots', 'Placer un mot avec une lettre rare (K, W, X, Y, Z) rapportant au moins 30 points', '✍️', 'gameplay'),
('friendly_rivalry', 'Rivalité Amicale', 'Jouer contre au moins 3 adversaires différents', '⚔️', 'special'),
('comeback_kid', 'Le Survivant', 'Gagner une partie avec moins de 10 points d''avance', '🛡️', 'special'),
('first_step', 'Premier Pas', 'Placer votre tout premier mot dans une partie', '👣', 'gameplay'),
('joker_master', 'L''Illusionniste', 'Placer un mot utilisant un joker (tuile blanche)', '🎭', 'gameplay'),
('long_word', 'Linguiste Émérite', 'Poser un mot de 8 lettres ou plus', '📚', 'gameplay'),
('half_century', 'Demi-Siècle', 'Marquer au moins 50 points en un seul coup', '🏅', 'gameplay'),
('veteran', 'Vétéran', 'Jouer 50 parties complètes', '🎖️', 'milestone'),
('elite_player', 'Légende du Club', 'Marquer plus de 500 points dans une seule partie', '🌌', 'milestone'),
('serial_winner', 'Série Victorieuse', 'Remporter 3 victoires consécutives', '🔥', 'special'),
('night_owl', 'Oiseau de Nuit', 'Terminer une partie entre 23h et 5h du matin', '🦉', 'special'),
('ips_master', 'Stratège', 'Atteindre un classement de 500 IPS ou plus', '📈', 'milestone'),
('sharp_mind', 'Esprit Vif', 'Résoudre 5 puzzles quotidiens', '💡', 'special'),
('puzzle_expert', 'Maître des Énigmes', 'Résoudre 15 puzzles quotidiens', '🔮', 'special'),
('chatty', 'Bavard', 'Envoyer un message dans le chat d''une partie', '💬', 'special')
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_achievements;
DROP TABLE IF EXISTS achievements;
-- +goose StatementEnd
