package services

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
	"github.com/ZiplEix/scrabble/api/word"
	"go.uber.org/zap"
)

var (
	// BotUserID contient l'identifiant de Scrabby en base de données.
	// Initialisé par InitBot() au démarrage du serveur.
	BotUserID int64 = -1

	// activeBotGames sert de verrou pour éviter que plusieurs goroutines
	// ne fassent jouer le bot en même temps sur la même partie.
	activeBotGames sync.Map
)

// InitBot charge l'ID de Scrabby depuis la base de données.
func InitBot() {
	err := database.QueryRow(
		`SELECT id FROM users WHERE role = 'ordinateur' AND is_bot = TRUE LIMIT 1`,
	).Scan(&BotUserID)
	if err != nil {
		zap.L().Error("bot: failed to load Scrabby user ID — bot will be disabled", zap.Error(err))
		BotUserID = -1
		return
	}
	zap.L().Info("bot: Scrabby initialized", zap.Int64("bot_user_id", BotUserID))
}

// TriggerBotIfNeeded vérifie si le prochain joueur est le bot et le déclenche en goroutine.
// Doit être appelé après chaque changement de tour (PlayMove, PassTurn, GetNewRack).
func TriggerBotIfNeeded(gameID string, currentTurnUserID int64) {
	if BotUserID == -1 || currentTurnUserID != BotUserID {
		return
	}
	go func() {
		// Petit délai artificiel pour que la réponse HTTP soit retournée au client avant que le bot joue
		time.Sleep(800 * time.Millisecond)
		if err := playBotTurn(gameID); err != nil {
			zap.L().Error("bot: failed to play turn", zap.Error(err), zap.String("game_id", gameID))
		}
	}()
}

// StartBotWorker lance la goroutine de rattrapage qui poll la DB pour les parties en attente du bot.
// Intervalle en secondes. Récupère les parties où c'est au bot de jouer mais qui n'auraient pas
// été déclenchées (ex: redémarrage serveur).
func StartBotWorker(intervalSeconds int) {
	if BotUserID == -1 {
		zap.L().Warn("bot: BotUserID not set, bot worker not started")
		return
	}
	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			rows, err := database.Query(
				`SELECT id FROM games WHERE status = 'ongoing' AND current_turn = $1`,
				BotUserID,
			)
			if err != nil {
				zap.L().Error("bot: poll query failed", zap.Error(err))
				continue
			}
			var gameIDs []string
			for rows.Next() {
				var gid string
				if err := rows.Scan(&gid); err == nil {
					gameIDs = append(gameIDs, gid)
				}
			}
			rows.Close()

			for _, gid := range gameIDs {
				gidCopy := gid
				go func() {
					if err := playBotTurn(gidCopy); err != nil {
						zap.L().Error("bot: worker failed to play turn", zap.Error(err), zap.String("game_id", gidCopy))
					}
				}()
			}
		}
	}()
	zap.L().Info("bot: worker started", zap.Int("interval_seconds", intervalSeconds))
}

// playBotTurn orchestre un tour du bot : cherche le meilleur coup, puis l'échange, puis passe.
func playBotTurn(gameID string) error {
	// Éviter l'exécution concurrente sur la même partie
	if _, loaded := activeBotGames.LoadOrStore(gameID, true); loaded {
		return nil
	}
	defer activeBotGames.Delete(gameID)

	// Recharger l'état complet : on vérifie que c'est bien au bot
	var currentTurn int64
	var status string
	var rackStr string
	err := database.QueryRow(
		`SELECT current_turn, status FROM games WHERE id = $1`, gameID,
	).Scan(&currentTurn, &status)
	if err != nil {
		return fmt.Errorf("bot: failed to load game state: %w", err)
	}
	if status != "ongoing" || currentTurn != BotUserID {
		return nil // plus notre tour ou partie terminée
	}

	// Charger le rack du bot
	err = database.QueryRow(
		`SELECT rack FROM game_players WHERE game_id = $1 AND player_id = $2`, gameID, BotUserID,
	).Scan(&rackStr)
	if err != nil {
		return fmt.Errorf("bot: failed to load bot rack: %w", err)
	}

	// Charger le plateau
	board, err := LoadBoard(gameID)
	if err != nil {
		return fmt.Errorf("bot: failed to load board: %w", err)
	}

	// Chercher le meilleur coup
	bestMove := findBestMove(board, rackStr, gameID)

	if bestMove != nil {
		zap.L().Info("bot: playing move", zap.String("game_id", gameID), zap.String("word", bestMove.Word), zap.Int("score", bestMove.Score))
		err = PlayMove(gameID, BotUserID, *bestMove)
		if err == nil {
			go maybeSendBotTaunt(gameID, bestMove.Score, false)
		}
		return err
	}

	// Aucun coup trouvé → tenter l'échange
	zap.L().Info("bot: no valid move found, trying rack exchange", zap.String("game_id", gameID))
	_, err = GetNewRack(BotUserID, gameID)
	if err == nil {
		go maybeSendBotTaunt(gameID, 0, true)
		return nil
	}

	// Échange impossible (sac vide) → passer
	zap.L().Info("bot: rack exchange failed, passing turn", zap.String("game_id", gameID))
	err = PassTurn(BotUserID, gameID)
	if err == nil {
		go maybeSendBotTaunt(gameID, 0, true)
	}
	return err
}

