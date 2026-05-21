package midgame

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/ZiplEix/scrabble/api/word"
)

const (
	BoardSize   = 15
	DefaultBag  = "AAAAAAAAAEEEEEEEEEEEEIIIIIIIIONNNNNNRRRRRRTTTTTTLLLLSSSSUDDDGGGMMMBBCCPPFFHHVVJQKWXYZ??"
	centerCoord = 7
)

type Direction int

const (
	Horizontal Direction = iota
	Vertical
)

type Hook struct {
	X      int
	Y      int
	Letter rune
}

type PlacedWord struct {
	Word      string
	X         int
	Y         int
	Direction Direction
	NewTiles  int
}

type Result struct {
	Board        [BoardSize][BoardSize]string
	PlayerRack   string
	RemainingBag string
	Words        []PlacedWord
}

type tile struct {
	X      int
	Y      int
	Letter rune
}

type placementCandidate struct {
	StartX   int
	StartY   int
	Dir      Direction
	Word     []rune
	WordStr  string
	NewTiles []tile
	Needed   []rune
}

type Generator struct {
	rng                *rand.Rand
	targetWords        int
	maxAttemptsPerHook int
	board              [BoardSize][BoardSize]string
	bag                []rune
	placed             []PlacedWord
}

func NewGenerator(targetWords int, seed int64) *Generator {
	if targetWords <= 0 {
		targetWords = 18
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	bag := []rune(DefaultBag)
	r.Shuffle(len(bag), func(i, j int) { bag[i], bag[j] = bag[j], bag[i] })

	return &Generator{
		rng:                r,
		targetWords:        targetWords,
		maxAttemptsPerHook: 250,
		bag:                bag,
	}
}

func (g *Generator) Generate() (*Result, error) {
	playerRack := string(g.drawLetters(7))

	if err := g.placeSeedWord(); err != nil {
		return nil, err
	}

	stalledRounds := 0
	for len(g.placed) < g.targetWords && stalledRounds < 16 {
		hooks := g.collectHooks()
		if len(hooks) == 0 {
			break
		}
		g.rng.Shuffle(len(hooks), func(i, j int) { hooks[i], hooks[j] = hooks[j], hooks[i] })

		placedThisRound := false
		for _, hook := range hooks {
			if g.tryPlaceAtHook(hook) {
				placedThisRound = true
				if len(g.placed) >= g.targetWords {
					break
				}
			}
		}

		if placedThisRound {
			stalledRounds = 0
		} else {
			stalledRounds++
		}
	}

	return &Result{
		Board:        g.board,
		PlayerRack:   playerRack,
		RemainingBag: string(g.bag),
		Words:        append([]PlacedWord(nil), g.placed...),
	}, nil
}

func (g *Generator) placeSeedWord() error {
	for attempt := 0; attempt < 800; attempt++ {
		seedWord, ok := word.RandomWord(g.rng, 4, 7)
		if !ok {
			return fmt.Errorf("dictionary has no seed word between 4 and 7 letters")
		}
		runes := []rune(seedWord)
		hookIndex := g.rng.Intn(len(runes))

		dirs := []Direction{Horizontal, Vertical}
		g.rng.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
		for _, dir := range dirs {
			startX, startY := centerCoord, centerCoord
			if dir == Horizontal {
				startX = centerCoord - hookIndex
			} else {
				startY = centerCoord - hookIndex
			}

			cand, ok := g.buildCandidate(startX, startY, dir, runes)
			if !ok {
				continue
			}
			g.applyCandidate(cand)
			return nil
		}
	}
	return fmt.Errorf("failed to place seed word")
}

func (g *Generator) tryPlaceAtHook(hook Hook) bool {
	candidates := word.WordsContainingLetter(hook.Letter, 2, 10)
	if len(candidates) == 0 {
		return false
	}

	for i := 0; i < g.maxAttemptsPerHook; i++ {
		w := candidates[g.rng.Intn(len(candidates))]
		runes := []rune(w)
		indices := make([]int, 0, len(runes))
		for idx, r := range runes {
			if r == hook.Letter {
				indices = append(indices, idx)
			}
		}
		if len(indices) == 0 {
			continue
		}

		g.rng.Shuffle(len(indices), func(a, b int) { indices[a], indices[b] = indices[b], indices[a] })
		dirs := []Direction{Horizontal, Vertical}
		g.rng.Shuffle(len(dirs), func(a, b int) { dirs[a], dirs[b] = dirs[b], dirs[a] })

		for _, dir := range dirs {
			for _, idx := range indices {
				startX, startY := hook.X, hook.Y
				if dir == Horizontal {
					startX -= idx
				} else {
					startY -= idx
				}

				cand, ok := g.buildCandidate(startX, startY, dir, runes)
				if !ok {
					continue
				}
				g.applyCandidate(cand)
				return true
			}
		}
	}

	return false
}

func (g *Generator) buildCandidate(startX, startY int, dir Direction, runes []rune) (placementCandidate, bool) {
	cand := placementCandidate{
		StartX:  startX,
		StartY:  startY,
		Dir:     dir,
		Word:    append([]rune(nil), runes...),
		WordStr: string(runes),
	}

	dx, dy := 1, 0
	if dir == Vertical {
		dx, dy = 0, 1
	}

	endX := startX + (len(runes)-1)*dx
	endY := startY + (len(runes)-1)*dy
	if !inBounds(startX, startY) || !inBounds(endX, endY) {
		return placementCandidate{}, false
	}

	if inBounds(startX-dx, startY-dy) && g.board[startY-dy][startX-dx] != "" {
		return placementCandidate{}, false
	}
	if inBounds(endX+dx, endY+dy) && g.board[endY+dy][endX+dx] != "" {
		return placementCandidate{}, false
	}

	hasOverlap := false
	for i, r := range runes {
		x := startX + i*dx
		y := startY + i*dy
		existing := g.board[y][x]
		if existing == "" {
			cand.NewTiles = append(cand.NewTiles, tile{X: x, Y: y, Letter: r})
			cand.Needed = append(cand.Needed, r)
			continue
		}
		er := []rune(existing)[0]
		if er != r {
			return placementCandidate{}, false
		}
		hasOverlap = true
	}

	if len(cand.NewTiles) == 0 {
		return placementCandidate{}, false
	}

	if len(g.placed) > 0 && !hasOverlap {
		return placementCandidate{}, false
	}

	for _, t := range cand.NewTiles {
		pw := g.readWordWithVirtualTile(t.X, t.Y, opposite(dir), t.Letter)
		if len(pw) > 1 && !word.WordExists(pw) {
			return placementCandidate{}, false
		}
	}

	if !g.canConsume(cand.Needed) {
		return placementCandidate{}, false
	}

	return cand, true
}

func (g *Generator) applyCandidate(c placementCandidate) {
	for _, t := range c.NewTiles {
		g.board[t.Y][t.X] = string(t.Letter)
	}
	g.consume(c.Needed)
	g.placed = append(g.placed, PlacedWord{
		Word:      c.WordStr,
		X:         c.StartX,
		Y:         c.StartY,
		Direction: c.Dir,
		NewTiles:  len(c.NewTiles),
	})
}

func (g *Generator) collectHooks() []Hook {
	hooks := make([]Hook, 0)
	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			if g.board[y][x] == "" {
				continue
			}
			if !hasFreeNeighbor(g.board, x, y) {
				continue
			}
			hooks = append(hooks, Hook{X: x, Y: y, Letter: []rune(g.board[y][x])[0]})
		}
	}
	return hooks
}

