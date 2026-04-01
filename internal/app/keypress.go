package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mexirica/aptui/internal/ui"
)

func (a App) onKeypress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if a.importConfirm {
		return a.onImportConfirmKeypress(msg)
	}
	if a.removeConfirm {
		return a.onRemoveConfirmKeypress(msg)
	}
	if a.exportConfirm && msg.String() != "E" && msg.String() != "esc" {
		a.exportConfirm = false
	}
	if model, cmd, handled := a.dispatchErrorLog(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := a.onFileListKeypress(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := a.dispatchNavigation(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := a.dispatchSelection(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := a.dispatchPackageAction(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := a.switchTab(msg); handled {
		return model, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "h":
		return a.toggleHelp()
	case "/":
		return a.openSearch()
	case "esc":
		return a.clearFilterOrSearch()
	case "ctrl+r":
		return a.reloadPackages()
	case "t":
		return a.openTransactions()
	case "f":
		return a.openFetchMirrors()
	case "P":
		return a.openPPAView()
	case "D":
		return a.clearErrorLog()
	case "U":
		return a.runAptUpdate()
	case "E":
		return a.exportInstalledPackages()
	case "I":
		return a.importPackages()
	case "l":
		return a.openFileList()
	case "T":
		return a.toggleTheme()
	}

	return a, nil
}

func (a App) toggleHelp() (tea.Model, tea.Cmd) {
	a.help.ShowAll = !a.help.ShowAll
	return a, nil
}

func (a App) toggleTheme() (tea.Model, tea.Cmd) {
	a.hasDarkBG = !a.hasDarkBG
	a.themeForced = true
	ui.ApplyTheme(a.hasDarkBG)
	a.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	a.help.Styles.FullKey = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	a.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ui.ColorNormalText)
	a.help.Styles.FullDesc = lipgloss.NewStyle().Foreground(ui.ColorNormalText)
	a.help.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ui.ColorHelpSep)
	a.help.Styles.FullSeparator = lipgloss.NewStyle().Foreground(ui.ColorHelpSep)
	a.spinner.Style = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	return a, nil
}

func (a App) openSearch() (tea.Model, tea.Cmd) {
	a.searching = true
	a.filterQueryBeforeEdit = a.filterQuery
	cmd := a.searchInput.Focus()
	a.searchInput.SetValue(a.filterQuery)
	return a, cmd
}

func (a App) clearFilterOrSearch() (tea.Model, tea.Cmd) {
	if a.exportConfirm {
		a.exportConfirm = false
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
		return a, nil
	}
	if a.filterQuery == "" {
		return a, nil
	}
	a.filterQuery = ""
	a.applyFilter()
	a.selectedIdx = 0
	a.scrollOffset = 0
	a.status = fmt.Sprintf("%d packages ", len(a.filtered))
	return a, a.updateSelectionCmd()
}

func (a App) runAptUpdate() (tea.Model, tea.Cmd) {
	a.loading = true
	a.pendingExecOp = "update"
	a.pendingExecCount = 1
	a.status = "Running apt update..."
	return a, aptUpdateCmd()
}

func (a App) reloadPackages() (tea.Model, tea.Cmd) {
	a.loading = true
	a.filterQuery = ""
	a.status = "Reloading..."
	return a, reloadAllPackages
}

func (a App) openTransactions() (tea.Model, tea.Cmd) {
	a.transactionView = true
	a.transactionItems = a.transactionStore.All()
	a.transactionIdx = 0
	a.transactionOffset = 0
	a.transactionDeps = nil
	a.status = fmt.Sprintf("%d transactions | esc back | z undo | x redo ", len(a.transactionItems))
	var cmd tea.Cmd
	if len(a.transactionItems) > 0 {
		cmd = loadTransactionDepsCmd(0, a.transactionItems[0].Packages)
	}
	return a, cmd
}

func (a App) openFetchMirrors() (tea.Model, tea.Cmd) {
	a.fetchView = true
	a.fetchMirrors = nil
	a.fetchSelected = make(map[int]bool)
	a.fetchIdx = 0
	a.fetchOffset = 0
	a.fetchTesting = true
	a.loading = true
	a.status = "Detecting distro and fetching mirror list..."
	return a, tea.Batch(a.spinner.Tick, fetchMirrorListCmd())
}

func (a App) openPPAView() (tea.Model, tea.Cmd) {
	a.ppaView = true
	a.ppaItems = nil
	a.ppaIdx = 0
	a.ppaOffset = 0
	a.ppaAdding = false
	a.loading = true
	a.status = "Loading PPA repositories..."
	return a, tea.Batch(a.spinner.Tick, listPPAsCmd())
}
