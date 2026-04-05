package filter

import (
	"testing"
)

func TestParseEmpty(t *testing.T) {
	f := Parse("")
	if !f.IsEmpty() {
		t.Error("empty query should produce empty filter")
	}
}

func TestParseSection(t *testing.T) {
	f := Parse("section:utils")
	if f.Section != "utils" {
		t.Errorf("expected section 'utils', got '%s'", f.Section)
	}
}

func TestParseSectionAlias(t *testing.T) {
	f := Parse("sec:libs")
	if f.Section != "libs" {
		t.Errorf("expected section 'libs', got '%s'", f.Section)
	}
}

func TestParseArch(t *testing.T) {
	f := Parse("arch:amd64")
	if f.Architecture != "amd64" {
		t.Errorf("expected arch 'amd64', got '%s'", f.Architecture)
	}
}

func TestParseSizeGt(t *testing.T) {
	f := Parse("size>10MB")
	if f.Size == nil {
		t.Fatal("expected size filter")
	}
	if f.Size.Op != SizeGt {
		t.Errorf("expected SizeGt, got %d", f.Size.Op)
	}
	if f.Size.KB != 10*1024 {
		t.Errorf("expected %d kB, got %d", 10*1024, f.Size.KB)
	}
}

func TestParseSizeLt(t *testing.T) {
	f := Parse("size<5MB")
	if f.Size == nil {
		t.Fatal("expected size filter")
	}
	if f.Size.Op != SizeLt {
		t.Errorf("expected SizeLt, got %d", f.Size.Op)
	}
	if f.Size.KB != 5*1024 {
		t.Errorf("expected %d kB, got %d", 5*1024, f.Size.KB)
	}
}

func TestParseSizeGe(t *testing.T) {
	f := Parse("size>=100kB")
	if f.Size == nil {
		t.Fatal("expected size filter")
	}
	if f.Size.Op != SizeGe {
		t.Errorf("expected SizeGe, got %d", f.Size.Op)
	}
	if f.Size.KB != 100 {
		t.Errorf("expected 100 kB, got %d", f.Size.KB)
	}
}

func TestParseSizeColonVariant(t *testing.T) {
	f := Parse("size:>2GB")
	if f.Size == nil {
		t.Fatal("expected size filter")
	}
	if f.Size.Op != SizeGt {
		t.Errorf("expected SizeGt, got %d", f.Size.Op)
	}
	if f.Size.KB != 2*1024*1024 {
		t.Errorf("expected %d kB, got %d", 2*1024*1024, f.Size.KB)
	}
}

func TestParseInstalled(t *testing.T) {
	f := Parse("installed")
	if f.Installed == nil || !*f.Installed {
		t.Error("expected installed=true")
	}
}

func TestParseNotInstalled(t *testing.T) {
	f := Parse("!installed")
	if f.Installed == nil || *f.Installed {
		t.Error("expected installed=false")
	}
}

func TestParseUpgradable(t *testing.T) {
	f := Parse("upgradable")
	if f.Upgradable == nil || !*f.Upgradable {
		t.Error("expected upgradable=true")
	}
}

func TestParseName(t *testing.T) {
	f := Parse("name:vim")
	if f.Name != "vim" {
		t.Errorf("expected name 'vim', got '%s'", f.Name)
	}
}

func TestParseVersion(t *testing.T) {
	f := Parse("ver:2.0")
	if f.Version != "2.0" {
		t.Errorf("expected version '2.0', got '%s'", f.Version)
	}
}

func TestParseDescription(t *testing.T) {
	f := Parse("desc:editor")
	if f.Description != "editor" {
		t.Errorf("expected description 'editor', got '%s'", f.Description)
	}
}

func TestParseMultiple(t *testing.T) {
	f := Parse("section:utils arch:amd64 size>10MB installed")
	if f.Section != "utils" {
		t.Errorf("section: expected 'utils', got '%s'", f.Section)
	}
	if f.Architecture != "amd64" {
		t.Errorf("arch: expected 'amd64', got '%s'", f.Architecture)
	}
	if f.Size == nil || f.Size.Op != SizeGt || f.Size.KB != 10*1024 {
		t.Error("size filter mismatch")
	}
	if f.Installed == nil || !*f.Installed {
		t.Error("expected installed=true")
	}
}

