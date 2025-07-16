package word

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

//go:embed "fr.txt"
var dictFile embed.FS

var (
	dictionary = make(map[string]struct{})
	once       sync.Once
)

func init() {
	initDict()
}

func WordExists(word string) bool {
	cleanedWord := removeAccents(strings.ToUpper(strings.TrimSpace(word)))
	_, found := dictionary[cleanedWord]
	return found
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
				dictionary[word] = struct{}{}
			}
		}
		if err := scanner.Err(); err != nil {
			panic(fmt.Errorf("error reading dictionary file: %w", err))
		}
		end := time.Now()
		fmt.Printf("Dictionary loaded with %d words in %s\n", len(dictionary), end.Sub(start))
	})
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
