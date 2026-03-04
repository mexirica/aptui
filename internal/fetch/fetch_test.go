package fetch

import (
	"testing"
	"time"
)

func TestFormatLatency(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{50 * time.Millisecond, "50 ms"},
		{999 * time.Millisecond, "999 ms"},
		{1500 * time.Millisecond, "1.5 s"},
		{3 * time.Second, "3.0 s"},
		{0, "0 ms"},
	}
	for _, tt := range tests {
		got := FormatLatency(tt.input)
		if got != tt.expected {
			t.Errorf("FormatLatency(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestLimitMirrors(t *testing.T) {
	mirrors := make([]Mirror, 100)
	for i := range mirrors {
		mirrors[i] = Mirror{URL: "http://example.com/" + string(rune('a'+i%26))}
	}

	// Limiting to more than available returns all
	result := LimitMirrors(mirrors, 200)
	if len(result) != 100 {
		t.Errorf("expected 100 mirrors, got %d", len(result))
	}

	// Limiting to less should return at most max
	result = LimitMirrors(mirrors, 10)
	if len(result) > 10 {
		t.Errorf("expected at most 10 mirrors, got %d", len(result))
	}
	if len(result) == 0 {
		t.Error("expected at least 1 mirror")
	}
}

func TestLimitMirrorsEmpty(t *testing.T) {
	result := LimitMirrors(nil, 10)
	if len(result) != 0 {
		t.Errorf("expected 0 mirrors from nil input, got %d", len(result))
	}
}

func TestScoreMirrors(t *testing.T) {
	mirrors := []Mirror{
		{URL: "http://fast.com/", Latency: 50 * time.Millisecond, Status: "ok"},
		{URL: "http://slow.com/", Latency: 500 * time.Millisecond, Status: "ok"},
		{URL: "http://err.com/", Latency: 0, Status: "error"},
		{URL: "http://medium.com/", Latency: 200 * time.Millisecond, Status: "ok"},
	}

	scored := ScoreMirrors(mirrors)

	// Only "ok" mirrors should be in the result
	if len(scored) != 3 {
		t.Fatalf("expected 3 scored mirrors (excluding error), got %d", len(scored))
	}

	// Should be sorted by latency (fastest first)
	if scored[0].URL != "http://fast.com/" {
		t.Errorf("expected fastest first, got %s", scored[0].URL)
	}
	if scored[len(scored)-1].URL != "http://slow.com/" {
		t.Errorf("expected slowest last, got %s", scored[len(scored)-1].URL)
	}

	// First should have highest score
	if scored[0].Score <= scored[len(scored)-1].Score {
		t.Errorf("first mirror should have higher score: %d vs %d",
			scored[0].Score, scored[len(scored)-1].Score)
	}
}

func TestScoreMirrorsEmpty(t *testing.T) {
	scored := ScoreMirrors(nil)
	if len(scored) != 0 {
		t.Errorf("expected 0 scored mirrors, got %d", len(scored))
	}
}

func TestScoreMirrorsAllErrors(t *testing.T) {
	mirrors := []Mirror{
		{URL: "http://a.com/", Status: "error"},
		{URL: "http://b.com/", Status: "error"},
	}
	scored := ScoreMirrors(mirrors)
	if len(scored) != 0 {
		t.Errorf("expected 0 mirrors when all errored, got %d", len(scored))
	}
}

func TestBaseDistro(t *testing.T) {
	tests := []struct {
		distro   Distro
		expected string
	}{
		{Distro{ID: "ubuntu"}, "ubuntu"},
		{Distro{ID: "pop"}, "ubuntu"},
		{Distro{ID: "linuxmint"}, "ubuntu"},
		{Distro{ID: "elementary"}, "ubuntu"},
		{Distro{ID: "zorin"}, "ubuntu"},
		{Distro{ID: "neon"}, "ubuntu"},
		{Distro{ID: "debian"}, "debian"},
		{Distro{ID: "kali"}, "debian"},
		{Distro{ID: "mx"}, "debian"},
	}
	for _, tt := range tests {
		got := baseDistro(tt.distro)
		if got != tt.expected {
			t.Errorf("baseDistro(%s) = %q, want %q", tt.distro.ID, got, tt.expected)
		}
	}
}

func TestDefaultMirrors(t *testing.T) {
	ubuntu := defaultUbuntuMirrors()
	if len(ubuntu) == 0 {
		t.Error("expected at least 1 default Ubuntu mirror")
	}
	for _, m := range ubuntu {
		if m.URL == "" {
			t.Error("mirror URL should not be empty")
		}
		if m.Status != "pending" {
			t.Errorf("default mirror status should be 'pending', got '%s'", m.Status)
		}
	}

	debian := defaultDebianMirrors()
	if len(debian) == 0 {
		t.Error("expected at least 1 default Debian mirror")
	}
}

func TestMirrorStruct(t *testing.T) {
	m := Mirror{
		URL:     "http://test.example.com/ubuntu/",
		Country: "BR",
		Latency: 150 * time.Millisecond,
		Status:  "ok",
		Score:   85,
		Active:  true,
	}
	if m.URL != "http://test.example.com/ubuntu/" {
		t.Errorf("unexpected URL: %s", m.URL)
	}
	if !m.Active {
		t.Error("expected active=true")
	}
}

func TestTestMirrorsChan(t *testing.T) {
	// Test with empty mirrors - channel should close immediately
	ch := TestMirrorsChan(nil)
	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 results for nil mirrors, got %d", count)
	}
}

func TestDistroStruct(t *testing.T) {
	d := Distro{
		ID:       "ubuntu",
		Codename: "noble",
		Name:     "Ubuntu 24.04 LTS",
	}
	if d.ID != "ubuntu" || d.Codename != "noble" {
		t.Errorf("unexpected distro: %+v", d)
	}
}