func (g *Generator) drawLetters(n int) []rune {
	if n <= 0 || len(g.bag) == 0 {
		return nil
	}
	if n > len(g.bag) {
		n = len(g.bag)
	}
	out := make([]rune, 0, n)
	for i := 0; i < n; i++ {
		idx := g.rng.Intn(len(g.bag))
		out = append(out, g.bag[idx])
		g.bag = append(g.bag[:idx], g.bag[idx+1:]...)
	}
	return out
}

func (g *Generator) canConsume(letters []rune) bool {
	count := make(map[rune]int)
	for _, r := range g.bag {
		count[r]++
	}
	for _, r := range letters {
		if count[r] > 0 {
			count[r]--
			continue
		}
		if count['?'] > 0 {
			count['?']--
			continue
		}
		return false
	}
	return true
}

func (g *Generator) consume(letters []rune) {
	for _, r := range letters {
		if g.removeFromBag(r) {
			continue
		}
		_ = g.removeFromBag('?')
	}
}

func (g *Generator) removeFromBag(target rune) bool {
	for i, r := range g.bag {
		if r == target {
			g.bag = append(g.bag[:i], g.bag[i+1:]...)
			return true
		}
	}
	return false
}

func (g *Generator) readWordWithVirtualTile(x, y int, dir Direction, letter rune) string {
	dx, dy := 1, 0
	if dir == Vertical {
		dx, dy = 0, 1
	}

	sx, sy := x, y
	for inBounds(sx-dx, sy-dy) && g.board[sy-dy][sx-dx] != "" {
		sx -= dx
		sy -= dy
	}

	var b strings.Builder
	cx, cy := sx, sy
	for inBounds(cx, cy) {
		if cx == x && cy == y {
			b.WriteRune(letter)
			cx += dx
			cy += dy
			continue
		}
		if g.board[cy][cx] == "" {
			break
		}
		b.WriteString(g.board[cy][cx])
		cx += dx
		cy += dy
	}
	return b.String()
}

func RenderBoard(board [BoardSize][BoardSize]string) string {
	var b strings.Builder
	b.WriteString("    ")
	for x := 0; x < BoardSize; x++ {
		b.WriteString(fmt.Sprintf("%2d", x))
	}
	b.WriteString("\n")
	for y := 0; y < BoardSize; y++ {
		b.WriteString(fmt.Sprintf("%2d |", y))
		for x := 0; x < BoardSize; x++ {
			cell := board[y][x]
			if cell == "" {
				b.WriteString(" .")
			} else {
				b.WriteString(" " + cell)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func inBounds(x, y int) bool {
	return x >= 0 && x < BoardSize && y >= 0 && y < BoardSize
}

func hasFreeNeighbor(board [BoardSize][BoardSize]string, x, y int) bool {
	neighbors := [][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}}
	for _, n := range neighbors {
		nx, ny := n[0], n[1]
		if inBounds(nx, ny) && board[ny][nx] == "" {
			return true
		}
	}
	return false
}

func opposite(d Direction) Direction {
	if d == Horizontal {
		return Vertical
	}
	return Horizontal
}
