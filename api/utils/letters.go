package utils

import "math/rand"

func ShuffleRunes(r []rune) {
	n := len(r)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		r[i], r[j] = r[j], r[i]
	}
}

func DrawLetters(bag *[]rune, n int) []rune {
	count := min(n, len(*bag))
	drawn := make([]rune, count)

	copy(drawn, (*bag)[:count])
	*bag = (*bag)[count:]

	return drawn
}

func DrawLettersFromString(bagStr string, n int) ([]string, string) {
	bag := []rune(bagStr)
	ShuffleRunes(bag)

	drawn := DrawLetters(&bag, n)

	drawnStr := make([]string, len(drawn))
	for i, r := range drawn {
		drawnStr[i] = string(r)
	}
	return drawnStr, string(bag)
}
