package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mexirica/aptui/internal/apt"
	"github.com/mexirica/aptui/internal/errlog"
	"github.com/mexirica/aptui/internal/fetch"
	"github.com/mexirica/aptui/internal/filter"
	"github.com/mexirica/aptui/internal/history"
	"github.com/mexirica/aptui/internal/model"
	"github.com/mexirica/aptui/internal/ui"
)

func newTestApp() App {
	a := New()
	a.width = 120
	a.height = 40
	a.loading = false
	return a
}

func TestNewApp(t *testing.T) {
	a := New()
	if a.upgradableMap == nil {
		t.Error("upgradableMap should be initialized")
	}
	if a.selected == nil {
		t.Error("selected should be initialized")
	}
	if a.infoCache == nil {
		t.Error("infoCache should be initialized")
	}
	if !a.loading {
		t.Error("app should start in loading state")
	}
	if a.status != "Loading packages..." {
		t.Errorf("unexpected initial status: %s", a.status)
	}
	if a.transactionStore == nil {
		t.Error("transactionStore should be initialized")
	}
}

func TestApplyFilterAll(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true, Upgradable: true},
		{Name: "curl", Installed: false},
	}
	a.activeTab = tabAll
	a.filterQuery = ""
	a.applyFilter()

	if len(a.filtered) != 3 {
		t.Errorf("expected 3 packages on All tab, got %d", len(a.filtered))
	}
}

func TestApplyFilterInstalledTab(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true, Upgradable: true},
		{Name: "curl", Installed: false},
	}
	a.activeTab = tabInstalled
	a.filterQuery = ""
	a.applyFilter()

	if len(a.filtered) != 2 {
		t.Errorf("expected 2 installed packages, got %d", len(a.filtered))
	}
	for _, p := range a.filtered {
		if !p.Installed {
			t.Errorf("non-installed package in Installed tab: %s", p.Name)
		}
	}
}

func TestApplyFilterUpgradableTab(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true, Upgradable: true},
		{Name: "curl", Installed: false},
	}
	a.activeTab = tabUpgradable
	a.filterQuery = ""
	a.applyFilter()

	if len(a.filtered) != 1 {
		t.Errorf("expected 1 upgradable package, got %d", len(a.filtered))
	}
	if a.filtered[0].Name != "git" {
		t.Errorf("expected 'git', got '%s'", a.filtered[0].Name)
	}
}

func TestApplyFilterFuzzySearch(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true},
		{Name: "curl", Installed: false},
		{Name: "htop", Installed: true},
	}
	a.activeTab = tabAll
	a.filterQuery = "vim"
	a.applyFilter()

	if len(a.filtered) == 0 {
		t.Error("expected at least 1 result for 'vim'")
	}
	if a.filtered[0].Name != "vim" {
		t.Errorf("expected 'vim' as top result, got '%s'", a.filtered[0].Name)
	}
}

func TestApplyFilterResetsSelection(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}
	a.selectedIdx = 2
	a.scrollOffset = 1
	a.applyFilter()

	if a.selectedIdx != 0 {
		t.Errorf("expected selectedIdx reset to 0, got %d", a.selectedIdx)
	}
	if a.scrollOffset != 0 {
		t.Errorf("expected scrollOffset reset to 0, got %d", a.scrollOffset)
	}
}

func TestListHeight(t *testing.T) {
	a := newTestApp()
	h := a.packageListHeight()
	if h < 5 {
		t.Errorf("listHeight should be at least 5, got %d", h)
	}
}

func TestStackedDetailPanelHeight(t *testing.T) {
	a := newTestApp()
	a.sideBySide = false
	h := a.stackedDetailPanelHeight()
	if h < 5 {
		t.Errorf("expected stackedDetailPanelHeight >= 5, got %d", h)
	}
}

func TestAdjustScroll(t *testing.T) {
	a := newTestApp()
	a.allPackages = make([]model.Package, 100)
	a.filtered = a.allPackages

	// Scroll down past viewport
	a.selectedIdx = 50
	a.scrollOffset = 0
	a.adjustPackageScroll()
	if a.scrollOffset == 0 {
		t.Error("scrollOffset should have been adjusted for selectedIdx=50")
	}

	// Scroll back up
	a.selectedIdx = 0
	a.adjustPackageScroll()
	if a.scrollOffset != 0 {
		t.Errorf("scrollOffset should be 0 when selectedIdx=0, got %d", a.scrollOffset)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	a := newTestApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	app := m.(App)
	if app.width != 200 || app.height != 50 {
		t.Errorf("expected 200x50, got %dx%d", app.width, app.height)
	}
}

func TestToggleSelection(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}
	a.selectedIdx = 0

	// Toggle select
	m, _ := a.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	app := m.(App)
	if !app.selected["vim"] {
		t.Error("vim should be selected after space")
	}

	// Toggle deselect
	m, _ = app.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	app = m.(App)
	if app.selected["vim"] {
		t.Error("vim should be deselected after second space")
	}
}

func TestSelectAll(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}

	// Select all
	m, _ := a.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	app := m.(App)
	if len(app.selected) != 3 {
		t.Errorf("expected 3 selected, got %d", len(app.selected))
	}

	// Toggle again to deselect all
	m, _ = app.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	app = m.(App)
	if len(app.selected) != 0 {
		t.Errorf("expected 0 selected after toggle, got %d", len(app.selected))
	}
}

func TestNavigationDown(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}
	a.selectedIdx = 0

	m, _ := a.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	app := m.(App)
	if app.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 after j, got %d", app.selectedIdx)
	}
}

func TestNavigationUp(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}
	a.selectedIdx = 2

	m, _ := a.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	app := m.(App)
	if app.selectedIdx != 1 {
		t.Errorf("expected selectedIdx=1 after k, got %d", app.selectedIdx)
	}
}

func TestNavigationBounds(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim"}, {Name: "git"},
	}

	// Can't go above 0
	a.selectedIdx = 0
	m, _ := a.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	app := m.(App)
	if app.selectedIdx != 0 {
		t.Errorf("should stay at 0, got %d", app.selectedIdx)
	}

	// Can't go below len-1
	a.selectedIdx = 1
	m, _ = a.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	app = m.(App)
	if app.selectedIdx != 1 {
		t.Errorf("should stay at 1, got %d", app.selectedIdx)
	}
}

func TestTabSwitching(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true, Upgradable: true},
		{Name: "curl", Installed: false},
	}
	a.applyFilter()

	if a.activeTab != tabAll {
		t.Errorf("expected tabAll initially, got %d", a.activeTab)
	}

	// Press tab -> tabInstalled
	m, _ := a.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app := m.(App)
	if app.activeTab != tabInstalled {
		t.Errorf("expected tabInstalled, got %d", app.activeTab)
	}

	// Press tab again -> tabUpgradable
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app = m.(App)
	if app.activeTab != tabUpgradable {
		t.Errorf("expected tabUpgradable, got %d", app.activeTab)
	}

	// Press tab again -> tabCleanup
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app = m.(App)
	if app.activeTab != tabCleanup {
		t.Errorf("expected tabCleanup, got %d", app.activeTab)
	}

	// Press tab again -> tabErrorLog
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app = m.(App)
	if app.activeTab != tabErrorLog {
		t.Errorf("expected tabErrorLog, got %d", app.activeTab)
	}

	// Press tab again -> tabTransactions
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app = m.(App)
	if app.activeTab != tabTransactions {
		t.Errorf("expected tabTransactions, got %d", app.activeTab)
	}

	// Press tab again -> tabRepos
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app = m.(App)
	if app.activeTab != tabRepos {
		t.Errorf("expected tabRepos, got %d", app.activeTab)
	}

	// Press tab again -> back to tabAll
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app = m.(App)
	if app.activeTab != tabAll {
		t.Errorf("expected tabAll, got %d", app.activeTab)
	}
}

func TestTransactionViewToggle(t *testing.T) {
	a := newTestApp()

	// Tab through to tabTransactions (5 tabs forward)
	var m tea.Model = a
	for i := 0; i < 5; i++ {
		m, _ = m.(App).Update(tea.KeyPressMsg{Code: tea.KeyTab})
	}
	app := m.(App)
	if app.activeTab != tabTransactions {
		t.Errorf("expected tabTransactions after 5 tabs, got %d", app.activeTab)
	}
}

func TestSearchMode(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}
	a.applyFilter()

	// Enter search mode
	m, _ := a.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	app := m.(App)
	if !app.searching {
		t.Error("expected searching=true after '/'")
	}

	// Cancel search with esc
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	app = m.(App)
	if app.searching {
		t.Error("expected searching=false after esc")
	}
}

func TestHelpToggle(t *testing.T) {
	a := newTestApp()
	if a.help.ShowAll {
		t.Error("help should start collapsed")
	}

	m, _ := a.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	app := m.(App)
	if !app.help.ShowAll {
		t.Error("expected help.ShowAll=true after 'h'")
	}

	m, _ = app.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	app = m.(App)
	if app.help.ShowAll {
		t.Error("expected help.ShowAll=false after second 'h'")
	}
}

func TestViewNotEmpty(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Version: "8.2"},
	}
	a.applyFilter()

	v := a.View()
	if v.Content == "" {
		t.Error("View should not be empty")
	}
}

func TestViewLoadingState(t *testing.T) {
	a := newTestApp()
	a.width = 0

	v := a.View()
	if v.Content != fmt.Sprintf("Updating and loading packages %s", a.spinner.View()) {
		t.Errorf("expected 'Updating and loading packages ...' when width=0, got %q", v.Content)
	}
}

func TestAllPackagesMsg(t *testing.T) {
	a := newTestApp()

	msg := allPackagesMsg{
		bulkInfo: map[string]apt.PackageInfo{
			"vim":  {Version: "8.2", Section: "editors", Architecture: "amd64"},
			"git":  {Version: "2.40", Section: "vcs", Architecture: "amd64"},
			"curl": {Version: "7.88", Section: "web", Architecture: "amd64"},
			"htop": {Version: "3.2", Section: "utils", Architecture: "amd64"},
		},
		installed:  []model.Package{{Name: "vim", Installed: true, Version: "8.2"}},
		upgradable: []model.Package{{Name: "vim", Installed: true, Upgradable: true, NewVersion: "9.0"}},
		err:        nil,
	}

	m, _ := a.Update(msg)
	app := m.(App)

	if app.loading {
		t.Error("loading should be false after allPackagesMsg")
	}
	if len(app.allPackages) != 4 {
		t.Errorf("expected 4 packages, got %d", len(app.allPackages))
	}
	if len(app.upgradableMap) != 1 {
		t.Errorf("expected 1 upgradable, got %d", len(app.upgradableMap))
	}
	if !app.allNamesLoaded {
		t.Error("allNamesLoaded should be true after allPackagesMsg")
	}
	if app.installedCount != 1 {
		t.Errorf("expected installedCount=1, got %d", app.installedCount)
	}
}

func TestAllPackagesMsgError(t *testing.T) {
	a := newTestApp()

	msg := allPackagesMsg{
		err: fmt.Errorf("test error"),
	}

	m, _ := a.Update(msg)
	app := m.(App)

	if app.loading {
		t.Error("loading should be false after error")
	}
	if app.status == "" {
		t.Error("status should contain error message")
	}
}

func TestExecFinishedMsg(t *testing.T) {
	a := newTestApp()
	a.pendingExecOp = "install"
	a.pendingExecPkgs = []string{"vim"}
	a.pendingExecCount = 1
	a.loading = true

	msg := execFinishedMsg{op: "install", name: "vim", err: nil}
	m, _ := a.Update(msg)
	app := m.(App)

	if app.pendingExecCount != 0 {
		t.Errorf("pendingExecCount should be 0, got %d", app.pendingExecCount)
	}
}

func TestOptimisticUpdateInstall(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: false},
		{Name: "git", Installed: true},
		{Name: "curl", Installed: false},
	}
	a.rebuildIndex()
	a.installedCount = 1
	a.pendingExecOp = "install"
	a.pendingExecPkgs = []string{"vim", "curl"}
	a.pendingExecCount = 1
	a.loading = true

	msg := execFinishedMsg{op: "install", name: "vim curl", err: nil}
	m, _ := a.Update(msg)
	app := m.(App)

	if !app.allPackages[0].Installed {
		t.Error("vim should be marked as installed after optimistic update")
	}
	if !app.allPackages[2].Installed {
		t.Error("curl should be marked as installed after optimistic update")
	}
	if app.installedCount != 3 {
		t.Errorf("expected installedCount=3, got %d", app.installedCount)
	}
}

func TestOptimisticUpdateRemove(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true},
	}
	a.rebuildIndex()
	a.installedCount = 2
	a.pendingExecOp = "remove"
	a.pendingExecPkgs = []string{"vim"}
	a.pendingExecCount = 1
	a.loading = true

	msg := execFinishedMsg{op: "remove", name: "vim", err: nil}
	m, _ := a.Update(msg)
	app := m.(App)

	if app.allPackages[0].Installed {
		t.Error("vim should not be installed after remove")
	}
	if !app.allPackages[1].Installed {
		t.Error("git should still be installed")
	}
	if app.installedCount != 1 {
		t.Errorf("expected installedCount=1, got %d", app.installedCount)
	}
}

func TestOptimisticUpdateUpgrade(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Upgradable: true, Version: "8.2", NewVersion: "9.0"},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{
		"vim": {Name: "vim", NewVersion: "9.0"},
	}
	a.installedCount = 1
	a.pendingExecOp = "upgrade"
	a.pendingExecPkgs = []string{"vim"}
	a.pendingExecCount = 1
	a.loading = true

	msg := execFinishedMsg{op: "upgrade", name: "vim", err: nil}
	m, _ := a.Update(msg)
	app := m.(App)

	if app.allPackages[0].Upgradable {
		t.Error("vim should not be upgradable after upgrade")
	}
	if app.allPackages[0].Version != "9.0" {
		t.Errorf("expected version '9.0', got '%s'", app.allPackages[0].Version)
	}
	if len(app.upgradableMap) != 0 {
		t.Errorf("expected empty upgradableMap, got %d entries", len(app.upgradableMap))
	}
}

func TestOptimisticUpdateSkippedOnFailure(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: false},
	}
	a.rebuildIndex()
	a.installedCount = 0
	a.pendingExecOp = "install"
	a.pendingExecPkgs = []string{"vim"}
	a.pendingExecCount = 1
	a.loading = true

	msg := execFinishedMsg{op: "install", name: "vim", err: fmt.Errorf("permission denied")}
	m, _ := a.Update(msg)
	app := m.(App)

	if app.allPackages[0].Installed {
		t.Error("vim should NOT be installed after failed install")
	}
	if app.installedCount != 0 {
		t.Errorf("expected installedCount=0 after failure, got %d", app.installedCount)
	}
}

func TestOptimisticUpdateCleanupAll(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, NewVersion: "9.0", Upgradable: true},
		{Name: "git", Installed: true},
		{Name: "curl", Installed: false},
	}
	a.rebuildIndex()
	a.installedCount = 2
	a.upgradableMap = map[string]model.Package{
		"vim": {Name: "vim", NewVersion: "9.0"},
	}
	a.autoremovable = []string{"vim", "git"}
	a.autoremovableSet = map[string]bool{"vim": true, "git": true}

	a.applyOptimisticUpdate("cleanup-all", []string{"vim", "git"})

	if a.allPackages[0].Installed {
		t.Error("vim should not be installed after cleanup-all")
	}
	if a.allPackages[1].Installed {
		t.Error("git should not be installed after cleanup-all")
	}
	if a.allPackages[0].NewVersion != "" {
		t.Errorf("vim NewVersion should be cleared, got %q", a.allPackages[0].NewVersion)
	}
	if a.allPackages[0].Upgradable {
		t.Error("vim should not be upgradable after cleanup-all")
	}
	if a.installedCount != 0 {
		t.Errorf("expected installedCount=0, got %d", a.installedCount)
	}
	if a.autoremovable != nil {
		t.Error("autoremovable should be nil after cleanup-all")
	}
	if len(a.autoremovableSet) != 0 {
		t.Errorf("autoremovableSet should be empty, got %d", len(a.autoremovableSet))
	}
}

func TestRebuildIndex(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}
	a.rebuildIndex()

	if len(a.pkgIndex) != 3 {
		t.Errorf("expected 3 entries in pkgIndex, got %d", len(a.pkgIndex))
	}
	if idx, ok := a.pkgIndex["git"]; !ok || idx != 1 {
		t.Errorf("expected pkgIndex[git]=1, got %d (ok=%v)", idx, ok)
	}
}

