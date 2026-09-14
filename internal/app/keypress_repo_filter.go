package app

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/mexirica/aptui/internal/filter"
	"github.com/mexirica/aptui/internal/model"
)

func (a App) onRepoFilterKeypress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "esc":
		a.repoFilterView = false
		a.repoFilterItems = nil
		return a, nil
	case "j", "down":
		if a.repoFilterIdx < len(a.repoFilterItems)-1 {
			a.repoFilterIdx++
			a.adjustRepoFilterScroll()
		}
		return a, nil
	case "k", "up":
		if a.repoFilterIdx > 0 {
			a.repoFilterIdx--
			a.adjustRepoFilterScroll()
		}
		return a, nil
	case "ctrl+d", "pgdown":
		a.repoFilterIdx += a.repoFilterListHeight()
		if a.repoFilterIdx >= len(a.repoFilterItems) {
			a.repoFilterIdx = len(a.repoFilterItems) - 1
		}
		a.adjustRepoFilterScroll()
		return a, nil
	case "ctrl+u", "pgup":
		a.repoFilterIdx -= a.repoFilterListHeight()
		if a.repoFilterIdx < 0 {
			a.repoFilterIdx = 0
		}
		a.adjustRepoFilterScroll()
		return a, nil
	case "enter":
		return a.applyRepoFilter()
	}
	return a, nil
}

func (a App) applyRepoFilter() (tea.Model, tea.Cmd) {
	if len(a.repoFilterItems) == 0 || a.repoFilterIdx >= len(a.repoFilterItems) {
		a.repoFilterView = false
		return a, nil
	}
	selected := a.repoFilterItems[a.repoFilterIdx]
	a.repoFilterView = false
	a.repoFilterItems = nil

	if selected == "" {
		a.filterQuery = removeRepoToken(a.filterQuery)
	} else {
		a.filterQuery = replaceRepoToken(a.filterQuery, selected)
	}

	a.applyFilter(false)
	a.status = ""
	return a, a.updateSelectionCmd()
}

func (a App) openRepoFilter() (tea.Model, tea.Cmd) {
	repos := reposFromPackages(a.allPackages)
	if len(repos) == 0 {
		a.status = "No repository metadata available. Run apt update first."
		return a, nil
	}

	items := append([]string{""}, repos...)

	a.repoFilterItems = items
	a.repoFilterIdx = 0
	a.repoFilterOffset = 0

	if current := currentRepoFilter(a.filterQuery); current != "" {
		for i, r := range items {
			if r == current {
				a.repoFilterIdx = i
				break
			}
		}
	}

	a.repoFilterView = true
	return a, nil
}

// reposFromPackages extracts sorted unique individual origins from all packages.
func reposFromPackages(pkgs []model.Package) []string {
	seen := make(map[string]bool)
	for _, p := range pkgs {
		if p.Origin == "" {
			continue
		}
		for _, part := range strings.Split(p.Origin, "; ") {
			part = strings.TrimSpace(part)
			if part != "" && !seen[part] {
				seen[part] = true
			}
		}
	}
	result := make([]string, 0, len(seen))
	for r := range seen {
		result = append(result, r)
	}
	sort.Strings(result)
	return result
}

// currentRepoFilter extracts the repo: value from a filter query string,
// handling both plain and quoted forms: repo:foo or repo:"foo bar".
func currentRepoFilter(q string) string {
	af := filter.Parse(q)
	return af.Origin
}

// splitQueryTokens splits q on whitespace, treating quoted strings as single tokens.
// Unlike strings.Fields, a value like repo:"foo bar" or repo:'foo bar' is kept as one token.
func splitQueryTokens(q string) []string {
	var tokens []string
	var cur strings.Builder
	inQuote := false
	quoteChar := rune(0)
	escaped := false
	for _, r := range q {
		if inQuote {
			cur.WriteRune(r)
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == quoteChar {
				inQuote = false
			}
		} else if r == '"' || r == '\'' {
			inQuote = true
			quoteChar = r
			cur.WriteRune(r)
		} else if r == ' ' || r == '\t' {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

// removeRepoToken strips repo:/origin: tokens from q.
func removeRepoToken(q string) string {
	var out []string
	for _, t := range splitQueryTokens(q) {
		lower := strings.ToLower(t)
		if !strings.HasPrefix(lower, "repo:") && !strings.HasPrefix(lower, "origin:") {
			out = append(out, t)
		}
	}
	return strings.Join(out, " ")
}

// replaceRepoToken replaces any existing repo:/origin: token with repo:"value",
// or appends one. Quotes are added when needed for spaces or escaped characters.
func replaceRepoToken(q string, repo string) string {
	var quoted string
	if strings.ContainsAny(repo, " \t\"\\") {
		escaped := strings.ReplaceAll(repo, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		quoted = `repo:"` + escaped + `"`
	} else {
		quoted = "repo:" + repo
	}

	tokens := splitQueryTokens(q)
	found := false
	for i, t := range tokens {
		lower := strings.ToLower(t)
		if strings.HasPrefix(lower, "repo:") || strings.HasPrefix(lower, "origin:") {
			tokens[i] = quoted
			found = true
			break
		}
	}
	if !found {
		tokens = append(tokens, quoted)
	}
	return strings.Join(tokens, " ")
}

func (a *App) adjustRepoFilterScroll() {
	h := a.repoFilterListHeight()
	if a.repoFilterIdx < a.repoFilterOffset {
		a.repoFilterOffset = a.repoFilterIdx
	}
	if a.repoFilterIdx >= a.repoFilterOffset+h {
		a.repoFilterOffset = a.repoFilterIdx - h + 1
	}
}

func (a App) repoFilterListHeight() int {
	h := a.height - 10
	if h < 3 {
		h = 3
	}
	return h
}
