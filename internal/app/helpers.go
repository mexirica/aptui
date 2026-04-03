package app

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mexirica/aptui/internal/filter"
	"github.com/mexirica/aptui/internal/fuzzy"
	"github.com/mexirica/aptui/internal/model"
	"github.com/mexirica/aptui/internal/ui"
)

type scoredPackage struct {
	pkg   model.Package
	score int
}

// applyComponentStyles refreshes help and spinner styles after a theme change.
func (a *App) applyComponentStyles() {
	a.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	a.help.Styles.FullKey = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	a.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ui.ColorNormalText)
	a.help.Styles.FullDesc = lipgloss.NewStyle().Foreground(ui.ColorNormalText)
	a.help.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ui.ColorHelpSep)
	a.help.Styles.FullSeparator = lipgloss.NewStyle().Foreground(ui.ColorHelpSep)
	a.spinner.Style = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
}

// tabStyle returns the appropriate style for a tab given the current state.
func (a App) tabStyle(t tabDef) lipgloss.Style {
	if t.kind == a.activeTab {
		return ui.TabActiveStyle
	}
	if t.kind == tabUpgradable && len(a.upgradableMap) > 0 {
		return ui.TabNotifyStyle
	}
	if t.kind == tabCleanup && len(a.autoremovable) > 0 {
		return ui.TabNotifyStyle
	}
	if t.kind == tabErrorLog && a.errlogStore.Count() > 0 {
		return ui.TabNotifyStyle
	}
	return ui.TabInactiveStyle
}

// activateTab switches to the given tab and returns the commands to refresh the view.
func (a *App) activateTab() tea.Cmd {
	if a.activeTab == tabErrorLog {
		a.errlogItems = a.errlogStore.All()
		a.errlogIdx = 0
		a.errlogOffset = 0
		a.status = ""
		return nil
	}
	if a.activeTab == tabTransactions {
		a.transactionItems = a.transactionStore.All()
		a.transactionIdx = 0
		a.transactionOffset = 0
		a.transactionDeps = nil
		a.status = ""
		var cmd tea.Cmd
		if len(a.transactionItems) > 0 {
			cmd = loadTransactionDepsCmd(0, a.transactionItems[0].Packages)
		}
		return cmd
	}
	if a.activeTab == tabRepos {
		a.ppaItems = nil
		a.ppaIdx = 0
		a.ppaOffset = 0
		a.ppaAdding = false
		a.loading = true
		a.status = "Loading repositories..."
		return tea.Batch(a.spinner.Tick, listPPAsCmd())
	}
	a.applyFilter()
	cmd := a.updateSelectionCmd()
	a.status = fmt.Sprintf("%d packages ", len(a.filtered))
	return cmd
}