func TestInstallAlreadyInstalled(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim", Installed: true},
	}
	a.selectedIdx = 0

	m, _ := a.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	app := m.(App)

	if app.loading {
		t.Error("should not be loading when package already installed")
	}
	if app.status == "" {
		t.Error("should show already installed message")
	}
}

func TestRemoveNotInstalled(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim", Installed: false},
	}
	a.selectedIdx = 0

	m, _ := a.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	app := m.(App)

	if app.loading {
		t.Error("should not be loading when package not installed")
	}
}

func TestRemoveConfirmation(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim", Installed: true},
	}
	a.selectedIdx = 0

	// Press 'r'
	m, _ := a.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	app := m.(App)

	if !app.removeConfirm {
		t.Error("expected removeConfirm to be true after pressing 'r'")
	}
	if app.loading {
		t.Error("should not be loading yet, waiting for confirmation")
	}
	if !app.removeCancelFocus {
		t.Error("expected default focus on Cancel")
	}

	// Press Enter (should cancel since Cancel is focused)
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app = m.(App)
	if app.removeConfirm {
		t.Error("expected removeConfirm to be false after pressing Enter on Cancel")
	}

	// Press 'r' again
	m, _ = app.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	app = m.(App)

	// Switch focus to Confirm
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	app = m.(App)
	if app.removeCancelFocus {
		t.Error("expected focus to switch to Confirm")
	}

	// Press Enter (should confirm)
	m, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	app = m.(App)
	if app.removeConfirm {
		t.Error("expected removeConfirm to be false after confirmation")
	}
	if !app.loading {
		t.Error("expected app to be loading after confirmation")
	}
}

func TestPurgeConfirmation(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim", Installed: true},
	}
	a.selectedIdx = 0

	// Press 'p'
	m, _ := a.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	app := m.(App)

	if !app.removeConfirm {
		t.Error("expected removeConfirm to be true after pressing 'p'")
	}
	if app.removeOp != "purge" {
		t.Errorf("expected removeOp to be 'purge', got %q", app.removeOp)
	}

	// Confirm via 'y'
	m, _ = app.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	app = m.(App)
	if app.removeConfirm {
		t.Error("expected removeConfirm to be false after confirmation")
	}
	if !app.loading {
		t.Error("expected app to be loading after confirmation")
	}
	if app.pendingExecOp != "purge" {
		t.Errorf("expected pendingExecOp to be 'purge', got %q", app.pendingExecOp)
	}
}

func TestMultipleRemoveConfirmation(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true},
	}
	a.selected = map[string]bool{"vim": true, "git": true}

	// Press 'r'
	m, _ := a.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	app := m.(App)

	if !app.removeConfirm {
		t.Error("expected removeConfirm to be true after pressing 'r'")
	}
	if len(app.removeToProcess) != 2 {
		t.Errorf("expected 2 packages to remove, got %d", len(app.removeToProcess))
	}

	// Cancel via 'n'
	m, _ = app.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	app = m.(App)
	if app.removeConfirm {
		t.Error("expected removeConfirm to be false after cancel")
	}
	if app.loading {
		t.Error("should not be loading after cancel")
	}
}

func TestUpgradeNotUpgradable(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "vim", Installed: true, Upgradable: false},
	}
	a.selectedIdx = 0

	m, _ := a.Update(tea.KeyPressMsg{Code: 'u', Text: "u"})
	app := m.(App)

	if app.loading {
		t.Error("should not be loading when package not upgradable")
	}
}

func TestFetchViewToggle(t *testing.T) {
	a := newTestApp()

	// Open fetch view
	m, _ := a.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})
	app := m.(App)
	if !app.fetchView {
		t.Error("expected fetchView=true after 'f'")
	}
}

func TestAdjustFetchScroll(t *testing.T) {
	a := newTestApp()
	a.fetchIdx = 50
	a.fetchOffset = 0
	a.adjustMirrorScroll()
	if a.fetchOffset == 0 {
		t.Error("fetchOffset should adjust when fetchIdx is past viewport")
	}
}

func TestAdjustTransactionScroll(t *testing.T) {
	a := newTestApp()
	a.transactionIdx = 50
	a.transactionOffset = 0
	a.adjustTransactionScroll()
	if a.transactionOffset == 0 {
		t.Error("transactionOffset should adjust when transactionIdx is past viewport")
	}
}

func TestTabDefsOrder(t *testing.T) {
	if len(tabDefs) != 7 {
		t.Fatalf("expected 7 tab definitions, got %d", len(tabDefs))
	}
	expected := []struct {
		kind tabKind
		name string
	}{
		{tabAll, "All"},
		{tabInstalled, "Installed"},
		{tabUpgradable, "Upgradable"},
		{tabCleanup, "Cleanup"},
		{tabErrorLog, "Errors"},
		{tabTransactions, "Transactions"},
		{tabRepos, "Repos"},
	}
	for i, e := range expected {
		if tabDefs[i].kind != e.kind {
			t.Errorf("tabDefs[%d].kind = %d, want %d", i, tabDefs[i].kind, e.kind)
		}
		if tabDefs[i].name != e.name {
			t.Errorf("tabDefs[%d].name = %q, want %q", i, tabDefs[i].name, e.name)
		}
		if tabDefs[i].label == "" {
			t.Errorf("tabDefs[%d].label should not be empty", i)
		}
	}
}

func TestTabStyleActive(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabAll

	got := a.tabStyle(tabDefs[0]).Render("X")
	want := ui.TabActiveStyle.Render("X")
	if got != want {
		t.Errorf("expected TabActiveStyle for the active tab, got %q vs %q", got, want)
	}
}

func TestTabStyleInactive(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabAll

	got := a.tabStyle(tabDefs[1]).Render("X")
	want := ui.TabInactiveStyle.Render("X")
	if got != want {
		t.Errorf("expected TabInactiveStyle for an inactive tab, got %q vs %q", got, want)
	}
}

func TestTabStyleNotify(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabAll
	a.upgradableMap = map[string]model.Package{"vim": {Name: "vim"}}

	got := a.tabStyle(tabDefs[2]).Render("X")
	want := ui.TabNotifyStyle.Render("X")
	if got != want {
		t.Errorf("expected TabNotifyStyle for upgradable tab when upgradable packages exist, got %q vs %q", got, want)
	}
}

func TestTabStyleUpgradableActiveNoNotify(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabUpgradable
	a.upgradableMap = map[string]model.Package{"vim": {Name: "vim"}}

	got := a.tabStyle(tabDefs[2]).Render("X")
	want := ui.TabActiveStyle.Render("X")
	if got != want {
		t.Errorf("expected TabActiveStyle for active upgradable tab, got %q vs %q", got, want)
	}
}

func TestActivateTabSetsStatus(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true},
	}
	a.activeTab = tabInstalled
	a.activateTab()

	if !strings.Contains(a.status, "2 packages") {
		t.Errorf("expected status to mention package count, got %q", a.status)
	}
}

func TestActivateTabResetsSelection(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim"}, {Name: "git"}, {Name: "curl"},
	}
	a.selectedIdx = 2
	a.scrollOffset = 1
	a.activeTab = tabAll
	a.activateTab()

	if a.selectedIdx != 0 {
		t.Errorf("expected selectedIdx=0 after activateTab, got %d", a.selectedIdx)
	}
	if a.scrollOffset != 0 {
		t.Errorf("expected scrollOffset=0 after activateTab, got %d", a.scrollOffset)
	}
}

func TestRenderTabBarContainsAllLabels(t *testing.T) {
	a := newTestApp()
	bar := a.renderTabBar()

	for _, td := range tabDefs {
		if !strings.Contains(bar, strings.TrimSpace(td.label)) {
			t.Errorf("renderTabBar missing label %q", td.label)
		}
	}
}

func TestSwitchTabBackward(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim"}}

	if a.activeTab != tabAll {
		t.Fatal("expected initial tab to be tabAll")
	}

	m, _, handled := a.switchTab(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if !handled {
		t.Fatal("expected switchTab to handle shift+tab")
	}
	app := m.(App)
	if app.activeTab != tabRepos {
		t.Errorf("expected tabRepos after shift+tab from tabAll, got %d", app.activeTab)
	}

	m, _, _ = app.switchTab(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	app = m.(App)
	if app.activeTab != tabTransactions {
		t.Errorf("expected tabTransactions after shift+tab from tabRepos, got %d", app.activeTab)
	}
}

func TestSearchBarYPositive(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
	}
	a.applyFilter()
	y := a.searchBarY()
	if y <= 0 || y >= a.height {
		t.Errorf("searchBarY=%d should be between 1 and %d", y, a.height-1)
	}
}

func TestUpgradeAllNoUpgradable(t *testing.T) {
	a := newTestApp()
	a.upgradableMap = map[string]model.Package{}

	m, _ := a.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	app := m.(App)

	if app.loading {
		t.Error("should not be loading when no upgradable packages")
	}
	if !strings.Contains(app.status, "No upgradable") {
		t.Errorf("expected 'No upgradable' in status, got %q", app.status)
	}
}

func TestUpgradeAllSetsState(t *testing.T) {
	a := newTestApp()
	a.upgradableMap = map[string]model.Package{
		"vim": {Name: "vim", Upgradable: true},
		"git": {Name: "git", Upgradable: true},
	}

	m, _ := a.upgradeAllPackages()
	app := m.(App)

	if !app.loading {
		t.Error("should be loading after upgradeAll")
	}
	if app.pendingExecOp != "upgrade-all" {
		t.Errorf("expected pendingExecOp='upgrade-all', got %q", app.pendingExecOp)
	}
	if len(app.pendingExecPkgs) != 2 {
		t.Errorf("expected 2 pending packages, got %d", len(app.pendingExecPkgs))
	}
	if !strings.Contains(app.status, "2 packages") {
		t.Errorf("expected status to mention 2 packages, got %q", app.status)
	}
}

func TestOnTabClickSameTab(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim"}}
	a.activeTab = tabAll

	m, cmd := a.onTabClick(0)
	app := m.(App)

	if app.activeTab != tabAll {
		t.Error("expected tab to stay on tabAll when clicking active tab")
	}
	if cmd != nil {
		t.Error("expected nil cmd when clicking already-active tab")
	}
}

func TestHoldListMsg(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true},
		{Name: "curl", Installed: false},
	}
	a.rebuildIndex()
	a.applyFilter()

	msg := holdListMsg{names: []string{"vim"}, err: nil}
	m, _ := a.Update(msg)
	app := m.(App)

	if !app.heldSet["vim"] {
		t.Error("vim should be in heldSet")
	}
	if app.heldSet["git"] {
		t.Error("git should not be in heldSet")
	}
	if !app.allPackages[0].Held {
		t.Error("vim should have Held=true")
	}
	if app.allPackages[1].Held {
		t.Error("git should have Held=false")
	}
}

func TestHoldListMsgError(t *testing.T) {
	a := newTestApp()
	msg := holdListMsg{err: fmt.Errorf("apt-mark error")}
	m, _ := a.Update(msg)
	app := m.(App)

	if len(app.heldSet) != 0 {
		t.Error("heldSet should be empty on error")
	}
}

func TestHoldSelectedPackage(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
	}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	a.heldSet = make(map[string]bool)

	m, cmd := a.holdSelectedPackages()
	app := m.(App)

	if !app.loading {
		t.Error("should be loading after hold")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestUnholdSelectedPackage(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Held: true},
	}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	a.heldSet = map[string]bool{"vim": true}

	m, cmd := a.holdSelectedPackages()
	app := m.(App)

	if !app.loading {
		t.Error("should be loading after unhold")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestHoldNotInstalledPackage(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: false},
	}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	a.heldSet = make(map[string]bool)

	m, _ := a.holdSelectedPackages()
	app := m.(App)

	if app.loading {
		t.Error("should not be loading for non-installed package")
	}
	if !strings.Contains(app.status, "not installed") {
		t.Errorf("expected 'not installed' status, got '%s'", app.status)
	}
}

func TestHeldFlagPropagatedOnLoad(t *testing.T) {
	a := newTestApp()
	a.heldSet = map[string]bool{"vim": true}

	msg := allPackagesMsg{
		bulkInfo:   map[string]apt.PackageInfo{"vim": {Version: "8.2"}},
		installed:  []model.Package{{Name: "vim", Installed: true, Version: "8.2"}},
		upgradable: nil,
		err:        nil,
	}

	m, _ := a.Update(msg)
	app := m.(App)

	idx := app.pkgIndex["vim"]
	if !app.allPackages[idx].Held {
		t.Error("vim should have Held=true after allPackagesMsg when in heldSet")
	}
}

func TestUpgradeBlockedForHeldPackage(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Upgradable: true, Held: true, Version: "8.2", NewVersion: "9.0"},
	}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	a.heldSet = map[string]bool{"vim": true}

	m, cmd := a.upgradeSelectedPackages()
	app := m.(App)

	if app.loading {
		t.Error("should not start upgrade for held package")
	}
	if !strings.Contains(app.status, "held") {
		t.Errorf("expected status to mention 'held', got '%s'", app.status)
	}
	if cmd != nil {
		t.Error("expected nil cmd when upgrade is blocked")
	}
}

func TestUpgradeAllSkipsHeldPackages(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Upgradable: true, Held: true, Version: "8.2", NewVersion: "9.0"},
		{Name: "git", Installed: true, Upgradable: true, Version: "2.34", NewVersion: "2.40"},
	}
	a.rebuildIndex()
	a.applyFilter()
	a.upgradableMap = map[string]model.Package{
		"vim": {Name: "vim", NewVersion: "9.0"},
		"git": {Name: "git", NewVersion: "2.40"},
	}
	a.heldSet = map[string]bool{"vim": true}

	m, _ := a.upgradeAllPackages()
	app := m.(App)

	if len(app.pendingExecPkgs) != 1 {
		t.Errorf("expected 1 package to upgrade, got %d", len(app.pendingExecPkgs))
	}
	if len(app.pendingExecPkgs) == 1 && app.pendingExecPkgs[0] != "git" {
		t.Errorf("expected 'git' to upgrade, got '%s'", app.pendingExecPkgs[0])
	}
}

func TestRemoveEssentialPackageBlocked(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "base-files", Installed: true, Essential: true},
	}
	a.essentialSet = map[string]bool{"base-files": true}
	a.selectedIdx = 0

	m, _ := a.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	app := m.(App)

	if app.loading {
		t.Error("should not be loading when trying to remove essential package")
	}
	if !strings.Contains(app.status, "essential") {
		t.Errorf("status should mention essential, got %q", app.status)
	}
}

func TestPurgeEssentialPackageBlocked(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "base-files", Installed: true, Essential: true},
	}
	a.essentialSet = map[string]bool{"base-files": true}
	a.selectedIdx = 0

	m, _ := a.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	app := m.(App)

	if app.loading {
		t.Error("should not be loading when trying to purge essential package")
	}
	if !strings.Contains(app.status, "essential") {
		t.Errorf("status should mention essential, got %q", app.status)
	}
}

func TestRemoveBatchSkipsEssential(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "base-files", Installed: true, Essential: true},
		{Name: "vim", Installed: true},
	}
	a.essentialSet = map[string]bool{"base-files": true}
	a.selected = map[string]bool{"base-files": true, "vim": true}

	m, cmd := a.removeSelectedPackages()
	app := m.(App)

	if !app.removeConfirm {
		t.Error("expected removeConfirm to be true")
	}
	if len(app.removeToProcess) != 1 {
		t.Fatalf("expected 1 package to process, got %d", len(app.removeToProcess))
	}
	if app.removeToProcess[0] != "vim" {
		t.Errorf("expected 'vim', got '%s'", app.removeToProcess[0])
	}
	if cmd != nil {
		t.Error("expected no command before confirmation")
	}

	// Now confirm
	m, cmd = app.confirmRemoval()
	app = m.(App)

	if cmd == nil {
		t.Error("expected command for non-essential package after confirmation")
	}
	if len(app.pendingExecPkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(app.pendingExecPkgs))
	}
	if app.pendingExecPkgs[0] != "vim" {
		t.Errorf("expected 'vim', got '%s'", app.pendingExecPkgs[0])
	}
}

func TestRemoveAllEssentialBlocked(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{
		{Name: "base-files", Installed: true, Essential: true},
		{Name: "dpkg", Installed: true, Essential: true},
	}
	a.essentialSet = map[string]bool{"base-files": true, "dpkg": true}
	a.selected = map[string]bool{"base-files": true, "dpkg": true}

	m, cmd := a.removeSelectedPackages()
	app := m.(App)

	if cmd != nil {
		t.Error("should not return command when all selected are essential")
	}
	if !strings.Contains(app.status, "essential") {
		t.Errorf("status should mention essential, got %q", app.status)
	}
}

