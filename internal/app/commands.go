package app

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/mexirica/aptui/internal/apt"
	"github.com/mexirica/aptui/internal/fetch"
	"github.com/mexirica/aptui/internal/model"
	"github.com/mexirica/aptui/internal/portpkg"
)

// execProcessWithStderr wraps tea.ExecProcess and captures stderr so we can
// display a meaningful error message after bubbletea restores the terminal.
func execProcessWithStderr(cmd *exec.Cmd, op, name string) tea.Cmd {
	var buf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &buf)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return execFinishedMsg{op: op, name: name, err: err, stderr: strings.TrimSpace(buf.String())}
	})
}

func purgeBatchCmd(names []string) tea.Cmd {
	cmd := apt.PurgeBatchCmd(names)
	return execProcessWithStderr(cmd, "purge", strings.Join(names, " "))
}

func reloadAllPackages() tea.Msg {
	type bulkResult struct {
		info map[string]apt.PackageInfo
	}
	type pkgResult struct {
		pkgs []model.Package
		err  error
	}
	type manualResult struct {
		set map[string]bool
		err error
	}

	bulkCh := make(chan bulkResult, 1)
	installedCh := make(chan pkgResult, 1)
	upgradableCh := make(chan pkgResult, 1)
	manualCh := make(chan manualResult, 1)

	go func() {
		info := apt.LoadAllAvailableInfo()
		bulkCh <- bulkResult{info}
	}()
	go func() {
		p, err := apt.ListInstalled()
		installedCh <- pkgResult{p, err}
	}()
	go func() {
		p, err := apt.ListUpgradable()
		upgradableCh <- pkgResult{p, err}
	}()
	go func() {
		set, err := apt.ListManual()
		manualCh <- manualResult{set, err}
	}()

	br := <-bulkCh
	ir := <-installedCh
	ur := <-upgradableCh
	mr := <-manualCh

	knownNames := make([]string, 0, len(br.info)+len(ir.pkgs)+len(ur.pkgs))
	for name := range br.info {
		knownNames = append(knownNames, name)
	}
	for _, p := range ir.pkgs {
		knownNames = append(knownNames, p.Name)
	}
	for _, p := range ur.pkgs {
		knownNames = append(knownNames, p.Name)
	}
	systemPinned, pinErr := apt.ListSystemPinned(knownNames)

	if ir.err != nil {
		return allPackagesMsg{err: ir.err}
	}
	return allPackagesMsg{
		bulkInfo:     br.info,
		installed:    ir.pkgs,
		upgradable:   ur.pkgs,
		manualSet:    mr.set,
		systemPinned: systemPinned,
		err:          nil,
		manualErr:    mr.err,
		pinErr:       pinErr,
	}
}

func aptUpdateCmd() tea.Cmd {
	cmd := apt.UpdateCmd()
	return execProcessWithStderr(cmd, "update", "apt")
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func silentUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		_ = apt.SilentUpdate()
		names, _ := apt.ListAllNames()
		pkgs, _ := apt.ListUpgradable()
		return silentUpdateDoneMsg{names: names, upgradable: pkgs}
	}
}

func searchPackagesCmd(query string) tea.Cmd {
	return func() tea.Msg {
		pkgs, err := apt.SearchPackages(query)
		return searchResultMsg{pkgs, err}
	}
}

func showPackageDetailCmd(name string, version string) tea.Cmd {
	return func() tea.Msg {
		info, err := apt.ShowPackage(name, version)
		return detailLoadedMsg{name: name, version: version, info: info, err: err}
	}
}

func loadTransactionDepsCmd(txIdx int, packages []string) tea.Cmd {
	return func() tea.Msg {
		seen := make(map[string]bool)
		for _, pkg := range packages {
			seen[pkg] = true
		}
		allDeps := []string{}
		for _, pkg := range packages {
			deps, err := apt.GetDependencies(pkg)
			if err != nil {
				continue
			}
			for _, d := range deps {
				if !seen[d] {
					seen[d] = true
					allDeps = append(allDeps, d)
				}
			}
		}
		return depsLoadedMsg{txIdx: txIdx, deps: allDeps}
	}
}

func upgradeAllPackagesCmd(names []string, recommends, suggests, includePhased bool) tea.Cmd {
	cmd := apt.DistUpgradeCmd(recommends, suggests, includePhased)
	return execProcessWithStderr(cmd, "upgrade-all", strings.Join(names, " "))
}

func detectPhasedCmd() tea.Cmd {
	return func() tea.Msg {
		phased, _ := apt.PhasedPackages()
		return phasedDetectedMsg{names: phased}
	}
}

func installBatchCmd(names []string, recommends, suggests bool) tea.Cmd {
	cmd := apt.InstallBatchCmd(names, recommends, suggests)
	return execProcessWithStderr(cmd, "install", strings.Join(names, " "))
}

