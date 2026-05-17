package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/mexirica/aptui/internal/ui"
)

func (a App) onVersionKeypress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "esc":
		a.versionView = false
		a.versionItems = nil
		a.versionPkg = ""
		a.status = fmt.Sprintf("%d packages ", len(a.filtered))
		return a, nil
	case "j", "down":
		if a.versionIdx < len(a.versionItems)-1 {
			a.versionIdx++
			a.adjustVersionScroll()
		}
		return a, nil
	case "k", "up":
		if a.versionIdx > 0 {
			a.versionIdx--
			a.adjustVersionScroll()
		}
		return a, nil
	case "ctrl+d", "pgdown":
		a.versionIdx += a.versionListHeight()
		if a.versionIdx >= len(a.versionItems) {
			a.versionIdx = len(a.versionItems) - 1
		}
		if a.versionIdx < 0 {
			a.versionIdx = 0
		}
		a.adjustVersionScroll()
		return a, nil
	case "ctrl+u", "pgup":
		a.versionIdx -= a.versionListHeight()
		if a.versionIdx < 0 {
			a.versionIdx = 0
		}
		a.adjustVersionScroll()
		return a, nil
	case "enter":
		return a.installSelectedVersion()
	}
	return a, nil
}

func (a App) installSelectedVersion() (tea.Model, tea.Cmd) {
	if len(a.versionItems) == 0 || a.versionIdx >= len(a.versionItems) {
		return a, nil
	}
	ver := a.versionItems[a.versionIdx]
	if ver.Installed {
		a.status = fmt.Sprintf("Version %s is already installed.", ver.Version)
		return a, nil
	}

	// Determine current version for undo tracking
	currentVer := ""
	isDowngrade := false
	if idx, ok := a.pkgIndex[a.versionPkg]; ok {
		pkg := a.allPackages[idx]
		if pkg.Installed {
			currentVer = pkg.Version
			// apt-cache policy lists versions newest-first.
			// If we find currentVer before the target, the target is older → downgrade.
			for _, v := range a.versionItems {
				if v.Version == currentVer {
					isDowngrade = true
					break
				}
				if v.Version == ver.Version {
					isDowngrade = false
					break
				}
			}
		}
	}

	a.versionPrevVer = currentVer
	a.versionIsDowngrade = isDowngrade
	a.pendingExecOp = "install-version"
	a.pendingExecPkgs = []string{a.versionPkg}
	a.pendingExecVersion = ver.Version
	a.pendingExecCount = 1
	a.loading = true
	a.versionView = false

	opLabel := "Installing"
	if isDowngrade {
		opLabel = "Downgrading"
	}
	a.status = fmt.Sprintf("%s %s to version %s...", opLabel, a.versionPkg, ver.Version)

	return a, installVersionCmd(a.versionPkg, ver.Version, a.installRecommends, a.installSuggests)
}

func (a App) onVersionListLoaded(msg versionListMsg) (tea.Model, tea.Cmd) {
	a.loading = false
	if msg.err != nil {
		a.errlogStore.Log("version-list", fmt.Sprintf("%s: %v", msg.name, msg.err))
		a.status = ui.ErrorStyle.Render(fmt.Sprintf("Error loading versions: %v", msg.err))
		return a, nil
	}
	if len(msg.versions) == 0 {
		a.status = fmt.Sprintf("No versions found for %s in configured repositories.", msg.name)
		return a, nil
	}
	if len(msg.versions) == 1 {
		a.status = fmt.Sprintf("Only 1 version available for %s in configured repositories.", msg.name)
		return a, nil
	}
	a.versionView = true
	a.versionItems = msg.versions
	a.versionPkg = msg.name
	a.versionIdx = 0
	a.versionOffset = 0
	a.status = fmt.Sprintf("%d versions available for %s", len(msg.versions), msg.name)
	return a, nil
}

func (a *App) adjustVersionScroll() {
	maxVisible := a.versionListHeight()
	if a.versionIdx < a.versionOffset {
		a.versionOffset = a.versionIdx
	}
	if a.versionIdx >= a.versionOffset+maxVisible {
		a.versionOffset = a.versionIdx - maxVisible + 1
	}
}

func (a App) versionListHeight() int {
	h := a.height - 10
	if h < 5 {
		h = 5
	}
	return h
}
