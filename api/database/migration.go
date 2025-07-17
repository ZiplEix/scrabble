package database

import "fmt"

func Migrate() error {
	if err := createUserTable(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := migrateAddUserRole(); err != nil {
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

	if err := createReportTable(); err != nil {
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

func migrateAddUserRole() error {
	query := `
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user';
	`

	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to alter users table to add role column: %w", err)
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

func createReportTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS reports (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE SET NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'open', -- open, in_progress, resolved, rejected
			priority TEXT DEFAULT 'normal',     -- optional: low, normal, high, urgent
			type TEXT DEFAULT 'bug',            -- optional: bug, suggestion, feedback, etc.
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_update_report_updated_at'
			) THEN
				CREATE TRIGGER trigger_update_report_updated_at
				BEFORE UPDATE ON reports
				FOR EACH ROW
				EXECUTE FUNCTION update_updated_at_column();
			END IF;
		END;
		$$;
	`

	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to create reports table: %w", err)
	}
	return nil
}