func TestMatchSection(t *testing.T) {
	f := Parse("section:utils")
	p := PackageData{Section: "utils", Name: "test"}
	if !f.Match(p) {
		t.Error("should match package in utils section")
	}
	p2 := PackageData{Section: "libs", Name: "test"}
	if f.Match(p2) {
		t.Error("should not match package in libs section")
	}
}

func TestMatchSectionContains(t *testing.T) {
	f := Parse("section:util")
	p := PackageData{Section: "utils", Name: "test"}
	if !f.Match(p) {
		t.Error("section filter should use contains matching")
	}
}

func TestMatchArch(t *testing.T) {
	f := Parse("arch:amd64")
	p := PackageData{Architecture: "amd64", Name: "test"}
	if !f.Match(p) {
		t.Error("should match amd64")
	}
	p2 := PackageData{Architecture: "arm64", Name: "test"}
	if f.Match(p2) {
		t.Error("should not match arm64")
	}
}

func TestMatchSize(t *testing.T) {
	f := Parse("size>5MB")
	p := PackageData{Size: "10.0 MB", Name: "test"}
	if !f.Match(p) {
		t.Error("10MB should be > 5MB")
	}
	p2 := PackageData{Size: "3.0 MB", Name: "test"}
	if f.Match(p2) {
		t.Error("3MB should not be > 5MB")
	}
}

func TestMatchSizeUnknown(t *testing.T) {
	f := Parse("size>5MB")
	p := PackageData{Size: "-", Name: "test"}
	if f.Match(p) {
		t.Error("unknown size should not match size filter")
	}
}

func TestMatchInstalled(t *testing.T) {
	f := Parse("installed")
	if !f.Match(PackageData{Installed: true, Name: "a"}) {
		t.Error("should match installed package")
	}
	if f.Match(PackageData{Installed: false, Name: "b"}) {
		t.Error("should not match non-installed package")
	}
}

func TestMatchCombined(t *testing.T) {
	f := Parse("section:editors arch:amd64 installed")
	p := PackageData{
		Name:         "vim",
		Section:      "editors",
		Architecture: "amd64",
		Installed:    true,
	}
	if !f.Match(p) {
		t.Error("should match all criteria")
	}
	p2 := PackageData{
		Name:         "vim",
		Section:      "editors",
		Architecture: "arm64",
		Installed:    true,
	}
	if f.Match(p2) {
		t.Error("should not match wrong architecture")
	}
}

