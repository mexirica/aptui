package components

import (
	"strings"
	"testing"
	"time"

	"github.com/mexirica/aptui/internal/apt"
	"github.com/mexirica/aptui/internal/errlog"
	"github.com/mexirica/aptui/internal/fetch"
	"github.com/mexirica/aptui/internal/history"
	"github.com/mexirica/aptui/internal/model"
)

func TestRenderPackageListEmpty(t *testing.T) {
	result := RenderPackageList(nil, 0, 0, 10, 120, nil)
	if !strings.Contains(result, "No packages found") {
		t.Error("empty list should show 'no packages' message")
	}
}

func TestRenderPackageListWithPackages(t *testing.T) {
	pkgs := []model.Package{
		{Name: "vim", Version: "8.2", Installed: true, Size: "9.8 MB"},
		{Name: "git", Version: "2.34", Installed: true, Upgradable: true, NewVersion: "2.40", Size: "3.2 MB"},
		{Name: "curl", Installed: false},
	}

	result := RenderPackageList(pkgs, 0, 0, 10, 120, nil)
	if result == "" {
		t.Error("rendered list should not be empty")
	}
	if !strings.Contains(result, "Name") {
		t.Error("should contain Name header")
	}
	if !strings.Contains(result, "Version") {
		t.Error("should contain Version header")
	}
}

func TestRenderPackageListSelectedIndex(t *testing.T) {
	pkgs := []model.Package{
		{Name: "vim", Version: "8.2", Installed: true},
		{Name: "git", Version: "2.34", Installed: true},
	}

	result := RenderPackageList(pkgs, 1, 0, 10, 120, nil)
	if !strings.Contains(result, "\u258c") {
		t.Error("selected item should show cursor")
	}
}

func TestRenderPackageListWithSelection(t *testing.T) {
	pkgs := []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true},
	}

	selected := map[string]bool{"vim": true}
	result := RenderPackageList(pkgs, 0, 0, 10, 120, selected)
	if !strings.Contains(result, "[x]") {
		t.Error("selected package should show [x]")
	}
	if !strings.Contains(result, "[ ]") {
		t.Error("unselected package should show [ ]")
	}
}

func TestRenderPackageListOffset(t *testing.T) {
	pkgs := make([]model.Package, 50)
	for i := range pkgs {
		pkgs[i] = model.Package{Name: "pkg-" + string(rune('a'+i%26))}
	}

	result := RenderPackageList(pkgs, 10, 5, 10, 120, nil)
	lines := strings.Split(result, "\n")
	if len(lines) < 10 {
		t.Errorf("expected at least 10 lines, got %d", len(lines))
	}
}

func TestRenderPackageDetailEmpty(t *testing.T) {
	result := RenderPackageDetail("", 120, 10, 1)
	if !strings.Contains(result, "No package selected") {
		t.Error("empty detail should show placeholder message")
	}
}

func TestRenderPackageDetailWithInfo(t *testing.T) {
	info := "Package: vim\nVersion: 2:8.2.4919-1ubuntu1\nStatus: Installed\nSection: editors\nInstalled-Size: 3984\nMaintainer: Debian Vim Maintainers\nArchitecture: amd64\nDepends: vim-common\nDescription: Vi IMproved\nHomepage: https://www.vim.org"

	result := RenderPackageDetail(info, 120, 10, 1)
	if result == "" {
		t.Error("detail should not be empty")
	}
	if !strings.Contains(result, "vim") {
		t.Error("should contain package name")
	}
}

func TestRenderPackageDetailMaxLines(t *testing.T) {
	info := "Package: vim\nVersion: 1.0\nSection: editors\nInstalled-Size: 100\nMaintainer: Test\nArchitecture: amd64\nDepends: libc6\nDescription: Test package\nHomepage: https://example.com\nStatus: Installed"

	result := RenderPackageDetail(info, 120, 3, 1)
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	if len(lines) > 3 {
		t.Errorf("expected at most 3 lines, got %d", len(lines))
	}
}

func TestRenderQueryPrompt(t *testing.T) {
	result := RenderQueryPrompt("vim", false)
	if !strings.Contains(result, "vim") {
		t.Error("query prompt should contain query")
	}
	if !strings.Contains(result, "\u276f") {
		t.Error("query prompt should contain prompt char")
	}
}

