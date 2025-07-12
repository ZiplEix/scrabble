package database

import "fmt"

func Migrate() error {
	if err := createUserTable(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := createGameTable(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := createGamePlayersTable(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := createGameMovesTable(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Database migration completed successfully")
	return nil
}

func createUserTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT now()
		);
	`

	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	return nil
}

func createGameTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS games (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			created_by INT REFERENCES users(id),
			status TEXT NOT NULL DEFAULT 'ongoing',
			current_turn INT REFERENCES users(id),
			board JSONB NOT NULL,
			available_letters TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT now()
		);
	`

	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	return nil
}

func createGamePlayersTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS game_players (
			game_id UUID REFERENCES games(id),
			player_id INT REFERENCES users(id),
			rack TEXT NOT NULL,
			position INT NOT NULL,
			score INT NOT NULL DEFAULT 0,
			PRIMARY KEY (game_id, player_id)
		);
	`

	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to create game_players table: %w", err)
	}
	return nil
}

func createGameMovesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS game_moves (
			id SERIAL PRIMARY KEY,
			game_id UUID REFERENCES games(id),
			player_id INT REFERENCES users(id),
			move JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT now()
		);
	`

	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to create game_moves table: %w", err)
	}
	return nil
}
