package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mexirica/aptui/internal/filter"
)

// Stacked layout: info panel (5 rows) sits above the list panel.
// tabBar(1) + gap(1) + infoPanel(5) + gap(1) = 8, then border(1) + header(1) + sep(1).
const (
	packageListHeaderY = 8  // first row inside list panel (header)
	packageListStartY  = 10 // first package item row
)

func (a App) onMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	a.exportConfirm = false
	m := msg.Mouse()

	switch msg.(type) {
	case tea.MouseWheelMsg:
		// Scroll on tabs that use their own lists.
		if a.activeTab == tabTransactions {
			return a.onTransactionScroll(m.Button)
		}
		if a.activeTab == tabRepos {
			return a.onPPAScroll(m.Button)
		}
		if a.activeTab == tabErrorLog {
			return a.onErrorLogScroll(m.Button)
		}
		if m.Button == tea.MouseWheelUp {
			a.selectedIdx -= 3
			if a.selectedIdx < 0 {
				a.selectedIdx = 0
			}
			a.adjustPackageScroll()
			return a, a.updateSelectionCmd()
		}
		if m.Button == tea.MouseWheelDown {
			a.selectedIdx += 3
			if a.selectedIdx >= len(a.filtered) {
				a.selectedIdx = len(a.filtered) - 1
			}
			if a.selectedIdx < 0 {
				a.selectedIdx = 0
			}
			a.adjustPackageScroll()
			return a, a.updateSelectionCmd()
		}
		return a, nil

	case tea.MouseClickMsg:
		if m.Button != tea.MouseLeft {
			return a, nil
		}

		y := m.Y

		// Click on tab bar (row 0) → switch tab
		if y == 0 {
			return a.onTabClick(m.X)
		}

		// Transactions/Repos/ErrorLog tabs: delegate to per-tab click handlers.
		if a.activeTab == tabTransactions {
			return a.onTransactionClick(m)
		}
		if a.activeTab == tabRepos {
			return a.onPPAClick(m)
		}
		if a.activeTab == tabErrorLog {
			return a.onErrorLogClick(m)
		}

		if a.sideBySide {
			return a.onSideBySideClick(m)
		}

		// Click on column header/separator area → toggle sort
		if y >= packageListHeaderY && y < packageListStartY {
			return a.onHeaderClick(m.X - 1) // -1 for left panel border
		}

		if y == a.searchBarY() && !a.searching {
			return a.openSearch()
		}

		row := y - packageListStartY
		if row < 0 || row >= a.packageListHeight() {
			return a, nil
		}

		idx := a.scrollOffset + row
		if idx < 0 || idx >= len(a.filtered) {
			return a, nil
		}

		// If clicking the already-selected row, toggle its selection (check/uncheck)
		if idx == a.selectedIdx {
			if a.selected == nil {
				a.selected = make(map[string]bool)
			}
			pkg := a.filtered[idx]
			if a.selected[pkg.Name] {
				delete(a.selected, pkg.Name)
			} else {
				a.selected[pkg.Name] = true
			}
			a.status = fmt.Sprintf("%d selected ", len(a.selected))
			return a, nil
		}

		// Move cursor to clicked row
		a.selectedIdx = idx
		a.adjustPackageScroll()
		return a, a.updateSelectionCmd()
	}

	return a, nil
}

func (a App) onTabClick(x int) (tea.Model, tea.Cmd) {
	labels := a.tabLabels()
	pos := 0
	for i, tab := range tabDefs {
		w := lipgloss.Width(a.tabStyle(tab).Render(labels[i]))
		if x >= pos && x < pos+w {
			if tab.kind == a.activeTab {
				return a, nil
			}
			a.activeTab = tab.kind
			cmd := a.activateTab()
			return a, cmd
		}
		pos += w
	}
	return a, nil
}

// onHeaderClick maps an X coordinate to a column and toggles sorting.
func (a App) onHeaderClick(x int) (tea.Model, tea.Cmd) {
	prefixW := 11
	available := a.width - prefixW - 4
	if available < 40 {
		available = 40
	}
	colName := available * 50 / 100
	colVersion := available * 35 / 100
	if colName < 20 {
		colName = 20
	}
	if colVersion < 12 {
		colVersion = 12
	}

	// Column boundaries (accounting for prefix and 2-char gaps)
	nameStart := prefixW
	nameEnd := nameStart + colName
	versionStart := nameEnd + 2
	versionEnd := versionStart + colVersion
	sizeStart := versionEnd + 2

	var clicked filter.SortColumn
	switch {
	case x >= nameStart && x < nameEnd:
		clicked = filter.SortName
	case x >= versionStart && x < versionEnd:
		clicked = filter.SortVersion
	case x >= sizeStart:
		clicked = filter.SortSize
	default:
		return a, nil
	}

	// Toggle: same column → flip direction; different column → ascending
	if a.sortColumn == clicked {
		if a.sortDesc {
			// Already descending → clear sort
			a.sortColumn = filter.SortNone
			a.sortDesc = false
		} else {
			a.sortDesc = true
		}
	} else {
		a.sortColumn = clicked
		a.sortDesc = false
	}

	a.applyFilter()
	return a, a.updateSelectionCmd()
}