// applyFilter rebuilds the filtered list from allPackages based on active tab,
// advanced filter, and search query. Uses fuzzy scoring when a search query is active.
func (a *App) applyFilter() {
	var source []model.Package
	switch a.activeTab {
	case tabInstalled:
		for _, p := range a.allPackages {
			if p.Installed {
				source = append(source, p)
			}
		}
	case tabUpgradable:
		for _, p := range a.allPackages {
			if p.Upgradable {
				source = append(source, p)
			}
		}
	case tabCleanup:
		for _, p := range a.allPackages {
			if a.autoremovableSet[p.Name] {
				source = append(source, p)
			}
		}
	default:
		source = make([]model.Package, len(a.allPackages))
		copy(source, a.allPackages)
	}

	af := filter.Parse(a.filterQuery)

	// Apply structured filter criteria (section:, arch:, size>, etc.)
	if af.Section != "" || af.Architecture != "" || af.Size != nil ||
		af.Installed != nil || af.Upgradable != nil ||
		af.Name != "" || af.Version != "" || af.Description != "" {
		var filtered []model.Package
		for _, p := range source {
			pd := filter.PackageData{
				Name:         p.Name,
				Version:      p.Version,
				NewVersion:   p.NewVersion,
				Size:         p.Size,
				Description:  p.Description,
				Installed:    p.Installed,
				Upgradable:   p.Upgradable,
				Section:      p.Section,
				Architecture: p.Architecture,
			}
			if af.Match(pd) {
				filtered = append(filtered, p)
			}
		}
		source = filtered
	}

	// Apply fuzzy search on free text (unrecognized tokens)
	freeText := af.FreeText
	if freeText == "" {
		a.filtered = source
	} else {
		minScore := fuzzy.MinQuality(len(freeText))
		var scored []scoredPackage
		for _, p := range source {
			nameRes := fuzzy.Score(freeText, p.Name)
			descRes := fuzzy.Score(freeText, p.Description)

			s := 0
			matched := false
			if nameRes.Matched {
				matched = true
				s = nameRes.Score + 50
			}
			if descRes.Matched && descRes.Score > s {
				matched = true
				s = descRes.Score
			}

			if matched && s >= minScore {
				scored = append(scored, scoredPackage{pkg: p, score: s})
			}
		}
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})

		a.filtered = make([]model.Package, len(scored))
		for i, sp := range scored {
			a.filtered[i] = sp.pkg
		}
	}

	// Apply sorting: click-based sort takes priority over filter-based sort
	sortCol := af.OrderBy
	sortDesc := af.OrderDesc
	if a.sortColumn != filter.SortNone {
		sortCol = a.sortColumn
		sortDesc = a.sortDesc
	}
	if sortCol != filter.SortNone {
		sort.SliceStable(a.filtered, func(i, j int) bool {
			pi, pj := a.filtered[i], a.filtered[j]

			// Push packages with unknown data to the end
			iEmpty, jEmpty := sortFieldEmpty(pi, sortCol), sortFieldEmpty(pj, sortCol)
			if iEmpty != jEmpty {
				return !iEmpty // non-empty comes first
			}
			if iEmpty && jEmpty {
				return false // both empty, keep original order
			}

			var less bool
			switch sortCol {
			case filter.SortName:
				less = strings.ToLower(pi.Name) < strings.ToLower(pj.Name)
			case filter.SortVersion:
				less = pi.Version < pj.Version
			case filter.SortSize:
				less = filter.ParseSizeToKB(pi.Size) < filter.ParseSizeToKB(pj.Size)
			case filter.SortSection:
				less = strings.ToLower(pi.Section) < strings.ToLower(pj.Section)
			case filter.SortArchitecture:
				less = strings.ToLower(pi.Architecture) < strings.ToLower(pj.Architecture)
			default:
				return false
			}
			if sortDesc {
				return !less
			}
			return less
		})
	}

	if len(a.pinnedSet) > 0 && sortCol == filter.SortNone && freeText == "" {
		sort.SliceStable(a.filtered, func(i, j int) bool {
			pi := a.pinnedSet[a.filtered[i].Name]
			pj := a.pinnedSet[a.filtered[j].Name]
			return pi && !pj
		})
	}

	a.selectedIdx = 0
	a.scrollOffset = 0
}

// effectiveSortInfo returns the active sort state, preferring click-based sort
// over filter-based sort.
func (a App) effectiveSortInfo() filter.SortInfo {
	if a.sortColumn != filter.SortNone {
		return filter.SortInfo{Column: a.sortColumn, Desc: a.sortDesc}
	}
	af := filter.Parse(a.filterQuery)
	return filter.SortInfo{Column: af.OrderBy, Desc: af.OrderDesc}
}

func sortFieldEmpty(p model.Package, col filter.SortColumn) bool {
	switch col {
	case filter.SortName:
		return p.Name == ""
	case filter.SortVersion:
		return p.Version == "" && p.NewVersion == ""
	case filter.SortSize:
		return p.Size == "" || p.Size == "-"
	case filter.SortSection:
		return p.Section == ""
	case filter.SortArchitecture:
		return p.Architecture == ""
	default:
		return false
	}
}

