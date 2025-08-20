package database

import "time"

// Message represents a chat message stored in the database.
type Message struct {
	ID        int64      `db:"id"`
	GameID    string     `db:"game_id"`
	UserID    int64      `db:"user_id"`
	Content   string     `db:"content"`
	Meta      []byte     `db:"meta"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