func TestRenderQueryPromptFocused(t *testing.T) {
	result := RenderQueryPrompt("test", true)
	if !strings.Contains(result, "\u2588") {
		t.Error("focused query prompt should show cursor block")
	}
}

func TestRenderQueryPromptEmpty(t *testing.T) {
	result := RenderQueryPrompt("", false)
	if result == "" {
		t.Error("empty query prompt should still render")
	}
}

func TestRenderStatusBar(t *testing.T) {
	result := RenderStatusBar("test status", 120)
	if !strings.Contains(result, "test status") {
		t.Error("status bar should contain the status text")
	}
}

func TestRenderStatusBarEmpty(t *testing.T) {
	result := RenderStatusBar("", 80)
	// Status bar renders with style even if content is empty
	_ = result
}

func TestRenderTransactionListEmpty(t *testing.T) {
	result := RenderTransactionList(nil, 0, 0, 10, 120)
	if !strings.Contains(result, "No transaction") {
		t.Error("empty transaction list should show 'No transaction' message")
	}
}

func TestRenderTransactionListWithItems(t *testing.T) {
	items := []history.Transaction{
		{ID: 1, Operation: history.OpInstall, Packages: []string{"vim"}, Success: true},
		{ID: 2, Operation: history.OpRemove, Packages: []string{"nano"}, Success: false},
	}

	result := RenderTransactionList(items, 0, 0, 10, 120)
	if result == "" {
		t.Error("rendered transaction list should not be empty")
	}
	if !strings.Contains(result, "ID") {
		t.Error("should contain ID header")
	}
}

func TestRenderTransactionDetail(t *testing.T) {
	tx := history.Transaction{
		ID:        1,
		Operation: history.OpInstall,
		Packages:  []string{"vim", "git", "curl"},
		Success:   true,
	}

	result := RenderTransactionDetail(tx, nil, 120, 10)
	if result == "" {
		t.Error("transaction detail should not be empty")
	}
	if !strings.Contains(result, "#1") {
		t.Error("should contain transaction ID")
	}
	if !strings.Contains(result, "install") {
		t.Error("should contain operation name")
	}
}

func TestRenderFetchHeader(t *testing.T) {
	d := fetch.Distro{Name: "Ubuntu 24.04", Codename: "noble"}
	result := RenderFetchHeader(d)
	if !strings.Contains(result, "Ubuntu 24.04") {
		t.Error("should contain distro name")
	}
	if !strings.Contains(result, "noble") {
		t.Error("should contain codename")
	}
}

func TestRenderFetchProgress(t *testing.T) {
	result := RenderFetchProgress(25, 50)
	if !strings.Contains(result, "50%") {
		t.Error("should show 50% progress")
	}
	if !strings.Contains(result, "25/50") {
		t.Error("should show 25/50 count")
	}
}

func TestRenderFetchProgressZero(t *testing.T) {
	result := RenderFetchProgress(0, 0)
	if !strings.Contains(result, "0%") {
		t.Error("should show 0% for empty totals")
	}
}

func TestRenderMirrorListEmpty(t *testing.T) {
	result := RenderMirrorList(nil, 0, 0, 10, 120, nil)
	if !strings.Contains(result, "No mirrors") {
		t.Error("empty mirror list should show message")
	}
}

func TestRenderMirrorListWithMirrors(t *testing.T) {
	mirrors := []fetch.Mirror{
		{URL: "http://archive.ubuntu.com/ubuntu/", Status: "ok", Latency: 50e6},
		{URL: "http://br.archive.ubuntu.com/ubuntu/", Status: "error"},
	}
	selected := map[int]bool{0: true}

	result := RenderMirrorList(mirrors, 0, 0, 10, 120, selected)
	if result == "" {
		t.Error("mirror list should not be empty")
	}
}

func TestRenderFetchHelp(t *testing.T) {
	result := RenderFetchHelp()
	if !strings.Contains(result, "space") {
		t.Error("should contain space key hint")
	}
	if !strings.Contains(result, "enter") {
		t.Error("should contain enter key hint")
	}
}

func TestRenderPackageListHeldBadge(t *testing.T) {
	pkgs := []model.Package{
		{Name: "vim", Version: "8.2", Installed: true, Held: true},
		{Name: "git", Version: "2.34", Installed: true},
	}

	result := RenderPackageList(pkgs, 0, 0, 10, 120, nil)
	if !strings.Contains(result, "⊝") {
		t.Error("held package should show lock badge")
	}
}