func TestParseSizeToKB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1.5 MB", 1536},
		{"324 kB", 324},
		{"2.1 GB", 2202009},
		{"-", 0},
		{"", 0},
		{"10.0 MB", 10240},
	}
	for _, tt := range tests {
		got := ParseSizeToKB(tt.input)
		if got != tt.expected {
			t.Errorf("ParseSizeToKB(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestFilterIsEmpty(t *testing.T) {
	f := Filter{}
	if !f.IsEmpty() {
		t.Error("zero-value filter should be empty")
	}
	f2 := Parse("installed")
	if f2.IsEmpty() {
		t.Error("filter with installed flag should not be empty")
	}
}

func TestParseOrderByNameAsc(t *testing.T) {
	f := Parse("order:name")
	if f.OrderBy != SortName {
		t.Errorf("expected SortName, got %d", f.OrderBy)
	}
	if f.OrderDesc {
		t.Error("expected ascending order by default")
	}
}

func TestParseOrderByNameDesc(t *testing.T) {
	f := Parse("order:name:desc")
	if f.OrderBy != SortName {
		t.Errorf("expected SortName, got %d", f.OrderBy)
	}
	if !f.OrderDesc {
		t.Error("expected descending order")
	}
}

func TestParseOrderBySizeAsc(t *testing.T) {
	f := Parse("order:size:asc")
	if f.OrderBy != SortSize {
		t.Errorf("expected SortSize, got %d", f.OrderBy)
	}
	if f.OrderDesc {
		t.Error("expected ascending order")
	}
}

func TestParseOrderByVersionDesc(t *testing.T) {
	f := Parse("order:ver:desc")
	if f.OrderBy != SortVersion {
		t.Errorf("expected SortVersion, got %d", f.OrderBy)
	}
	if !f.OrderDesc {
		t.Error("expected descending order")
	}
}

func TestParseOrderCombinedWithFilter(t *testing.T) {
	f := Parse("installed order:size:desc")
	if f.Installed == nil || !*f.Installed {
		t.Error("expected installed=true")
	}
	if f.OrderBy != SortSize {
		t.Errorf("expected SortSize, got %d", f.OrderBy)
	}
	if !f.OrderDesc {
		t.Error("expected descending order")
	}
}

func TestParseOrderIsNotEmpty(t *testing.T) {
	f := Parse("order:name")
	if f.IsEmpty() {
		t.Error("filter with order should not be empty")
	}
}

func TestSortByNameAsc(t *testing.T) {
	pkgs := []PackageData{
		{Name: "zsh"},
		{Name: "apt"},
		{Name: "nano"},
	}
	f := Filter{OrderBy: SortName}
	Sort(pkgs, f)
	if pkgs[0].Name != "apt" || pkgs[1].Name != "nano" || pkgs[2].Name != "zsh" {
		t.Errorf("unexpected order: %s, %s, %s", pkgs[0].Name, pkgs[1].Name, pkgs[2].Name)
	}
}

func TestSortByNameDesc(t *testing.T) {
	pkgs := []PackageData{
		{Name: "apt"},
		{Name: "zsh"},
		{Name: "nano"},
	}
	f := Filter{OrderBy: SortName, OrderDesc: true}
	Sort(pkgs, f)
	if pkgs[0].Name != "zsh" || pkgs[1].Name != "nano" || pkgs[2].Name != "apt" {
		t.Errorf("unexpected order: %s, %s, %s", pkgs[0].Name, pkgs[1].Name, pkgs[2].Name)
	}
}

func TestSortBySizeAsc(t *testing.T) {
	pkgs := []PackageData{
		{Name: "big", Size: "10.0 MB"},
		{Name: "small", Size: "100 kB"},
		{Name: "med", Size: "1.0 MB"},
	}
	f := Filter{OrderBy: SortSize}
	Sort(pkgs, f)
	if pkgs[0].Name != "small" || pkgs[1].Name != "med" || pkgs[2].Name != "big" {
		t.Errorf("unexpected order: %s, %s, %s", pkgs[0].Name, pkgs[1].Name, pkgs[2].Name)
	}
}

func TestSortBySizeDesc(t *testing.T) {
	pkgs := []PackageData{
		{Name: "big", Size: "10.0 MB"},
		{Name: "small", Size: "100 kB"},
		{Name: "med", Size: "1.0 MB"},
	}
	f := Filter{OrderBy: SortSize, OrderDesc: true}
	Sort(pkgs, f)
	if pkgs[0].Name != "big" || pkgs[1].Name != "med" || pkgs[2].Name != "small" {
		t.Errorf("unexpected order: %s, %s, %s", pkgs[0].Name, pkgs[1].Name, pkgs[2].Name)
	}
}

func TestSortNoneDoesNothing(t *testing.T) {
	pkgs := []PackageData{
		{Name: "zsh"},
		{Name: "apt"},
		{Name: "nano"},
	}
	f := Filter{OrderBy: SortNone}
	Sort(pkgs, f)
	if pkgs[0].Name != "zsh" || pkgs[1].Name != "apt" || pkgs[2].Name != "nano" {
		t.Errorf("SortNone should not change order: %s, %s, %s", pkgs[0].Name, pkgs[1].Name, pkgs[2].Name)
	}
}

func TestParseFreeText(t *testing.T) {
	f := Parse("vim")
	if f.FreeText != "vim" {
		t.Errorf("expected FreeText 'vim', got '%s'", f.FreeText)
	}
}

func TestParseFreeTextWithFilter(t *testing.T) {
	f := Parse("section:utils vim editor")
	if f.Section != "utils" {
		t.Errorf("expected section 'utils', got '%s'", f.Section)
	}
	if f.FreeText != "vim editor" {
		t.Errorf("expected FreeText 'vim editor', got '%s'", f.FreeText)
	}
}

func TestParseFreeTextEmpty(t *testing.T) {
	f := Parse("section:utils installed")
	if f.FreeText != "" {
		t.Errorf("expected empty FreeText, got '%s'", f.FreeText)
	}
}

func TestIsEmptyWithFreeText(t *testing.T) {
	f := Parse("vim")
	if f.IsEmpty() {
		t.Error("filter with free text should not be empty")
	}
}

func TestNeedsMetadata(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		expect bool
	}{
		{name: "section needs metadata", query: "section:utils", expect: true},
		{name: "arch needs metadata", query: "arch:amd64", expect: true},
		{name: "size needs metadata", query: "size>10MB", expect: true},
		{name: "installed no metadata", query: "installed", expect: false},
		{name: "name no metadata", query: "name:vim", expect: false},
		{name: "empty no metadata", query: "", expect: false},
		{name: "combined with section", query: "installed section:utils", expect: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Parse(tt.query)
			if f.NeedsMetadata() != tt.expect {
				t.Errorf("Parse(%q).NeedsMetadata() = %v, want %v", tt.query, f.NeedsMetadata(), tt.expect)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{name: "simple words", input: "a b c", expect: []string{"a", "b", "c"}},
		{name: "quoted string", input: `name:"foo bar"`, expect: []string{"name:foo bar"}},
		{name: "single quotes", input: "desc:'long description'", expect: []string{"desc:long description"}},
		{name: "mixed", input: `section:utils "free text"`, expect: []string{"section:utils", "free text"}},
		{name: "empty string", input: "", expect: nil},
		{name: "only spaces", input: "   ", expect: nil},
		{name: "multiple spaces between", input: "a   b", expect: []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenize(tt.input)
			if len(got) != len(tt.expect) {
				t.Errorf("tokenize(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.expect, len(tt.expect))
				return
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expect[i])
				}
			}
		})
	}
}

func TestParseSortColumn(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect SortColumn
	}{
		{name: "name", input: "name", expect: SortName},
		{name: "version", input: "version", expect: SortVersion},
		{name: "ver alias", input: "ver", expect: SortVersion},
		{name: "size", input: "size", expect: SortSize},
		{name: "section", input: "section", expect: SortSection},
		{name: "sec alias", input: "sec", expect: SortSection},
		{name: "arch", input: "arch", expect: SortArchitecture},
		{name: "architecture full", input: "architecture", expect: SortArchitecture},
		{name: "unknown", input: "unknown", expect: SortNone},
		{name: "empty", input: "", expect: SortNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSortColumn(tt.input)
			if got != tt.expect {
				t.Errorf("parseSortColumn(%q) = %d, want %d", tt.input, got, tt.expect)
			}
		})
	}
}

