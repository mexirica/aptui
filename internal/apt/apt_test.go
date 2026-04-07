package apt

import (
	"os"
	"strings"
	"testing"

	"github.com/mexirica/aptui/internal/model"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "-"},
		{"  ", "-"},
		{"0", "-"},
		{"500", "500 kB"},
		{"1024", "1.0 MB"},
		{"2048", "2.0 MB"},
		{"1048576", "1.0 GB"},
		{"2097152", "2.0 GB"},
		{"512", "512 kB"},
		{"1500", "1.5 MB"},
	}
	for _, tt := range tests {
		got := formatSize(tt.input)
		if got != tt.expected {
			t.Errorf("formatSize(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseDpkgOutput(t *testing.T) {
	input := `vim	8.2.4919	9876	Vi IMproved - enhanced vi editor
curl	7.88.1	456	command line tool for transferring data
`
	pkgs := parseDpkgOutput(input, true)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	vim := pkgs[0]
	if vim.Name != "vim" {
		t.Errorf("expected name 'vim', got '%s'", vim.Name)
	}
	if vim.Version != "8.2.4919" {
		t.Errorf("expected version '8.2.4919', got '%s'", vim.Version)
	}
	if !vim.Installed {
		t.Error("expected installed=true")
	}
	if vim.Description != "Vi IMproved - enhanced vi editor" {
		t.Errorf("unexpected description: %s", vim.Description)
	}

	curl := pkgs[1]
	if curl.Name != "curl" {
		t.Errorf("expected name 'curl', got '%s'", curl.Name)
	}
}

func TestParseDpkgOutputSkipsEmptyLines(t *testing.T) {
	input := `
vim	8.2	100	editor

curl	7.88	50	transfer tool
`
	pkgs := parseDpkgOutput(input, false)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Installed || pkgs[1].Installed {
		t.Error("expected installed=false")
	}
}

func TestParseDpkgOutputSkipsContinuationLines(t *testing.T) {
	input := `vim	8.2	100	editor
 this is a continuation line
curl	7.88	50	tool`
	pkgs := parseDpkgOutput(input, true)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (skipping continuation), got %d", len(pkgs))
	}
}

func TestParseDpkgOutputMinimalFields(t *testing.T) {
	input := `vim	8.2`
	pkgs := parseDpkgOutput(input, true)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "vim" || pkgs[0].Version != "8.2" {
		t.Errorf("unexpected: %+v", pkgs[0])
	}
	if pkgs[0].Size != "" {
		t.Errorf("expected empty size, got %s", pkgs[0].Size)
	}
}

func TestParseSearchOutput(t *testing.T) {
	// parseSearchOutput calls IsInstalled which requires dpkg-query.
	// We test just the parsing logic with a simple case.
	input := `vim - Vi IMproved
git - fast version control`

	// This will call IsInstalled which may fail on CI, but the parse logic itself should work.
	pkgs := parseSearchOutput(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "vim" {
		t.Errorf("expected name 'vim', got '%s'", pkgs[0].Name)
	}
	if pkgs[0].Description != "Vi IMproved" {
		t.Errorf("unexpected description: %s", pkgs[0].Description)
	}
}

func TestParseUpgradableOutput(t *testing.T) {
	input := `Listing... Done
vim/noble 2:9.1.0-1 amd64 [upgradable from: 2:8.2.4919-1]
curl/noble 8.5.0-1 amd64 [upgradable from: 7.88.1-1]`

	pkgs := parseUpgradableOutput(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	vim := pkgs[0]
	if vim.Name != "vim" {
		t.Errorf("expected name 'vim', got '%s'", vim.Name)
	}
	if !vim.Upgradable {
		t.Error("expected upgradable=true")
	}
	if !vim.Installed {
		t.Error("expected installed=true for upgradable")
	}
	if vim.NewVersion != "2:9.1.0-1" {
		t.Errorf("expected new version '2:9.1.0-1', got '%s'", vim.NewVersion)
	}
	if vim.Version != "2:8.2.4919-1" {
		t.Errorf("expected old version '2:8.2.4919-1', got '%s'", vim.Version)
	}
	if vim.SecurityUpdate {
		t.Error("expected SecurityUpdate=false for non-security repo")
	}
}

func TestParseUpgradableOutputSecurityUpdate(t *testing.T) {
	input := `Listing... Done
vim/noble-security 2:9.1.0-1 amd64 [upgradable from: 2:8.2.4919-1]
curl/noble 8.5.0-1 amd64 [upgradable from: 7.88.1-1]
libssl3/noble-security,noble-updates 3.0.14-1 amd64 [upgradable from: 3.0.13-1]
websecurity-tools/my-cybersecurity-ppa 1.2.0-1 amd64 [upgradable from: 1.1.0-1]`

	pkgs := parseUpgradableOutput(input)
	if len(pkgs) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(pkgs))
	}

	vim := pkgs[0]
	if !vim.SecurityUpdate {
		t.Error("expected SecurityUpdate=true for security repo")
	}

	curl := pkgs[1]
	if curl.SecurityUpdate {
		t.Error("expected SecurityUpdate=false for non-security repo")
	}

	libssl := pkgs[2]
	if !libssl.SecurityUpdate {
		t.Error("expected SecurityUpdate=true for comma-separated repo with security origin")
	}

	websecTools := pkgs[3]
	if websecTools.SecurityUpdate {
		t.Error("expected SecurityUpdate=false for PPA containing 'security' substring")
	}
}

func TestParseUpgradableOutputSkipsListing(t *testing.T) {
	input := `Listing... Done`
	pkgs := parseUpgradableOutput(input)
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseShowEntry(t *testing.T) {
	info := `Package: vim
Version: 2:8.2.4919-1ubuntu1
Installed-Size: 3984
Architecture: amd64
Depends: vim-common, libc6
Description: Vi IMproved - enhanced vi editor

Package: vim-tiny
Version: 2:8.2.4919-1ubuntu1
Installed-Size: 800
`
	pi := ParseShowEntry(info)
	if pi.Version != "2:8.2.4919-1ubuntu1" {
		t.Errorf("expected version, got '%s'", pi.Version)
	}
	if pi.Size == "" || pi.Size == "-" {
		t.Errorf("expected formatted size, got '%s'", pi.Size)
	}
	if pi.Description != "Vi IMproved - enhanced vi editor" {
		t.Errorf("expected description, got '%s'", pi.Description)
	}
}

func TestParseShowEntryEmpty(t *testing.T) {
	pi := ParseShowEntry("")
	if pi.Version != "" {
		t.Errorf("expected empty version for empty input, got '%s'", pi.Version)
	}
}

// TestPackageModelFields verifies Package struct fields work correctly.
func TestPackageModelFields(t *testing.T) {
	pkg := model.Package{
		Name:        "test-pkg",
		Version:     "1.0",
		Size:        "100 kB",
		Description: "A test package",
		Installed:   true,
		Upgradable:  true,
		NewVersion:  "2.0",
	}
	if pkg.Name != "test-pkg" {
		t.Errorf("unexpected name: %s", pkg.Name)
	}
	if !pkg.Installed || !pkg.Upgradable {
		t.Error("expected installed and upgradable")
	}
	if pkg.NewVersion != "2.0" {
		t.Errorf("expected new version 2.0, got %s", pkg.NewVersion)
	}
}

func TestParseDpkgOutputDeduplicatesMultiArch(t *testing.T) {
	input := `libc6	2.39-0ubuntu8	14000	GNU C Library	libs	amd64
libc6	2.39-0ubuntu8	7000	GNU C Library	libs	i386
vim	8.2.4919	9876	Vi IMproved	editors	amd64
`
	pkgs := parseDpkgOutput(input, true)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (deduped), got %d", len(pkgs))
	}
	if pkgs[0].Name != "libc6" {
		t.Errorf("expected first package 'libc6', got '%s'", pkgs[0].Name)
	}
	if pkgs[0].Architecture != "amd64" {
		t.Errorf("expected architecture 'amd64', got '%s'", pkgs[0].Architecture)
	}
	if pkgs[1].Name != "vim" {
		t.Errorf("expected second package 'vim', got '%s'", pkgs[1].Name)
	}
}

func TestParseUpgradableOutputDeduplicatesMultiArch(t *testing.T) {
	input := `Listing... Done
libc6/noble 2.39-1 amd64 [upgradable from: 2.39-0]
libc6/noble 2.39-1 i386 [upgradable from: 2.39-0]
vim/noble 2:9.1.0-1 amd64 [upgradable from: 2:8.2.4919-1]`

	pkgs := parseUpgradableOutput(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (deduped), got %d", len(pkgs))
	}
	if pkgs[0].Name != "libc6" {
		t.Errorf("expected 'libc6', got '%s'", pkgs[0].Name)
	}
	if pkgs[1].Name != "vim" {
		t.Errorf("expected 'vim', got '%s'", pkgs[1].Name)
	}
}

func TestParseSearchOutputDeduplicates(t *testing.T) {
	input := `libc6 - GNU C Library: Shared libraries
libc6 - GNU C Library: Shared libraries
vim - Vi IMproved - enhanced vi editor`

	pkgs := parseSearchOutput(input)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (deduped), got %d", len(pkgs))
	}
	if pkgs[0].Name != "libc6" {
		t.Errorf("expected 'libc6', got '%s'", pkgs[0].Name)
	}
	if pkgs[1].Name != "vim" {
		t.Errorf("expected 'vim', got '%s'", pkgs[1].Name)
	}
}

