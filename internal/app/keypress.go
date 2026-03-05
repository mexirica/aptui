package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mexirica/gpm/internal/fetch"
	"github.com/mexirica/gpm/internal/history"
	"github.com/mexirica/gpm/internal/ui"
)

func (a App) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := a.searchInput.Value()
		a.searching = false
		a.searchInput.Blur()
		a.filterQuery = query
		if query == "" {
			a.applyFilter()
			a.status = fmt.Sprintf("%d packages ", len(a.filtered))
			if len(a.filtered) > 0 {
				return a, showDetailCmd(a.filtered[0].Name)
			}
			return a, nil
		}
		if len(a.filtered) == 0 {
			a.loading = true
			a.status = fmt.Sprintf("Searching '%s' via apt-cache...", query)
			return a, searchCmd(query)
		}
		a.status = fmt.Sprintf("%d packages matching '%s'", len(a.filtered), query)
		return a, tea.Batch(showDetailCmd(a.filtered[0].Name), a.loadVisibleVersionsCmd())
	case "esc":
		a.searching = false
		a.searchInput.Blur()
		a.filterQuery = ""
		a.applyFilter()
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
		return a, nil
	default:
		var cmd tea.Cmd
		a.searchInput, cmd = a.searchInput.Update(msg)
		a.filterQuery = a.searchInput.Value()
		a.applyFilter()
		a.status = fmt.Sprintf("%d matching ", len(a.filtered))
		var detailCmd tea.Cmd
		if len(a.filtered) > 0 {
			detailCmd = showDetailCmd(a.filtered[0].Name)
		}
		return a, tea.Batch(cmd, detailCmd, a.loadVisibleVersionsCmd())
	}
}

