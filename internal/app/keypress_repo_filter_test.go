package app

import (
	"strings"
	"testing"

	"github.com/mexirica/aptui/internal/filter"
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