func TestParsePackageFileDescription(t *testing.T) {
	content := "Package: testpkg\nVersion: 1.0\nInstalled-Size: 100\nSection: utils\nArchitecture: amd64\nDescription: A test package\nDescription-md5: abc123\n\nPackage: localized\nVersion: 2.0\nDescription-pt_BR: Descricao em portugues\nDescription: English description\n"
	dir := t.TempDir()
	path := dir + "/test_Packages"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	info := make(map[string]PackageInfo)
	parsePackageFile(path, info)

	pi, ok := info["testpkg"]
	if !ok {
		t.Fatal("expected testpkg in info")
	}
	if pi.Description != "A test package" {
		t.Errorf("expected 'A test package', got '%s'", pi.Description)
	}

	pi2, ok := info["localized"]
	if !ok {
		t.Fatal("expected localized in info")
	}
	if pi2.Description != "Descricao em portugues" {
		t.Errorf("expected first description preserved, got '%s'", pi2.Description)
	}
}

func TestParseShowEntryIgnoresDescriptionMd5(t *testing.T) {
	info := "Package: vim\nVersion: 1.0\nDescription: Real description\nDescription-md5: abc123\n"
	pi := ParseShowEntry(info)
	if pi.Description != "Real description" {
		t.Errorf("expected 'Real description', got '%s'", pi.Description)
	}
}

