package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/ZiplEix/scrabble/api/midgame"
	"github.com/ZiplEix/scrabble/api/word"
)

func main() {
	var (
		seed      int64
		target    int
		showWords bool
	)

	flag.Int64Var(&seed, "seed", time.Now().UnixNano(), "random seed")
	flag.IntVar(&target, "words", 18, "target number of words on board")
	flag.BoolVar(&showWords, "show-words", true, "print placed words list")
	flag.Parse()

	gen := midgame.NewGenerator(target, seed)
	result, err := gen.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Dictionary words: %d\n", word.AllWordsCount())
	fmt.Printf("Seed: %d\n", seed)
	fmt.Printf("Target words: %d\n", target)
	fmt.Printf("Placed words: %d\n", len(result.Words))
	fmt.Printf("Player rack (drawn before generation): %s\n", result.PlayerRack)
	fmt.Printf("Remaining bag letters: %d\n\n", len([]rune(result.RemainingBag)))

	fmt.Println(midgame.RenderBoard(result.Board))

	if showWords {
		fmt.Println("Placed words:")
		for i, w := range result.Words {
			dir := "H"
			if w.Direction == midgame.Vertical {
				dir = "V"
			}
			fmt.Printf("%2d. %-15s start=(%d,%d) dir=%s new_tiles=%d\n", i+1, w.Word, w.X, w.Y, dir, w.NewTiles)
		}
	}
}