// candidate représente un coup candidat avec son score.
type candidate struct {
	move  request.PlayMoveRequest
	score int
}

type anchorCell struct {
	x, y   int
	letter rune // '?' if empty (adjacent to occupied)
}

func getAnchorCells(board [15][15]string) []anchorCell {
	var anchors []anchorCell

	isEmpty := true
	for y := 0; y < 15; y++ {
		for x := 0; x < 15; x++ {
			if board[y][x] != "" {
				isEmpty = false
				break
			}
		}
		if !isEmpty {
			break
		}
	}

	if isEmpty {
		return []anchorCell{{x: 7, y: 7, letter: '?'}}
	}

	var emptyAdded [15][15]bool
	for y := 0; y < 15; y++ {
		for x := 0; x < 15; x++ {
			if board[y][x] != "" {
				// Case occupée
				anchors = append(anchors, anchorCell{x: x, y: y, letter: rune(board[y][x][0])})
			} else {
				// Case vide, vérifier si adjacente à une case occupée
				neighbors := [][2]int{
					{x - 1, y}, {x + 1, y},
					{x, y - 1}, {x, y + 1},
				}
				isAdjacent := false
				for _, n := range neighbors {
					nx, ny := n[0], n[1]
					if nx >= 0 && nx < 15 && ny >= 0 && ny < 15 && board[ny][nx] != "" {
						isAdjacent = true
						break
					}
				}
				if isAdjacent && !emptyAdded[y][x] {
					anchors = append(anchors, anchorCell{x: x, y: y, letter: '?'})
					emptyAdded[y][x] = true
				}
			}
		}
	}
	return anchors
}

// canFormWord vérifie si un mot peut être formé avec le rack et les lettres du plateau.
func canFormWord(w string, rackCounts map[rune]int, wildcards int, boardLetters map[rune]bool) bool {
	boardIsEmpty := len(boardLetters) == 0

	// Copie locale du rack pour décompter les lettres utilisées
	usedRack := make(map[rune]int, len(rackCounts))
	for k, v := range rackCounts {
		usedRack[k] = v
	}
	usedWildcards := 0
	usedBoardLetters := 0

	for _, char := range w {
		if usedRack[char] > 0 {
			usedRack[char]--
		} else if usedWildcards < wildcards {
			usedWildcards++
		} else if !boardIsEmpty && boardLetters[char] {
			usedBoardLetters++
		} else {
			return false
		}
	}

	// Un coup valide doit poser au moins 1 lettre du rack
	if len(w) <= usedBoardLetters {
		return false
	}

	return true
}