func TestParseShowEntryEssential(t *testing.T) {
	info := "Package: base-files\nVersion: 12\nEssential: yes\nDescription: Debian base system miscellaneous files\n"
	pi := ParseShowEntry(info)
	if !pi.Essential {
		t.Error("expected Essential=true for package with 'Essential: yes'")
	}
}

func TestParseShowEntryNotEssential(t *testing.T) {
	info := "Package: vim\nVersion: 8.2\nDescription: Vi IMproved\n"
	pi := ParseShowEntry(info)
	if pi.Essential {
		t.Error("expected Essential=false for package without 'Essential: yes'")
	}
}

func TestParsePackageFileEssential(t *testing.T) {
	content := "Package: base-files\nVersion: 12\nEssential: yes\nInstalled-Size: 400\nSection: admin\nArchitecture: amd64\nDescription: Debian base system files\n\nPackage: vim\nVersion: 8.2\nInstalled-Size: 3000\nSection: editors\nArchitecture: amd64\nDescription: Vi IMproved\n"
	dir := t.TempDir()
	path := dir + "/test_Packages"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	info := make(map[string]PackageInfo)
	parsePackageFile(path, info)

	pi, ok := info["base-files"]
	if !ok {
		t.Fatal("expected base-files in info")
	}
	if !pi.Essential {
		t.Error("expected base-files to be Essential")
	}

	pi2, ok := info["vim"]
	if !ok {
		t.Fatal("expected vim in info")
	}
	if pi2.Essential {
		t.Error("expected vim to not be Essential")
	}
}

func TestToggleSourcesFileMultiStanza(t *testing.T) {
	// Simulates ubuntu.sources with two stanzas (Ubuntu 24.04+).
	content := `Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble noble-updates
Components: main restricted universe multiverse

Types: deb
URIs: http://security.ubuntu.com/ubuntu/
Suites: noble-security
Components: main restricted universe multiverse`

	ppa := PPA{
		URL: "http://security.ubuntu.com/ubuntu/",
	}

	// Disable the security stanza only.
	result := toggleSourcesFile(content, ppa, false)

	stanzas := splitDEB822Stanzas(result)
	if len(stanzas) != 2 {
		t.Fatalf("expected 2 stanzas, got %d", len(stanzas))
	}

	// First stanza must remain enabled (no Enabled field = default true).
	if !stanzas[0].Enabled {
		t.Error("first stanza should remain enabled")
	}
	if strings.Contains(stanzas[0].Raw, "Enabled:") {
		t.Error("first stanza should NOT have an Enabled field")
	}

	// Second stanza must be disabled.
	if stanzas[1].Enabled {
		t.Error("second stanza should be disabled")
	}

	// Re-enable the security stanza.
	result2 := toggleSourcesFile(result, ppa, true)
	stanzas2 := splitDEB822Stanzas(result2)
	if !stanzas2[0].Enabled {
		t.Error("first stanza should still be enabled after re-enable")
	}
	if !stanzas2[1].Enabled {
		t.Error("second stanza should be re-enabled")
	}
}

