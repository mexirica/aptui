package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (a App) onRemoveConfirmKeypress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "right", "h", "l", "tab":
		a.removeCancelFocus = !a.removeCancelFocus
		return a, nil
	case "enter":
		if a.removeCancelFocus {
			return a.cancelRemoval()
		}
		return a.confirmRemoval()
	case "y":
		return a.confirmRemoval()
	case "n", "esc":
		return a.cancelRemoval()
	}
	return a, nil
}

func (a App) confirmRemoval() (tea.Model, tea.Cmd) {
	op := a.removeOp
	names := a.removeToProcess
	count := len(names)

	a.removeConfirm = false
	a.removeToProcess = nil
	a.removeOp = ""

	a.pendingExecOp = op
	a.pendingExecPkgs = names
	a.pendingExecCount = 1
	a.loading = true
	if op == "purge" {
		a.status = fmt.Sprintf("Purging %d packages...", count)
	} else {
		a.status = fmt.Sprintf("Removing %d packages...", count)
	}
	a.selected = make(map[string]bool)

	if op == "purge" {
		return a, purgeBatchCmd(names)
	}
	return a, removeBatchCmd(names)
}

func (a App) cancelRemoval() (tea.Model, tea.Cmd) {
	a.removeConfirm = false
	a.removeToProcess = nil
	a.removeOp = ""
	if len(a.selected) > 0 {
		a.status = fmt.Sprintf("%d selected ", len(a.selected))
	} else {
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
	}
	return a, nil
}