func TestEssentialFlagSetFromBulkInfo(t *testing.T) {
	a := newTestApp()

	msg := allPackagesMsg{
		bulkInfo: map[string]apt.PackageInfo{
			"base-files": {Version: "12", Essential: true},
			"vim":        {Version: "8.2"},
		},
		installed: []model.Package{{Name: "base-files", Installed: true}},
	}

	m, _ := a.Update(msg)
	app := m.(App)

	if !app.essentialSet["base-files"] {
		t.Error("base-files should be in essentialSet")
	}
	if app.essentialSet["vim"] {
		t.Error("vim should not be in essentialSet")
	}

	var baseFiles model.Package
	for _, p := range app.allPackages {
		if p.Name == "base-files" {
			baseFiles = p
			break
		}
	}
	if !baseFiles.Essential {
		t.Error("base-files package should have Essential=true")
	}
}

// ──────────────────────────────────────────────────────────
// Side-by-side layout tests
// ──────────────────────────────────────────────────────────

func TestToggleLayoutKey(t *testing.T) {
	a := newTestApp()
	a.sideBySide = false

	m, _ := a.Update(tea.KeyPressMsg{Code: 'L', Text: "L"})
	app := m.(App)
	if !app.sideBySide {
		t.Error("expected sideBySide=true after pressing L on wide terminal")
	}

	m, _ = app.Update(tea.KeyPressMsg{Code: 'L', Text: "L"})
	app = m.(App)
	if app.sideBySide {
		t.Error("expected sideBySide=false after pressing L again")
	}
}

func TestToggleLayoutNarrowTerminal(t *testing.T) {
	a := newTestApp()
	a.width = 80
	a.sideBySide = false

	m, _ := a.Update(tea.KeyPressMsg{Code: 'L', Text: "L"})
	app := m.(App)
	if app.sideBySide {
		t.Error("should not enable sideBySide when width < sideMinWidth")
	}
}

func TestWindowSizeDisablesSideBySide(t *testing.T) {
	a := newTestApp()
	a.sideBySide = true

	m, _ := a.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	app := m.(App)
	if app.sideBySide {
		t.Error("sideBySide should be auto-disabled when window shrinks below sideMinWidth")
	}
}

func TestSideListWidthProportion(t *testing.T) {
	a := newTestApp()
	a.width = 200

	leftW := a.sideListWidth()
	rightW := a.sideDetailWidth()
	if leftW+rightW != a.width {
		t.Errorf("left(%d)+right(%d) should equal width(%d)", leftW, rightW, a.width)
	}
	if leftW != 200*sideSplitPct/100 {
		t.Errorf("expected left=%d, got %d", 200*sideSplitPct/100, leftW)
	}
}

func TestSideMainPanelHeightMinimum(t *testing.T) {
	a := newTestApp()
	a.height = 10
	a.sideBySide = true

	h := a.sideMainPanelHeight()
	if h < 7 {
		t.Errorf("sideMainPanelHeight should be at least 7, got %d", h)
	}
}

func TestPackageListHeightSideBySide(t *testing.T) {
	a := newTestApp()
	a.sideBySide = true

	h := a.packageListHeight()
	if h < 5 {
		t.Errorf("packageListHeight in sideBySide should be at least 5, got %d", h)
	}
}

func TestSearchBarYSideBySide(t *testing.T) {
	a := newTestApp()
	a.sideBySide = true
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.applyFilter()

	y := a.searchBarY()
	// Info panel is now above the main panels, directly after tabBar + gap.
	expected := 3
	if y != expected {
		t.Errorf("searchBarY in sideBySide=%d, expected %d", y, expected)
	}
}

func TestNewAppSideBySideDefault(t *testing.T) {
	a := New()
	if !a.sideBySide {
		t.Error("new app should default to sideBySide=true")
	}
}

// ──────────────────────────────────────────────────────────
// renderTitledPanel tests
// ──────────────────────────────────────────────────────────

func TestRenderTitledPanelContainsTitle(t *testing.T) {
	panel := renderTitledPanel("MyTitle", "", "content", 40, 5)
	if !strings.Contains(panel, "MyTitle") {
		t.Error("panel should contain the title text")
	}
}

func TestRenderTitledPanelContainsRightText(t *testing.T) {
	panel := renderTitledPanel("Title", "3/10", "content", 40, 5)
	if !strings.Contains(panel, "3/10") {
		t.Error("panel should contain the right text")
	}
}

func TestRenderTitledPanelBorders(t *testing.T) {
	panel := renderTitledPanel("Title", "", "hello", 30, 4)
	if !strings.Contains(panel, "╭") || !strings.Contains(panel, "╮") {
		t.Error("panel should have top corners ╭ and ╮")
	}
	if !strings.Contains(panel, "╰") || !strings.Contains(panel, "╯") {
		t.Error("panel should have bottom corners ╰ and ╯")
	}
	if !strings.Contains(panel, "│") {
		t.Error("panel should have side borders │")
	}
}

func TestRenderTitledPanelHeight(t *testing.T) {
	panel := renderTitledPanel("T", "", "a\nb\nc", 30, 6)
	lines := strings.Split(panel, "\n")
	if len(lines) != 6 {
		t.Errorf("expected 6 lines (height), got %d", len(lines))
	}
}

func TestRenderTitledPanelTruncatesContent(t *testing.T) {
	// 10 lines of content in a panel with height=5 (3 inner lines)
	content := strings.Repeat("line\n", 10)
	panel := renderTitledPanel("T", "", content, 30, 5)
	lines := strings.Split(panel, "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
}

// ──────────────────────────────────────────────────────────
// Tab tests for Transactions and Repos
// ──────────────────────────────────────────────────────────

func TestActivateTabTransactions(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabTransactions

	a.activateTab()
	if a.transactionIdx != 0 {
		t.Errorf("expected transactionIdx=0, got %d", a.transactionIdx)
	}
	if a.transactionOffset != 0 {
		t.Errorf("expected transactionOffset=0, got %d", a.transactionOffset)
	}
	if a.status != "" {
		t.Errorf("expected empty status for transactions tab, got %q", a.status)
	}
}

func TestActivateTabRepos(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabRepos

	a.activateTab()
	if !a.loading {
		t.Error("expected loading=true after activating repos tab")
	}
	if a.status != "Loading repositories..." {
		t.Errorf("unexpected status: %q", a.status)
	}
	if a.ppaIdx != 0 {
		t.Errorf("expected ppaIdx=0, got %d", a.ppaIdx)
	}
}

func TestTransactionTabKeypress(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabTransactions

	// Tab should switch tabs even in transaction tab
	m, _ := a.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app := m.(App)
	if app.activeTab != tabRepos {
		t.Errorf("expected tabRepos after tab from transactions, got %d", app.activeTab)
	}
}

func TestReposTabKeypress(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabRepos

	// Tab should switch tabs even in repos tab
	m, _ := a.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	app := m.(App)
	if app.activeTab != tabAll {
		t.Errorf("expected tabAll after tab from repos, got %d", app.activeTab)
	}
}

// ──────────────────────────────────────────────────────────
// Removed keybinding tests
// ──────────────────────────────────────────────────────────

func TestTKeyDoesNotOpenTransactionView(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim"}}
	a.applyFilter()

	m, _ := a.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	app := m.(App)
	// t key should not switch to transactions tab from main view
	if app.activeTab != tabAll {
		t.Errorf("expected to remain on tabAll after 't', got %d", app.activeTab)
	}
}

func TestPKeyDoesNotOpenPPAView(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim"}}
	a.applyFilter()

	m, _ := a.Update(tea.KeyPressMsg{Code: 'P', Text: "P"})
	app := m.(App)
	if app.activeTab != tabAll {
		t.Errorf("expected to remain on tabAll after 'P', got %d", app.activeTab)
	}
}

func TestFullHelpNoTransactionOrPPAKeys(t *testing.T) {
	keys := model.Keys
	groups := keys.FullHelp()

	for _, group := range groups {
		for _, b := range group {
			help := b.Help()
			if help.Key == "t" && help.Desc == "transactions" {
				t.Error("FullHelp should not contain 't/transactions' binding")
			}
			if help.Key == "P" && help.Desc == "repos" {
				t.Error("FullHelp should not contain 'P/repos' binding")
			}
			if help.Key == "z" && help.Desc == "undo" {
				t.Error("FullHelp should not contain 'z/undo' binding")
			}
			if help.Key == "x" && help.Desc == "redo" {
				t.Error("FullHelp should not contain 'x/redo' binding")
			}
		}
	}
}

// ──────────────────────────────────────────────────────────
// View rendering tests
// ──────────────────────────────────────────────────────────

func TestRenderSideBySideNotEmpty(t *testing.T) {
	a := newTestApp()
	a.sideBySide = true
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: false},
	}
	a.applyFilter()

	tabBar := a.renderTabBar()
	out := a.renderSideBySide(a.width, tabBar)
	if out == "" {
		t.Error("renderSideBySide should not return empty string")
	}
	if !strings.Contains(out, "Package List") {
		t.Error("renderSideBySide should contain 'Package List' panel title")
	}
	if !strings.Contains(out, "Package Detail") {
		t.Error("renderSideBySide should contain 'Package Detail' panel title")
	}
	if !strings.Contains(out, "Search") {
		t.Error("renderSideBySide should contain 'Search' panel")
	}
	if !strings.Contains(out, "Status") {
		t.Error("renderSideBySide should contain 'Status' panel")
	}
	if !strings.Contains(out, "Keys") {
		t.Error("renderSideBySide should contain 'Keys' panel")
	}
}

func TestRenderStackedNotEmpty(t *testing.T) {
	a := newTestApp()
	a.sideBySide = false
	a.width = 80
	a.height = 24
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: false},
	}
	a.applyFilter()

	tabBar := a.renderTabBar()
	out := a.renderStacked(a.width, tabBar)
	if out == "" {
		t.Error("renderStacked should not return empty string")
	}
	if !strings.Contains(out, "Package List") {
		t.Error("renderStacked should contain 'Package List' panel title")
	}
	if !strings.Contains(out, "Package Detail") {
		t.Error("renderStacked should contain 'Package Detail' panel title")
	}
	if !strings.Contains(out, "Search / Status") {
		t.Error("renderStacked should contain 'Search / Status' panel title")
	}
	if !strings.Contains(out, "Keys") {
		t.Error("renderStacked should contain 'Keys' panel")
	}
}

func TestStackedListPanelHeight(t *testing.T) {
	a := newTestApp()
	a.sideBySide = false
	a.width = 80
	a.height = 24

	h := a.stackedListPanelHeight()
	if h < 7 {
		t.Errorf("stackedListPanelHeight should be at least 7, got %d", h)
	}
}

func TestViewSideBySideMode(t *testing.T) {
	a := newTestApp()
	a.sideBySide = true
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.applyFilter()

	view := a.View()
	if view.Content == "" {
		t.Error("View in sideBySide mode should not be empty")
	}
}

func TestRenderTransactionViewNewDesign(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabTransactions
	a.activateTab()

	tabBar := a.renderTabBar()
	out := a.renderTransactionView(a.width, tabBar)
	if !strings.Contains(out, "Transactions") {
		t.Error("transaction view should contain titled panel 'Transactions'")
	}
	if !strings.Contains(out, "Details") {
		t.Error("transaction view should contain titled panel 'Details'")
	}
	// Should contain tab bar
	for _, td := range tabDefs {
		if !strings.Contains(out, strings.TrimSpace(td.label)) {
			t.Errorf("transaction view should contain tab label %q", td.label)
		}
	}
}

func TestRenderPPAViewNewDesign(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabRepos
	a.loading = false
	a.ppaItems = nil

	tabBar := a.renderTabBar()
	out := a.renderPPAView(a.width, tabBar)
	if !strings.Contains(out, "Repositories") {
		t.Error("PPA view should contain titled panel 'Repositories'")
	}
	if !strings.Contains(out, "Repo Detail") {
		t.Error("PPA view should contain titled panel 'Repo Detail'")
	}
	// Should contain tab bar
	for _, td := range tabDefs {
		if !strings.Contains(out, strings.TrimSpace(td.label)) {
			t.Errorf("PPA view should contain tab label %q", td.label)
		}
	}
}

func TestViewDispatchTransactionTab(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabTransactions
	a.activateTab()

	view := a.View()
	if view.Content == "" {
		t.Error("View with tabTransactions should not be empty")
	}
	if !strings.Contains(view.Content, "Transactions") {
		t.Error("View with tabTransactions should render Transactions panel")
	}
}

func TestViewDispatchReposTab(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabRepos
	a.loading = false

	view := a.View()
	if view.Content == "" {
		t.Error("View with tabRepos should not be empty")
	}
	if !strings.Contains(view.Content, "Repositories") {
		t.Error("View with tabRepos should render Repositories panel")
	}
}

// ──────────────────────────────────────────────────────────
// Side-by-side detail rendering tests
// ──────────────────────────────────────────────────────────

func TestRenderSideBasicDetailContainsFields(t *testing.T) {
	a := newTestApp()
	pkg := model.Package{
		Name:         "vim",
		Version:      "8.2",
		Installed:    true,
		Section:      "editors",
		Architecture: "amd64",
		Description:  "Vi IMproved",
	}
	out := a.renderPanelBasicDetail(pkg, 60)
	for _, field := range []string{"Name", "Version", "Status", "Section", "Architecture", "Description"} {
		if !strings.Contains(out, field) {
			t.Errorf("renderPanelBasicDetail should contain field %q", field)
		}
	}
	if !strings.Contains(out, "Installed") {
		t.Error("renderPanelBasicDetail should show 'Installed' status for installed package")
	}
}

func TestRenderSideBasicDetailNotInstalled(t *testing.T) {
	a := newTestApp()
	pkg := model.Package{
		Name:    "curl",
		Version: "7.0",
	}
	out := a.renderPanelBasicDetail(pkg, 60)
	if !strings.Contains(out, "Not installed") {
		t.Error("renderPanelBasicDetail should show 'Not installed' for non-installed package")
	}
}

func TestRenderSideBasicDetailUpgradable(t *testing.T) {
	a := newTestApp()
	pkg := model.Package{
		Name:       "git",
		Version:    "2.30",
		Installed:  true,
		Upgradable: true,
		NewVersion: "2.40",
	}
	out := a.renderPanelBasicDetail(pkg, 60)
	if !strings.Contains(out, "Upgrade available") {
		t.Error("renderPanelBasicDetail should show 'Upgrade available' for upgradable package")
	}
	if !strings.Contains(out, "New Version") {
		t.Error("renderPanelBasicDetail should show 'New Version' for upgradable package")
	}
}

func TestScrollDetailContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		maxLines   int
		offset     int
		wantOffset int
		wantMax    int
		wantLines  int
	}{
		{
			name:       "no scrolling needed",
			content:    "line1\nline2\nline3\n",
			maxLines:   10,
			offset:     0,
			wantOffset: 0,
			wantMax:    0,
			wantLines:  3,
		},
		{
			name:       "scrolling at top",
			content:    "l1\nl2\nl3\nl4\nl5\nl6\nl7\nl8\nl9\nl10\n",
			maxLines:   3,
			offset:     0,
			wantOffset: 0,
			wantMax:    7,
			wantLines:  3,
		},
		{
			name:       "scrolling in middle",
			content:    "l1\nl2\nl3\nl4\nl5\nl6\nl7\nl8\nl9\nl10\n",
			maxLines:   3,
			offset:     3,
			wantOffset: 3,
			wantMax:    7,
			wantLines:  3,
		},
		{
			name:       "offset clamped to max",
			content:    "l1\nl2\nl3\nl4\nl5\n",
			maxLines:   3,
			offset:     100,
			wantOffset: 2,
			wantMax:    2,
			wantLines:  3,
		},
		{
			name:       "single line",
			content:    "only line\n",
			maxLines:   5,
			offset:     0,
			wantOffset: 0,
			wantMax:    0,
			wantLines:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, offset, maxOff := scrollDetailContent(tt.content, tt.maxLines, tt.offset)
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
			if maxOff != tt.wantMax {
				t.Errorf("maxOffset = %d, want %d", maxOff, tt.wantMax)
			}
			lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
			if len(lines) != tt.wantLines {
				t.Errorf("got %d lines, want %d", len(lines), tt.wantLines)
			}
		})
	}
}

func TestFriendlyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{name: "nil error", err: nil, expected: "unknown error"},
		{name: "simple error", err: fmt.Errorf("something failed"), expected: "something failed"},
		{name: "wrapped error", err: fmt.Errorf("wrap: %w", fmt.Errorf("inner")), expected: "wrap: inner"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := friendlyError(tt.err)
			if got != tt.expected {
				t.Errorf("friendlyError() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSortFieldEmpty(t *testing.T) {
	tests := []struct {
		name   string
		pkg    model.Package
		col    filter.SortColumn
		expect bool
	}{
		{name: "empty name", pkg: model.Package{}, col: filter.SortName, expect: true},
		{name: "has name", pkg: model.Package{Name: "vim"}, col: filter.SortName, expect: false},
		{name: "empty version", pkg: model.Package{}, col: filter.SortVersion, expect: true},
		{name: "has version", pkg: model.Package{Version: "1.0"}, col: filter.SortVersion, expect: false},
		{name: "has new version", pkg: model.Package{NewVersion: "2.0"}, col: filter.SortVersion, expect: false},
		{name: "empty size", pkg: model.Package{}, col: filter.SortSize, expect: true},
		{name: "dash size", pkg: model.Package{Size: "-"}, col: filter.SortSize, expect: true},
		{name: "has size", pkg: model.Package{Size: "5 MB"}, col: filter.SortSize, expect: false},
		{name: "empty section", pkg: model.Package{}, col: filter.SortSection, expect: true},
		{name: "has section", pkg: model.Package{Section: "utils"}, col: filter.SortSection, expect: false},
		{name: "empty arch", pkg: model.Package{}, col: filter.SortArchitecture, expect: true},
		{name: "has arch", pkg: model.Package{Architecture: "amd64"}, col: filter.SortArchitecture, expect: false},
		{name: "SortNone always false", pkg: model.Package{}, col: filter.SortNone, expect: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortFieldEmpty(tt.pkg, tt.col)
			if got != tt.expect {
				t.Errorf("sortFieldEmpty(%v, %d) = %v, want %v", tt.pkg.Name, tt.col, got, tt.expect)
			}
		})
	}
}

func TestEffectiveSortInfo(t *testing.T) {
	tests := []struct {
		name        string
		sortColumn  filter.SortColumn
		sortDesc    bool
		filterQuery string
		wantCol     filter.SortColumn
		wantDesc    bool
	}{
		{
			name:        "explicit sort column",
			sortColumn:  filter.SortName,
			sortDesc:    true,
			filterQuery: "",
			wantCol:     filter.SortName,
			wantDesc:    true,
		},
		{
			name:        "from filter query",
			sortColumn:  filter.SortNone,
			filterQuery: "order:size:desc",
			wantCol:     filter.SortSize,
			wantDesc:    true,
		},
		{
			name:        "explicit overrides filter",
			sortColumn:  filter.SortVersion,
			sortDesc:    false,
			filterQuery: "order:name:desc",
			wantCol:     filter.SortVersion,
			wantDesc:    false,
		},
		{
			name:        "no sort",
			sortColumn:  filter.SortNone,
			filterQuery: "",
			wantCol:     filter.SortNone,
			wantDesc:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.sortColumn = tt.sortColumn
			a.sortDesc = tt.sortDesc
			a.filterQuery = tt.filterQuery
			info := a.effectiveSortInfo()
			if info.Column != tt.wantCol {
				t.Errorf("Column = %d, want %d", info.Column, tt.wantCol)
			}
			if info.Desc != tt.wantDesc {
				t.Errorf("Desc = %v, want %v", info.Desc, tt.wantDesc)
			}
		})
	}
}

// ── Update handler tests ─────────────────────────────────────────────

func TestOnSilentUpdateDone_NewPackagesAndUpgradable(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Version: "1.0"},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{}
	a.infoCache = map[string]apt.PackageInfo{
		"git": {Version: "2.0", Size: "5000 kB", Section: "vcs"},
	}
	a.pinnedSet = map[string]bool{}
	a.statusLock = time.Time{} // expired

	msg := silentUpdateDoneMsg{
		names:      []string{"git"},
		upgradable: []model.Package{{Name: "vim", NewVersion: "1.1"}},
	}
	m, _ := a.onSilentUpdateDone(msg)
	app := m.(App)

	if len(app.allPackages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(app.allPackages))
	}
	git := app.allPackages[app.pkgIndex["git"]]
	if git.NewVersion != "2.0" {
		t.Errorf("git NewVersion = %q, want %q", git.NewVersion, "2.0")
	}
	vim := app.allPackages[app.pkgIndex["vim"]]
	if !vim.Upgradable || vim.NewVersion != "1.1" {
		t.Errorf("vim should be upgradable with NewVersion=1.1, got Upgradable=%v, NewVersion=%q", vim.Upgradable, vim.NewVersion)
	}
}

func TestOnSilentUpdateDone_NoChange(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{"vim": {Name: "vim", NewVersion: "2.0"}}

	msg := silentUpdateDoneMsg{
		names:      nil,
		upgradable: []model.Package{{Name: "vim", NewVersion: "2.0"}},
	}
	m, cmd := a.onSilentUpdateDone(msg)
	_ = m.(App)
	if cmd != nil {
		t.Error("should return nil cmd when nothing changed")
	}
}

func TestOnSearchResultLoaded_Success(t *testing.T) {
	a := newTestApp()
	a.loading = true
	a.filterQuery = "vim"
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Version: "1.0", Size: "1000 kB"},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{
		"vim-gtk": {Name: "vim-gtk", NewVersion: "2.0", SecurityUpdate: true},
	}
	a.infoCache = map[string]apt.PackageInfo{}

	msg := searchResultMsg{
		pkgs: []model.Package{
			{Name: "vim"},
			{Name: "vim-gtk"},
		},
	}
	m, _ := a.onSearchResultLoaded(msg)
	app := m.(App)

	if app.loading {
		t.Error("loading should be false after search result")
	}
	if len(app.filtered) != 2 {
		t.Fatalf("expected 2 results, got %d", len(app.filtered))
	}
	if !app.filtered[0].Installed {
		t.Error("vim should be Installed after enrichment")
	}
	if app.filtered[0].Version != "1.0" {
		t.Errorf("vim Version = %q, want %q", app.filtered[0].Version, "1.0")
	}
	if !app.filtered[1].Upgradable {
		t.Error("vim-gtk should be Upgradable after enrichment")
	}
	if !strings.Contains(app.status, "2 results") {
		t.Errorf("status = %q, want contains '2 results'", app.status)
	}
}

func TestOnSearchResultLoaded_Error(t *testing.T) {
	a := newTestApp()
	a.loading = true
	msg := searchResultMsg{err: errors.New("search failed")}
	m, _ := a.onSearchResultLoaded(msg)
	app := m.(App)

	if app.loading {
		t.Error("loading should be false")
	}
	if !strings.Contains(app.status, "Error") {
		t.Errorf("status should contain Error, got %q", app.status)
	}
}

func TestOnPackageDetailLoaded_Success(t *testing.T) {
	a := newTestApp()
	a.infoCache = map[string]apt.PackageInfo{}
	a.filtered = []model.Package{{Name: "vim"}}
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()

	info := "Package: vim\nVersion: 8.2\nInstalled-Size: 3000\nSection: editors\nArchitecture: amd64\nDescription: Vi IMproved"
	msg := detailLoadedMsg{name: "vim", info: info}
	m, _ := a.onPackageDetailLoaded(msg)
	app := m.(App)

	if app.detailInfo != info {
		t.Error("detailInfo should be set")
	}
	if app.detailName != "vim" {
		t.Errorf("detailName = %q, want %q", app.detailName, "vim")
	}
	if _, ok := app.infoCache["vim"]; !ok {
		t.Error("infoCache should contain vim")
	}
}

func TestOnPackageDetailLoaded_Error(t *testing.T) {
	a := newTestApp()
	a.infoCache = map[string]apt.PackageInfo{}
	msg := detailLoadedMsg{name: "vim", err: errors.New("not found")}
	m, _ := a.onPackageDetailLoaded(msg)
	app := m.(App)

	if !strings.Contains(app.detailInfo, "Error") {
		t.Errorf("detailInfo should contain Error, got %q", app.detailInfo)
	}
}

func TestOnDepsLoaded(t *testing.T) {
	a := newTestApp()
	a.transactionIdx = 2
	tests := []struct {
		name     string
		txIdx    int
		deps     []string
		wantDeps bool
	}{
		{"matching index", 2, []string{"libc6", "libgcc"}, true},
		{"non-matching index", 1, []string{"libc6"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.transactionDeps = nil
			msg := depsLoadedMsg{txIdx: tt.txIdx, deps: tt.deps}
			m, _ := a.onDepsLoaded(msg)
			app := m.(App)
			if tt.wantDeps && len(app.transactionDeps) == 0 {
				t.Error("expected deps to be set")
			}
			if !tt.wantDeps && len(app.transactionDeps) > 0 {
				t.Error("expected deps to remain nil")
			}
		})
	}
}

func TestOnAutoremovableLoaded(t *testing.T) {
	tests := []struct {
		name      string
		msg       autoremovableMsg
		activeTab tabKind
		wantNames int
	}{
		{
			name:      "success with names",
			msg:       autoremovableMsg{names: []string{"pkg1", "pkg2"}, err: nil},
			activeTab: tabAll,
			wantNames: 2,
		},
		{
			name:      "error clears list",
			msg:       autoremovableMsg{err: errors.New("fail")},
			activeTab: tabCleanup,
			wantNames: 0,
		},
		{
			name:      "success on cleanup tab",
			msg:       autoremovableMsg{names: []string{"pkg1"}, err: nil},
			activeTab: tabCleanup,
			wantNames: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.activeTab = tt.activeTab
			a.allPackages = []model.Package{{Name: "pkg1"}, {Name: "pkg2"}}
			a.rebuildIndex()
			a.statusLock = time.Time{}

			m, _ := a.onAutoremovableLoaded(tt.msg)
			app := m.(App)
			if len(app.autoremovable) != tt.wantNames {
				t.Errorf("autoremovable len = %d, want %d", len(app.autoremovable), tt.wantNames)
			}
		})
	}
}

func TestOnHeldListLoaded(t *testing.T) {
	tests := []struct {
		name    string
		msg     holdListMsg
		wantSet int
	}{
		{"success", holdListMsg{names: []string{"vim", "git"}}, 2},
		{"error", holdListMsg{err: errors.New("fail")}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.allPackages = []model.Package{{Name: "vim"}, {Name: "git"}}
			a.rebuildIndex()
			m, _ := a.onHeldListLoaded(tt.msg)
			app := m.(App)
			if len(app.heldSet) != tt.wantSet {
				t.Errorf("heldSet len = %d, want %d", len(app.heldSet), tt.wantSet)
			}
		})
	}
}

func TestOnHoldFinished(t *testing.T) {
	tests := []struct {
		name        string
		holdPending int
		err         error
		wantLoading bool
	}{
		{"last hold success", 1, nil, false},
		{"last hold error", 1, errors.New("fail"), false},
		{"pending hold", 2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.holdPending = tt.holdPending
			a.loading = true
			a.holdFailed = false
			msg := holdFinishedMsg{op: "hold", err: tt.err}
			m, _ := a.onHoldFinished(msg)
			app := m.(App)
			if app.loading != tt.wantLoading {
				t.Errorf("loading = %v, want %v", app.loading, tt.wantLoading)
			}
		})
	}
}

func TestOnPPAListLoaded(t *testing.T) {
	tests := []struct {
		name    string
		msg     ppaListMsg
		ppaIdx  int
		wantLen int
		wantIdx int
		wantErr bool
	}{
		{
			name:    "success",
			msg:     ppaListMsg{ppas: []apt.PPA{{Name: "ppa1"}, {Name: "ppa2"}}},
			ppaIdx:  0,
			wantLen: 2,
			wantIdx: 0,
		},
		{
			name:    "error",
			msg:     ppaListMsg{err: errors.New("fail")},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "clamp index",
			msg:     ppaListMsg{ppas: []apt.PPA{{Name: "ppa1"}}},
			ppaIdx:  5,
			wantLen: 1,
			wantIdx: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.loading = true
			a.ppaIdx = tt.ppaIdx
			m, _ := a.onPPAListLoaded(tt.msg)
			app := m.(App)
			if len(app.ppaItems) != tt.wantLen {
				t.Errorf("ppaItems len = %d, want %d", len(app.ppaItems), tt.wantLen)
			}
			if tt.wantLen > 0 && app.ppaIdx != tt.wantIdx {
				t.Errorf("ppaIdx = %d, want %d", app.ppaIdx, tt.wantIdx)
			}
		})
	}
}

func TestOnPPAToggled(t *testing.T) {
	tests := []struct {
		name    string
		msg     ppaToggleMsg
		wantErr bool
	}{
		{"success", ppaToggleMsg{name: "ppa:user/repo", action: "enabled"}, false},
		{"error", ppaToggleMsg{name: "ppa:user/repo", err: errors.New("fail")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.loading = true
			m, _ := a.onPPAToggled(tt.msg)
			app := m.(App)
			if tt.wantErr && !strings.Contains(app.status, "Error") {
				t.Errorf("status should contain Error, got %q", app.status)
			}
			if !tt.wantErr && !strings.Contains(app.status, "✔") {
				t.Errorf("status should contain ✔, got %q", app.status)
			}
		})
	}
}

func TestOnMirrorApplyResult(t *testing.T) {
	tests := []struct {
		name    string
		msg     fetchApplyMsg
		wantErr bool
	}{
		{"success", fetchApplyMsg{}, false},
		{"error", fetchApplyMsg{err: errors.New("write error")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.fetchView = true
			m, _ := a.onMirrorApplyResult(tt.msg)
			app := m.(App)
			if app.fetchView {
				t.Error("fetchView should be false after apply")
			}
			if tt.wantErr && !strings.Contains(app.status, "Error") {
				t.Errorf("status should contain Error, got %q", app.status)
			}
			if !tt.wantErr && !strings.Contains(app.status, "Mirrors saved") {
				t.Errorf("status should contain success, got %q", app.status)
			}
		})
	}
}

func TestOnMirrorListLoaded_Error(t *testing.T) {
	a := newTestApp()
	a.fetchView = true
	a.loading = true
	msg := fetchMirrorsMsg{err: errors.New("no mirrors")}
	m, _ := a.onMirrorListLoaded(msg)
	app := m.(App)
	if app.fetchView {
		t.Error("fetchView should be false on error")
	}
	if app.loading {
		t.Error("loading should be false on error")
	}
}

func TestOnMirrorTestResult_Progress(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = []fetch.Mirror{
		{URL: "http://m1.example.com", Status: ""},
		{URL: "http://m2.example.com", Status: ""},
	}
	a.fetchTested = 0
	a.fetchTotal = 2

	msg := fetchTestResultMsg{
		result: fetch.TestResult{Index: 0, Latency: 100 * time.Millisecond},
		done:   false,
	}
	m, _ := a.onMirrorTestResult(msg)
	app := m.(App)
	if app.fetchMirrors[0].Status != "ok" {
		t.Errorf("mirror 0 status = %q, want %q", app.fetchMirrors[0].Status, "ok")
	}
	if app.fetchTested != 1 {
		t.Errorf("fetchTested = %d, want 1", app.fetchTested)
	}
}

func TestOnMirrorTestResult_Done(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = []fetch.Mirror{
		{URL: "http://m1.example.com", Latency: 50 * time.Millisecond},
	}
	a.fetchTesting = true
	a.loading = true

	msg := fetchTestResultMsg{done: true}
	m, _ := a.onMirrorTestResult(msg)
	app := m.(App)
	if app.fetchTesting {
		t.Error("fetchTesting should be false when done")
	}
}

func TestOnMirrorTestResult_Error(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = []fetch.Mirror{
		{URL: "http://m1.example.com"},
	}
	a.fetchTested = 0
	a.fetchTotal = 1
	msg := fetchTestResultMsg{
		result: fetch.TestResult{Index: 0, Err: errors.New("timeout")},
		done:   false,
	}
	m, _ := a.onMirrorTestResult(msg)
	app := m.(App)
	if app.fetchMirrors[0].Status != "error" {
		t.Errorf("mirror 0 status = %q, want %q", app.fetchMirrors[0].Status, "error")
	}
}

