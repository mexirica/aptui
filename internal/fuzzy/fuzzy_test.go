package fuzzy

import (
	"testing"
)

func TestScoreExactMatch(t *testing.T) {
	res := Score("htop", "htop")
	if !res.Matched {
		t.Fatal("expected exact match to be Matched")
	}
	// Exact match should get the highest bonuses (300+200+100)
	if res.Score < 600 {
		t.Errorf("exact match score too low: %d", res.Score)
	}
}

func TestScorePrefixMatch(t *testing.T) {
	res := Score("ht", "htop")
	if !res.Matched {
		t.Fatal("expected prefix match")
	}
	// Should get prefix bonus (+200) and substring bonus (+100)
	if res.Score < 200 {
		t.Errorf("prefix match score too low: %d", res.Score)
	}
}

func TestScoreSubstringMatch(t *testing.T) {
	res := Score("top", "htop")
	if !res.Matched {
		t.Fatal("expected substring match")
	}
	if res.Score < 100 {
		t.Errorf("substring match score too low: %d", res.Score)
	}
}

func TestScoreNoMatch(t *testing.T) {
	res := Score("xyz", "htop")
	if res.Matched {
		t.Fatal("expected no match")
	}
}

func TestScoreEmptyPattern(t *testing.T) {
	res := Score("", "anything")
	if !res.Matched {
		t.Fatal("empty pattern should always match")
	}
	if res.Score != 0 {
		t.Errorf("empty pattern score should be 0, got %d", res.Score)
	}
}

func TestScoreCaseInsensitive(t *testing.T) {
	res := Score("HTOP", "htop")
	if !res.Matched {
		t.Fatal("case-insensitive match should work")
	}
	if res.Score < 600 {
		t.Errorf("case-insensitive exact match score too low: %d", res.Score)
	}
}

func TestScoreWordBoundaryBonus(t *testing.T) {
	// Word boundary bonus: matching at boundary positions (after separator)
	// "core" in "git-core" starts at boundary so bonus applies
	res := Score("core", "git-core")
	if !res.Matched {
		t.Fatal("expected word boundary match")
	}
	// Compare with "core" in "abcoredef" (no boundary)
	res2 := Score("core", "abcoredef")
	if !res2.Matched {
		t.Fatal("expected match without boundary")
	}
	if res.Score <= res2.Score {
		t.Errorf("word boundary should score higher: %d vs %d", res.Score, res2.Score)
	}
}

func TestScoreShorterTargetPreferred(t *testing.T) {
	// "top" in "htop" (short) vs "top" in "libhttp-top-handler-perl" (long)
	short := Score("top", "htop")
	long := Score("top", "libhttp-top-handler-perl")
	if !short.Matched || !long.Matched {
		t.Fatal("both should match")
	}
	if short.Score <= long.Score {
		t.Errorf("shorter target should score higher: %d vs %d", short.Score, long.Score)
	}
}

func TestScorePositions(t *testing.T) {
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
}

func TestScoreConsecutiveBonus(t *testing.T) {
	// "abc" consecutive in "abcdef" should score higher than scattered "abc" in "aXbXc"
	consec := Score("abc", "abcdef")
	scattered := Score("abc", "aXbXcXX")
	if !consec.Matched || !scattered.Matched {
		t.Fatal("both should match")
	}
	if consec.Score <= scattered.Score {
		t.Errorf("consecutive should score higher: %d vs %d", consec.Score, scattered.Score)
	}
}

func TestMinQuality(t *testing.T) {
	// MinQuality should return a positive value proportional to pattern length
	q1 := MinQuality(1)
	q4 := MinQuality(4)
	if q1 <= 0 {
		t.Errorf("MinQuality(1) should be positive: %d", q1)
	}
	if q4 <= q1 {
		t.Errorf("MinQuality(4) should be > MinQuality(1): %d vs %d", q4, q1)
	}
	// Should be 30 * len
	if q4 != 120 {
		t.Errorf("MinQuality(4) expected 120, got %d", q4)
	}
}

func TestScoreScatteredBelowMinQuality(t *testing.T) {
	// Very scattered match should score below MinQuality threshold
	res := Score("abcd", "aXXXXXXXXbXXXXXXXXcXXXXXXXXd")
	minQ := MinQuality(4)
	if res.Matched && res.Score >= minQ {
		t.Errorf("very scattered match should be below MinQuality: score=%d, min=%d", res.Score, minQ)
	}
}

func TestIsSeparator(t *testing.T) {
	separators := []rune{' ', '-', '_', '.', '/', ':'}
	for _, r := range separators {
		if !isSeparator(r) {
			t.Errorf("%q should be a separator", r)
		}
	}
	nonSeparators := []rune{'a', 'Z', '0', '!', '@'}
	for _, r := range nonSeparators {
		if isSeparator(r) {
			t.Errorf("%q should not be a separator", r)
		}
	}
}