// onSideBySideClick handles mouse clicks in side-by-side layout.
// In this layout the list panel starts at Y=1 (top border) and the list
// items begin at Y=5 (border + title + header + separator).
func (a App) onSideBySideClick(m tea.Mouse) (tea.Model, tea.Cmd) {
	leftW := a.sideListWidth()

	// Only handle clicks in the left (list) panel
	if m.X >= leftW {
		return a, nil
	}

	y := m.Y

	// Side-by-side layout: info row (5 rows) sits above the main row.
	// tabBar(1) + gap(1) + infoRow(5) + gap(1) = 8, then border(1) + header(1) + sep(1).
	const sideListHeaderY = 8  // header row inside list panel
	const sideListStartY = 10  // first package item row

	// Click on search bar area → open search
	if y == a.searchBarY() && !a.searching {
		return a.openSearch()
	}

	// Column header click → sort toggle
	if y >= sideListHeaderY && y < sideListStartY {
		return a.onHeaderClick(m.X - 1) // -1 for left border
	}

	row := y - sideListStartY
	if row < 0 || row >= a.packageListHeight() {
		return a, nil
	}

	idx := a.scrollOffset + row
	if idx < 0 || idx >= len(a.filtered) {
		return a, nil
	}

	if idx == a.selectedIdx {
		if a.selected == nil {
			a.selected = make(map[string]bool)
		}
		pkg := a.filtered[idx]
		if a.selected[pkg.Name] {
			delete(a.selected, pkg.Name)
		} else {
			a.selected[pkg.Name] = true
		}
		a.status = fmt.Sprintf("%d selected ", len(a.selected))
		return a, nil
	}

	a.selectedIdx = idx
	a.adjustPackageScroll()
	return a, a.updateSelectionCmd()
}

// --- Scroll handlers for non-package tabs ---

func (a App) onTransactionScroll(btn tea.MouseButton) (tea.Model, tea.Cmd) {
	switch btn {
	case tea.MouseWheelUp:
		a.transactionIdx -= 3
		if a.transactionIdx < 0 {
			a.transactionIdx = 0
		}
	case tea.MouseWheelDown:
		a.transactionIdx += 3
		if a.transactionIdx >= len(a.transactionItems) {
			a.transactionIdx = len(a.transactionItems) - 1
		}
		if a.transactionIdx < 0 {
			a.transactionIdx = 0
		}
	default:
		return a, nil
	}
	a.adjustTransactionScroll()
	a.transactionDeps = nil
	if len(a.transactionItems) > 0 && a.transactionIdx < len(a.transactionItems) {
		return a, loadTransactionDepsCmd(a.transactionIdx, a.transactionItems[a.transactionIdx].Packages)
	}
	return a, nil
}

func (a App) onPPAScroll(btn tea.MouseButton) (tea.Model, tea.Cmd) {
	switch btn {
	case tea.MouseWheelUp:
		a.ppaIdx -= 3
		if a.ppaIdx < 0 {
			a.ppaIdx = 0
		}
	case tea.MouseWheelDown:
		a.ppaIdx += 3
		if a.ppaIdx >= len(a.ppaItems) {
			a.ppaIdx = len(a.ppaItems) - 1
		}
		if a.ppaIdx < 0 {
			a.ppaIdx = 0
		}
	default:
		return a, nil
	}
	a.adjustPPAScroll()
	return a, nil
}

func (a App) onErrorLogScroll(btn tea.MouseButton) (tea.Model, tea.Cmd) {
	switch btn {
	case tea.MouseWheelUp:
		a.errlogIdx -= 3
		if a.errlogIdx < 0 {
			a.errlogIdx = 0
		}
	case tea.MouseWheelDown:
		a.errlogIdx += 3
		if a.errlogIdx >= len(a.errlogItems) {
			a.errlogIdx = len(a.errlogItems) - 1
		}
		if a.errlogIdx < 0 {
			a.errlogIdx = 0
		}
	default:
		return a, nil
	}
	a.adjustErrorLogScroll()
	return a, nil
}

// --- Click handlers for non-package tabs ---
// Layout: tabBar(Y=0) + gap(Y=1) + panelBorder(Y=2) + header(Y=3) + items(Y=4+).
// PPA list has an extra separator line, so items start at Y=5.

const (
	simpleListStartY = 4 // first item row for transactions / error log
	ppaListStartY    = 5 // first item row for PPAs (header + separator)
)

func (a App) onTransactionClick(m tea.Mouse) (tea.Model, tea.Cmd) {
	leftW := a.width / 2
	if m.X >= leftW {
		return a, nil
	}
	row := m.Y - simpleListStartY
	if row < 0 || row >= a.transactionListHeight() {
		return a, nil
	}
	idx := a.transactionOffset + row
	if idx < 0 || idx >= len(a.transactionItems) {
		return a, nil
	}
	a.transactionIdx = idx
	a.adjustTransactionScroll()
	a.transactionDeps = nil
	return a, loadTransactionDepsCmd(a.transactionIdx, a.transactionItems[a.transactionIdx].Packages)
}

func (a App) onPPAClick(m tea.Mouse) (tea.Model, tea.Cmd) {
	leftW := a.width / 2
	if m.X >= leftW {
		return a, nil
	}
	row := m.Y - ppaListStartY
	if row < 0 || row >= a.packageListHeight() {
		return a, nil
	}
	idx := a.ppaOffset + row
	if idx < 0 || idx >= len(a.ppaItems) {
		return a, nil
	}
	a.ppaIdx = idx
	a.adjustPPAScroll()
	return a, nil
}

func (a App) onErrorLogClick(m tea.Mouse) (tea.Model, tea.Cmd) {
	leftW := a.width / 2
	if m.X >= leftW {
		return a, nil
	}
	row := m.Y - simpleListStartY
	if row < 0 || row >= a.errorLogListHeight() {
		return a, nil
	}
	idx := a.errlogOffset + row
	if idx < 0 || idx >= len(a.errlogItems) {
		return a, nil
	}
	a.errlogIdx = idx
	a.adjustErrorLogScroll()
	return a, nil
}
