package fuzzy

import (
	"testing"
)

func TestScoreMatching(t *testing.T) {
	cases := []struct {
		name      string
		pattern   string
		target    string
		wantMatch bool
		minScore  int
	}{
		{"exact match", "htop", "htop", true, 600},
		{"prefix match", "ht", "htop", true, 200},
		{"substring match", "top", "htop", true, 100},
		{"no match", "xyz", "htop", false, 0},
		{"empty pattern", "", "anything", true, 0},
		{"case-insensitive", "HTOP", "htop", true, 600},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			res := Score(tt.pattern, tt.target)
			if res.Matched != tt.wantMatch {
				t.Fatalf("expected Matched=%v, got %v", tt.wantMatch, res.Matched)
			}
			if tt.wantMatch && res.Score < tt.minScore {
				t.Errorf("score too low: got %d, want at least %d", res.Score, tt.minScore)
			}
			if tt.name == "empty pattern" && res.Score != 0 {
				t.Errorf("empty pattern score should be 0, got %d", res.Score)
			}
		})
	}
}

func TestScoreWordBoundaryBonus(t *testing.T) {
	t.Run("word boundary bonus", func(t *testing.T) {
		res := Score("core", "git-core")
		if !res.Matched {
			t.Fatal("expected word boundary match")
		}
		res2 := Score("core", "abcoredef")
		if !res2.Matched {
			t.Fatal("expected match without boundary")
		}
		if res.Score <= res2.Score {
			t.Errorf("word boundary should score higher: %d vs %d", res.Score, res2.Score)
		}
	})
}

func TestScoreShorterTargetPreferred(t *testing.T) {
	t.Run("shorter target preferred", func(t *testing.T) {
		short := Score("top", "htop")
		long := Score("top", "libhttp-top-handler-perl")
		if !short.Matched || !long.Matched {
			t.Fatal("both should match")
		}
		if short.Score <= long.Score {
			t.Errorf("shorter target should score higher: %d vs %d", short.Score, long.Score)
		}
	})
}

func TestScorePositions(t *testing.T) {
	t.Run("positions", func(t *testing.T) {
		res := Score("abc", "aXbXc")
		if !res.Matched {
			t.Fatal("expected match")
		}
		if len(res.Positions) != 3 {
			t.Fatalf("expected 3 positions, got %d", len(res.Positions))
		}
		if res.Positions[0] != 0 || res.Positions[1] != 2 || res.Positions[2] != 4 {
			t.Errorf("unexpected positions: %v", res.Positions)
		}
	})
}

func TestScoreConsecutiveBonus(t *testing.T) {
	t.Run("consecutive bonus", func(t *testing.T) {
		consec := Score("abc", "abcdef")
		scattered := Score("abc", "aXbXcXX")
		if !consec.Matched || !scattered.Matched {
			t.Fatal("both should match")
		}
		if consec.Score <= scattered.Score {
			t.Errorf("consecutive should score higher: %d vs %d", consec.Score, scattered.Score)
		}
	})
}

func TestMinQuality(t *testing.T) {
	t.Run("min quality", func(t *testing.T) {
		q1 := MinQuality(1)
		q4 := MinQuality(4)
		if q1 <= 0 {
			t.Errorf("MinQuality(1) should be positive: %d", q1)
		}
		if q4 <= q1 {
			t.Errorf("MinQuality(4) should be > MinQuality(1): %d vs %d", q4, q1)
		}
		if q4 != 120 {
			t.Errorf("MinQuality(4) expected 120, got %d", q4)
		}
	})
}

func TestScoreScatteredBelowMinQuality(t *testing.T) {
	t.Run("scattered below min quality", func(t *testing.T) {
		res := Score("abcd", "aXXXXXXXXbXXXXXXXXcXXXXXXXXd")
		minQ := MinQuality(4)
		if res.Matched && res.Score >= minQ {
			t.Errorf("very scattered match should be below MinQuality: score=%d, min=%d", res.Score, minQ)
		}
	})
}

func TestIsSeparator(t *testing.T) {
	t.Run("separators", func(t *testing.T) {
		separators := []rune{' ', '-', '_', '.', '/', ':'}
		for _, r := range separators {
			if !isSeparator(r) {
				t.Errorf("%q should be a separator", r)
			}
		}
	})
	t.Run("non-separators", func(t *testing.T) {
		nonSeparators := []rune{'a', 'Z', '0', '!', '@'}
		for _, r := range nonSeparators {
			if isSeparator(r) {
				t.Errorf("%q should not be a separator", r)
			}
		}
	})
}
