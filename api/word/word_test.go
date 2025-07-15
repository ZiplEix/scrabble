package word

import (
	"testing"
)

func TestWordExists(t *testing.T) {
	tests := []struct {
		word    string
		expects bool
	}{
		{"chien", true},        // mot courant
		{"ChIeN", true},        // casse
		{"était", true},        // mot avec accent
		{"ÉTÉ", true},          // majuscule + accent
		{"nrogvnerojv", false}, // mot inventé
		{"", false},            // vide
		{"  chien  ", true},    // espaces
		{"caféteria", true},    // mot avec é
		{"électricité", true},  // accent complexe
	}

	for _, test := range tests {
		ok := WordExists(test.word)
		if ok != test.expects {
			t.Errorf("WordExists(%q) = %v; want %v", test.word, ok, test.expects)
		}
	}
}
