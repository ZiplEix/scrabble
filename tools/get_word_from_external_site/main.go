package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

// FetchHTML récupère le contenu HTML d'une URL et renvoie les octets en UTF-8.
// La requête utilise le contexte fourni (généralement avec un timeout).
func FetchHTML(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// User-Agent "navigateur" pour éviter certains blocages
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoFetcher/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("réponse HTTP non OK: %s", resp.Status)
	}

	// Lire tout le corps
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Déterminer l'encodage et convertir en UTF-8 si nécessaire
	enc, _, _ := charset.DetermineEncoding(raw, resp.Header.Get("Content-Type"))
	reader := transform.NewReader(bytes.NewReader(raw), enc.NewDecoder())

	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return utf8Data, nil
}

// ExtractInnerHTMLFromSpans retourne l'HTML interne des balises <span> contenant la classe donnée.
func ExtractInnerHTMLFromSpans(htmlBytes []byte, class string) ([]string, error) {
	root, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, err
	}

	var results []string

	var hasClass = func(n *html.Node, want string) bool {
		for _, a := range n.Attr {
			if a.Key == "class" {
				// classes séparées par des espaces
				for _, c := range strings.Fields(a.Val) {
					if c == want {
						return true
					}
				}
			}
		}
		return false
	}

	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" && hasClass(n, class) {
			var buf bytes.Buffer
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				_ = html.Render(&buf, c)
			}
			results = append(results, buf.String())
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(root)

	return results, nil
}

func appendInFile(spans []string) {
	// Split sur les espaces et append chaque string dans api/word/fr.txt
	text := strings.Join(spans, " ")
	// Normaliser les retours à la ligne et tabulations en espaces
	replacer := strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	text = replacer.Replace(text)
	// Split par espaces (en éliminant les vides)
	// Note: strings.Fields split sur tout espace blanc, ce qui convient ici
	tokens := strings.Fields(text)

	// Trouver le chemin vers api/word/fr.txt en remontant depuis le CWD
	target := findPathUpwards("api/word/fr.txt", 6)
	if target == "" {
		log.Fatal("Impossible de localiser api/word/fr.txt en remontant l'arborescence")
	}

	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	bw := bufio.NewWriter(f)
	for _, tok := range tokens {
		if tok == "" {
			continue
		}
		if _, err := bw.WriteString(strings.ToLower(tok) + "\n"); err != nil {
			log.Fatal(err)
		}
	}
	if err := bw.Flush(); err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stderr, "Ajouté %d entrées dans %s\n", len(tokens), target)
}

func main() {
	// Timeout par requête (pas global)
	timeout := time.Duration(20 * time.Second)

	url := "https://www.listesdemots.net/touslesmots.htm"

	utf8Data, err := FetchHTMLWithRetry(url, timeout, 3)
	if err != nil {
		log.Fatal(err)
	}

	// Extraire uniquement le contenu HTML des balises span.mt
	spans, err := ExtractInnerHTMLFromSpans(utf8Data, "mt")
	if err != nil {
		log.Fatal(err)
	}

	appendInFile(spans)

	for i := 2; i <= 1548; i++ {
		fmt.Printf("Traitement de la page %d\n", i)

		url = fmt.Sprintf("https://www.listesdemots.net/touslesmotspage%d.htm", i)

		utf8Data, err := FetchHTMLWithRetry(url, timeout, 3)
		if err != nil {
			log.Fatal(err)
		}

		// Extraire uniquement le contenu HTML des balises span.mt
		spans, err := ExtractInnerHTMLFromSpans(utf8Data, "mt")
		if err != nil {
			log.Fatal(err)
		}

		appendInFile(spans)
	}
}

// FetchHTMLWithRetry exécute des tentatives de récupération avec un timeout par requête.
func FetchHTMLWithRetry(url string, perRequestTimeout time.Duration, attempts int) ([]byte, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 1; i <= attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), perRequestTimeout)
		data, err := FetchHTML(ctx, url)
		cancel()
		if err == nil {
			return data, nil
		}
		lastErr = err
		// backoff simple
		if i < attempts {
			time.Sleep(time.Duration(i*i) * 300 * time.Millisecond)
		}
	}
	return nil, lastErr
}

// findPathUpwards cherche un chemin relatif (relPath) en remontant jusqu'à maxUp niveaux
// depuis le répertoire de travail courant. Retourne le chemin absolu trouvé ou "" si non trouvé.
func findPathUpwards(relPath string, maxUp int) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i <= maxUp; i++ {
		candidate := filepath.Join(dir, relPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir { // atteint la racine
			break
		}
		dir = parent
	}
	return ""
}