func TestMatchSizeLe(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		size    string
		matches bool
	}{
		{name: "5MB <= 10MB", query: "size<=10MB", size: "5.0 MB", matches: true},
		{name: "15MB not <= 10MB", query: "size<=10MB", size: "15.0 MB", matches: false},
		{name: "exact 10MB <= 10MB", query: "size<=10MB", size: "10.0 MB", matches: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Parse(tt.query)
			p := PackageData{Name: "test", Size: tt.size}
			if f.Match(p) != tt.matches {
				t.Errorf("Parse(%q).Match(size=%q) = %v, want %v", tt.query, tt.size, !tt.matches, tt.matches)
			}
		})
	}
}

func TestMatchSizeEq(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		size    string
		matches bool
	}{
		{name: "exact match", query: "size=10MB", size: "10.0 MB", matches: true},
		{name: "not equal", query: "size=10MB", size: "5.0 MB", matches: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Parse(tt.query)
			p := PackageData{Name: "test", Size: tt.size}
			if f.Match(p) != tt.matches {
				t.Errorf("Parse(%q).Match(size=%q) = %v, want %v", tt.query, tt.size, !tt.matches, tt.matches)
			}
		})
	}
}

func TestMatchNameFilter(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		pkgName string
		matches bool
	}{
		{name: "contains match", query: "name:vim", pkgName: "vim-enhanced", matches: true},
		{name: "exact match", query: "name:vim", pkgName: "vim", matches: true},
		{name: "no match", query: "name:vim", pkgName: "nano", matches: false},
		{name: "case insensitive", query: "name:VIM", pkgName: "vim", matches: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Parse(tt.query)
			p := PackageData{Name: tt.pkgName}
			if f.Match(p) != tt.matches {
				t.Errorf("Parse(%q).Match(name=%q) = %v, want %v", tt.query, tt.pkgName, !tt.matches, tt.matches)
			}
		})
	}
}

