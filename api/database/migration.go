package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ZiplEix/scrabble/api/word"
)

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

	if err := migrateAddPassCount(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := migrateAddGameEndInfo(); err != nil {
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

	if err := createPushSubscriptionTable(); err != nil {
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

func migrateAddPassCount() error {
	query := `
		ALTER TABLE games
		ADD COLUMN IF NOT EXISTS pass_count INT NOT NULL DEFAULT 0;
	`
	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to add pass_count to games: %w", err)
	}
	return nil
}

func migrateAddGameEndInfo() error {
	query := `
        ALTER TABLE games
        ADD COLUMN IF NOT EXISTS winner_username TEXT,
        ADD COLUMN IF NOT EXISTS ended_at TIMESTAMP;
    `
	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to add end-of-game info to games table: %w", err)
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

func createPushSubscriptionTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS push_subscriptions (
			user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			subscription JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`
	_, err := Query(query)
	if err != nil {
		return fmt.Errorf("failed to create push_subscriptions table: %w", err)
	}
	return nil
}

func EndGameBackTracker() error {
	rows, err := Query(`
        SELECT id, pass_count, available_letters
        FROM games
    `)
	if err != nil {
		return fmt.Errorf("failed to select games for backfill: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			gameID       string
			passCount    int
			availLetters string
		)
		if err := rows.Scan(&gameID, &passCount, &availLetters); err != nil {
			return err
		}

		// Nombre de joueurs dans la partie
		var playerCount int
		if err := QueryRow(
			`SELECT COUNT(*) FROM game_players WHERE game_id = $1`,
			gameID,
		).Scan(&playerCount); err != nil {
			return err
		}

		// On ouvre une transaction pour appliquer finishGame
		tx, err := DB.Begin()
		if err != nil {
			return err
		}

		ended := false
		var lastPlayerID int64

		// 2.a) fin par passes successives ?
		if passCount >= playerCount*2 {
			ended = true
			// lastPlayerID = 0 → finishGame fera juste les retraits de racks
		} else if len(availLetters) == 0 {
			// 2.b) fin par épuisement du sac + rack vide ?
			// on recherche un joueur dont le rack est vide
			err := tx.QueryRow(
				`SELECT player_id FROM game_players WHERE game_id=$1 AND rack = '' LIMIT 1`,
				gameID,
			).Scan(&lastPlayerID)
			if err != nil && err != sql.ErrNoRows {
				tx.Rollback()
				return err
			}
			if lastPlayerID != 0 {
				ended = true
			}
		}

		if ended {
			// appliquer les pénalités et bonus, et écrire ended_at + winner_username
			if err := finishGame(tx, gameID, lastPlayerID); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// pas terminé → on annule la tx
			tx.Rollback()
			continue
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return rows.Err()
}

func rackPoints(rack string) int {
	pts := 0
	for _, c := range rack {
		pts += word.LetterValues[strings.ToUpper(string(c))]
	}
	return pts
}

func finishGame(tx *sql.Tx, gameID string, lastPlayerID int64) error {
	rows, err := tx.Query(
		`SELECT player_id, rack FROM game_players WHERE game_id = $1`, gameID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	totalLeftover := 0
	for rows.Next() {
		var pid int64
		var rack string
		if err := rows.Scan(&pid, &rack); err != nil {
			return err
		}
		lp := rackPoints(rack)
		// retire les points non joués
		if _, err := tx.Exec(
			`UPDATE game_players SET score = score - $1 WHERE game_id = $2 AND player_id = $3`,
			lp, gameID, pid,
		); err != nil {
			return err
		}
		// cumule pour le bonus
		if pid != lastPlayerID {
			totalLeftover += lp
		}
	}
	// bonus pour le finisseur (si lastPlayerID != 0)
	if lastPlayerID != 0 && totalLeftover > 0 {
		if _, err := tx.Exec(
			`UPDATE game_players SET score = score + $1 WHERE game_id = $2 AND player_id = $3`,
			totalLeftover, gameID, lastPlayerID,
		); err != nil {
			return err
		}
	}

	// on marque la partie terminée, stocke le vainqueur et l'heure
	var winnerUsername sql.NullString
	if lastPlayerID != 0 {
		// on récupère le username du vainqueur
		if err := tx.QueryRow(
			`SELECT username FROM users WHERE id = $1`, lastPlayerID,
		).Scan(&winnerUsername); err != nil {
			return err
		}
	}
	_, err = tx.Exec(
		`UPDATE games
           SET status           = 'ended',
               winner_username  = $1,
               ended_at         = now()
         WHERE id = $2`,
		winnerUsername.String, gameID,
	)
	if err != nil {
		return err
	}

	return nil
}
