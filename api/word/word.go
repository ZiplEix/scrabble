package word

import (
	"bufio"
	"embed"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

//go:embed "fr.txt"
var dictFile embed.FS

var (
	dictionary          = make(map[string]struct{})
	allWords            []string
	wordsByLength       = make(map[int][]string)
	wordsByLetter       = make(map[rune][]string)
	wordsByLetterLength = make(map[rune]map[int][]string)
	once                sync.Once
)

func init() {
	initDict()
}

func WordExists(word string) bool {
	cleanedWord := removeAccents(strings.ToUpper(strings.TrimSpace(word)))
	_, found := dictionary[cleanedWord]
	return found
}

// WordsContainingLetter retourne les mots contenant une lettre donnée,
// filtrés par longueur min/max. Le résultat est indexé, pas une recherche
// complète dans tout le dictionnaire.
func WordsContainingLetter(letter rune, minLen, maxLen int) []string {
	if minLen <= 0 {
		minLen = 1
	}
	if maxLen > 0 && maxLen < minLen {
		return []string{}
	}

	key := normalizeRune(letter)
	if key == 0 {
		return []string{}
	}

	byLen, ok := wordsByLetterLength[key]
	if !ok {
		return []string{}
	}

	if maxLen <= 0 {
		out := make([]string, 0, len(wordsByLetter[key]))
		for _, w := range wordsByLetter[key] {
			if len(w) >= minLen {
				out = append(out, w)
			}
		}
		return out
	}

	out := make([]string, 0)
	for l := minLen; l <= maxLen; l++ {
		if words, exists := byLen[l]; exists {
			out = append(out, words...)
		}
	}
	return out
}

// RandomWord retourne un mot aléatoire dans une plage de longueur.
func RandomWord(rng *rand.Rand, minLen, maxLen int) (string, bool) {
	if rng == nil {
		return "", false
	}
	if minLen <= 0 {
		minLen = 1
	}
	if maxLen > 0 && maxLen < minLen {
		return "", false
	}

	candidates := make([]string, 0)
	if maxLen <= 0 {
		for l, words := range wordsByLength {
			if l >= minLen {
				candidates = append(candidates, words...)
			}
		}
	} else {
		for l := minLen; l <= maxLen; l++ {
			if words, exists := wordsByLength[l]; exists {
				candidates = append(candidates, words...)
			}
		}
	}

	if len(candidates) == 0 {
		return "", false
	}

	return candidates[rng.Intn(len(candidates))], true
}

// AllWordsCount retourne la taille du dictionnaire.
func AllWordsCount() int {
	return len(allWords)
}

// AllWords retourne la liste de tous les mots du dictionnaire.
func AllWords() []string {
	return allWords
}

func initDict() {
	once.Do(func() {
		start := time.Now()
		f, err := dictFile.Open("fr.txt")
		if err != nil {
			panic(fmt.Errorf("failed to load embedded dictionary: %w", err))
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			word := strings.ToUpper(scanner.Text())
			word = removeAccents(strings.TrimSpace(word))
			if word != "" {
				if _, exists := dictionary[word]; exists {
					continue
				}
				dictionary[word] = struct{}{}
				allWords = append(allWords, word)

				wlen := len(word)
				wordsByLength[wlen] = append(wordsByLength[wlen], word)

				seenLetters := make(map[rune]struct{})
				for _, r := range word {
					r = normalizeRune(r)
					if r == 0 {
						continue
					}
					if _, seen := seenLetters[r]; seen {
						continue
					}
					seenLetters[r] = struct{}{}

					wordsByLetter[r] = append(wordsByLetter[r], word)
					if wordsByLetterLength[r] == nil {
						wordsByLetterLength[r] = make(map[int][]string)
					}
					wordsByLetterLength[r][wlen] = append(wordsByLetterLength[r][wlen], word)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			panic(fmt.Errorf("error reading dictionary file: %w", err))
		}
		end := time.Now()
		fmt.Printf("Dictionary loaded with %d words in %s\n", len(dictionary), end.Sub(start))
	})
}

func normalizeRune(r rune) rune {
	u := strings.ToUpper(removeAccents(string(r)))
	if u == "" {
		return 0
	}
	rr := []rune(u)
	if len(rr) == 0 {
		return 0
	}
	return rr[0]
}

func removeAccents(s string) string {
	t := norm.NFD.String(s)
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, t)
}