func (a App) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "q" || msg.String() == "ctrl+c":
		return a, tea.Quit

	case msg.String() == "h":
		a.help.ShowAll = !a.help.ShowAll
		return a, nil

	case msg.String() == "/":
		a.searching = true
		a.searchInput.Focus()
		a.searchInput.SetValue(a.filterQuery)
		return a, textinput.Blink

	case msg.String() == "esc":
		if a.filterQuery != "" {
			a.filterQuery = ""
			a.applyFilter()
			a.selectedIdx = 0
			a.scrollOffset = 0
			a.status = fmt.Sprintf("%d packages ", len(a.filtered))
			var cmds []tea.Cmd
			if len(a.filtered) > 0 {
				cmds = append(cmds, showDetailCmd(a.filtered[0].Name))
			}
			cmds = append(cmds, a.loadVisibleVersionsCmd())
			return a, tea.Batch(cmds...)
		}
		return a, nil

	case msg.String() == "j" || msg.String() == "down":
		if a.selectedIdx < len(a.filtered)-1 {
			a.selectedIdx++
			a.adjustScroll()
			return a, showDetailCmd(a.filtered[a.selectedIdx].Name)
		}
		return a, nil

	case msg.String() == "k" || msg.String() == "up":
		if a.selectedIdx > 0 {
			a.selectedIdx--
			a.adjustScroll()
			return a, showDetailCmd(a.filtered[a.selectedIdx].Name)
		}
		return a, nil

	case msg.String() == "ctrl+d" || msg.String() == "pgdown":
		a.selectedIdx += a.listHeight()
		if a.selectedIdx >= len(a.filtered) {
			a.selectedIdx = len(a.filtered) - 1
		}
		if a.selectedIdx < 0 {
			a.selectedIdx = 0
		}
		a.adjustScroll()
		var cmds []tea.Cmd
		if len(a.filtered) > 0 {
			cmds = append(cmds, showDetailCmd(a.filtered[a.selectedIdx].Name))
		}
		cmds = append(cmds, a.loadVisibleVersionsCmd())
		return a, tea.Batch(cmds...)

	case msg.String() == "ctrl+u" || msg.String() == "pgup":
		a.selectedIdx -= a.listHeight()
		if a.selectedIdx < 0 {
			a.selectedIdx = 0
		}
		a.adjustScroll()
		var cmds []tea.Cmd
		if len(a.filtered) > 0 {
			cmds = append(cmds, showDetailCmd(a.filtered[a.selectedIdx].Name))
		}
		cmds = append(cmds, a.loadVisibleVersionsCmd())
		return a, tea.Batch(cmds...)

	case msg.String() == " ":
		if len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) {
			pkg := a.filtered[a.selectedIdx]
			if a.selected == nil {
				a.selected = make(map[string]bool)
			}
			if a.selected[pkg.Name] {
				delete(a.selected, pkg.Name)
			} else {
				a.selected[pkg.Name] = true
			}
			a.status = fmt.Sprintf("%d selected ", len(a.selected))
			return a, nil
		}
		return a, nil

	case msg.String() == "A":
		if a.selected == nil {
			a.selected = make(map[string]bool)
		}
		allSelected := true
		for _, p := range a.filtered {
			if !a.selected[p.Name] {
				allSelected = false
				break
			}
		}
		if allSelected {
			a.selected = make(map[string]bool)
			a.status = "0 selected "
			return a, nil
		}
		for _, p := range a.filtered {
			a.selected[p.Name] = true
		}
		a.status = fmt.Sprintf("%d selected ", len(a.selected))
		return a, nil

	case msg.String() == "i":
		if len(a.selected) > 0 {
			var cmds []tea.Cmd
			var names []string
			for name := range a.selected {
				cmds = append(cmds, installCmd(name))
				names = append(names, name)
			}
			a.pendingExecOp = "install"
			a.pendingExecPkgs = names
			a.pendingExecCount = len(cmds)
			a.loading = true
			a.status = fmt.Sprintf("Installing %d packages...", len(cmds))
			a.selected = make(map[string]bool)
			return a, tea.Batch(cmds...)
		}
		if len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) {
			pkg := a.filtered[a.selectedIdx]
			if pkg.Installed {
				a.status = fmt.Sprintf("'%s' is already installed.", pkg.Name)
				return a, nil
			}
			a.pendingExecOp = "install"
			a.pendingExecPkgs = []string{pkg.Name}
			a.pendingExecCount = 1
			a.loading = true
			a.status = fmt.Sprintf("Installing %s...", pkg.Name)
			return a, installCmd(pkg.Name)
		}
		return a, nil

	case msg.String() == "r":
		if len(a.selected) > 0 {
			var cmds []tea.Cmd
			var names []string
			for name := range a.selected {
				cmds = append(cmds, removeCmd(name))
				names = append(names, name)
			}
			a.pendingExecOp = "remove"
			a.pendingExecPkgs = names
			a.pendingExecCount = len(cmds)
			a.loading = true
			a.status = fmt.Sprintf("Removing %d packages...", len(cmds))
			a.selected = make(map[string]bool)
			return a, tea.Batch(cmds...)
		}
		if len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) {
			pkg := a.filtered[a.selectedIdx]
			if !pkg.Installed {
				a.status = fmt.Sprintf("'%s' is not installed.", pkg.Name)
				return a, nil
			}
			a.pendingExecOp = "remove"
			a.pendingExecPkgs = []string{pkg.Name}
			a.pendingExecCount = 1
			a.loading = true
			a.status = fmt.Sprintf("Removing %s...", pkg.Name)
			return a, removeCmd(pkg.Name)
		}
		return a, nil

	case msg.String() == "u":
		if len(a.selected) > 0 {
			var cmds []tea.Cmd
			var names []string
			for name := range a.selected {
				cmds = append(cmds, upgradeCmd(name))
				names = append(names, name)
			}
			a.pendingExecOp = "upgrade"
			a.pendingExecPkgs = names
			a.pendingExecCount = len(cmds)
			a.loading = true
			a.status = fmt.Sprintf("Upgrading %d packages...", len(cmds))
			a.selected = make(map[string]bool)
			return a, tea.Batch(cmds...)
		}
		if len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) {
			pkg := a.filtered[a.selectedIdx]
			if !pkg.Upgradable {
				a.status = fmt.Sprintf("'%s' is already at the latest version.", pkg.Name)
				return a, nil
			}
			a.pendingExecOp = "upgrade"
			a.pendingExecPkgs = []string{pkg.Name}
			a.pendingExecCount = 1
			a.loading = true
			a.status = fmt.Sprintf("Upgrading %s...", pkg.Name)
			return a, upgradeCmd(pkg.Name)
		}
		return a, nil

	case msg.String() == "G":
		a.pendingExecOp = "upgrade-all"
		a.pendingExecPkgs = []string{"all"}
		a.pendingExecCount = 1
		a.loading = true
		a.status = "Upgrading ALL packages (sudo apt-get upgrade)..."
		return a, upgradeAllCmd()

	case msg.String() == "ctrl+r":
		a.loading = true
		a.filterQuery = ""
		a.status = "Reloading..."
		return a, loadAllCmd

	case msg.String() == "t":
		a.transactionView = true
		a.transactionItems = a.transactionStore.All()
		a.transactionIdx = 0
		a.transactionOffset = 0
		a.transactionDeps = nil
		a.status = fmt.Sprintf("%d transactions | esc back | z undo | x redo ", len(a.transactionItems))
		var cmd tea.Cmd
		if len(a.transactionItems) > 0 {
			cmd = loadTxDepsCmd(0, a.transactionItems[0].Packages)
		}
		return a, cmd

	case msg.String() == "f":
		a.fetchView = true
		a.fetchMirrors = nil
		a.fetchSelected = make(map[int]bool)
		a.fetchIdx = 0
		a.fetchOffset = 0
		a.fetchTesting = true
		a.loading = true
		a.status = "Detecting distro and fetching mirror list..."
		return a, tea.Batch(a.spinner.Tick, fetchMirrorsCmd())

	case msg.String() == "tab":
		a.activeTab = (a.activeTab + 1) % 3
		a.applyFilter()
		var cmds []tea.Cmd
		if len(a.filtered) > 0 {
			cmds = append(cmds, showDetailCmd(a.filtered[0].Name))
		}
		cmds = append(cmds, a.loadVisibleVersionsCmd())
		tabNames := []string{"All", "Installed", "Upgradable"}
		a.status = fmt.Sprintf("%d packages (%s) ", len(a.filtered), tabNames[a.activeTab])
		return a, tea.Batch(cmds...)

	case msg.String() == "shift+tab":
		a.activeTab = (a.activeTab + 2) % 3
		a.applyFilter()
		var cmds []tea.Cmd
		if len(a.filtered) > 0 {
			cmds = append(cmds, showDetailCmd(a.filtered[0].Name))
		}
		cmds = append(cmds, a.loadVisibleVersionsCmd())
		tabNames := []string{"All", "Installed", "Upgradable"}
		a.status = fmt.Sprintf("%d packages (%s) ", len(a.filtered), tabNames[a.activeTab])
		return a, tea.Batch(cmds...)
	}

	return a, nil
}