func removeBatchCmd(names []string) tea.Cmd {
	cmd := apt.RemoveBatchCmd(names)
	return execProcessWithStderr(cmd, "remove", strings.Join(names, " "))
}

func upgradeBatchCmd(names []string, recommends, suggests bool) tea.Cmd {
	cmd := apt.UpgradeBatchCmd(names, recommends, suggests)
	return execProcessWithStderr(cmd, "upgrade", strings.Join(names, " "))
}

func fetchMirrorListCmd() tea.Cmd {
	return func() tea.Msg {
		distro, err := fetch.DetectDistro()
		if err != nil {
			return fetchMirrorsMsg{err: err}
		}
		mirrors, err := fetch.FetchMirrorList(distro)
		if err != nil {
			return fetchMirrorsMsg{err: err}
		}
		return fetchMirrorsMsg{mirrors: mirrors, distro: distro}
	}
}

func awaitMirrorTestResult(ch <-chan fetch.TestResult) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return fetchTestResultMsg{done: true}
		}
		return fetchTestResultMsg{result: r, done: false}
	}
}

func loadAutoremovableCmd() tea.Cmd {
	return func() tea.Msg {
		names, err := apt.ListAutoremovable()
		return autoremovableMsg{names: names, err: err}
	}
}

func autoremoveAllCmd(names []string) tea.Cmd {
	cmd := apt.AutoRemoveCmd()
	return execProcessWithStderr(cmd, "cleanup-all", strings.Join(names, " "))
}

func listPPAsCmd() tea.Cmd {
	return func() tea.Msg {
		ppas, err := apt.ListAllRepos()
		return ppaListMsg{ppas: ppas, err: err}
	}
}

func addPPACmd(ppa string) tea.Cmd {
	cmd := apt.AddPPACmd(ppa)
	return execProcessWithStderr(cmd, "ppa-add", ppa)
}

func removePPACmd(ppa string) tea.Cmd {
	cmd := apt.RemovePPACmd(ppa)
	return execProcessWithStderr(cmd, "ppa-remove", ppa)
}

func togglePPACmd(ppa apt.PPA) tea.Cmd {
	return func() tea.Msg {
		enabled := !ppa.Enabled
		err := apt.SetPPAEnabled(ppa, enabled)
		action := "enabled"
		if !enabled {
			action = "disabled"
		}
		return ppaToggleMsg{name: ppa.Name, action: action, err: err}
	}
}

func loadHeldCmd() tea.Cmd {
	return func() tea.Msg {
		names, err := apt.ListHeld()
		return holdListMsg{names: names, err: err}
	}
}

func holdBatchCmd(names []string) tea.Cmd {
	return func() tea.Msg {
		err := apt.Hold(names)
		return holdFinishedMsg{op: "hold", names: names, err: err}
	}
}

func unholdBatchCmd(names []string) tea.Cmd {
	return func() tea.Msg {
		err := apt.Unhold(names)
		return holdFinishedMsg{op: "unhold", names: names, err: err}
	}
}

func exportPackagesCmd(packages []model.Package) tea.Cmd {
	return func() tea.Msg {
		var entries []portpkg.PackageEntry
		for _, p := range packages {
			if p.Installed {
				entries = append(entries, portpkg.PackageEntry{
					Name: p.Name,
				})
			}
		}
		path, err := portpkg.Export(entries)
		return exportFinishedMsg{path: path, err: err}
	}
}

func exportManualPackagesCmd(packages []model.Package) tea.Cmd {
	return func() tea.Msg {
		var entries []portpkg.PackageEntry
		for _, p := range packages {
			if p.Installed && p.ManuallyInstalled {
				entries = append(entries, portpkg.PackageEntry{
					Name: p.Name,
				})
			}
		}
		path, err := portpkg.Export(entries)
		return exportFinishedMsg{path: path, err: err}
	}
}

func loadFileListCmd(name string) tea.Cmd {
	return func() tea.Msg {
		files, err := apt.ListPackageFiles(name)
		return fileListLoadedMsg{name: name, files: files, err: err}
	}
}

func importPackagesCmd(path string) tea.Cmd {
	return func() tea.Msg {
		entries, resolvedPath, err := portpkg.Import(path)
		if err != nil {
			return importFinishedMsg{path: resolvedPath, err: err}
		}
		var names []string
		for _, e := range entries {
			names = append(names, e.Name)
		}
		return importFinishedMsg{names: names, path: resolvedPath}
	}
}

func loadVersionsCmd(name string) tea.Cmd {
	return func() tea.Msg {
		versions, err := apt.ListVersions(name)
		return versionListMsg{name: name, versions: versions, err: err}
	}
}

func installVersionCmd(name, version string, recommends, suggests bool) tea.Cmd {
	cmd := apt.InstallVersionCmd(name, version, recommends, suggests)
	return execProcessWithStderr(cmd, "install-version", name)
}