func TestOnExecFinished_Success(t *testing.T) {
	tests := []struct {
		name   string
		op     string
		wantIn string
	}{
		{"update", "update", "apt update completed"},
		{"cleanup-all", "cleanup-all", "Cleanup completed"},
		{"ppa-add", "ppa-add", "PPA"},
		{"ppa-remove", "ppa-remove", "PPA"},
		{"install", "install", "install vim completed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.pendingExecCount = 1
			a.pendingExecOp = tt.op
			a.allPackages = []model.Package{{Name: "vim", Installed: true}}
			a.rebuildIndex()
			a.fileListCache = map[string][]string{}
			msg := execFinishedMsg{op: tt.op, name: "vim"}
			m, _ := a.onExecFinished(msg)
			app := m.(App)
			if !strings.Contains(app.status, tt.wantIn) {
				t.Errorf("status = %q, want contains %q", app.status, tt.wantIn)
			}
		})
	}
}

func TestOnExecFinished_Error(t *testing.T) {
	a := newTestApp()
	a.pendingExecCount = 1
	a.pendingExecOp = "install"
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.fileListCache = map[string][]string{}
	msg := execFinishedMsg{op: "install", name: "vim", err: errors.New("permission denied")}
	m, _ := a.onExecFinished(msg)
	app := m.(App)
	if !strings.Contains(app.status, "Error") {
		t.Errorf("status should contain Error, got %q", app.status)
	}
}

func TestOnExecFinished_PendingBatch(t *testing.T) {
	a := newTestApp()
	a.pendingExecCount = 2
	a.pendingExecOp = "install"
	msg := execFinishedMsg{op: "install", name: "vim"}
	m, cmd := a.onExecFinished(msg)
	app := m.(App)
	if app.pendingExecCount != 1 {
		t.Errorf("pendingExecCount = %d, want 1", app.pendingExecCount)
	}
	if cmd != nil {
		t.Error("should return nil cmd while still pending")
	}
}

func TestOnAllPackagesLoaded_Error(t *testing.T) {
	a := newTestApp()
	a.loading = true
	msg := allPackagesMsg{err: errors.New("dpkg error")}
	m, _ := a.onAllPackagesLoaded(msg)
	app := m.(App)
	if app.loading {
		t.Error("loading should be false on error")
	}
	if !strings.Contains(app.status, "Error") {
		t.Errorf("status should contain Error, got %q", app.status)
	}
}

func TestOnAllPackagesLoaded_Success(t *testing.T) {
	a := newTestApp()
	a.loading = true
	a.heldSet = map[string]bool{"vim": true}
	a.pinnedSet = map[string]bool{"git": true}
	msg := allPackagesMsg{
		installed: []model.Package{
			{Name: "vim", Installed: true, Version: "1.0"},
			{Name: "curl", Installed: true, Version: "7.0"},
		},
		upgradable: []model.Package{
			{Name: "vim", NewVersion: "1.1"},
		},
		bulkInfo: map[string]apt.PackageInfo{
			"vim": {Version: "1.1", Size: "1000 kB", Section: "editors", Essential: true},
			"git": {Version: "2.0", Size: "5000 kB", Section: "vcs", Description: "git scm"},
		},
		manualSet: map[string]bool{"vim": true},
	}
	m, _ := a.onAllPackagesLoaded(msg)
	app := m.(App)
	if app.loading {
		t.Error("loading should be false")
	}
	idx := app.pkgIndex["vim"]
	vim := app.allPackages[idx]
	if !vim.Upgradable {
		t.Error("vim should be upgradable")
	}
	if !vim.Held {
		t.Error("vim should be held")
	}
	if !vim.Essential {
		t.Error("vim should be essential")
	}
	if !vim.ManuallyInstalled {
		t.Error("vim should be manually installed")
	}
	gitIdx := app.pkgIndex["git"]
	git := app.allPackages[gitIdx]
	if git.Installed {
		t.Error("git should not be installed")
	}
	if !git.Pinned {
		t.Error("git should be pinned")
	}
}

func TestOnExportFinished(t *testing.T) {
	tests := []struct {
		name    string
		msg     exportFinishedMsg
		wantErr bool
	}{
		{"success", exportFinishedMsg{path: "/tmp/packages.txt"}, false},
		{"error", exportFinishedMsg{err: errors.New("write error")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.loading = true
			m, _ := a.onExportFinished(tt.msg)
			app := m.(App)
			if tt.wantErr && !strings.Contains(app.status, "failed") {
				t.Errorf("status should contain failed, got %q", app.status)
			}
			if !tt.wantErr && !strings.Contains(app.status, "Exported") {
				t.Errorf("status should contain Exported, got %q", app.status)
			}
		})
	}
}

func TestOnImportFinished(t *testing.T) {
	tests := []struct {
		name    string
		msg     importFinishedMsg
		wantErr bool
	}{
		{"success", importFinishedMsg{names: []string{"vim", "git"}, path: "/tmp/packages.txt"}, false},
		{"error", importFinishedMsg{err: errors.New("parse error")}, true},
		{"empty", importFinishedMsg{names: []string{}, path: "/tmp/empty.txt"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.loading = true
			a.importingPath = true
			m, _ := a.onImportFinished(tt.msg)
			app := m.(App)
			if tt.wantErr && !strings.Contains(app.status, "failed") {
				t.Errorf("status should contain failed, got %q", app.status)
			}
		})
	}
}

func TestClearStatusMsg(t *testing.T) {
	a := newTestApp()
	a.pendingStatus = "5 packages"
	a.loading = false
	m, _ := a.Update(clearStatusMsg{})
	app := m.(App)
	if app.status != "5 packages" {
		t.Errorf("status = %q, want %q", app.status, "5 packages")
	}
	if app.pendingStatus != "" {
		t.Errorf("pendingStatus should be cleared, got %q", app.pendingStatus)
	}
}

func TestClearStatusMsg_WhileLoading(t *testing.T) {
	a := newTestApp()
	a.pendingStatus = "5 packages"
	a.loading = true
	m, _ := a.Update(clearStatusMsg{})
	app := m.(App)
	if app.status == "5 packages" {
		t.Error("should not apply pendingStatus while loading")
	}
}

func TestWindowSizeMsgSideBySide(t *testing.T) {
	a := newTestApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	app := m.(App)
	if app.width != 200 || app.height != 50 {
		t.Errorf("size = %dx%d, want 200x50", app.width, app.height)
	}
	if !app.sideBySide {
		t.Error("sideBySide should be true for width >= 120")
	}
}

func TestWindowSizeMsgNarrow(t *testing.T) {
	a := newTestApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	app := m.(App)
	if app.sideBySide {
		t.Error("sideBySide should be false for width < 120")
	}
}

// ── Helper function tests ────────────────────────────────────────────

func TestScrollDetailContentEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		maxLines   int
		offset     int
		wantLines  int
		wantOffset int
		wantMax    int
	}{
		{
			name:       "no scroll needed",
			content:    "line1\nline2\nline3\n",
			maxLines:   5,
			offset:     0,
			wantLines:  3,
			wantOffset: 0,
			wantMax:    0,
		},
		{
			name:       "scroll needed",
			content:    "a\nb\nc\nd\ne\nf\n",
			maxLines:   3,
			offset:     2,
			wantLines:  3,
			wantOffset: 2,
			wantMax:    3,
		},
		{
			name:       "offset clamped to max",
			content:    "a\nb\nc\n",
			maxLines:   2,
			offset:     10,
			wantLines:  2,
			wantOffset: 1,
			wantMax:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, offset, maxScroll := scrollDetailContent(tt.content, tt.maxLines, tt.offset)
			// Count non-empty lines in result
			lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
			if len(lines) > tt.wantLines {
				t.Errorf("got %d lines, want <= %d", len(lines), tt.wantLines)
			}
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
			if maxScroll != tt.wantMax {
				t.Errorf("maxScroll = %d, want %d", maxScroll, tt.wantMax)
			}
		})
	}
}

func TestTabStyle(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabAll
	a.upgradableMap = map[string]model.Package{"vim": {}}

	tests := []struct {
		name string
		tab  tabDef
	}{
		{"active tab", tabDefs[0]},        // tabAll - active
		{"upgradable notify", tabDefs[2]}, // tabUpgradable - has upgradable
		{"inactive tab", tabDefs[1]},      // tabInstalled - inactive
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := a.tabStyle(tt.tab)
			// Just ensure no panic and returns a valid style
			_ = style.Render("test")
		})
	}
}

func TestTabLabels(t *testing.T) {
	a := newTestApp()
	a.width = 200
	labels := a.tabLabels()
	if len(labels) != len(tabDefs) {
		t.Errorf("tabLabels len = %d, want %d", len(labels), len(tabDefs))
	}
}

func TestTabLabels_Narrow(t *testing.T) {
	a := newTestApp()
	a.width = 40 // very narrow
	labels := a.tabLabels()
	if len(labels) != len(tabDefs) {
		t.Errorf("tabLabels len = %d, want %d", len(labels), len(tabDefs))
	}
}

func TestActivateTab_ErrorLog(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabErrorLog
	a.errlogStore.Log("test", "error message")
	cmd := a.activateTab()
	if cmd != nil {
		t.Error("activateTab for error log should return nil cmd")
	}
	if len(a.errlogItems) == 0 {
		t.Error("errlogItems should be populated")
	}
}

func TestActivateTab_Transactions(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabTransactions
	a.transactionStore.Record("install", []string{"vim"}, true)
	cmd := a.activateTab()
	// Should return a cmd to load deps
	if cmd == nil {
		t.Error("activateTab for transactions with items should return cmd")
	}
	if len(a.transactionItems) == 0 {
		t.Error("transactionItems should be populated")
	}
}

func TestActivateTab_PackageTab(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabAll
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	cmd := a.activateTab()
	// cmd may be non-nil (updateSelectionCmd)
	_ = cmd
	if !strings.Contains(a.status, "packages") {
		t.Errorf("status = %q, want contains 'packages'", a.status)
	}
}

func TestLayoutHelpers(t *testing.T) {
	a := newTestApp()
	a.width = 160
	a.height = 40
	a.sideBySide = true

	if h := a.packageListHeight(); h < 5 {
		t.Errorf("packageListHeight = %d, want >= 5", h)
	}
	if h := a.sideDetailInnerHeight(); h < 3 {
		t.Errorf("sideDetailInnerHeight = %d, want >= 3", h)
	}
	if h := a.sideMainPanelHeight(); h < 7 {
		t.Errorf("sideMainPanelHeight = %d, want >= 7", h)
	}
	if w := a.sideListWidth(); w <= 0 {
		t.Errorf("sideListWidth = %d, want > 0", w)
	}
	if w := a.sideDetailWidth(); w <= 0 {
		t.Errorf("sideDetailWidth = %d, want > 0", w)
	}
}

func TestLayoutHelpers_Stacked(t *testing.T) {
	a := newTestApp()
	a.width = 80
	a.height = 40
	a.sideBySide = false

	if h := a.packageListHeight(); h < 5 {
		t.Errorf("packageListHeight (stacked) = %d, want >= 5", h)
	}
	if h := a.stackedListPanelHeight(); h < 7 {
		t.Errorf("stackedListPanelHeight = %d, want >= 7", h)
	}
	if h := a.stackedDetailPanelHeight(); h < 5 {
		t.Errorf("stackedDetailPanelHeight = %d, want >= 5", h)
	}
}

func TestTransactionListHeight(t *testing.T) {
	a := newTestApp()
	a.height = 40
	h := a.transactionListHeight()
	if h < 3 {
		t.Errorf("transactionListHeight = %d, want >= 3", h)
	}
}

func TestErrorLogListHeight(t *testing.T) {
	a := newTestApp()
	a.height = 40
	h := a.errorLogListHeight()
	if h < 3 {
		t.Errorf("errorLogListHeight = %d, want >= 3", h)
	}
}

func TestFileListHeight(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.width = 160
	a.sideBySide = true
	hSide := a.fileListHeight()
	if hSide < 1 {
		t.Errorf("fileListHeight (side) = %d, want >= 1", hSide)
	}

	a.sideBySide = false
	a.width = 80
	hStack := a.fileListHeight()
	if hStack < 1 {
		t.Errorf("fileListHeight (stacked) = %d, want >= 1", hStack)
	}
}

func TestAdjustPackageScroll(t *testing.T) {
	a := newTestApp()
	a.width = 120
	a.height = 40
	a.sideBySide = true
	a.allPackages = make([]model.Package, 100)
	a.selectedIdx = 50
	a.scrollOffset = 0
	a.adjustPackageScroll()
	if a.scrollOffset == 0 {
		t.Error("scrollOffset should have been adjusted for selectedIdx=50")
	}
}

func TestAdjustErrorLogScroll(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.errlogIdx = 50
	a.errlogOffset = 0
	a.adjustErrorLogScroll()
	if a.errlogOffset == 0 {
		t.Error("errlogOffset should have been adjusted")
	}
}

func TestAdjustPPAScroll(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.width = 120
	a.sideBySide = true
	a.ppaIdx = 50
	a.ppaOffset = 0
	a.adjustPPAScroll()
	if a.ppaOffset == 0 {
		t.Error("ppaOffset should have been adjusted")
	}
}

func TestAdjustFileListScroll(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.width = 120
	a.sideBySide = true
	a.fileListIdx = 50
	a.fileListOffset = 0
	a.adjustFileListScroll()
	if a.fileListOffset == 0 {
		t.Error("fileListOffset should have been adjusted")
	}
}

func TestDetailContentMaxScroll_EmptyFiltered(t *testing.T) {
	a := newTestApp()
	a.filtered = nil
	got := a.detailContentMaxScroll()
	if got != 0 {
		t.Errorf("detailContentMaxScroll = %d, want 0 for empty filtered", got)
	}
}

func TestDetailContentMaxScroll_OutOfBounds(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{{Name: "vim"}}
	a.selectedIdx = 5
	got := a.detailContentMaxScroll()
	if got != 0 {
		t.Errorf("detailContentMaxScroll = %d, want 0 for out-of-bounds", got)
	}
}

func TestDetailContentMaxScroll_WithDetailInfo(t *testing.T) {
	a := newTestApp()
	a.width = 160
	a.height = 20 // short terminal to force scrolling
	a.sideBySide = true
	a.filtered = []model.Package{{Name: "vim", Installed: true}}
	a.selectedIdx = 0
	// Generate content with many lines to exceed the visible area
	a.detailInfo = strings.Repeat("This is a line of detail content\n", 200)
	got := a.detailContentMaxScroll()
	if got <= 0 {
		t.Errorf("detailContentMaxScroll = %d, should be > 0 for long content", got)
	}
}

func TestApplyOptimisticUpdate_Install(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: false},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{}
	a.installedCount = 0
	a.applyOptimisticUpdate("install", []string{"vim"})
	if !a.allPackages[0].Installed {
		t.Error("vim should be installed after optimistic update")
	}
	if a.installedCount != 1 {
		t.Errorf("installedCount = %d, want 1", a.installedCount)
	}
}

func TestApplyOptimisticUpdate_Remove(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{}
	a.installedCount = 1
	a.applyOptimisticUpdate("remove", []string{"vim"})
	if a.allPackages[0].Installed {
		t.Error("vim should not be installed after remove")
	}
	if a.installedCount != 0 {
		t.Errorf("installedCount = %d, want 0", a.installedCount)
	}
}

func TestApplyOptimisticUpdate_Upgrade(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true, Version: "1.0", Upgradable: true, NewVersion: "1.1"},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{"vim": {Name: "vim", NewVersion: "1.1"}}
	a.applyOptimisticUpdate("upgrade", []string{"vim"})
	if a.allPackages[0].Upgradable {
		t.Error("vim should not be upgradable after upgrade")
	}
	if a.allPackages[0].Version != "1.1" {
		t.Errorf("vim Version = %q, want %q", a.allPackages[0].Version, "1.1")
	}
}

func TestApplyOptimisticUpdate_CleanupAll(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{
		{Name: "pkg1", Installed: true},
	}
	a.rebuildIndex()
	a.upgradableMap = map[string]model.Package{}
	a.autoremovable = []string{"pkg1"}
	a.autoremovableSet = map[string]bool{"pkg1": true}
	a.installedCount = 1
	a.applyOptimisticUpdate("cleanup-all", []string{"pkg1"})
	if len(a.autoremovable) != 0 {
		t.Error("autoremovable should be cleared after cleanup-all")
	}
}

func TestSearchBarY(t *testing.T) {
	a := newTestApp()
	y := a.searchBarY()
	if y != 3 {
		t.Errorf("searchBarY = %d, want 3", y)
	}
}

// ── Keypress handler tests ───────────────────────────────────────────