func (a App) handleFetchKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.fetchTesting {
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "ctrl+c" {
			a.fetchView = false
			a.fetchTesting = false
			a.loading = false
			a.status = "Fetch cancelled."
			return a, nil
		}
		return a, nil
	}

	switch {
	case msg.String() == "esc":
		a.fetchView = false
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
		return a, nil

	case msg.String() == "q" || msg.String() == "ctrl+c":
		return a, tea.Quit

	case msg.String() == "j" || msg.String() == "down":
		if a.fetchIdx < len(a.fetchMirrors)-1 {
			a.fetchIdx++
			a.adjustFetchScroll()
		}
		return a, nil

	case msg.String() == "k" || msg.String() == "up":
		if a.fetchIdx > 0 {
			a.fetchIdx--
			a.adjustFetchScroll()
		}
		return a, nil

	case msg.String() == "ctrl+d" || msg.String() == "pgdown":
		a.fetchIdx += a.listHeight()
		if a.fetchIdx >= len(a.fetchMirrors) {
			a.fetchIdx = len(a.fetchMirrors) - 1
		}
		if a.fetchIdx < 0 {
			a.fetchIdx = 0
		}
		a.adjustFetchScroll()
		return a, nil

	case msg.String() == "ctrl+u" || msg.String() == "pgup":
		a.fetchIdx -= a.listHeight()
		if a.fetchIdx < 0 {
			a.fetchIdx = 0
		}
		a.adjustFetchScroll()
		return a, nil

	case msg.String() == " ":
		if len(a.fetchMirrors) > 0 && a.fetchIdx < len(a.fetchMirrors) {
			if a.fetchSelected[a.fetchIdx] {
				delete(a.fetchSelected, a.fetchIdx)
			} else {
				a.fetchSelected[a.fetchIdx] = true
			}
			a.status = fmt.Sprintf("%d mirrors selected | enter: apply • esc: cancel", len(a.fetchSelected))
		}
		return a, nil

	case msg.String() == "enter":
		if len(a.fetchSelected) == 0 {
			a.status = ui.ErrorStyle.Render("Select at least one mirror (space to toggle).")
			return a, nil
		}
		for i := range a.fetchMirrors {
			a.fetchMirrors[i].Active = a.fetchSelected[i]
		}
		cmd := fetch.WriteSourcesListCmd(a.fetchMirrors, a.fetchDistro)
		return a, tea.ExecProcess(cmd, func(err error) tea.Msg {
			return fetchApplyMsg{err: err}
		})
	}
	return a, nil
}

