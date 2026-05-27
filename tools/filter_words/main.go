package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// State définit la structure persistée de notre progression
type State struct {
	Validated   map[string]bool `json:"validated"`
	Invalidated map[string]bool `json:"invalidated"`
}

var (
	stateFile = ".filter_words_state.json"
	state     = State{
		Validated:   make(map[string]bool),
		Invalidated: make(map[string]bool),
	}
	stateMutex sync.Mutex
)

// Normalisation des accents et passage en majuscules
func cleanWord(s string) string {
	t := norm.NFD.String(strings.ToUpper(strings.TrimSpace(s)))
	cleaned := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1 // Supprime les marques d'accentuation
		}
		// Conserver uniquement les lettres de A-Z et les tirets/apostrophes si besoin,
		// mais pour le Scrabble, on garde généralement A-Z et tirets
		return r
	}, t)
	return cleaned
}

// Charger l'état depuis le fichier JSON
func loadState() {
	f, err := os.Open(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("Erreur lors de la lecture du fichier d'état : %v. Démarrage à blanc.", err)
		return
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&state); err != nil {
		log.Printf("Erreur lors du décodage du fichier d'état : %v. Démarrage à blanc.", err)
		return
	}
	if state.Validated == nil {
		state.Validated = make(map[string]bool)
	}
	if state.Invalidated == nil {
		state.Invalidated = make(map[string]bool)
	}
	fmt.Printf("État chargé : %d mots validés, %d mots invalidés.\n", len(state.Validated), len(state.Invalidated))
}

// Sauvegarder l'état de façon atomique
func saveState() {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Erreur lors de la sérialisation de l'état : %v", err)
		return
	}

	tmpFile := stateFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		log.Printf("Erreur lors de l'écriture de l'état temporaire : %v", err)
		return
	}

	if err := os.Rename(tmpFile, stateFile); err != nil {
		log.Printf("Erreur lors du remplacement du fichier d'état : %v", err)
	}
}