// rebuildIndex rebuilds the package name to index mapping for O(1) lookups.
func (a *App) rebuildIndex() {
	a.pkgIndex = make(map[string]int, len(a.allPackages))
	for i, p := range a.allPackages {
		a.pkgIndex[p.Name] = i
	}
}

// applyOptimisticUpdate updates in-memory package state immediately after
// a successful operation, avoiding the need to wait for a full reload.
func (a *App) applyOptimisticUpdate(op string, pkgs []string) {
	switch op {
	case "install":
		for _, name := range pkgs {
			if idx, ok := a.pkgIndex[name]; ok {
				if !a.allPackages[idx].Installed {
					a.installedCount++
				}
				a.allPackages[idx].Installed = true
				a.allPackages[idx].Upgradable = false
				a.allPackages[idx].NewVersion = ""
				delete(a.upgradableMap, name)
			}
		}
	case "remove", "purge":
		for _, name := range pkgs {
			if idx, ok := a.pkgIndex[name]; ok {
				if a.allPackages[idx].Installed {
					a.installedCount--
				}
				a.allPackages[idx].Installed = false
				a.allPackages[idx].Upgradable = false
				a.allPackages[idx].NewVersion = ""
				delete(a.upgradableMap, name)
			}
		}
	case "upgrade", "upgrade-all":
		for _, name := range pkgs {
			if idx, ok := a.pkgIndex[name]; ok {
				if up, ok := a.upgradableMap[name]; ok {
					a.allPackages[idx].Version = up.NewVersion
				}
				a.allPackages[idx].Upgradable = false
				a.allPackages[idx].NewVersion = ""
				delete(a.upgradableMap, name)
			}
		}
	case "cleanup-all":
		for _, name := range pkgs {
			if idx, ok := a.pkgIndex[name]; ok {
				if a.allPackages[idx].Installed {
					a.installedCount--
				}
				a.allPackages[idx].Installed = false
				a.allPackages[idx].Upgradable = false
				a.allPackages[idx].NewVersion = ""
				delete(a.upgradableMap, name)
			}
		}
		a.autoremovable = nil
		a.autoremovableSet = make(map[string]bool)
	}
	a.applyFilter()
}

func (a *App) adjustPackageScroll() {
	h := a.packageListHeight()
	if a.selectedIdx < a.scrollOffset {
		a.scrollOffset = a.selectedIdx
	}
	if a.selectedIdx >= a.scrollOffset+h {
		a.scrollOffset = a.selectedIdx - h + 1
	}
}

func (a *App) adjustMirrorScroll() {
	h := a.packageListHeight()
	if a.fetchIdx < a.fetchOffset {
		a.fetchOffset = a.fetchIdx
	}
	if a.fetchIdx >= a.fetchOffset+h {
		a.fetchOffset = a.fetchIdx - h + 1
	}
}

func (a *App) adjustTransactionScroll() {
	h := a.transactionListHeight()
	if a.transactionIdx < a.transactionOffset {
		a.transactionOffset = a.transactionIdx
	}
	if a.transactionIdx >= a.transactionOffset+h {
		a.transactionOffset = a.transactionIdx - h + 1
	}
}

// searchBarY returns the Y coordinate of the search bar row.
func (a App) searchBarY() int {
	if a.sideBySide {
		// In side-by-side, the search panel is in the info row below the main panels.
		// Tab(1) + mainPanel + border of info row + title
		return 1 + a.sideMainPanelHeight() + 2
	}
	helpLines := strings.Count(a.help.View(a.keys), "\n") + 1
	if !a.loading && len(a.filtered) > 0 {
		detailLines := a.packageDetailHeight()
		if a.detailName == "" || a.detailInfo == "" {
			idx := a.selectedIdx
			if idx >= len(a.filtered) {
				idx = len(a.filtered) - 1
			}
			pkg := a.filtered[idx]
			detailLines = strings.Count(a.renderBasicDetail(pkg), "\n")
		}
		return a.height - 4 - detailLines - helpLines
	}
	return a.height - 3 - helpLines
}