func TestDispatchErrorLog_Navigation(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		startIdx   int
		items      int
		wantIdx    int
		wantHandle bool
	}{
		{"j moves down", "j", 0, 5, 1, true},
		{"down moves down", "down", 0, 5, 1, true},
		{"k moves up", "k", 2, 5, 1, true},
		{"up moves up", "up", 2, 5, 1, true},
		{"j at end stays", "j", 4, 5, 4, true},
		{"k at start stays", "k", 0, 5, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp()
			a.activeTab = tabErrorLog
			a.height = 40
			a.errlogItems = make([]errlog.Entry, tt.items)
			a.errlogIdx = tt.startIdx
			msg := tea.KeyPressMsg{Code: -1}
			switch tt.key {
			case "j":
				msg.Code = 'j'
			case "k":
				msg.Code = 'k'
			case "down":
				msg.Code = tea.KeyDown
			case "up":
				msg.Code = tea.KeyUp
			}
			m, _, handled := a.dispatchErrorLog(msg)
			app := m.(App)
			if handled != tt.wantHandle {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandle)
			}
			if app.errlogIdx != tt.wantIdx {
				t.Errorf("errlogIdx = %d, want %d", app.errlogIdx, tt.wantIdx)
			}
		})
	}
}

func TestDispatchErrorLog_NotErrorTab(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabAll
	msg := tea.KeyPressMsg{Code: 'j'}
	_, _, handled := a.dispatchErrorLog(msg)
	if handled {
		t.Error("should not handle when not on error log tab")
	}
}

func TestScrollErrorsDown(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.errlogItems = make([]errlog.Entry, 100)
	a.errlogIdx = 0
	m, _ := a.scrollErrorsDown()
	app := m.(App)
	if app.errlogIdx == 0 {
		t.Error("errlogIdx should advance after page down")
	}
}

func TestScrollErrorsUp(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.errlogItems = make([]errlog.Entry, 100)
	a.errlogIdx = 50
	m, _ := a.scrollErrorsUp()
	app := m.(App)
	if app.errlogIdx >= 50 {
		t.Error("errlogIdx should decrease after page up")
	}
}

func TestScrollErrorsUp_AtStart(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.errlogItems = make([]errlog.Entry, 10)
	a.errlogIdx = 0
	m, _ := a.scrollErrorsUp()
	app := m.(App)
	if app.errlogIdx != 0 {
		t.Errorf("errlogIdx = %d, want 0 when already at start", app.errlogIdx)
	}
}

func TestClearErrorLog(t *testing.T) {
	a := newTestApp()
	a.errlogStore.Log("test", "msg1")
	a.errlogStore.Log("test", "msg2")
	a.errlogItems = a.errlogStore.All()
	a.errlogIdx = 1
	m, _ := a.clearErrorLog()
	app := m.(App)
	if len(app.errlogItems) != 0 {
		t.Errorf("errlogItems len = %d, want 0", len(app.errlogItems))
	}
	if app.errlogIdx != 0 {
		t.Errorf("errlogIdx = %d, want 0", app.errlogIdx)
	}
}

func TestSelectNextTransaction(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.transactionStore.Record("install", []string{"vim"}, true)
	a.transactionStore.Record("remove", []string{"git"}, true)
	a.transactionItems = a.transactionStore.All()
	a.transactionIdx = 0
	m, cmd := a.selectNextTransaction()
	app := m.(App)
	if app.transactionIdx != 1 {
		t.Errorf("transactionIdx = %d, want 1", app.transactionIdx)
	}
	if cmd == nil {
		t.Error("should return cmd to load deps")
	}
}

func TestSelectNextTransaction_AtEnd(t *testing.T) {
	a := newTestApp()
	a.transactionItems = []history.Transaction{
		{ID: 1, Operation: "install", Packages: []string{"vim"}, Success: true},
	}
	a.transactionIdx = 0
	m, _ := a.selectNextTransaction()
	app := m.(App)
	if app.transactionIdx != 0 {
		t.Errorf("transactionIdx = %d, want 0 when at end", app.transactionIdx)
	}
}

func TestSelectPreviousTransaction(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.transactionStore.Record("install", []string{"vim"}, true)
	a.transactionStore.Record("remove", []string{"git"}, true)
	a.transactionItems = a.transactionStore.All()
	a.transactionIdx = 1
	m, cmd := a.selectPreviousTransaction()
	app := m.(App)
	if app.transactionIdx != 0 {
		t.Errorf("transactionIdx = %d, want 0", app.transactionIdx)
	}
	if cmd == nil {
		t.Error("should return cmd to load deps")
	}
}

func TestSelectPreviousTransaction_AtStart(t *testing.T) {
	a := newTestApp()
	a.transactionStore.Record("install", []string{"vim"}, true)
	a.transactionItems = a.transactionStore.All()
	a.transactionIdx = 0
	m, _ := a.selectPreviousTransaction()
	app := m.(App)
	if app.transactionIdx != 0 {
		t.Errorf("transactionIdx = %d, want 0", app.transactionIdx)
	}
}

func TestScrollTransactionsDown(t *testing.T) {
	a := newTestApp()
	a.height = 40
	for i := 0; i < 50; i++ {
		a.transactionStore.Record("install", []string{fmt.Sprintf("pkg%d", i)}, true)
	}
	a.transactionItems = a.transactionStore.All()
	a.transactionIdx = 0
	m, _ := a.scrollTransactionsDown()
	app := m.(App)
	if app.transactionIdx == 0 {
		t.Error("transactionIdx should advance after page down")
	}
}

func TestScrollTransactionsUp(t *testing.T) {
	a := newTestApp()
	a.height = 40
	for i := 0; i < 50; i++ {
		a.transactionStore.Record("install", []string{fmt.Sprintf("pkg%d", i)}, true)
	}
	a.transactionItems = a.transactionStore.All()
	a.transactionIdx = 30
	m, _ := a.scrollTransactionsUp()
	app := m.(App)
	if app.transactionIdx >= 30 {
		t.Error("transactionIdx should decrease after page up")
	}
}

func TestFileListKeypress_Close(t *testing.T) {
	a := newTestApp()
	a.fileListActive = true
	a.fileListItems = []string{"/usr/bin/vim", "/usr/share/vim/doc"}
	a.fileListPkg = "vim"
	a.filtered = []model.Package{{Name: "vim"}}
	msg := tea.KeyPressMsg{Code: 'l'}
	m, _, handled := a.onFileListKeypress(msg)
	app := m.(App)
	if !handled {
		t.Error("should handle 'l' key")
	}
	if app.fileListActive {
		t.Error("fileListActive should be false after close")
	}
}

func TestFileListKeypress_Inactive(t *testing.T) {
	a := newTestApp()
	a.fileListActive = false
	msg := tea.KeyPressMsg{Code: 'l'}
	_, _, handled := a.onFileListKeypress(msg)
	if handled {
		t.Error("should not handle when file list is inactive")
	}
}

func TestFileListKeypress_Navigate(t *testing.T) {
	a := newTestApp()
	a.height = 40
	a.width = 160
	a.sideBySide = true
	a.fileListActive = true
	a.fileListItems = []string{"/usr/bin/vim", "/usr/share/vim/doc", "/etc/vim/vimrc"}
	a.fileListIdx = 0

	// Move down
	msg := tea.KeyPressMsg{Code: 'J', Text: "J"}
	m, _, handled := a.onFileListKeypress(msg)
	if !handled {
		t.Error("J should be handled")
	}
	app := m.(App)
	if app.fileListIdx != 1 {
		t.Errorf("fileListIdx = %d, want 1", app.fileListIdx)
	}

	// Move up
	msg = tea.KeyPressMsg{Code: 'K', Text: "K"}
	m, _, handled = app.onFileListKeypress(msg)
	if !handled {
		t.Error("K should be handled")
	}
	app = m.(App)
	if app.fileListIdx != 0 {
		t.Errorf("fileListIdx = %d, want 0", app.fileListIdx)
	}
}

func TestOpenFileList_Empty(t *testing.T) {
	a := newTestApp()
	a.filtered = nil
	m, cmd := a.openFileList()
	_ = m.(App)
	if cmd != nil {
		t.Error("cmd should be nil for empty filtered list")
	}
}

func TestOpenFileList_Toggle(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{{Name: "vim"}}
	a.selectedIdx = 0
	a.fileListActive = true
	a.fileListPkg = "vim"
	m, _ := a.openFileList()
	app := m.(App)
	if app.fileListActive {
		t.Error("fileListActive should be toggled off")
	}
}

func TestOpenFileList_Cached(t *testing.T) {
	a := newTestApp()
	a.filtered = []model.Package{{Name: "vim"}}
	a.selectedIdx = 0
	a.fileListCache = map[string][]string{
		"vim": {"/usr/bin/vim", "/etc/vim/vimrc"},
	}
	m, cmd := a.openFileList()
	app := m.(App)
	if !app.fileListActive {
		t.Error("fileListActive should be true")
	}
	if len(app.fileListItems) != 2 {
		t.Errorf("fileListItems len = %d, want 2", len(app.fileListItems))
	}
	if cmd != nil {
		t.Error("cmd should be nil when using cache")
	}
}

func TestOnFileListLoaded_Success(t *testing.T) {
	a := newTestApp()
	a.fileListPkg = "vim"
	a.fileListActive = true
	a.fileListCache = map[string][]string{}
	msg := fileListLoadedMsg{name: "vim", files: []string{"/usr/bin/vim"}}
	m, _ := a.onFileListLoaded(msg)
	app := m.(App)
	if len(app.fileListItems) != 1 {
		t.Errorf("fileListItems len = %d, want 1", len(app.fileListItems))
	}
	if _, ok := app.fileListCache["vim"]; !ok {
		t.Error("should cache file list")
	}
}

func TestOnFileListLoaded_WrongPackage(t *testing.T) {
	a := newTestApp()
	a.fileListPkg = "git"
	a.fileListActive = true
	msg := fileListLoadedMsg{name: "vim", files: []string{"/usr/bin/vim"}}
	m, _ := a.onFileListLoaded(msg)
	app := m.(App)
	if len(app.fileListItems) != 0 {
		t.Error("should ignore file list for wrong package")
	}
}

func TestOnFileListLoaded_Error(t *testing.T) {
	a := newTestApp()
	a.fileListPkg = "vim"
	a.fileListActive = true
	msg := fileListLoadedMsg{name: "vim", err: errors.New("some error")}
	m, _ := a.onFileListLoaded(msg)
	app := m.(App)
	if app.fileListActive {
		t.Error("fileListActive should be false on error")
	}
}

func TestErrIsAptFileMissing(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"other error", errors.New("timeout"), false},
		{"apt-file missing", errors.New("apt-file is not installed"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errIsAptFileMissing(tt.err); got != tt.want {
				t.Errorf("errIsAptFileMissing() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ── View function tests ──────────────────────────────────────────────

func TestView_ZeroWidth(t *testing.T) {
	a := newTestApp()
	a.width = 0
	v := a.View()
	// Should return loading spinner view
	if v.Content == "" {
		t.Error("View should render something even with zero width")
	}
}

func TestView_SideBySide(t *testing.T) {
	a := newTestApp()
	a.width = 160
	a.height = 40
	a.sideBySide = true
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("View should render content")
	}
}

func TestView_Stacked(t *testing.T) {
	a := newTestApp()
	a.width = 80
	a.height = 40
	a.sideBySide = false
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("View should render content")
	}
}

func TestView_ErrorLogTab(t *testing.T) {
	a := newTestApp()
	a.width = 120
	a.height = 40
	a.activeTab = tabErrorLog
	a.errlogStore.Log("test", "error1")
	a.errlogItems = a.errlogStore.All()
	v := a.View()
	if v.Content == "" {
		t.Error("View should render error log")
	}
}

func TestView_TransactionsTab(t *testing.T) {
	a := newTestApp()
	a.width = 120
	a.height = 40
	a.activeTab = tabTransactions
	a.transactionStore.Record("install", []string{"vim"}, true)
	a.transactionItems = a.transactionStore.All()
	v := a.View()
	if v.Content == "" {
		t.Error("View should render transactions")
	}
}

func TestApplyComponentStyles(t *testing.T) {
	a := newTestApp()
	// Just ensure no panic
	a.applyComponentStyles()
}

func TestAdjustMirrorScroll(t *testing.T) {
	a := newTestApp()
	a.width = 120
	a.height = 40
	a.sideBySide = true
	a.fetchIdx = 50
	a.fetchOffset = 0
	a.adjustMirrorScroll()
	if a.fetchOffset == 0 {
		t.Error("fetchOffset should have been adjusted")
	}
}

// ── Fetch keypress tests ──

func TestCancelFetchTest(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantCancel bool
	}{
		{"esc cancels", "esc", true},
		{"q cancels", "q", true},
		{"ctrl+c cancels", "ctrl+c", true},
		{"j does nothing", "j", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.fetchView = true
			a.fetchTesting = true
			a.loading = true
			msg := tea.KeyPressMsg{Code: 0, Text: tc.key}
			if tc.key == "esc" {
				msg = tea.KeyPressMsg{Code: tea.KeyEsc}
			} else if tc.key == "ctrl+c" {
				msg = tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
			}
			m, _ := a.cancelFetchTest(msg)
			result := m.(App)
			if tc.wantCancel {
				if result.fetchView {
					t.Error("fetchView should be false")
				}
				if result.fetchTesting {
					t.Error("fetchTesting should be false")
				}
				if result.loading {
					t.Error("loading should be false")
				}
			} else {
				if !result.fetchView {
					t.Error("fetchView should still be true")
				}
			}
		})
	}
}

func TestCloseMirrorView(t *testing.T) {
	a := newTestApp()
	a.fetchView = true
	a.filtered = make([]model.Package, 5)
	m, _ := a.closeMirrorView()
	result := m.(App)
	if result.fetchView {
		t.Error("fetchView should be false")
	}
}

func TestSelectNextMirror(t *testing.T) {
	tests := []struct {
		name     string
		mirrors  int
		startIdx int
		wantIdx  int
	}{
		{"move down", 5, 0, 1},
		{"at end", 5, 4, 4},
		{"single mirror", 1, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.fetchMirrors = make([]fetch.Mirror, tc.mirrors)
			a.fetchIdx = tc.startIdx
			a.fetchSelected = make(map[int]bool)
			m, _ := a.selectNextMirror()
			result := m.(App)
			if result.fetchIdx != tc.wantIdx {
				t.Errorf("got fetchIdx=%d, want %d", result.fetchIdx, tc.wantIdx)
			}
		})
	}
}

func TestSelectPreviousMirror(t *testing.T) {
	tests := []struct {
		name     string
		startIdx int
		wantIdx  int
	}{
		{"move up", 3, 2},
		{"at start", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.fetchMirrors = make([]fetch.Mirror, 10)
			a.fetchIdx = tc.startIdx
			a.fetchSelected = make(map[int]bool)
			m, _ := a.selectPreviousMirror()
			result := m.(App)
			if result.fetchIdx != tc.wantIdx {
				t.Errorf("got fetchIdx=%d, want %d", result.fetchIdx, tc.wantIdx)
			}
		})
	}
}

func TestScrollMirrorsDown(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = make([]fetch.Mirror, 100)
	a.fetchIdx = 0
	a.fetchSelected = make(map[int]bool)
	m, _ := a.scrollMirrorsDown()
	result := m.(App)
	if result.fetchIdx == 0 {
		t.Error("fetchIdx should have advanced")
	}
}

func TestScrollMirrorsDown_Empty(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = nil
	a.fetchIdx = 0
	a.fetchSelected = make(map[int]bool)
	m, _ := a.scrollMirrorsDown()
	result := m.(App)
	if result.fetchIdx != 0 {
		t.Errorf("fetchIdx should be 0, got %d", result.fetchIdx)
	}
}

func TestScrollMirrorsUp(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = make([]fetch.Mirror, 100)
	a.fetchIdx = 50
	a.fetchSelected = make(map[int]bool)
	m, _ := a.scrollMirrorsUp()
	result := m.(App)
	if result.fetchIdx >= 50 {
		t.Error("fetchIdx should have decreased")
	}
}

func TestScrollMirrorsUp_AtStart(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = make([]fetch.Mirror, 100)
	a.fetchIdx = 0
	a.fetchSelected = make(map[int]bool)
	m, _ := a.scrollMirrorsUp()
	result := m.(App)
	if result.fetchIdx != 0 {
		t.Errorf("fetchIdx should stay 0, got %d", result.fetchIdx)
	}
}