func (a App) handleTransactionKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc" || msg.String() == "t":
		a.transactionView = false
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
		return a, nil

	case msg.String() == "q" || msg.String() == "ctrl+c":
		return a, tea.Quit

	case msg.String() == "h":
		a.help.ShowAll = !a.help.ShowAll
		return a, nil

	case msg.String() == "j" || msg.String() == "down":
		if a.transactionIdx < len(a.transactionItems)-1 {
			a.transactionIdx++
			a.adjustTransactionScroll()
			a.transactionDeps = nil
			return a, loadTxDepsCmd(a.transactionIdx, a.transactionItems[a.transactionIdx].Packages)
		}
		return a, nil

	case msg.String() == "k" || msg.String() == "up":
		if a.transactionIdx > 0 {
			a.transactionIdx--
			a.adjustTransactionScroll()
			a.transactionDeps = nil
			return a, loadTxDepsCmd(a.transactionIdx, a.transactionItems[a.transactionIdx].Packages)
		}
		return a, nil

	case msg.String() == "ctrl+d" || msg.String() == "pgdown":
		a.transactionIdx += a.txListHeight()
		if a.transactionIdx >= len(a.transactionItems) {
			a.transactionIdx = len(a.transactionItems) - 1
		}
		if a.transactionIdx < 0 {
			a.transactionIdx = 0
		}
		a.adjustTransactionScroll()
		a.transactionDeps = nil
		var cmd tea.Cmd
		if len(a.transactionItems) > 0 {
			cmd = loadTxDepsCmd(a.transactionIdx, a.transactionItems[a.transactionIdx].Packages)
		}
		return a, cmd

	case msg.String() == "ctrl+u" || msg.String() == "pgup":
		a.transactionIdx -= a.txListHeight()
		if a.transactionIdx < 0 {
			a.transactionIdx = 0
		}
		a.adjustTransactionScroll()
		a.transactionDeps = nil
		var cmd tea.Cmd
		if len(a.transactionItems) > 0 {
			cmd = loadTxDepsCmd(a.transactionIdx, a.transactionItems[a.transactionIdx].Packages)
		}
		return a, cmd

	case msg.String() == "z":
		if len(a.transactionItems) > 0 && a.transactionIdx < len(a.transactionItems) {
			tx := a.transactionItems[a.transactionIdx]
			if !tx.Success {
				a.status = ui.ErrorStyle.Render("Cannot undo a failed transaction.")
				return a, nil
			}
			undoOp := history.UndoOperation(tx.Operation)
			var cmds []tea.Cmd
			for _, pkg := range tx.Packages {
				switch undoOp {
				case history.OpRemove:
					cmds = append(cmds, removeCmd(pkg))
				case history.OpInstall:
					cmds = append(cmds, installCmd(pkg))
				}
			}
			a.pendingExecOp = string(undoOp)
			a.pendingExecPkgs = tx.Packages
			a.pendingExecCount = len(cmds)
			a.transactionView = false
			a.loading = true
			a.status = fmt.Sprintf("Undoing #%d (%s %d packages)...", tx.ID, undoOp, len(tx.Packages))
			return a, tea.Batch(cmds...)
		}
		return a, nil

	case msg.String() == "x":
		if len(a.transactionItems) > 0 && a.transactionIdx < len(a.transactionItems) {
			tx := a.transactionItems[a.transactionIdx]
			var cmds []tea.Cmd
			for _, pkg := range tx.Packages {
				switch tx.Operation {
				case history.OpInstall:
					cmds = append(cmds, installCmd(pkg))
				case history.OpRemove:
					cmds = append(cmds, removeCmd(pkg))
				case history.OpUpgrade:
					cmds = append(cmds, upgradeCmd(pkg))
				case history.OpUpgradeAll:
					cmds = append(cmds, upgradeAllCmd())
				}
			}
			a.pendingExecOp = string(tx.Operation)
			a.pendingExecPkgs = tx.Packages
			a.pendingExecCount = len(cmds)
			a.transactionView = false
			a.loading = true
			a.status = fmt.Sprintf("Redoing #%d (%s %d packages)...", tx.ID, tx.Operation, len(tx.Packages))
			return a, tea.Batch(cmds...)
		}
		return a, nil
	}

	return a, nil
}
