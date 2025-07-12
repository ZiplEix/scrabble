package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
)

func drawLetters(available *[]rune, count int) []rune {
	if len(*available) < count {
		count = len(*available)
	}

	drawn := make([]rune, 0, count)
	for i := 0; i < count; i++ {
		index := time.Now().UnixNano() % int64(len(*available))
		drawn = append(drawn, (*available)[index])
		*available = append((*available)[:index], (*available)[index+1:]...)
	}

	return drawn
}

func rackContains(rack string, letters []request.PlacedLetter) bool {
	rackCount := map[rune]int{}
	for _, r := range rack {
		rackCount[r]++
	}
	for _, l := range letters {
		c := rune(l.Char[0])
		if rackCount[c] == 0 {
			return false
		}
		rackCount[c]--
	}
	return true
}

func updatePlayerRack(gameID string, userID int64, rack string, played []request.PlacedLetter) (string, error) {
	// 1. Retirer les lettres jouées
	for _, l := range played {
		c := rune(l.Char[0])
		i := strings.IndexRune(rack, c)
		if i == -1 {
			return "", fmt.Errorf("letter %c not in rack", c)
		}
		rack = rack[:i] + rack[i+1:]
	}

	// 2. Tirer des nouvelles lettres du sac
	var available string
	err := database.QueryRow(`SELECT available_letters FROM games WHERE id = $1`, gameID).Scan(&available)
	if err != nil {
		return "", err
	}

	drawCount := 7 - len(rack)
	availableRunes := []rune(available)
	newLetters := drawLetters(&availableRunes, drawCount)

	// 3. Mettre à jour le sac
	_, err = database.Pool.Exec(context.Background(),
		`UPDATE games SET available_letters = $1 WHERE id = $2`, string(availableRunes), gameID)
	if err != nil {
		return "", err
	}

	return rack + string(newLetters), nil
}
