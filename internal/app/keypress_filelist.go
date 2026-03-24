package app

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/mexirica/aptui/internal/ui"
)

func (a App) onFileListKeypress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	if !a.fileListActive {
		return a, nil, false
	}

	switch msg.String() {
	case "l", "esc":
		a.fileListActive = false
		a.fileListItems = nil
		a.fileListPkg = ""
		a.fileListIdx = 0
		a.fileListOffset = 0
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
		return a, nil, true
	case "J", "shift+down":
		if a.fileListIdx < len(a.fileListItems)-1 {
			a.fileListIdx++
			a.adjustFileListScroll()
		}
		return a, nil, true
	case "K", "shift+up":
		if a.fileListIdx > 0 {
			a.fileListIdx--
			a.adjustFileListScroll()
		}
		return a, nil, true
	case "shift+pgdown", "shift+ctrl+d":
		a.fileListIdx += a.fileListHeight()
		if a.fileListIdx >= len(a.fileListItems) {
			a.fileListIdx = len(a.fileListItems) - 1
		}
		if a.fileListIdx < 0 {
			a.fileListIdx = 0
		}
		a.adjustFileListScroll()
		return a, nil, true
	case "shift+pgup", "shift+ctrl+u":
		a.fileListIdx -= a.fileListHeight()
		if a.fileListIdx < 0 {
			a.fileListIdx = 0
		}
		a.adjustFileListScroll()
		return a, nil, true
	}

	return a, nil, false
}

func (a App) openFileList() (tea.Model, tea.Cmd) {
	if len(a.filtered) == 0 || a.selectedIdx >= len(a.filtered) {
		return a, nil
	}
	pkg := a.filtered[a.selectedIdx]
	if a.fileListActive && a.fileListPkg == pkg.Name {
		// Toggle off
		a.fileListActive = false
		a.fileListItems = nil
		a.fileListPkg = ""
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
		return a, nil
	}
	a.fileListActive = true
	a.fileListPkg = pkg.Name
	a.fileListItems = nil
	a.fileListIdx = 0
	a.fileListOffset = 0
	a.status = ui.WarningStyle.Render(fmt.Sprintf("Loading file list for %s...", pkg.Name))
	return a, loadFileListCmd(pkg.Name)
}

func (a App) onFileListLoaded(msg fileListLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		a.fileListActive = false
		a.errlogStore.Log("file-list", msg.err.Error())
		a.status = ui.ErrorStyle.Render(fmt.Sprintf("File list error: %v", msg.err))
		return a, clearStatusAfter(5 * time.Second)
	}
	a.fileListItems = msg.files
	a.fileListIdx = 0
	a.fileListOffset = 0
	a.status = fmt.Sprintf("%d files in %s | Shift+↑↓ scroll | l/esc close",
		len(msg.files), msg.name)
	return a, nil
}