func TestToggleSourcesFileExistingEnabledField(t *testing.T) {
	content := `Types: deb
URIs: http://example.com/repo/
Suites: stable
Enabled: yes
Components: main

Types: deb
URIs: http://other.com/repo/
Suites: testing
Components: main`

	ppa := PPA{URL: "http://example.com/repo/"}

	result := toggleSourcesFile(content, ppa, false)
	stanzas := splitDEB822Stanzas(result)

	if stanzas[0].Enabled {
		t.Error("first stanza should be disabled")
	}
	if !stanzas[1].Enabled {
		t.Error("second stanza should remain enabled")
	}
}

func TestToggleSourcesFileNoMatch(t *testing.T) {
	content := `Types: deb
URIs: http://example.com/repo/
Suites: stable
Components: main`

	ppa := PPA{URL: "http://nonexistent.com/repo/"}
	result := toggleSourcesFile(content, ppa, false)

	if result != content {
		t.Error("content should be unchanged when no stanza matches")
	}
}

func TestValidatePPA(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid PPA", input: "ppa:deadsnakes/ppa", wantErr: false},
		{name: "valid PPA with dashes", input: "ppa:user-name/repo-name", wantErr: false},
		{name: "missing ppa prefix", input: "deadsnakes/ppa", wantErr: true},
		{name: "missing repo", input: "ppa:deadsnakes/", wantErr: true},
		{name: "missing user", input: "ppa:/ppa", wantErr: true},
		{name: "no slash", input: "ppa:deadsnakes", wantErr: true},
		{name: "empty string", input: "", wantErr: true},
		{name: "just ppa:", input: "ppa:", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePPA(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePPA(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestExtractPPAName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "launchpad.net URL",
			input:    "deb https://ppa.launchpad.net/deadsnakes/ppa/ubuntu noble main",
			expected: "ppa:deadsnakes/ppa",
		},
		{
			name:     "launchpadcontent.net URL",
			input:    "deb https://ppa.launchpadcontent.net/user/repo/ubuntu noble main",
			expected: "ppa:user/repo",
		},
		{
			name:     "no PPA URL",
			input:    "deb http://archive.ubuntu.com/ubuntu noble main",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPPAName(tt.input)
			if got != tt.expected {
				t.Errorf("extractPPAName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractPPAURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "launchpad URL",
			input:    "deb https://ppa.launchpad.net/deadsnakes/ppa/ubuntu noble main",
			expected: "https://ppa.launchpad.net/deadsnakes/ppa/ubuntu",
		},
		{
			name:     "no PPA URL",
			input:    "deb http://archive.ubuntu.com/ubuntu noble main",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPPAURL(tt.input)
			if got != tt.expected {
				t.Errorf("extractPPAURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractRepoURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard deb line",
			input:    "deb http://archive.ubuntu.com/ubuntu noble main",
			expected: "http://archive.ubuntu.com/ubuntu",
		},
		{
			name:     "https URL",
			input:    "deb https://deb.debian.org/debian bookworm main",
			expected: "https://deb.debian.org/debian",
		},
		{
			name:     "no URL",
			input:    "some text without url",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRepoURL(tt.input)
			if got != tt.expected {
				t.Errorf("extractRepoURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard Ubuntu deb line",
			input:    "deb http://archive.ubuntu.com/ubuntu noble main",
			expected: "archive.ubuntu.com noble",
		},
		{
			name:     "Debian deb line",
			input:    "deb https://deb.debian.org/debian bookworm main",
			expected: "deb.debian.org bookworm",
		},
		{
			name:     "no URL",
			input:    "something without url",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRepoName(tt.input)
			if got != tt.expected {
				t.Errorf("extractRepoName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSplitDEB822Stanzas(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
	}{
		{
			name: "two stanzas",
			content: `Types: deb
URIs: http://example.com/
Suites: stable

Types: deb
URIs: http://other.com/
Suites: testing`,
			wantCount: 2,
		},
		{
			name: "single stanza",
			content: `Types: deb
URIs: http://example.com/
Suites: stable`,
			wantCount: 1,
		},
		{
			name:      "empty content",
			content:   "",
			wantCount: 0,
		},
		{
			name: "three stanzas with blank lines",
			content: `Types: deb
URIs: http://a.com/

Types: deb
URIs: http://b.com/

Types: deb
URIs: http://c.com/`,
			wantCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stanzas := splitDEB822Stanzas(tt.content)
			if len(stanzas) != tt.wantCount {
				t.Errorf("splitDEB822Stanzas() returned %d stanzas, want %d", len(stanzas), tt.wantCount)
			}
		})
	}
}

func TestParseDEB822Stanza(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantURI     string
		wantSuites  string
		wantEnabled bool
		wantTypes   string
	}{
		{
			name:        "standard stanza",
			raw:         "Types: deb\nURIs: http://example.com/\nSuites: stable\nComponents: main",
			wantURI:     "http://example.com/",
			wantSuites:  "stable",
			wantEnabled: true,
			wantTypes:   "deb",
		},
		{
			name:        "disabled stanza",
			raw:         "Types: deb\nURIs: http://example.com/\nSuites: stable\nEnabled: no",
			wantURI:     "http://example.com/",
			wantSuites:  "stable",
			wantEnabled: false,
			wantTypes:   "deb",
		},
		{
			name:        "enabled explicitly",
			raw:         "Types: deb\nURIs: http://example.com/\nSuites: stable\nEnabled: yes",
			wantURI:     "http://example.com/",
			wantSuites:  "stable",
			wantEnabled: true,
			wantTypes:   "deb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := parseDEB822Stanza(tt.raw)
			if s.URI != tt.wantURI {
				t.Errorf("URI = %q, want %q", s.URI, tt.wantURI)
			}
			if s.Suites != tt.wantSuites {
				t.Errorf("Suites = %q, want %q", s.Suites, tt.wantSuites)
			}
			if s.Enabled != tt.wantEnabled {
				t.Errorf("Enabled = %v, want %v", s.Enabled, tt.wantEnabled)
			}
			if s.Types != tt.wantTypes {
				t.Errorf("Types = %q, want %q", s.Types, tt.wantTypes)
			}
		})
	}
}

func TestExtractSourcesRepoName(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
		expected string
	}{
		{
			name:     "with URI and Suites",
			content:  "URIs: http://archive.ubuntu.com/ubuntu\nSuites: noble noble-updates",
			filename: "ubuntu.sources",
			expected: "archive.ubuntu.com noble",
		},
		{
			name:     "no URI falls back to filename",
			content:  "Types: deb\nSuites: stable",
			filename: "debian.sources",
			expected: "debian",
		},
		{
			name:     "URI without suites",
			content:  "URIs: https://deb.debian.org/debian",
			filename: "debian.sources",
			expected: "deb.debian.org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSourcesRepoName(tt.content, tt.filename)
			if got != tt.expected {
				t.Errorf("extractSourcesRepoName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "normal lines", input: "a\nb\nc", expected: []string{"a", "b", "c"}},
		{name: "empty lines skipped", input: "a\n\nb\n\nc", expected: []string{"a", "b", "c"}},
		{name: "whitespace trimmed", input: "  a  \n  b  ", expected: []string{"a", "b"}},
		{name: "empty string", input: "", expected: nil},
		{name: "only whitespace", input: "   \n   ", expected: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("splitLines(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.expected, len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("splitLines(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestToggleListFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		ppa     PPA
		enabled bool
		check   func(string)
	}{
		{
			name:    "disable PPA in list file",
			content: "deb https://ppa.launchpad.net/deadsnakes/ppa/ubuntu noble main\n",
			ppa:     PPA{Name: "ppa:deadsnakes/ppa", IsPPA: true},
			enabled: false,
			check: func(result string) {
				if !strings.Contains(result, "# deb") {
					t.Errorf("expected commented line, got %q", result)
				}
			},
		},
		{
			name:    "enable PPA in list file",
			content: "# deb https://ppa.launchpad.net/deadsnakes/ppa/ubuntu noble main\n",
			ppa:     PPA{Name: "ppa:deadsnakes/ppa", IsPPA: true},
			enabled: true,
			check: func(result string) {
				if strings.Contains(result, "# deb") {
					t.Errorf("expected uncommented line, got %q", result)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toggleListFile(tt.content, tt.ppa, tt.enabled)
			tt.check(result)
		})
	}
}

func TestParseShowEntryMultipleEntries(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantVersion string
		wantSection string
	}{
		{
			name:        "only first entry parsed",
			input:       "Package: vim\nVersion: 8.2\nSection: editors\n\nPackage: vim\nVersion: 9.0\nSection: editors\n",
			wantVersion: "8.2",
			wantSection: "editors",
		},
		{
			name:        "with architecture",
			input:       "Package: curl\nVersion: 7.88\nArchitecture: amd64\n",
			wantVersion: "7.88",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := ParseShowEntry(tt.input)
			if pi.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", pi.Version, tt.wantVersion)
			}
			if tt.wantSection != "" && pi.Section != tt.wantSection {
				t.Errorf("Section = %q, want %q", pi.Section, tt.wantSection)
			}
		})
	}
}

func TestParseShowEntryNewFields(t *testing.T) {
	tests := []struct {
		name           string
		info           string
		wantRecommends string
		wantSuggests   string
		wantSource     string
		wantPriority   string
	}{
		{
			name:           "all fields present",
			info:           "Package: vim\nVersion: 2:9.1.0-1\nPriority: optional\nSection: editors\nSource: vim (2:9.1.0-1)\nArchitecture: amd64\nRecommends: vim-runtime\nSuggests: ctags, vim-doc\nDescription: Vi IMproved\n",
			wantRecommends: "vim-runtime",
			wantSuggests:   "ctags, vim-doc",
			wantSource:     "vim (2:9.1.0-1)",
			wantPriority:   "optional",
		},
		{
			name:           "fields empty",
			info:           "Package: minimal\nVersion: 1.0\nDescription: Minimal package\n",
			wantRecommends: "",
			wantSuggests:   "",
			wantSource:     "",
			wantPriority:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := ParseShowEntry(tt.info)
			if pi.Recommends != tt.wantRecommends {
				t.Errorf("Recommends = %q, want %q", pi.Recommends, tt.wantRecommends)
			}
			if pi.Suggests != tt.wantSuggests {
				t.Errorf("Suggests = %q, want %q", pi.Suggests, tt.wantSuggests)
			}
			if pi.Source != tt.wantSource {
				t.Errorf("Source = %q, want %q", pi.Source, tt.wantSource)
			}
			if pi.Priority != tt.wantPriority {
				t.Errorf("Priority = %q, want %q", pi.Priority, tt.wantPriority)
			}
		})
	}
}

func TestPPAStruct(t *testing.T) {
	tests := []struct {
		name string
		ppa  PPA
	}{
		{
			name: "PPA type",
			ppa: PPA{
				Name:    "ppa:deadsnakes/ppa",
				URL:     "https://ppa.launchpad.net/deadsnakes/ppa/ubuntu",
				File:    "/etc/apt/sources.list.d/deadsnakes.list",
				Enabled: true,
				IsPPA:   true,
			},
		},
		{
			name: "standard repo type",
			ppa: PPA{
				Name:    "archive.ubuntu.com noble",
				URL:     "http://archive.ubuntu.com/ubuntu",
				File:    "/etc/apt/sources.list",
				Enabled: true,
				IsPPA:   false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ppa.Name == "" {
				t.Error("PPA Name should not be empty")
			}
			if tt.ppa.URL == "" {
				t.Error("PPA URL should not be empty")
			}
		})
	}
}

func TestPackageInfoStruct(t *testing.T) {
	tests := []struct {
		name string
		info PackageInfo
	}{
		{
			name: "full info",
			info: PackageInfo{
				Version:      "8.2",
				Size:         "9.8 MB",
				Section:      "editors",
				Architecture: "amd64",
				Description:  "Vi IMproved",
				Essential:    false,
			},
		},
		{
			name: "essential package info",
			info: PackageInfo{
				Version:   "12",
				Essential: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.info // verify struct construction
		})
	}
}

func TestParseListFile(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantCount int
	}{
		{
			name:      "standard deb lines",
			data:      "deb http://archive.ubuntu.com/ubuntu noble main\ndeb http://archive.ubuntu.com/ubuntu noble-updates main\n",
			wantCount: 2,
		},
		{
			name:      "commented line",
			data:      "# deb http://archive.ubuntu.com/ubuntu noble main\ndeb http://archive.ubuntu.com/ubuntu noble-updates main\n",
			wantCount: 2,
		},
		{
			name:      "non-deb lines skipped",
			data:      "some random text\n# comment\ndeb http://archive.ubuntu.com/ubuntu noble main\n",
			wantCount: 1,
		},
		{
			name:      "empty data",
			data:      "",
			wantCount: 0,
		},
		{
			name:      "PPA line",
			data:      "deb https://ppa.launchpad.net/deadsnakes/ppa/ubuntu noble main\n",
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := make(map[string]bool)
			repos := parseListFile(tt.data, "/etc/apt/sources.list", seen)
			if len(repos) != tt.wantCount {
				t.Errorf("parseListFile() returned %d repos, want %d", len(repos), tt.wantCount)
			}
		})
	}
}

func TestInstallBatchCmd(t *testing.T) {
	tests := []struct {
		name       string
		names      []string
		recommends bool
		suggests   bool
		wantArgs   []string
		noWantArgs []string
	}{
		{
			name:       "with recommends and suggests",
			names:      []string{"vim", "git"},
			recommends: true,
			suggests:   true,
			wantArgs:   []string{"--install-recommends", "--install-suggests", "vim", "git"},
			noWantArgs: []string{"--no-install-recommends", "--no-install-suggests"},
		},
		{
			name:       "without recommends or suggests",
			names:      []string{"curl"},
			recommends: false,
			suggests:   false,
			wantArgs:   []string{"--no-install-recommends", "--no-install-suggests", "curl"},
			noWantArgs: []string{"--install-recommends", "--install-suggests"},
		},
		{
			name:       "recommends only",
			names:      []string{"htop"},
			recommends: true,
			suggests:   false,
			wantArgs:   []string{"--install-recommends", "--no-install-suggests", "htop"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := InstallBatchCmd(tt.names, tt.recommends, tt.suggests)
			args := strings.Join(cmd.Args, " ")
			for _, want := range tt.wantArgs {
				if !strings.Contains(args, want) {
					t.Errorf("expected %q in args %q", want, args)
				}
			}
			for _, noWant := range tt.noWantArgs {
				if strings.Contains(args, noWant) {
					t.Errorf("unexpected %q in args %q", noWant, args)
				}
			}
		})
	}
}

func TestUpgradeBatchCmd(t *testing.T) {
	tests := []struct {
		name       string
		names      []string
		recommends bool
		suggests   bool
		wantArg    string
	}{
		{
			name:       "with recommends",
			names:      []string{"vim"},
			recommends: true,
			suggests:   false,
			wantArg:    "--only-upgrade",
		},
		{
			name:       "without recommends",
			names:      []string{"git"},
			recommends: false,
			suggests:   false,
			wantArg:    "--no-install-recommends",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := UpgradeBatchCmd(tt.names, tt.recommends, tt.suggests)
			args := strings.Join(cmd.Args, " ")
			if !strings.Contains(args, tt.wantArg) {
				t.Errorf("expected %q in args %q", tt.wantArg, args)
			}
			if !strings.Contains(args, "--only-upgrade") {
				t.Errorf("expected --only-upgrade in args %q", args)
			}
		})
	}
}

func TestDistUpgradeCmd(t *testing.T) {
	tests := []struct {
		name       string
		recommends bool
		suggests   bool
		wantArg    string
	}{
		{
			name:       "with both",
			recommends: true,
			suggests:   true,
			wantArg:    "--install-recommends",
		},
		{
			name:       "without both",
			recommends: false,
			suggests:   false,
			wantArg:    "--no-install-recommends",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := DistUpgradeCmd(tt.recommends, tt.suggests)
			args := strings.Join(cmd.Args, " ")
			if !strings.Contains(args, "dist-upgrade") {
				t.Errorf("expected 'dist-upgrade' in args %q", args)
			}
			if !strings.Contains(args, tt.wantArg) {
				t.Errorf("expected %q in args %q", tt.wantArg, args)
			}
		})
	}
}

func TestRemoveBatchCmd(t *testing.T) {
	tests := []struct {
		name  string
		names []string
	}{
		{name: "single package", names: []string{"vim"}},
		{name: "multiple packages", names: []string{"vim", "git", "curl"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := RemoveBatchCmd(tt.names)
			args := strings.Join(cmd.Args, " ")
			if !strings.Contains(args, "remove") {
				t.Errorf("expected 'remove' in args %q", args)
			}
			for _, name := range tt.names {
				if !strings.Contains(args, name) {
					t.Errorf("expected %q in args %q", name, args)
				}
			}
		})
	}
}

func TestPurgeBatchCmd(t *testing.T) {
	tests := []struct {
		name  string
		names []string
	}{
		{name: "single package", names: []string{"vim"}},
		{name: "multiple packages", names: []string{"vim", "git"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := PurgeBatchCmd(tt.names)
			args := strings.Join(cmd.Args, " ")
			if !strings.Contains(args, "purge") {
				t.Errorf("expected 'purge' in args %q", args)
			}
			for _, name := range tt.names {
				if !strings.Contains(args, name) {
					t.Errorf("expected %q in args %q", name, args)
				}
			}
		})
	}
}

func TestAddPPACmd(t *testing.T) {
	tests := []struct {
		name string
		ppa  string
	}{
		{name: "standard PPA", ppa: "ppa:deadsnakes/ppa"},
		{name: "another PPA", ppa: "ppa:user/repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := AddPPACmd(tt.ppa)
			args := strings.Join(cmd.Args, " ")
			if !strings.Contains(args, "add-apt-repository") {
				t.Errorf("expected 'add-apt-repository' in args %q", args)
			}
			if !strings.Contains(args, tt.ppa) {
				t.Errorf("expected %q in args %q", tt.ppa, args)
			}
			if strings.Contains(args, "--remove") {
				t.Errorf("AddPPACmd should not contain --remove, got %q", args)
			}
		})
	}
}

func TestRemovePPACmd(t *testing.T) {
	tests := []struct {
		name string
		ppa  string
	}{
		{name: "standard PPA", ppa: "ppa:deadsnakes/ppa"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := RemovePPACmd(tt.ppa)
			args := strings.Join(cmd.Args, " ")
			if !strings.Contains(args, "add-apt-repository") {
				t.Errorf("expected 'add-apt-repository' in args %q", args)
			}
			if !strings.Contains(args, "--remove") {
				t.Errorf("expected '--remove' in args %q", args)
			}
			if !strings.Contains(args, tt.ppa) {
				t.Errorf("expected %q in args %q", tt.ppa, args)
			}
		})
	}
}

func TestUpdateCmd(t *testing.T) {
	cmd := UpdateCmd()
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "apt-get") || !strings.Contains(args, "update") {
		t.Errorf("expected apt-get update in args %q", args)
	}
}

func TestAutoRemoveCmd(t *testing.T) {
	cmd := AutoRemoveCmd()
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "autoremove") {
		t.Errorf("expected autoremove in args %q", args)
	}
}

func TestParsePackageFileEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/empty_Packages"
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	info := make(map[string]PackageInfo)
	parsePackageFile(path, info)
	if len(info) != 0 {
		t.Errorf("expected 0 entries from empty file, got %d", len(info))
	}
}

