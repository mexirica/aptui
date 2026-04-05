package fetch

import (
	"fmt"
	"strings"
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

func TestWriteSourcesListCmd(t *testing.T) {
	tests := []struct {
		name       string
		mirrors    []Mirror
		distro     Distro
		wantInArgs string
	}{
		{
			name:       "ubuntu mirrors",
			mirrors:    []Mirror{{URL: "http://archive.ubuntu.com/ubuntu/", Active: true}},
			distro:     Distro{ID: "ubuntu", Codename: "noble", Name: "Ubuntu 24.04"},
			wantInArgs: "/etc/apt/sources.list.d/aptui-mirrors.list",
		},
		{
			name:       "debian mirrors",
			mirrors:    []Mirror{{URL: "https://deb.debian.org/debian/", Active: true}},
			distro:     Distro{ID: "debian", Codename: "bookworm", Name: "Debian 12"},
			wantInArgs: "/etc/apt/sources.list.d/aptui-mirrors.list",
		},
		{
			name:       "inactive mirrors skipped",
			mirrors:    []Mirror{{URL: "http://skip.com/", Active: false}, {URL: "http://use.com/", Active: true}},
			distro:     Distro{ID: "ubuntu", Codename: "noble", Name: "Ubuntu 24.04"},
			wantInArgs: "/etc/apt/sources.list.d/aptui-mirrors.list",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := WriteSourcesListCmd(tt.mirrors, tt.distro)
			if cmd == nil {
				t.Fatal("expected non-nil command")
			}
			found := false
			for _, arg := range cmd.Args {
				if arg == tt.wantInArgs {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected %q in command args %v", tt.wantInArgs, cmd.Args)
			}
			if cmd.Stdin == nil {
				t.Error("expected non-nil stdin (content)")
			}
		})
	}
}

func TestWriteSourcesListCmdUbuntuContent(t *testing.T) {
	mirrors := []Mirror{{URL: "http://archive.ubuntu.com/ubuntu/", Active: true}}
	d := Distro{ID: "ubuntu", Codename: "noble", Name: "Ubuntu 24.04"}
	cmd := WriteSourcesListCmd(mirrors, d)

	// Read the stdin content
	buf := make([]byte, 4096)
	n, _ := cmd.Stdin.Read(buf)
	content := string(buf[:n])

	tests := []struct {
		name     string
		contains string
	}{
		{name: "has header", contains: "# Generated by aptui"},
		{name: "has distro info", contains: "Ubuntu 24.04"},
		{name: "has main deb line", contains: "deb http://archive.ubuntu.com/ubuntu/ noble main restricted universe multiverse"},
		{name: "has updates", contains: "noble-updates"},
		{name: "has security", contains: "noble-security"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(content, tt.contains) {
				t.Errorf("content should contain %q, got:\n%s", tt.contains, content)
			}
		})
	}
}

func TestWriteSourcesListCmdDebianContent(t *testing.T) {
	mirrors := []Mirror{{URL: "https://deb.debian.org/debian/", Active: true}}
	d := Distro{ID: "debian", Codename: "bookworm", Name: "Debian 12"}
	cmd := WriteSourcesListCmd(mirrors, d)

	buf := make([]byte, 4096)
	n, _ := cmd.Stdin.Read(buf)
	content := string(buf[:n])

	tests := []struct {
		name     string
		contains string
	}{
		{name: "has main deb line", contains: "deb https://deb.debian.org/debian/ bookworm main contrib non-free non-free-firmware"},
		{name: "has updates", contains: "bookworm-updates"},
		{name: "no security line for debian", contains: "non-free-firmware"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(content, tt.contains) {
				t.Errorf("content should contain %q, got:\n%s", tt.contains, content)
			}
		})
	}
}

func TestFetchMirrorListUnsupported(t *testing.T) {
	d := Distro{ID: "gentoo", Codename: "test", Name: "Gentoo"}
	_, err := FetchMirrorList(d)
	if err == nil {
		t.Error("expected error for unsupported distro")
	}
}

func TestTestResult(t *testing.T) {
	tests := []struct {
		name   string
		result TestResult
	}{
		{
			name:   "successful result",
			result: TestResult{Index: 0, Latency: 100 * time.Millisecond, Err: nil},
		},
		{
			name:   "error result",
			result: TestResult{Index: 1, Latency: 0, Err: fmt.Errorf("timeout")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Err != nil && tt.result.Latency != 0 {
				t.Error("error results should have zero latency")
			}
		})
	}
}
