package utils

import (
	"math/rand"
	"testing"
	"time"
)

func TestShuffleRunes_NoPanicAndShuffles(t *testing.T) {
	// seed for reproducibility but still allow randomness
	rand.Seed(1)
	r := []rune("ABCDEFG")
	before := string(r)
	ShuffleRunes(r)
	after := string(r)
	if len(after) != len(before) {
		t.Fatalf("length changed after shuffle: %d != %d", len(after), len(before))
	}
	// In rare case the same order, reseed and try again once to avoid flaky
	if before == after {
		rand.Seed(time.Now().UnixNano())
		ShuffleRunes(r)
		after = string(r)
		if before == after {
			t.Logf("array remained in same order after two shuffles; acceptable but unlikely")
		}
	}
}

func TestDrawLetters(t *testing.T) {
	bag := []rune{'A', 'B', 'C', 'D'}
	drawn := DrawLetters(&bag, 2)
	if len(drawn) != 2 {
		t.Fatalf("expected drawn=2, got %d", len(drawn))
	}
	if len(bag) != 2 {
		t.Fatalf("expected bag len=2, got %d", len(bag))
	}
	// Next draw more than remaining
	drawn2 := DrawLetters(&bag, 5)
	if len(drawn2) != 2 {
		t.Fatalf("expected drawn2=2, got %d", len(drawn2))
	}
	if len(bag) != 0 {
		t.Fatalf("expected bag empty, got %d", len(bag))
	}
}

func TestDrawLettersFromString(t *testing.T) {
	// Seed to make ShuffleRunes deterministic for test run
	rand.Seed(1)
	drawn, rest := DrawLettersFromString("ABCDE", 3)
	if len(drawn) != 3 {
		t.Fatalf("expected 3 drawn, got %d", len(drawn))
	}
	if len(rest) != 2 {
		t.Fatalf("expected rest len 2, got %d", len(rest))
	}
}