// Récupère le HTML d'un mot sur 1mot.net
func fetchWordHTML(ctx context.Context, word string) ([]byte, int, error) {
	url := fmt.Sprintf("https://1mot.net/%s", strings.ToLower(word))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, http.StatusNotFound, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("statut HTTP non OK : %d (%s)", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

// Analyse le HTML d'une page 1mot.net pour en extraire les infos
func parseHTML(htmlBytes []byte) (isValid bool, valids []string, invalids []string) {
	doc, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return false, nil, nil
	}

	// Récupérer le texte d'un nœud HTML de manière récursive
	var getNodeText func(*html.Node) string
	getNodeText = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		var text string
		// Pour les images (les lettres sous forme d'images sur 1mot.net), on peut essayer d'extraire leur attribut alt
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "alt" {
					text += attr.Val
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			text += getNodeText(c)
		}
		return text
	}

	// Chercher le H1 et les H4
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "h1" {
				h1Text := strings.ToLower(getNodeText(n))
				if strings.Contains(h1Text, "est valide au scrabble") {
					isValid = true
				}
			} else if n.Data == "h4" {
				h4Text := getNodeText(n)
				// Si c'est la section des mots valides tirés des définitions
				if strings.Contains(h4Text, "mots valides tirés des") {
					// Parcourir les nœuds suivants pour trouver le paragraphe <p class="mm">
					for sib := n.NextSibling; sib != nil; sib = sib.NextSibling {
						if sib.Type == html.ElementNode && sib.Data == "p" {
							// Extraire les mots de ce paragraphe
							valids = extractWordsFromParagraph(sib)
							break
						}
					}
				} else if strings.Contains(h4Text, "mots invalides tirés des") {
					// Parcourir les nœuds suivants pour trouver le paragraphe <p class="mm">
					for sib := n.NextSibling; sib != nil; sib = sib.NextSibling {
						if sib.Type == html.ElementNode && sib.Data == "p" {
							// Extraire les mots de ce paragraphe
							invalids = extractWordsFromParagraph(sib)
							break
						}
					}
				} else if strings.Contains(strings.ToLower(h4Text), "cousin") {
					// Parcourir les nœuds suivants pour trouver le paragraphe <p class="mm">
					for sib := n.NextSibling; sib != nil; sib = sib.NextSibling {
						if sib.Type == html.ElementNode && sib.Data == "p" {
							// Les cousins sont toujours des mots valides au Scrabble
							valids = append(valids, extractWordsFromParagraph(sib)...)
							break
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(doc)

	return isValid, valids, invalids
}

// Extrait les mots se trouvant dans les balises <a> au sein d'un paragraphe
func extractWordsFromParagraph(pNode *html.Node) []string {
	var words []string
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// On prend le texte à l'intérieur du lien
			var text string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					text += c.Data
				} else if c.Type == html.ElementNode && c.Data == "i" {
					// Parfois les mots invalides sont en italique <i>
					for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
						if cc.Type == html.TextNode {
							text += cc.Data
						}
					}
				}
			}
			cleaned := cleanWord(text)
			if cleaned != "" {
				words = append(words, cleaned)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(pNode)
	return words
}

// findPathUpwards cherche un chemin relatif en remontant jusqu'à maxUp niveaux
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
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func main() {
	log.Println("Démarrage du processus de filtrage et d'enrichissement...")

	// 1) Trouver le chemin vers api/word/fr.txt
	dictPath := findPathUpwards("api/word/fr.txt", 6)
	if dictPath == "" {
		log.Fatal("Impossible de localiser api/word/fr.txt")
	}
	fmt.Printf("Fichier dictionnaire trouvé : %s\n", dictPath)

	// 2) Charger l'état de la session précédente
	loadState()

	// 3) Charger et nettoyer la liste des mots à tester
	file, err := os.Open(dictPath)
	if err != nil {
		log.Fatalf("Impossible d'ouvrir le dictionnaire : %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	allInputWords := make(map[string]bool)
	for scanner.Scan() {
		w := cleanWord(scanner.Text())
		if w != "" {
			allInputWords[w] = true
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Erreur de lecture du dictionnaire : %v", err)
	}

	fmt.Printf("Chargé %d mots uniques à partir du dictionnaire d'entrée.\n", len(allInputWords))

	// Déterminer la liste des mots restant à tester
	stateMutex.Lock()
	var pendingWords []string
	for w := range allInputWords {
		if !state.Validated[w] && !state.Invalidated[w] {
			pendingWords = append(pendingWords, w)
		}
	}
	stateMutex.Unlock()

	fmt.Printf("Mots restants à tester : %d / %d (déjà validés: %d, déjà invalidés: %d)\n",
		len(pendingWords), len(allInputWords), len(state.Validated), len(state.Invalidated))

	if len(pendingWords) == 0 {
		fmt.Println("Tous les mots ont déjà été traités !")
		writeFinalDictionary(dictPath)
		return
	}

	// 4) Configurer le pool concurrent de workers
	concurrencyLimit := 15
	jobs := make(chan string, len(pendingWords))
	for _, w := range pendingWords {
		jobs <- w
	}
	close(jobs)

	var wg sync.WaitGroup
	// Sauvegarde automatique périodique toutes les 15 secondes
	stopAutoSave := make(chan struct{})
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				saveState()
			case <-stopAutoSave:
				return
			}
		}
	}()

	fmt.Printf("Lancement de %d workers en parallèle...\n", concurrencyLimit)
	for i := 0; i < concurrencyLimit; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for w := range jobs {
				// Avant de faire la requête, vérifier si le mot a été validé ou invalidé
				// à la volée par une autre goroutine (optimisation)
				stateMutex.Lock()
				if state.Validated[w] || state.Invalidated[w] {
					stateMutex.Unlock()
					continue
				}
				stateMutex.Unlock()

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				body, statusCode, err := fetchWordHTML(ctx, w)
				cancel()

				if err != nil {
					if statusCode == http.StatusNotFound {
						// 404 standard
						stateMutex.Lock()
						state.Invalidated[w] = true
						stateMutex.Unlock()
						fmt.Printf("[Worker %d] %s -> 404 (INVALIDÉ)\n", workerID, w)
					} else {
						// Autre erreur réseau : on loggue et on continue sans marquer comme invalidé
						log.Printf("[Worker %d] Erreur réseau pour le mot %s : %v", workerID, w, err)
						// Petit temps de pause pour éviter d'insister en cas de surcharge réseau
						time.Sleep(1 * time.Second)
					}
					continue
				}

				if statusCode == http.StatusNotFound {
					stateMutex.Lock()
					state.Invalidated[w] = true
					stateMutex.Unlock()
					fmt.Printf("[Worker %d] %s -> 404 (INVALIDÉ)\n", workerID, w)
					continue
				}

				// Parser le contenu de la page
				isValid, valids, invalids := parseHTML(body)

				stateMutex.Lock()
				if isValid {
					state.Validated[w] = true
					fmt.Printf("[Worker %d] %s -> VALIDE (%d valides associés, %d invalides associés)\n", workerID, w, len(valids), len(invalids))
					// Ajouter directement les mots valides associés à notre état sans les retester
					for _, v := range valids {
						if !state.Validated[v] {
							state.Validated[v] = true
						}
					}
				} else {
					state.Invalidated[w] = true
					fmt.Printf("[Worker %d] %s -> INVALIDE (%d invalides associés)\n", workerID, w, len(invalids))
					// Ajouter directement les mots invalides associés à notre état
					for _, inv := range invalids {
						if !state.Invalidated[inv] {
							state.Invalidated[inv] = true
						}
					}
				}
				stateMutex.Unlock()

				// Limite légère pour éviter le spam trop rapide
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	close(stopAutoSave)
	saveState()

	// 5) Réécrire le dictionnaire final homogénéisé
	writeFinalDictionary(dictPath)
}

// Écrit le dictionnaire final de façon ordonnée, homogène (Majuscules, sans accents) et sans doublons
func writeFinalDictionary(dictPath string) {
	fmt.Println("Phase finale : écriture et homogénéisation du dictionnaire fr.txt...")

	stateMutex.Lock()
	words := make([]string, 0, len(state.Validated))
	for w := range state.Validated {
		words = append(words, w)
	}
	stateMutex.Unlock()

	// Tri avec collation française
	coll := collate.New(language.French)
	coll.SortStrings(words)

	// Écriture atomique
	dir := filepath.Dir(dictPath)
	tmp, err := os.CreateTemp(dir, ".fr-sorted-*.tmp")
	if err != nil {
		log.Fatalf("Impossible de créer un fichier temporaire : %v", err)
	}
	defer os.Remove(tmp.Name())

	bw := bufio.NewWriter(tmp)
	for _, w := range words {
		if _, err := bw.WriteString(w + "\n"); err != nil {
			_ = tmp.Close()
			log.Fatalf("Erreur d'écriture : %v", err)
		}
	}
	if err := bw.Flush(); err != nil {
		_ = tmp.Close()
		log.Fatalf("Erreur de vidage du tampon : %v", err)
	}
	if err := tmp.Close(); err != nil {
		log.Fatalf("Erreur de fermeture : %v", err)
	}

	if err := os.Rename(tmp.Name(), dictPath); err != nil {
		log.Fatalf("Erreur lors du remplacement de fr.txt : %v", err)
	}

	fmt.Printf("Dictionnaire final fr.txt écrit avec succès : %d mots enregistrés (tout en majuscules sans accents).\n", len(words))
}