// Side-by-side layout constants and helpers.
const (
	sideInfoRowH = 5  // 3 inner lines + 2 border lines
	sideSplitPct = 60 // left panel percentage
	sideMinWidth = 120
)

func (a App) sideKeysRowH() int {
	helpLines := strings.Count(a.help.View(a.keys), "\n") + 1
	return helpLines + 2 // help lines + 2 border lines (title is now in border)
}

func (a App) sideListWidth() int {
	return a.width * sideSplitPct / 100
}

func (a App) sideDetailWidth() int {
	return a.width - a.sideListWidth()
}

func (a App) sideMainPanelHeight() int {
	// height minus: 1 (tab bar) + infoRow + keysRow
	h := a.height - 1 - sideInfoRowH - a.sideKeysRowH()
	if h < 7 {
		h = 7
	}
	return h
}

func (a App) packageListHeight() int {
	if a.sideBySide {
		// In side-by-side: inner height of main panel = panelH - 2 (border) - 2 (header+separator)
		h := a.sideMainPanelHeight() - 2 - 2
		if h < 5 {
			h = 5
		}
		return h
	}
	helpLines := strings.Count(a.help.View(a.keys), "\n") + 1
	h := a.height - a.packageDetailHeight() - 9 - helpLines
	if h < 5 {
		h = 5
	}
	return h
}

func (a App) packageDetailHeight() int {
	return 10
}

func (a App) sideDetailInnerHeight() int {
	return a.sideMainPanelHeight() - 2 // minus border
}

func (a App) transactionListHeight() int {
	helpLines := strings.Count(a.help.View(a.keys), "\n") + 1
	footerLines := 2 + helpLines
	innerH := a.height - 3 - footerLines
	if innerH < 5 {
		innerH = 5
	}
	mv := innerH - 1
	if mv < 3 {
		mv = 3
	}
	return mv
}

func (a *App) adjustErrorLogScroll() {
	h := a.errorLogListHeight()
	if a.errlogIdx < a.errlogOffset {
		a.errlogOffset = a.errlogIdx
	}
	if a.errlogIdx >= a.errlogOffset+h {
		a.errlogOffset = a.errlogIdx - h + 1
	}
}

func (a App) errorLogListHeight() int {
	helpLines := strings.Count(a.help.View(a.keys), "\n") + 1
	footerLines := 2 + helpLines
	innerH := a.height - 3 - footerLines
	if innerH < 5 {
		innerH = 5
	}
	mv := innerH - 1
	if mv < 3 {
		mv = 3
	}
	return mv
}

func (a *App) adjustPPAScroll() {
	h := a.packageListHeight()
	if a.ppaIdx < a.ppaOffset {
		a.ppaOffset = a.ppaIdx
	}
	if a.ppaIdx >= a.ppaOffset+h {
		a.ppaOffset = a.ppaIdx - h + 1
	}
}

func (a *App) adjustFileListScroll() {
	h := a.fileListHeight()
	if a.fileListIdx < a.fileListOffset {
		a.fileListOffset = a.fileListIdx
	}
	if a.fileListIdx >= a.fileListOffset+h {
		a.fileListOffset = a.fileListIdx - h + 1
	}
}

func (a App) fileListHeight() int {
	return a.packageDetailHeight()
}

func friendlyError(err error) string {
	if err == nil {
		return "unknown error"
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		if stderr != "" {
			return stderr
		}
		switch exitErr.ExitCode() {
		case 100:
			return "apt failed — check your sources list or network connection"
		case 1:
			return "operation failed — try running with sudo"
		default:
			return fmt.Sprintf("process exited with code %d", exitErr.ExitCode())
		}
	}
	return err.Error()
}
