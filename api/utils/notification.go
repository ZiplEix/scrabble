package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/ZiplEix/scrabble/api/database"
)

type Subscription struct {
	Endpoint       string            `json:"endpoint"`
	ExpirationTime *int64            `json:"expirationTime"`
	Keys           map[string]string `json:"keys"` // p256dh, auth
}

func SendNotificationToUserByID(userID int64, payload string) error {
	sub, err := GetPushSubscription(userID)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("no push subscription found for user ID %d", userID)
	}

	return SendNotification(*sub, payload)
}

func SendNotification(sub Subscription, message string) error {
	subJSON, _ := json.Marshal(sub)
	s := &webpush.Subscription{}
	_ = json.Unmarshal(subJSON, s)

	resp, err := webpush.SendNotification([]byte(message), s, &webpush.Options{
		Subscriber:      "mailto:admin@yourdomain.com",
		VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		TTL:             60,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func GetPushSubscription(userID int64) (*Subscription, error) {
	row := database.QueryRow(`SELECT subscription FROM push_subscriptions WHERE user_id = $1`, userID)

	var subJSON string
	err := row.Scan(&subJSON)
	if err != nil {
		return nil, err
	}

	var sub Subscription
	if err := json.Unmarshal([]byte(subJSON), &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}
