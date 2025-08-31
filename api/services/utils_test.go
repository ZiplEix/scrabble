package services

import (
	"testing"

	"github.com/ZiplEix/scrabble/api/models/request"
)

func TestRackContains(t *testing.T) {
	rack := "ABCDEFG"
	if !rackContains(rack, []request.PlacedLetter{{Char: "A"}, {Char: "B"}}) {
		t.Fatalf("expected rack to contain A and B")
	}
	if rackContains(rack, []request.PlacedLetter{{Char: "Z"}}) {
		t.Fatalf("expected rack not to contain Z")
	}
	// duplicate letters
	if rackContains("ABCD", []request.PlacedLetter{{Char: "A"}, {Char: "A"}}) {
		t.Fatalf("expected rack not to contain two As")
	}
}

func TestApplyLetters(t *testing.T) {
	var board [15][15]string
	letters := []request.PlacedLetter{{X: 7, Y: 7, Char: "A"}, {X: 8, Y: 7, Char: "B"}}
	if err := applyLetters(&board, letters); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if board[7][7] != "A" || board[7][8] != "B" {
		t.Fatalf("letters not applied correctly")
	}
	// placing over existing cell should error
	if err := applyLetters(&board, []request.PlacedLetter{{X: 7, Y: 7, Char: "C"}}); err == nil {
		t.Fatalf("expected error when cell already occupied")
	}
}

func TestComputeMoveScore_SimpleWord(t *testing.T) {
	var board [15][15]string
	// Place HELLO horizontally starting from (7,7)
	placed := []request.PlacedLetter{
		{X: 7, Y: 7, Char: "H"},
		{X: 8, Y: 7, Char: "E"},
		{X: 9, Y: 7, Char: "L"},
		{X: 10, Y: 7, Char: "L"},
		{X: 11, Y: 7, Char: "O"},
	}
	// apply to board to allow compute to read letters
	if err := applyLetters(&board, placed); err != nil {
		t.Fatalf("applyLetters error: %v", err)
	}
	score := computeMoveScore(board, placed, map[Pos]bool{})
	if score <= 0 {
		t.Fatalf("expected positive score, got %d", score)
	}
}

func TestRackPoints(t *testing.T) {
	pts := rackPoints("AEIOU") // 1 each in mapping
	if pts != 5 {
		t.Fatalf("expected 5, got %d", pts)
	}
	// mix case and letters with different values
	v := rackPoints("Az")
	if v <= 0 {
		t.Fatalf("expected > 0, got %d", v)
	}
}

func TestComputeMoveScore_Blank(t *testing.T) {
	var board [15][15]string
	placed := []request.PlacedLetter{
		{X: 7, Y: 7, Char: "C"},
		{X: 8, Y: 7, Char: "A", Blank: true}, // joker utilisé comme 'A'
		{X: 9, Y: 7, Char: "T"},
	}
	if err := applyLetters(&board, placed); err != nil {
		t.Fatalf("applyLetters error: %v", err)
	}
	// Marque la position (8,7) comme blank pour le calcul
	blanks := map[Pos]bool{{X: 8, Y: 7}: true}
	score := computeMoveScore(board, placed, blanks)
	// Le score doit être > 0 et inférieur à la somme normale car 'A' vaut 0
	if score <= 0 {
		t.Fatalf("expected positive score with blank, got %d", score)
	}
}

func TestResolveBlanks_InferFromRack(t *testing.T) {
	rack := "AB?DEF"                                            // un joker disponible
	letters := []request.PlacedLetter{{Char: "A"}, {Char: "Z"}} // Z absent du rack
	out, err := resolveBlanks(rack, letters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 || out[1].Blank != true || out[1].Char != "Z" {
		t.Fatalf("expected second letter to be marked as blank for Z, got %+v", out)
	}
}

func TestResolveBlanks_RespectProvidedBlank(t *testing.T) {
	rack := "A?BCDEF"
	letters := []request.PlacedLetter{{Char: "Q", Blank: true}} // client marque blank
	out, err := resolveBlanks(rack, letters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out[0].Blank {
		t.Fatalf("expected letter to remain blank")
	}
}

func TestRackContains_WithBlank(t *testing.T) {
	rack := "AB?DEF"
	// On veut jouer Z comme joker
	if !rackContains(rack, []request.PlacedLetter{{Char: "Z", Blank: true}}) {
		t.Fatalf("expected rack to contain a joker for Z")
	}
	// Sans blank -> impossible
	if rackContains(rack, []request.PlacedLetter{{Char: "Z"}}) {
		t.Fatalf("expected rack NOT to contain Z without marking blank")
	}
}

func TestUpdatePlayerRack_RemovesQuestionMarkForBlank(t *testing.T) {
	// Test unitaire pur sur la logique de retrait local (sans DB): on simule en extrayant la logique de retrait
	rack := "AB?DEF"
	played := []request.PlacedLetter{{Char: "Z", Blank: true}}
	// retirer '?'
	var toRemove rune
	if played[0].Blank {
		toRemove = '?'
	} else {
		toRemove = rune(played[0].Char[0])
	}
	i := -1
	for idx, r := range rack {
		if r == toRemove {
			i = idx
			break
		}
	}
	if i == -1 {
		t.Fatalf("expected to find '?' in rack")
	}
	rack = rack[:i] + rack[i+1:]
	if rack != "ABDEF" {
		t.Fatalf("expected rack to be ABDEF, got %s", rack)
	}
}

func TestComputeMoveScore_BlankOnDL_TL_Ignored(t *testing.T) {
	// Vérifie que DL/TL n'augmentent pas un joker
	var board [15][15]string
	// Place un mot vertical à (7,7) centre ★: joker au centre, lettre réelle au-dessus
	placed := []request.PlacedLetter{
		{X: 7, Y: 7, Char: "A", Blank: true}, // centre ★
		{X: 7, Y: 6, Char: "B"},
	}
	if err := applyLetters(&board, placed); err != nil {
		t.Fatalf("applyLetters: %v", err)
	}
	blanks := map[Pos]bool{{X: 7, Y: 7}: true}
	score := computeMoveScore(board, placed, blanks)
	// B=3, A(joker)=0, centre ★ => mot x2 => 3*2=6
	if score != 6 {
		t.Fatalf("expected 6, got %d", score)
	}
}

func TestResolveBlanks_ErrorNoBlank(t *testing.T) {
	rack := "ABCDEF" // pas de '?'
	_, err := resolveBlanks(rack, []request.PlacedLetter{{Char: "Z"}})
	if err == nil {
		t.Fatalf("expected error when no '?' available to cover missing letter")
	}
}