func TestMatchVersionFilter(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		version    string
		newVersion string
		matches    bool
	}{
		{name: "version contains", query: "ver:2.0", version: "2.0.1", matches: true},
		{name: "version no match", query: "ver:3.0", version: "2.0.1", matches: false},
		{name: "uses NewVersion if set", query: "ver:3.0", version: "2.0.1", newVersion: "3.0.0", matches: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Parse(tt.query)
			p := PackageData{Name: "pkg", Version: tt.version, NewVersion: tt.newVersion}
			if f.Match(p) != tt.matches {
				t.Errorf("Parse(%q).Match(ver=%q, new=%q) = %v, want %v",
					tt.query, tt.version, tt.newVersion, !tt.matches, tt.matches)
			}
		})
	}
}

func TestMatchDescriptionFilter(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		desc    string
		matches bool
	}{
		{name: "contains match", query: "desc:editor", desc: "text editor for terminal", matches: true},
		{name: "no match", query: "desc:browser", desc: "text editor for terminal", matches: false},
		{name: "case insensitive", query: "desc:Editor", desc: "text editor for terminal", matches: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Parse(tt.query)
			p := PackageData{Name: "pkg", Description: tt.desc}
			if f.Match(p) != tt.matches {
				t.Errorf("Parse(%q).Match(desc=%q) = %v, want %v", tt.query, tt.desc, !tt.matches, tt.matches)
			}
		})
	}
}

func TestMatchUpgradable(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		upgradable bool
		matches    bool
	}{
		{name: "upgradable match", query: "upgradable", upgradable: true, matches: true},
		{name: "upgradable no match", query: "upgradable", upgradable: false, matches: false},
		{name: "not upgradable match", query: "!upgradable", upgradable: false, matches: true},
		{name: "not upgradable no match", query: "!upgradable", upgradable: true, matches: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Parse(tt.query)
			p := PackageData{Name: "pkg", Upgradable: tt.upgradable}
			if f.Match(p) != tt.matches {
				t.Errorf("Parse(%q).Match(upgradable=%v) = %v, want %v", tt.query, tt.upgradable, !tt.matches, tt.matches)
			}
		})
	}
}

func TestSortBySection(t *testing.T) {
	pkgs := []PackageData{
		{Name: "vim", Section: "editors"},
		{Name: "apt", Section: "admin"},
		{Name: "zsh", Section: "shells"},
	}
	f := Filter{OrderBy: SortSection}
	Sort(pkgs, f)
	if pkgs[0].Section != "admin" || pkgs[1].Section != "editors" || pkgs[2].Section != "shells" {
		t.Errorf("unexpected order: %s, %s, %s", pkgs[0].Section, pkgs[1].Section, pkgs[2].Section)
	}
}

func TestSortByArchitecture(t *testing.T) {
	pkgs := []PackageData{
		{Name: "a", Architecture: "i386"},
		{Name: "b", Architecture: "amd64"},
		{Name: "c", Architecture: "arm64"},
	}
	f := Filter{OrderBy: SortArchitecture}
	Sort(pkgs, f)
	if pkgs[0].Architecture != "amd64" || pkgs[1].Architecture != "arm64" || pkgs[2].Architecture != "i386" {
		t.Errorf("unexpected order: %s, %s, %s", pkgs[0].Architecture, pkgs[1].Architecture, pkgs[2].Architecture)
	}
}