func TestToggleMirrorSelection(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = make([]fetch.Mirror, 5)
	a.fetchIdx = 2
	a.fetchSelected = make(map[int]bool)

	// Select
	m, _ := a.toggleMirrorSelection()
	result := m.(App)
	if !result.fetchSelected[2] {
		t.Error("mirror 2 should be selected")
	}

	// Deselect
	m, _ = result.toggleMirrorSelection()
	result = m.(App)
	if result.fetchSelected[2] {
		t.Error("mirror 2 should be deselected")
	}
}

func TestToggleMirrorSelection_Empty(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = nil
	a.fetchSelected = make(map[int]bool)
	m, _ := a.toggleMirrorSelection()
	result := m.(App)
	if len(result.fetchSelected) != 0 {
		t.Error("no mirrors should be selected")
	}
}

func TestApplySelectedMirrors_NoSelection(t *testing.T) {
	a := newTestApp()
	a.fetchMirrors = make([]fetch.Mirror, 5)
	a.fetchSelected = make(map[int]bool)
	m, _ := a.applySelectedMirrors()
	result := m.(App)
	if !strings.Contains(result.status, "Select at least one mirror") {
		t.Error("should show error about no selection")
	}
}

func TestOnFetchKeypress_Dispatch(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		startIdx int
		wantIdx  int
	}{
		{"j moves down", "j", 0, 1},
		{"k moves up", "k", 3, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.fetchView = true
			a.fetchTesting = false
			a.fetchMirrors = make([]fetch.Mirror, 10)
			a.fetchIdx = tc.startIdx
			a.fetchSelected = make(map[int]bool)
			msg := tea.KeyPressMsg{Code: 0, Text: tc.key}
			m, _ := a.onFetchKeypress(msg)
			result := m.(App)
			if result.fetchIdx != tc.wantIdx {
				t.Errorf("got fetchIdx=%d, want %d", result.fetchIdx, tc.wantIdx)
			}
		})
	}
}

func TestOnFetchKeypress_EscCloses(t *testing.T) {
	a := newTestApp()
	a.fetchView = true
	a.fetchTesting = false
	a.fetchMirrors = make([]fetch.Mirror, 5)
	a.fetchSelected = make(map[int]bool)
	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m, _ := a.onFetchKeypress(msg)
	result := m.(App)
	if result.fetchView {
		t.Error("fetchView should be false after esc")
	}
}

// ── PPA keypress tests ──

func TestSelectNextPPA(t *testing.T) {
	tests := []struct {
		name     string
		items    int
		startIdx int
		wantIdx  int
	}{
		{"move down", 5, 0, 1},
		{"at end", 3, 2, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.ppaItems = make([]apt.PPA, tc.items)
			a.ppaIdx = tc.startIdx
			m, _ := a.selectNextPPA()
			result := m.(App)
			if result.ppaIdx != tc.wantIdx {
				t.Errorf("got ppaIdx=%d, want %d", result.ppaIdx, tc.wantIdx)
			}
		})
	}
}

func TestSelectPreviousPPA(t *testing.T) {
	tests := []struct {
		name     string
		startIdx int
		wantIdx  int
	}{
		{"move up", 3, 2},
		{"at start", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.ppaItems = make([]apt.PPA, 5)
			a.ppaIdx = tc.startIdx
			m, _ := a.selectPreviousPPA()
			result := m.(App)
			if result.ppaIdx != tc.wantIdx {
				t.Errorf("got ppaIdx=%d, want %d", result.ppaIdx, tc.wantIdx)
			}
		})
	}
}

func TestScrollPPAsDown(t *testing.T) {
	a := newTestApp()
	a.ppaItems = make([]apt.PPA, 100)
	a.ppaIdx = 0
	m, _ := a.scrollPPAsDown()
	result := m.(App)
	if result.ppaIdx == 0 {
		t.Error("ppaIdx should have advanced")
	}
}

func TestScrollPPAsDown_Empty(t *testing.T) {
	a := newTestApp()
	a.ppaItems = nil
	a.ppaIdx = 0
	m, _ := a.scrollPPAsDown()
	result := m.(App)
	if result.ppaIdx != 0 {
		t.Errorf("ppaIdx should be 0, got %d", result.ppaIdx)
	}
}

func TestScrollPPAsUp(t *testing.T) {
	a := newTestApp()
	a.ppaItems = make([]apt.PPA, 100)
	a.ppaIdx = 50
	m, _ := a.scrollPPAsUp()
	result := m.(App)
	if result.ppaIdx >= 50 {
		t.Error("ppaIdx should have decreased")
	}
}

func TestScrollPPAsUp_AtStart(t *testing.T) {
	a := newTestApp()
	a.ppaItems = make([]apt.PPA, 100)
	a.ppaIdx = 0
	m, _ := a.scrollPPAsUp()
	result := m.(App)
	if result.ppaIdx != 0 {
		t.Errorf("ppaIdx should stay 0, got %d", result.ppaIdx)
	}
}

func TestStartAddPPA(t *testing.T) {
	a := newTestApp()
	m, cmd := a.startAddPPA()
	result := m.(App)
	if !result.ppaAdding {
		t.Error("ppaAdding should be true")
	}
	if !strings.Contains(result.status, "Enter PPA") {
		t.Error("status should prompt for PPA")
	}
	if cmd == nil {
		t.Error("should return focus command")
	}
}

func TestSubmitAddPPA_InvalidPPA(t *testing.T) {
	a := newTestApp()
	a.ppaAdding = true
	a.ppaInput.SetValue("invalid-ppa")
	m, _ := a.submitAddPPA()
	result := m.(App)
	// Should show error in status (ANSI styled)
	if result.status == "" {
		t.Error("status should contain error")
	}
}

func TestSubmitAddPPA_ValidPPA(t *testing.T) {
	a := newTestApp()
	a.ppaAdding = true
	a.ppaInput.SetValue("ppa:test/repo")
	m, cmd := a.submitAddPPA()
	result := m.(App)
	if result.ppaAdding {
		t.Error("ppaAdding should be false")
	}
	if !result.loading {
		t.Error("loading should be true")
	}
	if result.pendingExecOp != "ppa-add" {
		t.Errorf("pendingExecOp should be ppa-add, got %s", result.pendingExecOp)
	}
	if cmd == nil {
		t.Error("should return a command")
	}
}

func TestRemoveSelectedPPA_Empty(t *testing.T) {
	a := newTestApp()
	a.ppaItems = nil
	m, _ := a.removeSelectedPPA()
	result := m.(App)
	if result.loading {
		t.Error("should not be loading for empty list")
	}
}

func TestRemoveSelectedPPA_NotPPA(t *testing.T) {
	a := newTestApp()
	a.ppaItems = []apt.PPA{{Name: "ubuntu-main", IsPPA: false}}
	a.ppaIdx = 0
	m, _ := a.removeSelectedPPA()
	result := m.(App)
	if !strings.Contains(result.status, "only supported for PPA") {
		t.Error("should show error for non-PPA repo")
	}
}

func TestRemoveSelectedPPA_Success(t *testing.T) {
	a := newTestApp()
	a.ppaItems = []apt.PPA{{Name: "ppa:test/repo", IsPPA: true}}
	a.ppaIdx = 0
	m, cmd := a.removeSelectedPPA()
	result := m.(App)
	if !result.loading {
		t.Error("should be loading")
	}
	if result.pendingExecOp != "ppa-remove" {
		t.Errorf("pendingExecOp should be ppa-remove, got %s", result.pendingExecOp)
	}
	if cmd == nil {
		t.Error("should return a command")
	}
}

func TestToggleSelectedPPA_Empty(t *testing.T) {
	a := newTestApp()
	a.ppaItems = nil
	m, _ := a.toggleSelectedPPA()
	result := m.(App)
	if result.loading {
		t.Error("should not be loading for empty list")
	}
}

func TestToggleSelectedPPA_Enable(t *testing.T) {
	a := newTestApp()
	a.ppaItems = []apt.PPA{{Name: "ppa:test/repo", Enabled: false}}
	a.ppaIdx = 0
	m, cmd := a.toggleSelectedPPA()
	result := m.(App)
	if !result.loading {
		t.Error("should be loading")
	}
	if !strings.Contains(result.status, "Enabling") {
		t.Error("status should say Enabling")
	}
	if cmd == nil {
		t.Error("should return a command")
	}
}

func TestToggleSelectedPPA_Disable(t *testing.T) {
	a := newTestApp()
	a.ppaItems = []apt.PPA{{Name: "ppa:test/repo", Enabled: true}}
	a.ppaIdx = 0
	m, _ := a.toggleSelectedPPA()
	result := m.(App)
	if !strings.Contains(result.status, "Disabling") {
		t.Error("status should say Disabling")
	}
}

func TestOnPPAKeypress_Navigation(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		startIdx int
		wantIdx  int
	}{
		{"j moves down", "j", 0, 1},
		{"k moves up", "k", 3, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.activeTab = tabRepos
			a.ppaItems = make([]apt.PPA, 10)
			a.ppaIdx = tc.startIdx
			msg := tea.KeyPressMsg{Code: 0, Text: tc.key}
			m, _ := a.onPPAKeypress(msg)
			result := m.(App)
			if result.ppaIdx != tc.wantIdx {
				t.Errorf("got ppaIdx=%d, want %d", result.ppaIdx, tc.wantIdx)
			}
		})
	}
}

func TestOnPPAInputKeypress_Escape(t *testing.T) {
	a := newTestApp()
	a.ppaAdding = true
	a.ppaItems = make([]apt.PPA, 3)
	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m, _ := a.onPPAInputKeypress(msg)
	result := m.(App)
	if result.ppaAdding {
		t.Error("ppaAdding should be false after esc")
	}
}

func TestOnPPAInputKeypress_Enter(t *testing.T) {
	a := newTestApp()
	a.ppaAdding = true
	a.ppaInput.SetValue("ppa:test/repo")
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m, _ := a.onPPAInputKeypress(msg)
	result := m.(App)
	if result.ppaAdding {
		t.Error("ppaAdding should be false after enter with valid PPA")
	}
}

// ── Portpkg keypress tests ──

func TestExportInstalledPackages_Loading(t *testing.T) {
	a := newTestApp()
	a.loading = true
	m, cmd := a.exportInstalledPackages()
	result := m.(App)
	if cmd != nil {
		t.Error("should not return command when loading")
	}
	if strings.Contains(result.status, "Exporting") {
		t.Error("should not start export when loading")
	}
}

func TestExportManualPackages_Loading(t *testing.T) {
	a := newTestApp()
	a.loading = true
	m, _ := a.exportManualPackages()
	result := m.(App)
	_ = result // should be a noop when loading
}

func TestImportPackages_Loading(t *testing.T) {
	a := newTestApp()
	a.loading = true
	m, _ := a.importPackages()
	result := m.(App)
	if result.importingPath {
		t.Error("should not start import when loading")
	}
}

func TestImportPackages_StartsInput(t *testing.T) {
	a := newTestApp()
	a.loading = false
	m, cmd := a.importPackages()
	result := m.(App)
	if !result.importingPath {
		t.Error("importingPath should be true")
	}
	if !strings.Contains(result.status, "Enter file path") {
		t.Error("status should prompt for file path")
	}
	if cmd == nil {
		t.Error("should return focus command")
	}
}

func TestOnImportInputKeypress_Escape(t *testing.T) {
	a := newTestApp()
	a.importingPath = true
	a.filtered = make([]model.Package, 5)
	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m, _ := a.onImportInputKeypress(msg)
	result := m.(App)
	if result.importingPath {
		t.Error("importingPath should be false after esc")
	}
}

func TestOnImportInputKeypress_Enter(t *testing.T) {
	a := newTestApp()
	a.importingPath = true
	a.importInput.SetValue("/tmp/packages.txt")
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m, cmd := a.onImportInputKeypress(msg)
	result := m.(App)
	if result.importingPath {
		t.Error("importingPath should be false after enter")
	}
	if !strings.Contains(result.status, "Reading package list") {
		t.Errorf("status should say reading, got %s", result.status)
	}
	if cmd == nil {
		t.Error("should return command to import packages")
	}
}

func TestOnImportConfirmKeypress_Yes(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importToInstall = []string{"vim", "git"}
	a.importFromPath = "/tmp/packages.txt"
	msg := tea.KeyPressMsg{Code: 0, Text: "y"}
	m, cmd := a.onImportConfirmKeypress(msg)
	result := m.(App)
	if result.importConfirm {
		t.Error("importConfirm should be false")
	}
	if !result.loading {
		t.Error("loading should be true")
	}
	if result.pendingExecOp != "install" {
		t.Errorf("pendingExecOp should be install, got %s", result.pendingExecOp)
	}
	if cmd == nil {
		t.Error("should return install command")
	}
}

func TestOnImportConfirmKeypress_No(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importToInstall = []string{"vim"}
	msg := tea.KeyPressMsg{Code: 0, Text: "n"}
	m, _ := a.onImportConfirmKeypress(msg)
	result := m.(App)
	if result.importConfirm {
		t.Error("importConfirm should be false")
	}
	if result.importToInstall != nil {
		t.Error("importToInstall should be nil")
	}
}

func TestOnImportConfirmKeypress_Esc(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importToInstall = []string{"vim"}
	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m, _ := a.onImportConfirmKeypress(msg)
	result := m.(App)
	if result.importConfirm {
		t.Error("importConfirm should be false after esc")
	}
}

func TestOnImportConfirmKeypress_Details(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importToInstall = []string{"vim", "git", "curl"}
	msg := tea.KeyPressMsg{Code: 0, Text: "d"}
	m, _ := a.onImportConfirmKeypress(msg)
	result := m.(App)
	if !result.importDetails {
		t.Error("importDetails should be true")
	}
}

func TestOnImportConfirmKeypress_DetailsPagination(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importDetails = true
	// Create more than 15 packages to have multiple pages
	pkgs := make([]string, 30)
	for i := range pkgs {
		pkgs[i] = fmt.Sprintf("pkg-%d", i)
	}
	a.importToInstall = pkgs

	// Navigate right
	msg := tea.KeyPressMsg{Code: 0, Text: "right"}
	m, _ := a.onImportConfirmKeypress(msg)
	result := m.(App)
	if result.importDetailOffset != 1 {
		t.Errorf("importDetailOffset should be 1, got %d", result.importDetailOffset)
	}

	// Navigate left
	msg = tea.KeyPressMsg{Code: 0, Text: "left"}
	m, _ = result.onImportConfirmKeypress(msg)
	result = m.(App)
	if result.importDetailOffset != 0 {
		t.Errorf("importDetailOffset should be 0, got %d", result.importDetailOffset)
	}

	// d closes details
	msg = tea.KeyPressMsg{Code: 0, Text: "d"}
	m, _ = result.onImportConfirmKeypress(msg)
	result = m.(App)
	if result.importDetails {
		t.Error("importDetails should be false after d in detail mode")
	}
}

// ── Search keypress tests ──

func TestSubmitSearch_EmptyQuery(t *testing.T) {
	a := newTestApp()
	a.searching = true
	a.searchInput.SetValue("")
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.rebuildIndex()
	a.applyFilter()
	m, _ := a.submitSearch()
	result := m.(App)
	if result.searching {
		t.Error("searching should be false")
	}
	if !strings.Contains(result.status, "packages") {
		t.Errorf("status should show package count, got: %s", result.status)
	}
}

func TestSubmitSearch_WithResults(t *testing.T) {
	a := newTestApp()
	a.searching = true
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: true},
	}
	a.rebuildIndex()
	a.filterQuery = "vim"
	a.applyFilter()
	a.searchInput.SetValue("vim")
	m, _ := a.submitSearch()
	result := m.(App)
	if result.searching {
		t.Error("searching should be false")
	}
}

func TestCancelSearch(t *testing.T) {
	a := newTestApp()
	a.searching = true
	a.filterQueryBeforeEdit = "old-query"
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	m, _ := a.cancelSearch()
	result := m.(App)
	if result.searching {
		t.Error("searching should be false")
	}
	if result.filterQuery != "old-query" {
		t.Errorf("filterQuery should be restored to %q, got %q", "old-query", result.filterQuery)
	}
}

func TestUpdateSearchFilter(t *testing.T) {
	a := newTestApp()
	a.searching = true
	a.allPackages = []model.Package{
		{Name: "vim", Installed: true},
		{Name: "git", Installed: false},
	}
	a.rebuildIndex()
	a.applyFilter()
	msg := tea.KeyPressMsg{Code: 'v', Text: "v"}
	m, _ := a.updateSearchFilter(msg)
	result := m.(App)
	if !strings.Contains(result.status, "matching") {
		t.Errorf("status should contain 'matching', got %s", result.status)
	}
}