func TestRenderPackageListEssentialBadge(t *testing.T) {
	pkgs := []model.Package{
		{Name: "base-files", Version: "12", Installed: true, Essential: true},
		{Name: "vim", Version: "8.2", Installed: true},
	}

	result := RenderPackageList(pkgs, 0, 0, 10, 120, nil)
	if !strings.Contains(result, "◈") {
		t.Error("essential package should show shield badge")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     int // expected number of lines
	}{
		{name: "short text", text: "hello", maxWidth: 80, want: 1},
		{name: "exact width", text: "hello", maxWidth: 5, want: 1},
		{name: "wrap needed", text: "hello world this is long", maxWidth: 10, want: 3},
		{name: "empty text", text: "", maxWidth: 80, want: 1},
		{name: "zero width returns single", text: "hello", maxWidth: 0, want: 1},
		{name: "negative width returns single", text: "hello", maxWidth: -1, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := wrapText(tt.text, tt.maxWidth)
			if len(lines) != tt.want {
				t.Errorf("wrapText(%q, %d) returned %d lines, want %d", tt.text, tt.maxWidth, len(lines), tt.want)
			}
		})
	}
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantVals map[string]string
	}{
		{
			name:     "standard fields",
			input:    "Package: vim\nVersion: 8.2\nSection: editors\nArchitecture: amd64",
			wantKeys: []string{"Package", "Version", "Section", "Architecture"},
			wantVals: map[string]string{"Package": "vim", "Version": "8.2", "Section": "editors"},
		},
		{
			name:     "continuation line",
			input:    "Package: vim\nDescription: text editor\n for terminal use",
			wantKeys: []string{"Package", "Description"},
			wantVals: map[string]string{"Description": "text editor for terminal use"},
		},
		{
			name:     "empty input",
			input:    "",
			wantKeys: nil,
			wantVals: map[string]string{},
		},
		{
			name:     "only first entry",
			input:    "Package: vim\nVersion: 8.2\n\nPackage: nano\nVersion: 7.0",
			wantVals: map[string]string{"Package": "vim", "Version": "8.2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := parseFields(tt.input)
			for k, v := range tt.wantVals {
				if fields[k] != v {
					t.Errorf("parseFields(%q)[%q] = %q, want %q", tt.input, k, fields[k], v)
				}
			}
		})
	}
}