func TestParsePackageFileNonexistent(t *testing.T) {
	info := make(map[string]PackageInfo)
	parsePackageFile("/nonexistent/path/Packages", info)
	if len(info) != 0 {
		t.Errorf("expected 0 entries from nonexistent file, got %d", len(info))
	}
}

func TestParsePackageFileMultiplePackages(t *testing.T) {
	content := `Package: vim
Version: 8.2
Installed-Size: 3000
Section: editors
Architecture: amd64
Description: Vi IMproved

Package: git
Version: 2.40
Installed-Size: 50000
Section: vcs
Architecture: amd64
Description: fast version control

Package: curl
Version: 7.88
Installed-Size: 500
Section: web
Architecture: amd64
Description: transfer tool
`
	dir := t.TempDir()
	path := dir + "/multi_Packages"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	info := make(map[string]PackageInfo)
	parsePackageFile(path, info)
	if len(info) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(info))
	}
	if info["vim"].Section != "editors" {
		t.Errorf("vim section = %q, want 'editors'", info["vim"].Section)
	}
	if info["git"].Architecture != "amd64" {
		t.Errorf("git arch = %q, want 'amd64'", info["git"].Architecture)
	}
	if info["curl"].Description != "transfer tool" {
		t.Errorf("curl desc = %q, want 'transfer tool'", info["curl"].Description)
	}
}