func TestOnSearchKeypress_Enter(t *testing.T) {
	a := newTestApp()
	a.searching = true
	a.searchInput.SetValue("")
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m, _ := a.onSearchKeypress(msg)
	result := m.(App)
	if result.searching {
		t.Error("searching should be false after enter")
	}
}

func TestOnSearchKeypress_Esc(t *testing.T) {
	a := newTestApp()
	a.searching = true
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m, _ := a.onSearchKeypress(msg)
	result := m.(App)
	if result.searching {
		t.Error("searching should be false after esc")
	}
}

// ── Main keypress dispatch tests ──

func TestToggleHelp(t *testing.T) {
	a := newTestApp()
	initial := a.help.ShowAll
	m, _ := a.toggleHelp()
	result := m.(App)
	if result.help.ShowAll == initial {
		t.Error("help.ShowAll should be toggled")
	}
}

func TestToggleLayout(t *testing.T) {
	a := newTestApp()
	a.width = 200 // wide enough for side-by-side
	a.sideBySide = false
	m, _ := a.toggleLayout()
	result := m.(App)
	if !result.sideBySide {
		t.Error("sideBySide should be true after toggle")
	}
}

func TestToggleLayout_TooNarrow(t *testing.T) {
	a := newTestApp()
	a.width = 50
	a.sideBySide = false
	m, _ := a.toggleLayout()
	result := m.(App)
	if result.sideBySide {
		t.Error("sideBySide should stay false when too narrow")
	}
}

func TestToggleTheme(t *testing.T) {
	a := newTestApp()
	initial := a.hasDarkBG
	m, _ := a.toggleTheme()
	result := m.(App)
	if result.hasDarkBG == initial {
		t.Error("hasDarkBG should be toggled")
	}
	if !result.themeForced {
		t.Error("themeForced should be true")
	}
}

func TestToggleRecommends(t *testing.T) {
	a := newTestApp()
	a.installRecommends = false
	m, _ := a.toggleRecommends()
	result := m.(App)
	if !result.installRecommends {
		t.Error("installRecommends should be true")
	}
	if !strings.Contains(result.status, "ON") {
		t.Error("status should say ON")
	}
}

func TestToggleSuggests(t *testing.T) {
	a := newTestApp()
	a.installSuggests = false
	m, _ := a.toggleSuggests()
	result := m.(App)
	if !result.installSuggests {
		t.Error("installSuggests should be true")
	}
	if !strings.Contains(result.status, "ON") {
		t.Error("status should say ON")
	}
}

func TestOpenSearch(t *testing.T) {
	a := newTestApp()
	a.searching = false
	a.filterQuery = "vim"
	m, cmd := a.openSearch()
	result := m.(App)
	if !result.searching {
		t.Error("searching should be true")
	}
	if result.filterQueryBeforeEdit != "vim" {
		t.Errorf("filterQueryBeforeEdit should be %q", "vim")
	}
	if cmd == nil {
		t.Error("should return focus command")
	}
}

func TestClearFilterOrSearch_ExportConfirm(t *testing.T) {
	a := newTestApp()
	a.exportConfirm = true
	a.filtered = make([]model.Package, 5)
	m, _ := a.clearFilterOrSearch()
	result := m.(App)
	if result.exportConfirm {
		t.Error("exportConfirm should be false")
	}
}

func TestClearFilterOrSearch_ExportManualConfirm(t *testing.T) {
	a := newTestApp()
	a.exportManualConfirm = true
	a.filtered = make([]model.Package, 5)
	m, _ := a.clearFilterOrSearch()
	result := m.(App)
	if result.exportManualConfirm {
		t.Error("exportManualConfirm should be false")
	}
}

func TestClearFilterOrSearch_NoFilter(t *testing.T) {
	a := newTestApp()
	a.filterQuery = ""
	m, _ := a.clearFilterOrSearch()
	result := m.(App)
	_ = result // noop
}

func TestClearFilterOrSearch_WithFilter(t *testing.T) {
	a := newTestApp()
	a.filterQuery = "vim"
	a.allPackages = []model.Package{{Name: "vim"}, {Name: "git"}}
	a.rebuildIndex()
	a.applyFilter()
	m, _ := a.clearFilterOrSearch()
	result := m.(App)
	if result.filterQuery != "" {
		t.Error("filterQuery should be cleared")
	}
	if result.selectedIdx != 0 {
		t.Error("selectedIdx should be 0")
	}
}

func TestRunAptUpdate(t *testing.T) {
	a := newTestApp()
	m, cmd := a.runAptUpdate()
	result := m.(App)
	if !result.loading {
		t.Error("loading should be true")
	}
	if result.pendingExecOp != "update" {
		t.Errorf("pendingExecOp should be update, got %s", result.pendingExecOp)
	}
	if !strings.Contains(result.status, "apt update") {
		t.Error("status should say apt update")
	}
	if cmd == nil {
		t.Error("should return a command")
	}
}

func TestReloadPackages(t *testing.T) {
	a := newTestApp()
	a.filterQuery = "vim"
	m, cmd := a.reloadPackages()
	result := m.(App)
	if !result.loading {
		t.Error("loading should be true")
	}
	if result.filterQuery != "" {
		t.Error("filterQuery should be cleared")
	}
	if cmd == nil {
		t.Error("should return command")
	}
}

func TestOpenFetchMirrors(t *testing.T) {
	a := newTestApp()
	m, cmd := a.openFetchMirrors()
	result := m.(App)
	if !result.fetchView {
		t.Error("fetchView should be true")
	}
	if !result.fetchTesting {
		t.Error("fetchTesting should be true")
	}
	if !result.loading {
		t.Error("loading should be true")
	}
	if result.fetchSelected == nil {
		t.Error("fetchSelected should be initialized")
	}
	if cmd == nil {
		t.Error("should return batch command")
	}
}

// ── Navigation and selection dispatch tests ──

func TestDispatchNavigation_Down(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	msg := tea.KeyPressMsg{Code: 0, Text: "j"}
	m, _, handled := a.dispatchNavigation(msg)
	if !handled {
		t.Error("should handle j key")
	}
	result := m.(App)
	if result.selectedIdx != 1 {
		t.Errorf("selectedIdx should be 1, got %d", result.selectedIdx)
	}
}

func TestDispatchNavigation_Up(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 2
	msg := tea.KeyPressMsg{Code: 0, Text: "k"}
	m, _, handled := a.dispatchNavigation(msg)
	if !handled {
		t.Error("should handle k key")
	}
	result := m.(App)
	if result.selectedIdx != 1 {
		t.Errorf("selectedIdx should be 1, got %d", result.selectedIdx)
	}
}

func TestDispatchNavigation_Unhandled(t *testing.T) {
	a := newTestApp()
	msg := tea.KeyPressMsg{Code: 0, Text: "x"}
	_, _, handled := a.dispatchNavigation(msg)
	if handled {
		t.Error("x should not be handled")
	}
}

func TestDispatchSelection_Space(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	msg := tea.KeyPressMsg{Code: 0, Text: "space"}
	m, _, handled := a.dispatchSelection(msg)
	if !handled {
		t.Error("space should be handled")
	}
	result := m.(App)
	if !result.selected["vim"] {
		t.Error("vim should be selected")
	}
}

func TestToggleSelectAll(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim"}, {Name: "git"}}
	a.rebuildIndex()
	a.applyFilter()
	a.selected = make(map[string]bool)

	// Select all
	m, _ := a.toggleSelectAll()
	result := m.(App)
	if len(result.selected) != 2 {
		t.Errorf("should have 2 selected, got %d", len(result.selected))
	}

	// Deselect all
	m, _ = result.toggleSelectAll()
	result = m.(App)
	if len(result.selected) != 0 {
		t.Errorf("should have 0 selected, got %d", len(result.selected))
	}
}

func TestDispatchPackageAction_Install(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim", Installed: false}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	msg := tea.KeyPressMsg{Code: 0, Text: "i"}
	m, cmd, handled := a.dispatchPackageAction(msg)
	if !handled {
		t.Error("i should be handled")
	}
	result := m.(App)
	if !result.loading {
		t.Error("should be loading")
	}
	if result.pendingExecOp != "install" {
		t.Errorf("expected install op, got %s", result.pendingExecOp)
	}
	if cmd == nil {
		t.Error("should return cmd")
	}
}

func TestInstallSelectedPackages_AlreadyInstalled(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	m, _ := a.installSelectedPackages()
	result := m.(App)
	if strings.Contains(result.status, "already installed") == false {
		t.Error("should indicate package is already installed")
	}
}

func TestRemoveSelectedPackages_NotInstalled(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim", Installed: false}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	m, _ := a.removeSelectedPackages()
	result := m.(App)
	if !strings.Contains(result.status, "not installed") {
		t.Error("should indicate package is not installed")
	}
}

func TestRemoveSelectedPackages_Essential(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "base-files", Installed: true, Essential: true}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	a.essentialSet = map[string]bool{"base-files": true}
	m, _ := a.removeSelectedPackages()
	result := m.(App)
	if !strings.Contains(result.status, "essential") {
		t.Error("should indicate package is essential")
	}
}

func TestRemoveSelectedPackages_ShowsConfirm(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	a.essentialSet = make(map[string]bool)
	m, _ := a.removeSelectedPackages()
	result := m.(App)
	if !result.removeConfirm {
		t.Error("removeConfirm should be true")
	}
	if result.removeOp != "remove" {
		t.Errorf("removeOp should be remove, got %s", result.removeOp)
	}
}

func TestUpgradeSelectedPackages(t *testing.T) {
	a := newTestApp()
	a.allPackages = []model.Package{{Name: "vim", Installed: true, Upgradable: true}}
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	a.upgradableMap = map[string]model.Package{"vim": {Name: "vim", Upgradable: true}}
	m, cmd := a.upgradeSelectedPackages()
	result := m.(App)
	if !result.loading {
		t.Error("should be loading")
	}
	if cmd == nil {
		t.Error("should return command")
	}
}

func TestScrollPackagesDown(t *testing.T) {
	a := newTestApp()
	pkgs := make([]model.Package, 100)
	for i := range pkgs {
		pkgs[i] = model.Package{Name: fmt.Sprintf("pkg-%d", i)}
	}
	a.allPackages = pkgs
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 0
	m, _ := a.scrollPackagesDown()
	result := m.(App)
	if result.selectedIdx == 0 {
		t.Error("selectedIdx should have advanced")
	}
}

func TestScrollPackagesUp(t *testing.T) {
	a := newTestApp()
	pkgs := make([]model.Package, 100)
	for i := range pkgs {
		pkgs[i] = model.Package{Name: fmt.Sprintf("pkg-%d", i)}
	}
	a.allPackages = pkgs
	a.rebuildIndex()
	a.applyFilter()
	a.selectedIdx = 50
	m, _ := a.scrollPackagesUp()
	result := m.(App)
	if result.selectedIdx >= 50 {
		t.Error("selectedIdx should have decreased")
	}
}

// ── View rendering tests ──

func TestRenderFetchView(t *testing.T) {
	a := newTestApp()
	a.fetchView = true
	a.fetchTesting = true
	a.fetchMirrors = nil
	a.fetchSelected = make(map[int]bool)
	v := a.View()
	if v.Content == "" {
		t.Error("fetch view should produce content")
	}
}

func TestRenderFetchView_WithMirrors(t *testing.T) {
	a := newTestApp()
	a.fetchView = true
	a.fetchTesting = false
	a.fetchMirrors = []fetch.Mirror{
		{URL: "http://mirror1.example.com", Country: "US", Score: 100, Status: "ok"},
		{URL: "http://mirror2.example.com", Country: "DE", Score: 80, Status: "ok"},
	}
	a.fetchIdx = 0
	a.fetchOffset = 0
	a.fetchSelected = make(map[int]bool)
	v := a.View()
	if v.Content == "" {
		t.Error("fetch view with mirrors should produce content")
	}
}

func TestRenderPPAView(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabRepos
	a.ppaItems = []apt.PPA{
		{Name: "ppa:test/repo", URL: "http://ppa.launchpad.net/test/repo/ubuntu", Enabled: true, IsPPA: true, File: "/etc/apt/sources.list.d/test.list"},
	}
	a.ppaIdx = 0
	v := a.View()
	if v.Content == "" {
		t.Error("PPA view should produce content")
	}
}

func TestRenderTransactionView(t *testing.T) {
	a := newTestApp()
	a.activeTab = tabTransactions
	a.transactionItems = []history.Transaction{
		{Operation: "install", Timestamp: time.Now(), Packages: []string{"vim"}},
	}
	v := a.View()
	if v.Content == "" {
		t.Error("transaction view should produce content")
	}
}

func TestApplyImportConfirmOverlay(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importToInstall = []string{"vim", "git"}
	a.importFromPath = "/tmp/packages.txt"
	page := strings.Repeat("x", a.width) + "\n"
	page = strings.Repeat(page, a.height)
	result := a.applyImportConfirmOverlay(page, a.width)
	if result == "" {
		t.Error("overlay should produce output")
	}
}

func TestApplyImportConfirmOverlay_Details(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importDetails = true
	pkgs := make([]string, 20)
	for i := range pkgs {
		pkgs[i] = fmt.Sprintf("pkg-%d", i)
	}
	a.importToInstall = pkgs
	a.importFromPath = "/tmp/packages.txt"
	page := strings.Repeat("x", a.width) + "\n"
	page = strings.Repeat(page, a.height)
	result := a.applyImportConfirmOverlay(page, a.width)
	if result == "" {
		t.Error("overlay with details should produce output")
	}
}

func TestApplyRemoveConfirmOverlay(t *testing.T) {
	a := newTestApp()
	a.removeConfirm = true
	a.removeOp = "remove"
	a.removeToProcess = []string{"vim"}
	a.removeCancelFocus = true
	page := strings.Repeat("x", a.width) + "\n"
	page = strings.Repeat(page, a.height)
	result := a.applyRemoveConfirmOverlay(page, a.width)
	if result == "" {
		t.Error("remove overlay should produce output")
	}
}

func TestApplyRemoveConfirmOverlay_Purge(t *testing.T) {
	a := newTestApp()
	a.removeConfirm = true
	a.removeOp = "purge"
	a.removeToProcess = []string{"vim"}
	a.removeCancelFocus = false
	page := strings.Repeat("x", a.width) + "\n"
	page = strings.Repeat(page, a.height)
	result := a.applyRemoveConfirmOverlay(page, a.width)
	if !strings.Contains(result, "Purge") {
		t.Error("should contain Purge for purge op")
	}
}

func TestView_ImportConfirmOverlay(t *testing.T) {
	a := newTestApp()
	a.importConfirm = true
	a.importToInstall = []string{"vim"}
	a.importFromPath = "/tmp/packages.txt"
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("view with import overlay should produce content")
	}
}

func TestView_RemoveConfirmOverlay(t *testing.T) {
	a := newTestApp()
	a.removeConfirm = true
	a.removeOp = "remove"
	a.removeToProcess = []string{"vim"}
	a.removeCancelFocus = true
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("view with remove overlay should produce content")
	}
}

func TestRenderPanelFileList(t *testing.T) {
	a := newTestApp()
	a.fileListActive = true
	a.fileListPkg = "vim"
	a.fileListItems = []string{"/usr/bin/vim", "/usr/share/vim/vimrc"}
	a.fileListIdx = 0
	a.fileListOffset = 0
	result := a.renderPanelFileList(100, 20)
	if result == "" {
		t.Error("should render file list content")
	}
}

func TestView_FileListActive(t *testing.T) {
	a := newTestApp()
	a.fileListActive = true
	a.fileListPkg = "vim"
	a.fileListItems = []string{"/usr/bin/vim", "/usr/share/vim/vimrc"}
	a.fileListIdx = 0
	a.fileListOffset = 0
	a.allPackages = []model.Package{{Name: "vim", Installed: true}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("view with file list active should produce content")
	}
}

func TestView_Loading(t *testing.T) {
	a := newTestApp()
	a.loading = true
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("loading view should produce content")
	}
}

func TestView_Searching(t *testing.T) {
	a := newTestApp()
	a.searching = true
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("view while searching should produce content")
	}
}

func TestView_ImportingPath(t *testing.T) {
	a := newTestApp()
	a.importingPath = true
	a.allPackages = []model.Package{{Name: "vim"}}
	a.rebuildIndex()
	a.applyFilter()
	v := a.View()
	if v.Content == "" {
		t.Error("view while importing path should produce content")
	}
}