func TestExtractFirstEntry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single entry",
			input: "Package: vim\nVersion: 8.2",
			want:  "Package: vim\nVersion: 8.2",
		},
		{
			name:  "two entries",
			input: "Package: vim\nVersion: 8.2\n\nPackage: nano\nVersion: 7.0",
			want:  "Package: vim\nVersion: 8.2",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "only newlines",
			input: "\n\n\n",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFirstEntry(tt.input)
			if got != tt.want {
				t.Errorf("extractFirstEntry(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{name: "shorter string", input: "hi", width: 5, want: "hi   "},
		{name: "exact width", input: "hello", width: 5, want: "hello"},
		{name: "longer string truncated", input: "hello world", width: 5, want: "hello"},
		{name: "empty string", input: "", width: 3, want: "   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padRight(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestRenderPPAListEmpty(t *testing.T) {
	result := RenderPPAList(nil, 0, 0, 10, 120)
	if !strings.Contains(result, "No repositories found") {
		t.Error("empty PPA list should show 'No repositories found' message")
	}
}

func TestRenderPPAListWithPPAs(t *testing.T) {
	ppas := []apt.PPA{
		{Name: "ppa:deadsnakes/ppa", URL: "https://ppa.launchpad.net/deadsnakes/ppa/ubuntu", Enabled: true, IsPPA: true},
		{Name: "archive.ubuntu.com noble", URL: "http://archive.ubuntu.com/ubuntu", Enabled: true, IsPPA: false},
		{Name: "ppa:disabled/repo", URL: "https://ppa.launchpad.net/disabled/repo/ubuntu", Enabled: false, IsPPA: true},
	}

	tests := []struct {
		name     string
		selected int
		check    func(string)
	}{
		{
			name:     "first selected",
			selected: 0,
			check: func(result string) {
				if !strings.Contains(result, "▌") {
					t.Error("should contain cursor for selected item")
				}
			},
		},
		{
			name:     "contains PPA type",
			selected: 0,
			check: func(result string) {
				if !strings.Contains(result, "PPA") {
					t.Error("PPA item should show 'PPA' type")
				}
			},
		},
		{
			name:     "contains repo type",
			selected: 1,
			check: func(result string) {
				if !strings.Contains(result, "repo") {
					t.Error("non-PPA should show 'repo' type")
				}
			},
		},
		{
			name:     "disabled shows cross",
			selected: 0,
			check: func(result string) {
				if !strings.Contains(result, "✘") {
					t.Error("disabled PPA should show ✘")
				}
			},
		},
		{
			name:     "enabled shows check",
			selected: 0,
			check: func(result string) {
				if !strings.Contains(result, "✔") {
					t.Error("enabled PPA should show ✔")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPPAList(ppas, tt.selected, 0, 10, 120)
			tt.check(result)
		})
	}
}

func TestRenderPPAHelp(t *testing.T) {
	result := RenderPPAHelp()
	tests := []struct {
		name     string
		contains string
	}{
		{name: "add key", contains: "a:"},
		{name: "remove key", contains: "r:"},
		{name: "enable key", contains: "e:"},
		{name: "esc key", contains: "esc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(result, tt.contains) {
				t.Errorf("PPA help should contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestRenderErrorLogListEmpty(t *testing.T) {
	result := RenderErrorLogList(nil, 0, 0, 10, 120)
	if !strings.Contains(result, "No errors logged") {
		t.Error("empty error log should show 'No errors logged' message")
	}
}

func TestRenderErrorLogListWithEntries(t *testing.T) {
	entries := []errlog.Entry{
		{ID: 1, Source: "apt-install", Message: "dependency issue", Timestamp: time.Now()},
		{ID: 2, Source: "apt-remove", Message: "package not found", Timestamp: time.Now()},
	}

	tests := []struct {
		name     string
		selected int
		check    func(string)
	}{
		{
			name:     "first selected has cursor",
			selected: 0,
			check: func(result string) {
				if !strings.Contains(result, "▌") {
					t.Error("selected entry should show cursor")
				}
			},
		},
		{
			name:     "contains ID header",
			selected: 0,
			check: func(result string) {
				if !strings.Contains(result, "ID") {
					t.Error("should contain ID header")
				}
			},
		},
		{
			name:     "contains Source header",
			selected: 0,
			check: func(result string) {
				if !strings.Contains(result, "Source") {
					t.Error("should contain Source header")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderErrorLogList(entries, tt.selected, 0, 10, 120)
			tt.check(result)
		})
	}
}

func TestRenderErrorLogDetail(t *testing.T) {
	tests := []struct {
		name  string
		entry errlog.Entry
		check func(string)
	}{
		{
			name: "shows all fields",
			entry: errlog.Entry{
				ID:        1,
				Source:    "apt-install",
				Message:   "dependency conflict",
				Timestamp: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
			},
			check: func(result string) {
				if !strings.Contains(result, "#1") {
					t.Error("should contain entry ID")
				}
				if !strings.Contains(result, "apt-install") {
					t.Error("should contain source")
				}
				if !strings.Contains(result, "dependency conflict") {
					t.Error("should contain message")
				}
			},
		},
		{
			name: "long message wraps",
			entry: errlog.Entry{
				ID:      2,
				Source:  "fetch",
				Message: strings.Repeat("x", 200),
			},
			check: func(result string) {
				if !strings.Contains(result, "fetch") {
					t.Error("should contain source")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderErrorLogDetail(tt.entry, 80)
			tt.check(result)
		})
	}
}

func TestRenderPPAListOffset(t *testing.T) {
	ppas := make([]apt.PPA, 30)
	for i := range ppas {
		ppas[i] = apt.PPA{
			Name:    "ppa:test/repo",
			URL:     "http://example.com/",
			Enabled: true,
			IsPPA:   true,
		}
	}

	result := RenderPPAList(ppas, 5, 3, 5, 120)
	if result == "" {
		t.Error("PPA list with offset should not be empty")
	}
}

func TestRenderErrorLogListOffset(t *testing.T) {
	entries := make([]errlog.Entry, 30)
	for i := range entries {
		entries[i] = errlog.Entry{
			ID:      i + 1,
			Source:  "test",
			Message: "error message",
		}
	}

	result := RenderErrorLogList(entries, 5, 3, 5, 120)
	if result == "" {
		t.Error("error log list with offset should not be empty")
	}
}
