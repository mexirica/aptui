package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mexirica/aptui/internal/filter"
	"github.com/mexirica/aptui/internal/model"
)

func TestReplaceRepoTokenEscapesQuotedValues(t *testing.T) {
	repo := "Local \"Quoted\" Repo"
	q := replaceRepoToken("installed python", repo)
	if !strings.Contains(q, `repo:"Local \"Quoted\" Repo"`) {
		t.Fatalf("expected escaped quoted repo token, got %q", q)
	}

	if got := currentRepoFilter(q); got != repo {
		t.Fatalf("currentRepoFilter()=%q, want %q", got, repo)
	}

	af := filter.Parse(q)
	if af.Origin != repo {
		t.Fatalf("filter.Parse origin=%q, want %q", af.Origin, repo)
	}
}

func TestRemoveRepoTokenHandlesQuotedValue(t *testing.T) {
	q := `installed repo:"archive.ubuntu.com/ubuntu noble/main" python`
	got := removeRepoToken(q)
	if strings.Contains(got, "repo:") || strings.Contains(got, "origin:") {
		t.Fatalf("expected repo token removed, got %q", got)
	}
	if !strings.Contains(got, "installed") || !strings.Contains(got, "python") {
		t.Fatalf("expected non-repo tokens preserved, got %q", got)
	}
}

func TestRemoveRepoTokenHandlesWrappedQuotedToken(t *testing.T) {
	q := `"repo:apt.pop-os.org/ubuntu noble/main" installed`
	got := removeRepoToken(q)
	if strings.Contains(got, "repo:") || strings.Contains(got, "origin:") {
		t.Fatalf("expected wrapped repo token removed, got %q", got)
	}
	if got != "installed" {
		t.Fatalf("expected only installed token preserved, got %q", got)
	}
}

func TestRemoveRepoTokenHandlesEscapedWhitespace(t *testing.T) {
	q := `repo:apt.pop-os.org/ubuntu\ noble/main installed`
	got := removeRepoToken(q)
	if strings.Contains(got, "repo:") || strings.Contains(got, "origin:") {
		t.Fatalf("expected escaped repo token removed, got %q", got)
	}
	if got != "installed" {
		t.Fatalf("expected only installed token preserved, got %q", got)
	}
}

func TestOpenRepoFilterRestoresSelectionVisible(t *testing.T) {
	a := newTestApp()
	a.height = 16 // repoFilterListHeight = 6
	a.filterQuery = `repo:"repo-12"`

	pkgs := make([]model.Package, 0, 20)
	for i := 0; i < 20; i++ {
		pkgs = append(pkgs, model.Package{
			Name:   fmt.Sprintf("pkg-%02d", i),
			Origin: fmt.Sprintf("repo-%02d", i),
		})
	}
	a.allPackages = pkgs

	m, _ := a.openRepoFilter()
	app := m.(App)

	if !app.repoFilterView {
		t.Fatal("repoFilterView should be active")
	}
	if app.repoFilterIdx <= 0 {
		t.Fatalf("expected restored repo selection index > 0, got %d", app.repoFilterIdx)
	}
	if app.repoFilterIdx < app.repoFilterOffset || app.repoFilterIdx >= app.repoFilterOffset+app.repoFilterListHeight() {
		t.Fatalf("selected index %d should be visible within offset %d and height %d", app.repoFilterIdx, app.repoFilterOffset, app.repoFilterListHeight())
	}
}
