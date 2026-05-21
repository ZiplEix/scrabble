package services

import (
	"strings"

	"github.com/ZiplEix/scrabble/api/database"
)

// GetDictionaryDefinition retrieves the definition of a word.
// It returns the JSON bytes of the definition, or sql.ErrNoRows if not found.
func GetDictionaryDefinition(word string) (string, error) {
	upperWord := strings.ToUpper(strings.TrimSpace(word))
	var definitionsJson string
	err := database.QueryRow(`
		SELECT definitions FROM dictionary_definitions WHERE word = $1
	`, upperWord).Scan(&definitionsJson)

	if err != nil {
		return "", err
	}
	return definitionsJson, nil
}

// SaveDictionaryDefinition saves the definition of a word.
// definitionsJson is the JSON string containing the structured definition.
func SaveDictionaryDefinition(word string, definitionsJson string) error {
	upperWord := strings.ToUpper(strings.TrimSpace(word))
	_, err := database.Exec(`
		INSERT INTO dictionary_definitions (word, definitions, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (word) DO UPDATE SET definitions = $2, created_at = NOW()
	`, upperWord, definitionsJson)

	return err
}