func TestSortByVersion(t *testing.T) {
	pkgs := []PackageData{
		{Name: "c", Version: "3.0"},
		{Name: "a", Version: "1.0"},
		{Name: "b", Version: "2.0"},
	}
	f := Filter{OrderBy: SortVersion}
	Sort(pkgs, f)
	if pkgs[0].Version != "1.0" || pkgs[1].Version != "2.0" || pkgs[2].Version != "3.0" {
		t.Errorf("unexpected order: %s, %s, %s", pkgs[0].Version, pkgs[1].Version, pkgs[2].Version)
	}
}

func TestSortEmptyFieldsLast(t *testing.T) {
	pkgs := []PackageData{
		{Name: "empty", Section: ""},
		{Name: "vim", Section: "editors"},
		{Name: "apt", Section: "admin"},
	}
	f := Filter{OrderBy: SortSection}
	Sort(pkgs, f)
	if pkgs[2].Name != "empty" {
		t.Errorf("empty section should be last, got %q at position 2", pkgs[2].Name)
	}
}

func TestPdFieldEmpty(t *testing.T) {
	tests := []struct {
		name   string
		pkg    PackageData
		col    SortColumn
		expect bool
	}{
		{name: "empty name", pkg: PackageData{}, col: SortName, expect: true},
		{name: "non-empty name", pkg: PackageData{Name: "a"}, col: SortName, expect: false},
		{name: "empty version", pkg: PackageData{}, col: SortVersion, expect: true},
		{name: "has new version", pkg: PackageData{NewVersion: "2.0"}, col: SortVersion, expect: false},
		{name: "empty size", pkg: PackageData{}, col: SortSize, expect: true},
		{name: "dash size", pkg: PackageData{Size: "-"}, col: SortSize, expect: true},
		{name: "has size", pkg: PackageData{Size: "5 MB"}, col: SortSize, expect: false},
		{name: "empty section", pkg: PackageData{}, col: SortSection, expect: true},
		{name: "has section", pkg: PackageData{Section: "utils"}, col: SortSection, expect: false},
		{name: "empty arch", pkg: PackageData{}, col: SortArchitecture, expect: true},
		{name: "has arch", pkg: PackageData{Architecture: "amd64"}, col: SortArchitecture, expect: false},
		{name: "SortNone always false", pkg: PackageData{}, col: SortNone, expect: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pdFieldEmpty(tt.pkg, tt.col)
			if got != tt.expect {
				t.Errorf("pdFieldEmpty(%v, %d) = %v, want %v", tt.pkg, tt.col, got, tt.expect)
			}
		})
	}
}

func TestParseNotUpgradable(t *testing.T) {
	f := Parse("!upgradable")
	if f.Upgradable == nil || *f.Upgradable {
		t.Error("expected upgradable=false")
	}
}

func TestParseDescriptionAlias(t *testing.T) {
	f := Parse("description:browser")
	if f.Description != "browser" {
		t.Errorf("expected description 'browser', got '%s'", f.Description)
	}
}

func TestParseArchitectureAlias(t *testing.T) {
	f := Parse("architecture:arm64")
	if f.Architecture != "arm64" {
		t.Errorf("expected architecture 'arm64', got '%s'", f.Architecture)
	}
}

func TestParseVersionAlias(t *testing.T) {
	f := Parse("version:3.0")
	if f.Version != "3.0" {
		t.Errorf("expected version '3.0', got '%s'", f.Version)
	}
}

func TestContainsFold(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		expect bool
	}{
		{name: "exact match", s: "hello", substr: "hello", expect: true},
		{name: "case insensitive", s: "Hello", substr: "hello", expect: true},
		{name: "substring", s: "hello world", substr: "world", expect: true},
		{name: "no match", s: "hello", substr: "world", expect: false},
		{name: "empty substr", s: "hello", substr: "", expect: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsFold(tt.s, tt.substr)
			if got != tt.expect {
				t.Errorf("containsFold(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.expect)
			}
		})
	}
}