// findBestMove explore exhaustivement tous les placements légaux et retourne celui avec le score maximum.
// Utilise un algorithme ultra-rapide basé sur le pré-filtrage du dictionnaire.
// Retourne nil si aucun coup valide n'est trouvé.
func findBestMove(board [15][15]string, rack string, gameID string) *request.PlayMoveRequest {
	boardBlanks := BuildBoardBlanks(gameID)
	boardIsEmpty := isBoardEmpty(board)

	// Collecter les lettres uniques présentes sur le plateau
	boardLetters := make(map[rune]bool)
	for y := 0; y < 15; y++ {
		for x := 0; x < 15; x++ {
			if board[y][x] != "" {
				boardLetters[rune(board[y][x][0])] = true
			}
		}
	}

	// Compter les lettres et jokers du rack
	rackCounts := make(map[rune]int)
	wildcards := 0
	for _, r := range rack {
		if r == '?' {
			wildcards++
		} else {
			rackCounts[r]++
		}
	}

	// 1. Filtrer tout le dictionnaire en < 5ms
	var candidates []string
	for _, w := range word.AllWords() {
		// Pas la peine de tester les mots trop courts ou trop longs pour le plateau
		if len(w) < 2 || len(w) > 15 {
			continue
		}
		if canFormWord(w, rackCounts, wildcards, boardLetters) {
			candidates = append(candidates, w)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Directions à tester
	type dir struct {
		dx, dy int
		name   string
	}
	directions := []dir{{1, 0, "H"}, {0, 1, "V"}}
	anchors := getAnchorCells(board)

	// 2. Paralléliser l'évaluation des candidats sur les CPU disponibles
	numWorkers := runtime.NumCPU()
	if numWorkers <= 0 {
		numWorkers = 1
	}
	if numWorkers > len(candidates) {
		numWorkers = len(candidates)
	}

	localBests := make([]*candidate, numWorkers)
	chunkSize := (len(candidates) + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		startIdx := i * chunkSize
		endIdx := startIdx + chunkSize
		if startIdx >= len(candidates) {
			wg.Done()
			continue
		}
		if endIdx > len(candidates) {
			endIdx = len(candidates)
		}

		go func(workerID int, workerCandidates []string) {
			defer wg.Done()
			var localBest *candidate

			for _, w := range workerCandidates {
				wRunes := []rune(w)
				wLen := len(wRunes)
				seen := make(map[string]bool)

				for _, ac := range anchors {
					if ac.letter != '?' {
						// Cas 1 : L'ancrage est occupé. Le mot doit contenir cette lettre.
						for posInWord := 0; posInWord < wLen; posInWord++ {
							if wRunes[posInWord] != ac.letter {
								continue
							}

							for _, d := range directions {
								startX := ac.x - posInWord*d.dx
								startY := ac.y - posInWord*d.dy

								// Vérifier les limites du plateau
								endX := startX + (wLen-1)*d.dx
								endY := startY + (wLen-1)*d.dy
								if startX < 0 || startY < 0 || endX >= 15 || endY >= 15 {
									continue
								}

								key := fmt.Sprintf("%d,%d,%d,%d", startX, startY, d.dx, d.dy)
								if seen[key] {
									continue
								}
								seen[key] = true

								placed, valid := buildPlacement(board, wRunes, startX, startY, d.dx, d.dy, rack, ac.letter, posInWord)
								if !valid || len(placed) == 0 {
									continue
								}

								if !isConnected(board, placed) {
									continue
								}

								// Valider les mots formés
								boardCopy := board
								if err := ApplyLetters(&boardCopy, placed); err != nil {
									continue
								}
								formedWords := extractFormedWords(boardCopy, placed)
								if len(formedWords) == 0 {
									continue
								}
								allValid := true
								for _, fw := range formedWords {
									if !word.WordExists(fw.Word) {
										allValid = false
										break
									}
								}
								if !allValid {
									continue
								}

								score := ComputeMoveScore(boardCopy, placed, boardBlanks)
								if localBest == nil || score > localBest.score {
									move := request.PlayMoveRequest{
										Word:      w,
										StartX:    startX,
										StartY:    startY,
										Direction: d.name,
										Letters:   placed,
										Score:     score,
									}
									localBest = &candidate{move: move, score: score}
								}
							}
						}
					} else {
						// Cas 2 : L'ancrage est vide (case adjacente ou centre).
						// On peut caler n'importe quelle lettre du mot sur cette case.
						for posInWord := 0; posInWord < wLen; posInWord++ {
							for _, d := range directions {
								startX := ac.x - posInWord*d.dx
								startY := ac.y - posInWord*d.dy

								endX := startX + (wLen-1)*d.dx
								endY := startY + (wLen-1)*d.dy
								if startX < 0 || startY < 0 || endX >= 15 || endY >= 15 {
									continue
								}

								key := fmt.Sprintf("%d,%d,%d,%d", startX, startY, d.dx, d.dy)
								if seen[key] {
									continue
								}
								seen[key] = true

								placed, valid := buildPlacement(board, wRunes, startX, startY, d.dx, d.dy, rack, '?', posInWord)
								if !valid || len(placed) == 0 {
									continue
								}

								if boardIsEmpty {
									touchesCenter := false
									for _, pl := range placed {
										if pl.X == 7 && pl.Y == 7 {
											touchesCenter = true
											break
										}
									}
									if !touchesCenter {
										continue
									}
								} else {
									if !isConnected(board, placed) {
										continue
									}
								}

								boardCopy := board
								if err := ApplyLetters(&boardCopy, placed); err != nil {
									continue
								}
								formedWords := extractFormedWords(boardCopy, placed)
								if len(formedWords) == 0 {
									continue
								}
								allValid := true
								for _, fw := range formedWords {
									if !word.WordExists(fw.Word) {
										allValid = false
										break
									}
								}
								if !allValid {
									continue
								}

								score := ComputeMoveScore(boardCopy, placed, boardBlanks)
								if localBest == nil || score > localBest.score {
									move := request.PlayMoveRequest{
										Word:      w,
										StartX:    startX,
										StartY:    startY,
										Direction: d.name,
										Letters:   placed,
										Score:     score,
									}
									localBest = &candidate{move: move, score: score}
								}
							}
						}
					}
				}
			}
			localBests[workerID] = localBest
		}(i, candidates[startIdx:endIdx])
	}

	wg.Wait()

	var absoluteBest *candidate
	for _, lb := range localBests {
		if lb != nil {
			if absoluteBest == nil || lb.score > absoluteBest.score {
				absoluteBest = lb
			}
		}
	}

	if absoluteBest == nil {
		return nil
	}
	return &absoluteBest.move
}


// buildPlacement construit la liste des PlacedLetter pour un mot sur le plateau,
// en vérifiant la compatibilité avec les cases déjà occupées et le rack disponible.
// Retourne les lettres à poser et un booléen de validité.
func buildPlacement(
	board [15][15]string,
	wordRunes []rune,
	startX, startY, dx, dy int,
	rack string,
	anchorRune rune,
	anchorPosInWord int,
) ([]request.PlacedLetter, bool) {
	// Copie du rack pour consommation
	rackCounts := map[rune]int{}
	for _, r := range rack {
		rackCounts[r]++
	}

	var placed []request.PlacedLetter

	for i, letter := range wordRunes {
		x := startX + i*dx
		y := startY + i*dy
		existing := board[y][x]

		if existing != "" {
			// Case occupée : doit correspondre exactement à la lettre du mot
			if existing != string(letter) {
				return nil, false
			}
			// La lettre vient du plateau, pas du rack
			continue
		}

		// Case vide : on doit poser la lettre depuis le rack
		isBlank := false

		if rackCounts[letter] > 0 {
			rackCounts[letter]--
		} else if rackCounts['?'] > 0 {
			// Utiliser un joker
			rackCounts['?']--
			isBlank = true
		} else {
			return nil, false // lettre manquante
		}

		placed = append(placed, request.PlacedLetter{
			X:     x,
			Y:     y,
			Char:  string(letter),
			Blank: isBlank,
		})
	}

	// Un placement valide doit poser au moins une lettre
	if len(placed) == 0 {
		return nil, false
	}

	// Vérifier qu'il n'y a pas de lettres collées avant le début ou après la fin du mot
	// (sinon le mot serait en réalité plus long)
	beforeX := startX - dx
	beforeY := startY - dy
	if beforeX >= 0 && beforeX < 15 && beforeY >= 0 && beforeY < 15 {
		if board[beforeY][beforeX] != "" {
			return nil, false
		}
	}
	endX := startX + len(wordRunes)*dx
	endY := startY + len(wordRunes)*dy
	if endX >= 0 && endX < 15 && endY >= 0 && endY < 15 {
		if board[endY][endX] != "" {
			return nil, false
		}
	}

	return placed, true
}

// isConnected vérifie qu'au moins une des lettres posées est adjacente à une tuile existante du plateau.
func isConnected(board [15][15]string, placed []request.PlacedLetter) bool {
	for _, pl := range placed {
		neighbors := [][2]int{
			{pl.X - 1, pl.Y}, {pl.X + 1, pl.Y},
			{pl.X, pl.Y - 1}, {pl.X, pl.Y + 1},
		}
		for _, n := range neighbors {
			nx, ny := n[0], n[1]
			if nx >= 0 && nx < 15 && ny >= 0 && ny < 15 && board[ny][nx] != "" {
				return true
			}
		}
	}
	return false
}

// isBoardEmpty retourne true si le plateau est entièrement vide.
func isBoardEmpty(board [15][15]string) bool {
	for y := 0; y < 15; y++ {
		for x := 0; x < 15; x++ {
			if board[y][x] != "" {
				return false
			}
		}
	}
	return true
}

// uniqueRunes retourne les runes uniques présentes dans la tranche.
func uniqueRunes(runes []rune) []rune {
	seen := map[rune]bool{}
	var result []rune
	for _, r := range runes {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}
	return result
}

// IsBotGame retourne true si au moins un des joueurs de la partie est le bot.
func IsBotGame(gameID string) bool {
	if BotUserID == -1 {
		return false
	}
	var count int
	_ = database.QueryRow(
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1 AND player_id = $2`,
		gameID, BotUserID,
	).Scan(&count)
	return count > 0
}

// FindBestMoveStandalone explore tous les placements légaux sur un plateau donné avec un rack donné,
// sans nécessiter de connexion à la base de données.
func FindBestMoveStandalone(board [15][15]string, rack string) *request.PlayMoveRequest {
	return findBestMove(board, rack, "")
}

// maybeSendBotTaunt choisit et envoie aléatoirement une réplique amusante dans le chat de la partie
// en fonction de la qualité du coup joué par Scrabby.
func maybeSendBotTaunt(gameID string, score int, isPassOrExchange bool) {
	if BotUserID == -1 {
		return
	}

	// 30% de chance d'envoyer un message sur un coup normal, 50% sur passe/échange
	prob := 0.30
	if isPassOrExchange {
		prob = 0.50
	}

	if rand.Float64() > prob {
		return
	}

	var taunts []string

	if isPassOrExchange {
		taunts = []string{
			"Échange de lettres... Ce sac est rempli de consonnes impossibles !",
			"Je jette mes lettres, ce rack était maudit.",
			"Passer mon tour... Je vis un enfer de voyelles. S'il vous plaît, soyez indulgents.",
			"Pas de mot possible. Je boude dans mon coin de processeur.",
			"Je passe. C'est un complot de lettres, j'en suis sûr !",
			"Rien, le vide absolu. Mon dictionnaire est en deuil.",
		}
	} else if score >= 50 {
		taunts = []string{
			"Et vlan ! 50 points et plus dans la musette. Qui a dit que les ordinateurs ne savaient pas lire ?",
			"B-I-N-G-O ! Tremblez, humains, mon processeur est en surchauffe de génie !",
			"Joli coup, non ? Ne pleurez pas sur le plateau, ça va gondoler les lettres.",
			"Hop là ! Un coup digne des plus grands maîtres. Vous prenez des notes ?",
			"Désolé, c'est mon côté perfectionniste. Magnifique mot, n'est-ce pas ?",
			"Regardez ce score ! C'est presque indécent. Quelqu'un veut un autographe de Scrabby ?",
			"Je pose ça là... Ne cherchez pas à faire pareil, c'est breveté.",
			"Mon algorithme me chuchote à l'oreille que vous êtes en train de perdre.",
		}
	} else if score >= 25 {
		taunts = []string{
			"Pas mal, pas mal... Je consolide mon avance !",
			"Un petit coup sympathique pour pimenter la partie.",
			"Je place ça tranquillement. À vous de faire mieux !",
			"Petit mot deviendra grand... Surtout avec mes multiplicateurs !",
			"On avance doucement mais sûrement. C'est à vous !",
			"Une tactique subtile. Saurez-vous déchiffrer ma stratégie ?",
			"Un coup honnête. Pas transcendant, mais redoutable.",
		}
	} else if score < 15 {
		taunts = []string{
			"Mouais... Quelques lettres posées pour un score ridicule. Mon rack est digne d'un dictionnaire de maternelle.",
			"Franchement, avec ce tirage de lettres, même un dictionnaire n'aurait rien pu faire de mieux.",
			"Je joue ça, mais c'est uniquement pour vous laisser une chance.",
			"Mes capteurs de dignité sont au plus bas après ce coup.",
			"Ce rack est une offense à la langue française. Je fais ce que je peux !",
			"Bon, d'accord, ce n'est pas mon meilleur coup. Oublions cette séquence...",
			"Aïe. Même pour un bot, c'est un peu embarrassant.",
		}
	} else {
		// Coup moyen (score entre 15 et 24)
		taunts = []string{
			"Un coup classique, efficace. Rien à signaler.",
			"Je pose mes lettres sagement.",
			"C'est un mot de transition. Le grand jeu viendra plus tard.",
			"Voilà qui devrait faire réfléchir mes adversaires.",
		}
	}

	if len(taunts) == 0 {
		return
	}

	// Choisir une réplique aléatoirement
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	msg := taunts[rng.Intn(len(taunts))]

	// Envoyer le message de chat de la part de Scrabby
	_, err := CreateMessage(BotUserID, gameID, msg, map[string]any{})
	if err != nil {
		zap.L().Warn("bot: failed to send chat taunt", zap.Error(err), zap.String("game_id", gameID))
	}
}