func TestParseListFilePPAEnabled(t *testing.T) {
	data := "deb https://ppa.launchpad.net/deadsnakes/ppa/ubuntu noble main\n"
	seen := make(map[string]bool)
	repos := parseListFile(data, "/etc/apt/sources.list.d/test.list", seen)
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if !repos[0].Enabled {
		t.Error("expected Enabled=true for uncommented line")
	}
	if !repos[0].IsPPA {
		t.Error("expected IsPPA=true for launchpad URL")
	}
}

func TestParseListFileDisabledRepo(t *testing.T) {
	data := "# deb http://archive.ubuntu.com/ubuntu noble main\n"
	seen := make(map[string]bool)
	repos := parseListFile(data, "/etc/apt/sources.list", seen)
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Enabled {
		t.Error("expected Enabled=false for commented line")
	}
}

func TestExtractRepoNameWithBrackets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "deb line with arch brackets",
			input:    "deb [arch=amd64] http://archive.ubuntu.com/ubuntu noble main",
			expected: "archive.ubuntu.com noble",
		},
		{
			name:     "URL only no suite",
			input:    "deb http://example.com/",
			expected: "example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRepoName(tt.input)
			if got != tt.expected {
				t.Errorf("extractRepoName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseDpkgOutputAllFields(t *testing.T) {
	input := "vim\t8.2\t3000\tVi IMproved\teditors\tamd64\n"
	pkgs := parseDpkgOutput(input, true)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	p := pkgs[0]
	if p.Name != "vim" {
		t.Errorf("Name = %q", p.Name)
	}
	if p.Version != "8.2" {
		t.Errorf("Version = %q", p.Version)
	}
	if p.Description != "Vi IMproved" {
		t.Errorf("Description = %q", p.Description)
	}
	if p.Section != "editors" {
		t.Errorf("Section = %q", p.Section)
	}
	if p.Architecture != "amd64" {
		t.Errorf("Architecture = %q", p.Architecture)
	}
}
