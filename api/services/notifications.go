package services

import "github.com/ZiplEix/scrabble/api/database"

func PushSubscribe(userID int64, subBytes []byte) error {
	_, err := database.Exec(`
		INSERT INTO push_subscriptions (user_id, subscription, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE SET subscription = $2, created_at = NOW()
	`, userID, string(subBytes))

	if err != nil {
		return err
	}
	return nil
}

func PushUnsubscribe(userID int64) error {
	_, err := database.Exec(`
		DELETE FROM push_subscriptions WHERE user_id = $1
	`, userID)

	if err != nil {
		return err
	}
	return nil
}
