// Package app provides the main Bubbletea application model and logic for the aptui TUI.
package app

import (
	"os"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/mexirica/aptui/internal/apt"
	"github.com/mexirica/aptui/internal/errlog"
	"github.com/mexirica/aptui/internal/fetch"
	"github.com/mexirica/aptui/internal/filter"
	"github.com/mexirica/aptui/internal/history"
	"github.com/mexirica/aptui/internal/model"
	"github.com/mexirica/aptui/internal/pin"
	"github.com/mexirica/aptui/internal/portpkg"
	"github.com/mexirica/aptui/internal/ui"
)

type tabKind int

const (
	tabAll tabKind = iota
	tabInstalled
	tabUpgradable
	tabCleanup
	tabErrorLog
	tabTransactions
	tabRepos
)

type tabDef struct {
	label string
	kind  tabKind
	name  string
}

var tabDefs = []tabDef{
	{" ◉ All ", tabAll, "All"},
	{" ● Installed ", tabInstalled, "Installed"},
	{" ↑ Upgradable ", tabUpgradable, "Upgradable"},
	{" ◇ Cleanup ", tabCleanup, "Cleanup"},
	{" ✕ Errors ", tabErrorLog, "Errors"},
	{" ⟳ Transactions ", tabTransactions, "Transactions"},
	{" ◆ Repos ", tabRepos, "Repos"},
}

// App is the main Bubbletea model. It manages three views:
// the package list (default), the transaction history, and the mirror selector.
type App struct {
	allPackages   []model.Package
	filtered      []model.Package
	upgradableMap map[string]model.Package

	activeTab tabKind

	selectedIdx  int
	scrollOffset int

	detailInfo         string
	detailName         string
	detailScrollOffset int

	// Search state
	searchInput           textinput.Model
	searching             bool
	filterQuery           string
	filterQueryBeforeEdit string

	selected map[string]bool

	sortColumn filter.SortColumn
	sortDesc   bool

	transactionStore  *history.Store
	transactionItems  []history.Transaction
	transactionIdx    int
	transactionOffset int
	transactionDeps   []string
	pendingExecOp     string
	pendingExecPkgs   []string
	pendingExecCount  int
	pendingExecFailed bool

	fetchView     bool
	fetchDistro   fetch.Distro
	fetchMirrors  []fetch.Mirror
	fetchIdx      int
	fetchOffset   int
	fetchSelected map[int]bool
	fetchTesting  bool
	fetchTested   int
	fetchTotal    int
	fetchResultCh <-chan fetch.TestResult

	ppaItems  []apt.PPA
	ppaIdx    int
	ppaOffset int
	ppaAdding bool
	ppaInput  textinput.Model

	infoCache map[string]apt.PackageInfo
	pkgIndex  map[string]int

	autoremovable    []string
	autoremovableSet map[string]bool

	heldSet     map[string]bool
	holdPending int
	holdFailed  bool

	essentialSet map[string]bool

	pinStore  *pin.Store
	pinnedSet map[string]bool

	allNamesLoaded bool
	installedCount int

	importingPath       bool
	importInput         textinput.Model
	importConfirm       bool
	exportConfirm       bool
	exportManualConfirm bool
	importDetails       bool
	importDetailOffset  int
	importToInstall     []string
	importFromPath      string

	removeConfirm     bool
	removeToProcess   []string
	removeOp          string // "remove" or "purge"
	removeCancelFocus bool   // true if [Cancel] is focused, false if [Confirm] is focused

	errlogStore  *errlog.Store
	errlogItems  []errlog.Entry
	errlogIdx    int
	errlogOffset int

	fileListActive bool
	fileListPkg    string
	fileListItems  []string
	fileListIdx    int
	fileListOffset int
	fileListCache  map[string][]string

	installRecommends bool
	installSuggests   bool

	sideBySide bool

	spinner       spinner.Model
	help          help.Model
	keys          model.KeyMap
	hasDarkBG     bool
	themeForced   bool
	status        string
	statusLock    time.Time
	pendingStatus string
	loading       bool
	width         int
	height        int
}

func New() App {
	defaultDark := true
	themeForced := false
	if v := os.Getenv("APTUI_THEME"); v != "" {
		switch v {
		case "light":
			defaultDark = false
			themeForced = true
		case "dark":
			defaultDark = true
			themeForced = true
		}
	}

	ti := textinput.New()
	ti.Placeholder = "Search or filter: section: arch: size> installed ..."
	ti.CharLimit = 200
	ti.SetWidth(80)

	pi := textinput.New()
	pi.Placeholder = "ppa:user/repository"
	pi.CharLimit = 100
	pi.SetWidth(50)

	ii := textinput.New()
	ii.Placeholder = portpkg.DefaultPath()
	ii.CharLimit = 300
	ii.SetWidth(80)

	s := spinner.New()
	s.Spinner = spinner.Dot

	h := help.New()

	ps := pin.Load()

	ui.ApplyTheme(defaultDark)

	app := App{
		upgradableMap:     make(map[string]model.Package),
		selected:          make(map[string]bool),
		infoCache:         make(map[string]apt.PackageInfo),
		pkgIndex:          make(map[string]int),
		autoremovableSet:  make(map[string]bool),
		heldSet:           make(map[string]bool),
		essentialSet:      make(map[string]bool),
		fileListCache:     make(map[string][]string),
		installRecommends: true,
		sideBySide:        true,
		pinStore:          ps,
		pinnedSet:         ps.Set(),
		searchInput:       ti,
		ppaInput:          pi,
		importInput:       ii,
		spinner:           s,
		help:              h,
		keys:              model.Keys,
		hasDarkBG:         defaultDark,
		themeForced:       themeForced,
		status:            "Loading packages...",
		loading:           true,
		transactionStore:  history.Load(),
		errlogStore:       errlog.Load(),
	}
	app.applyComponentStyles()
	return app
}

func (a App) Init() tea.Cmd {
	return tea.Batch(a.spinner.Tick, reloadAllPackages, loadAutoremovableCmd(), loadHeldCmd(), tea.RequestBackgroundColor)
}
