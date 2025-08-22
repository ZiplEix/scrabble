package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <fichier>\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}
	path := os.Args[1]

	// 1) Lire le fichier
	in, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	// 2) Enlever doublons (après trim + normalisation NFC)
	seen := make(map[string]struct{})
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		line = norm.NFC.String(line)
		seen[line] = struct{}{}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	// 3) Mettre en slice et trier avec collation française
	words := make([]string, 0, len(seen))
	for w := range seen {
		words = append(words, w)
	}
	coll := collate.New(language.French)
	coll.SortStrings(words)

	// 4) Écrire de façon atomique (fichier temporaire + rename)
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".sorted-*.tmp")
	if err != nil {
		log.Fatal(err)
	}
	bw := bufio.NewWriter(tmp)
	for _, w := range words {
		if _, err := bw.WriteString(w + "\n"); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			log.Fatal(err)
		}
	}
	if err := bw.Flush(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		log.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		log.Fatal(err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		log.Fatal(err)
	}
}
