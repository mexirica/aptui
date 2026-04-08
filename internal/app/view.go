package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/mexirica/aptui/internal/fetch"
	"github.com/mexirica/aptui/internal/model"
	"github.com/mexirica/aptui/internal/ui"
	"github.com/mexirica/aptui/internal/ui/components"
)

func (a App) newView(s string) tea.View {
	v := tea.NewView(s)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (a App) View() tea.View {
	if a.width == 0 {
		return a.newView(fmt.Sprintf("Updating and loading packages %s", a.spinner.View()))
	}

	w := a.width

	if a.fetchView {
		return a.newView(a.renderFetchView(w))
	}

	tabBar := a.renderTabBar()

	if a.activeTab == tabRepos {
		return a.newView(a.renderPPAView(w, tabBar))
	}

	if a.activeTab == tabTransactions {
		return a.newView(a.renderTransactionView(w, tabBar))
	}

	if a.activeTab == tabErrorLog {
		return a.newView(a.renderErrorLogTab(w, tabBar))
	}

	if a.sideBySide {
		return a.newView(a.renderSideBySide(w, tabBar))
	}

	return a.newView(a.renderStacked(w, tabBar))
}

func (a App) applyImportConfirmOverlay(page string, w int) string {
	bg := lipgloss.NewLayer(page)

	yKey := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorWhite).Background(ui.ColorSuccess).Padding(0, 1).Render("y")
	nKey := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorWhite).Background(ui.ColorDanger).Padding(0, 1).Render("n")
	hintText := lipgloss.NewStyle().Foreground(ui.ColorSecondary)

	var box string
	if a.importDetails {
		detailTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.ColorWhite).
			Background(ui.ColorPrimary).
			Padding(0, 2).
			Render(" Packages to Install ")

		const perPage = 15
		total := len(a.importToInstall)
		totalPages := (total + perPage - 1) / perPage
		currentPage := a.importDetailOffset + 1

		start := a.importDetailOffset * perPage
		end := start + perPage
		if end > total {
			end = total
		}
		visible := a.importToInstall[start:end]

		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorWhite)
		var lines []string
		for _, name := range visible {
			lines = append(lines, "  "+nameStyle.Render(name))
		}

		pageInfo := lipgloss.NewStyle().Foreground(ui.ColorSecondary).Render(
			fmt.Sprintf("Page %d/%d (%d packages)", currentPage, totalPages, total),
		)

		dKey := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorWhite).Background(ui.ColorPrimary).Padding(0, 1).Render("d")
		lKey := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorWhite).Background(ui.ColorPrimary).Padding(0, 1).Render("←")
		rKey := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorWhite).Background(ui.ColorPrimary).Padding(0, 1).Render("→")
		hints := yKey + hintText.Render(" confirm  ") + nKey + hintText.Render(" cancel  ") + dKey + hintText.Render(" back  ") + lKey + rKey + hintText.Render(" page")

		parts := []string{detailTitle, "", pageInfo, ""}
		parts = append(parts, lines...)
		parts = append(parts, "", hints)
		detailContent := lipgloss.JoinVertical(lipgloss.Center, parts...)

		box = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorPrimary).
			Padding(1, 3).
			Align(lipgloss.Center).
			Foreground(ui.ColorWhite).
			Render(detailContent)
	} else {
		title := lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.ColorWhite).
			Background(ui.ColorPrimary).
			Padding(0, 2).
			Render(" Import Packages ")

		countStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorInfo)
		pathStyle := lipgloss.NewStyle().Foreground(ui.ColorSubtle)
		body := fmt.Sprintf(
			"%s packages to install from\n%s",
			countStyle.Render(fmt.Sprintf("%d", len(a.importToInstall))),
			pathStyle.Render(a.importFromPath),
		)

		dKey := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorWhite).Background(ui.ColorPrimary).Padding(0, 1).Render("d")
		hints := yKey + hintText.Render(" confirm  ") + nKey + hintText.Render(" cancel  ") + dKey + hintText.Render(" details")

		content := lipgloss.JoinVertical(lipgloss.Center, title, "", body, "", hints)

		box = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorPrimary).
			Padding(1, 3).
			Align(lipgloss.Center).
			Foreground(ui.ColorWhite).
			Render(content)
	}

	boxW := lipgloss.Width(box)
	boxH := lipgloss.Height(box)
	fg := lipgloss.NewLayer(box).
		X((w - boxW) / 2).
		Y((a.height - boxH) / 2).
		Z(1)
	return lipgloss.NewCompositor(bg, fg).Render()
}

func (a App) applyRemoveConfirmOverlay(page string, w int) string {
	bg := lipgloss.NewLayer(page)

	titleText := " Remove Packages "
	if a.removeOp == "purge" {
		titleText = " Purge Packages "
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.ColorWhite).
		Background(ui.ColorPrimary).
		Padding(0, 2).
		Render(titleText)

	countStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorInfo)
	body := fmt.Sprintf(
		"Are you sure you want to %s\n%s packages?",
		a.removeOp,
		countStyle.Render(fmt.Sprintf("%d", len(a.removeToProcess))),
	)

	confirmBtnStyle := lipgloss.NewStyle().Padding(0, 2).Margin(0, 1)
	cancelBtnStyle := lipgloss.NewStyle().Padding(0, 2).Margin(0, 1)

	if a.removeCancelFocus {
		cancelBtnStyle = cancelBtnStyle.Foreground(ui.ColorWhite).Background(ui.ColorPrimary).Bold(true)
		confirmBtnStyle = confirmBtnStyle.Foreground(ui.ColorSubtle).Background(ui.ColorDim)
	} else {
		confirmBtnStyle = confirmBtnStyle.Foreground(ui.ColorWhite).Background(ui.ColorDanger).Bold(true)
		cancelBtnStyle = cancelBtnStyle.Foreground(ui.ColorSubtle).Background(ui.ColorDim)
	}

	confirmBtn := confirmBtnStyle.Render("Confirm/Remove")
	cancelBtn := cancelBtnStyle.Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, confirmBtn, cancelBtn)

	content := lipgloss.JoinVertical(lipgloss.Center, title, "", body, "", buttons)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 3).
		Align(lipgloss.Center).
		Foreground(ui.ColorWhite).
		Render(content)

	boxW := lipgloss.Width(box)
	boxH := lipgloss.Height(box)
	fg := lipgloss.NewLayer(box).
		X((w - boxW) / 2).
		Y((a.height - boxH) / 2).
		Z(1)
	return lipgloss.NewCompositor(bg, fg).Render()
}

func (a App) renderTabBar() string {
	labels := a.tabLabels()
	var parts []string
	for i, t := range tabDefs {
		parts = append(parts, a.tabStyle(t).Render(labels[i]))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (a App) renderStacked(w int, tabBar string) string {
	listPanelH := a.stackedListPanelHeight()
	detailPanelH := a.stackedDetailPanelHeight()
	innerW := w - 2
	listInnerH := listPanelH - 2
	detailInnerH := detailPanelH - 2

	// ── Panel 1: Package List (full width) ──

	counterStyle := lipgloss.NewStyle().Foreground(ui.ColorSubtle)
	pos := a.selectedIdx + 1
	if len(a.filtered) == 0 {
		pos = 0
	}
	counterText := counterStyle.Render(fmt.Sprintf("%d/%d", pos, len(a.filtered)))

	var listContent string
	if a.loading {
		pad := listInnerH / 2
		loadingLine := fmt.Sprintf("Loading packages %s", a.spinner.View())
		centered := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Center).Render(loadingLine)
		listContent = strings.Repeat("\n", pad) + centered
	} else {
		si := a.effectiveSortInfo()
		listContent = components.RenderPackageList(a.filtered, a.selectedIdx, a.scrollOffset, a.packageListHeight(), innerW, a.selected, si)
	}
	listPanel := renderTitledPanel("Package List", counterText, listContent, w, listPanelH)

	// ── Panel 2: Package Detail (full width) ──

	detailTitle := "Package Detail"
	if a.fileListActive {
		detailTitle = fmt.Sprintf("Files: %s", a.fileListPkg)
	}

	var detailContent string
	if a.fileListActive {
		detailContent = a.renderPanelFileList(innerW, detailInnerH)
	} else if !a.loading && len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) && a.detailName != "" && a.detailInfo != "" {
		detailContent = scrollDetailView(components.RenderPackageDetail(enrichedDetailInfo(a.filtered[a.selectedIdx], a.detailInfo), innerW, 0, 1), detailInnerH, a.detailScrollOffset)
	} else if !a.loading && len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) {
		detailContent = scrollDetailView(a.renderPanelBasicDetail(a.filtered[a.selectedIdx], innerW), detailInnerH, a.detailScrollOffset)
	}
	detailPanel := renderTitledPanel(detailTitle, "", detailContent, w, detailPanelH)

	// ── Panel 3: Search / Status (merged, full width) ──

	var searchContent string
	if a.importingPath {
		searchContent = " Import path: " + a.importInput.View()
	} else if a.searching {
		searchContent = " " + a.searchInput.View()
	} else {
		if a.filterQuery != "" {
			promptStyle := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
			queryStyle := lipgloss.NewStyle().Foreground(ui.ColorDetailValue)
			searchContent = " " + promptStyle.Render("❯ ") + queryStyle.Render(a.filterQuery)
		} else {
			searchContent = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(" Press / to search or filter...")
		}
	}

	statusContent := a.status
	if statusContent == "" {
		statusContent = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(" Ready")
	} else {
		statusContent = " " + statusContent
	}
	settingsLine := a.renderInstallSettings()
	infoContent := searchContent + "\n" + statusContent + "\n" + settingsLine
	infoPanel := renderTitledPanel("Search / Status", "", infoContent, w, infoRowH)

	// ── Panel 4: Keys (full width) ──

	keysH := a.keysRowH()
	helpText := a.help.View(a.keys)
	keysPanel := renderTitledPanel("Keys", "", helpText, w, keysH)

	// ── Assemble ──

	page := tabBar + "\n\n" + infoPanel + "\n" + listPanel + "\n" + detailPanel + "\n" + keysPanel

	if a.importConfirm {
		page = a.applyImportConfirmOverlay(page, w)
	}
	if a.removeConfirm {
		page = a.applyRemoveConfirmOverlay(page, w)
	}

	return page
}

func (a App) renderFetchView(w int) string {
	header := components.RenderFetchHeader(a.fetchDistro)
	var statusParts []string
	counterStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	sel := len(a.fetchSelected)
	total := len(a.fetchMirrors)
	statusParts = append(statusParts, counterStyle.Render(fmt.Sprintf("  %d/%d mirrors selected", sel, total)))

	sep := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Render(strings.Repeat("─", w))
	statusParts = append(statusParts, sep)

	if !a.fetchTesting && len(a.fetchMirrors) > 0 && a.fetchIdx < len(a.fetchMirrors) {
		m := a.fetchMirrors[a.fetchIdx]
		lbl := lipgloss.NewStyle().Foreground(ui.ColorWhite).Bold(true).Width(14).Align(lipgloss.Right)
		sepChar := lipgloss.NewStyle().Foreground(ui.ColorMuted)
		val := lipgloss.NewStyle().Foreground(ui.ColorWhite)

		var detail strings.Builder
		fmt.Fprintf(&detail, "  %s %s %s\n", lbl.Render("URL"), sepChar.Render(":"), val.Render(m.URL))
		fmt.Fprintf(&detail, "  %s %s %s\n", lbl.Render("Latency"), sepChar.Render(":"), val.Render(fetch.FormatLatency(m.Latency)))
		fmt.Fprintf(&detail, "  %s %s %d\n", lbl.Render("Score"), sepChar.Render(":"), m.Score)
		statusParts = append(statusParts, detail.String())
	}

	helpLine := components.RenderFetchHelp()
	statusParts = append(statusParts, lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(helpLine))

	statusBarView := lipgloss.JoinVertical(lipgloss.Left, statusParts...)
	statusBarLines := strings.Count(statusBarView, "\n") + 1

	var upperView string
	if a.fetchTesting {
		progress := components.RenderFetchProgress(a.fetchTested, a.fetchTotal)
		progLine := fmt.Sprintf("%s %s", a.spinner.View(), progress)

		centeredProg := lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(progLine)

		headerLines := strings.Count(header, "\n") + 1
		availLines := a.height - headerLines - statusBarLines
		if availLines < 1 {
			availLines = 1
		}
		topPad := (availLines - 1) / 2
		if topPad < 0 {
			topPad = 0
		}

		upperView = header + "\n"
		upperView += strings.Repeat("\n", topPad)
		upperView += centeredProg + "\n"
		rem := availLines - topPad - 1
		if rem > 0 {
			upperView += strings.Repeat("\n", rem)
		}
	} else {
		listView := components.RenderMirrorList(a.fetchMirrors, a.fetchIdx, a.fetchOffset, a.packageListHeight(), w, a.fetchSelected)
		upperView = header + "\n" + listView
	}

	listLines := strings.Count(upperView, "\n")
	gap := a.height - listLines - statusBarLines
	if gap < 0 {
		gap = 0
	}

	return upperView + strings.Repeat("\n", gap) + statusBarView
}

func (a App) renderPPAView(w int, tabBar string) string {
	// Status bar + help
	var statusParts []string
	statusParts = append(statusParts, components.RenderStatusBar(a.status, w))
	helpLine := components.RenderPPAHelp()
	statusParts = append(statusParts, lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(helpLine))
	statusBarView := lipgloss.JoinVertical(lipgloss.Left, statusParts...)
	statusBarLines := strings.Count(statusBarView, "\n") + 1

	tabBarLines := strings.Count(tabBar, "\n") + 1
	panelH := a.height - tabBarLines - 2 - statusBarLines
	if panelH < 7 {
		panelH = 7
	}

	leftW := w / 2
	rightW := w - leftW
	innerLW := leftW - 2
	innerH := panelH - 2

	counterStyle := lipgloss.NewStyle().Foreground(ui.ColorSubtle)
	counterText := counterStyle.Render(fmt.Sprintf("%d repo(s)", len(a.ppaItems)))

	// Left panel: repo list
	var listContent string
	if a.loading {
		pad := innerH / 2
		loadingLine := fmt.Sprintf("Loading repositories %s", a.spinner.View())
		centered := lipgloss.NewStyle().Width(innerLW).Align(lipgloss.Center).Render(loadingLine)
		listContent = strings.Repeat("\n", pad) + centered
	} else {
		maxVisible := innerH - 1
		if maxVisible < 3 {
			maxVisible = 3
		}
		listContent = components.RenderPPAList(a.ppaItems, a.ppaIdx, a.ppaOffset, maxVisible, innerLW)
	}
	leftPanel := renderTitledPanel("Repositories", counterText, listContent, leftW, panelH)

	innerRW := rightW - 2
	var detailContent string
	if !a.loading && len(a.ppaItems) > 0 && a.ppaIdx < len(a.ppaItems) {
		p := a.ppaItems[a.ppaIdx]
		labelW := 12
		lbl := lipgloss.NewStyle().Foreground(ui.ColorWhite).Bold(true).Width(labelW).Align(lipgloss.Left)
		sepChar := lipgloss.NewStyle().Foreground(ui.ColorMuted)
		valStyle := lipgloss.NewStyle().Foreground(ui.ColorWhite)

		// prefix: 1 space + label(12) + space + colon + space = 16
		maxValW := innerRW - 16
		if maxValW < 10 {
			maxValW = 10
		}
		renderVal := func(s string, st lipgloss.Style) string {
			if lipgloss.Width(s) <= maxValW {
				return st.Render(s)
			}
			// wrap long values across multiple lines
			var lines []string
			indent := strings.Repeat(" ", 16)
			for len(s) > 0 {
				chunk := s
				if len(chunk) > maxValW {
					chunk = s[:maxValW]
				}
				lines = append(lines, st.Render(chunk))
				s = s[len(chunk):]
			}
			return strings.Join(lines, "\n"+indent)
		}

		var detail strings.Builder
		fmt.Fprintf(&detail, " %s %s %s\n", lbl.Render("Name"), sepChar.Render(":"), renderVal(p.Name, valStyle))
		fmt.Fprintf(&detail, " %s %s %s\n", lbl.Render("URL"), sepChar.Render(":"), renderVal(p.URL, valStyle))
		repoType := "Standard"
		if p.IsPPA {
			repoType = "PPA"
		}
		fmt.Fprintf(&detail, " %s %s %s\n", lbl.Render("Type"), sepChar.Render(":"), renderVal(repoType, valStyle))
		status := "Enabled"
		stStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
		if !p.Enabled {
			status = "Disabled"
			stStyle = lipgloss.NewStyle().Foreground(ui.ColorDanger).Bold(true)
		}
		fmt.Fprintf(&detail, " %s %s %s\n", lbl.Render("Status"), sepChar.Render(":"), renderVal(status, stStyle))
		fmt.Fprintf(&detail, " %s %s %s\n", lbl.Render("File"), sepChar.Render(":"), renderVal(p.File, valStyle))
		detailContent = detail.String()
	}

	// Input line for adding PPA
	if a.ppaAdding {
		if detailContent != "" {
			detailContent += "\n"
		}
		detailContent += " " + a.ppaInput.View()
	}

	rightPanel := renderTitledPanel("Repo Detail", "", detailContent, rightW, panelH)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	panelLines := strings.Count(panels, "\n")
	gap := a.height - tabBarLines - 2 - panelLines - statusBarLines
	if gap < 0 {
		gap = 0
	}

	return tabBar + "\n\n" + panels + strings.Repeat("\n", gap) + statusBarView
}

func (a App) renderTransactionView(w int, tabBar string) string {
	var statusParts []string
	counterStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	statusParts = append(statusParts, components.RenderStatusBar(a.status, w))
	statusParts = append(statusParts, counterStyle.Render(fmt.Sprintf("%d transactions | z undo | x redo ", len(a.transactionItems))))
	statusParts = append(statusParts, ui.HelpStyle.Render(a.help.View(a.keys)))
	statusBarView := lipgloss.JoinVertical(lipgloss.Left, statusParts...)
	statusBarLines := strings.Count(statusBarView, "\n") + 1

	tabBarLines := strings.Count(tabBar, "\n") + 1
	panelH := a.height - tabBarLines - 2 - statusBarLines
	if panelH < 7 {
		panelH = 7
	}
	leftW := w / 2
	rightW := w - leftW
	innerH := panelH - 2
	innerLW := leftW - 2
	innerRW := rightW - 2

	maxVisible := innerH - 1
	if maxVisible < 3 {
		maxVisible = 3
	}
	listContent := components.RenderTransactionList(a.transactionItems, a.transactionIdx, a.transactionOffset, maxVisible, innerLW)

	txCountText := lipgloss.NewStyle().Foreground(ui.ColorSubtle).Render(fmt.Sprintf("%d", len(a.transactionItems)))
	leftPanel := renderTitledPanel("Transactions", txCountText, listContent, leftW, panelH)

	detailContent := ""
	if len(a.transactionItems) > 0 && a.transactionIdx < len(a.transactionItems) {
		tx := a.transactionItems[a.transactionIdx]
		detailContent = components.RenderTransactionDetail(tx, a.transactionDeps, innerRW, innerH)
	}
	rightPanel := renderTitledPanel("Details", "", detailContent, rightW, panelH)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	panelLines := strings.Count(panels, "\n")
	gap := a.height - tabBarLines - 2 - panelLines - statusBarLines
	if gap < 0 {
		gap = 0
	}

	return tabBar + "\n\n" + panels + strings.Repeat("\n", gap) + statusBarView
}

func (a App) renderErrorLogTab(w int, tabBar string) string {
	var statusParts []string
	statusParts = append(statusParts, components.RenderStatusBar(a.status, w))
	statusParts = append(statusParts, ui.HelpStyle.Render(a.help.View(a.keys)))
	statusBarView := lipgloss.JoinVertical(lipgloss.Left, statusParts...)
	statusBarLines := strings.Count(statusBarView, "\n") + 1

	tabBarLines := strings.Count(tabBar, "\n") + 1
	panelH := a.height - tabBarLines - 2 - statusBarLines
	if panelH < 7 {
		panelH = 7
	}
	leftW := w / 2
	rightW := w - leftW

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary)

	innerH := panelH - 2
	innerLW := leftW - 2
	innerRW := rightW - 2

	maxVisible := innerH - 1
	if maxVisible < 3 {
		maxVisible = 3
	}
	listContent := components.RenderErrorLogList(a.errlogItems, a.errlogIdx, a.errlogOffset, maxVisible, innerLW)
	if lines := strings.Split(listContent, "\n"); len(lines) > innerH {
		listContent = strings.Join(lines[:innerH], "\n")
	}
	leftPanel := clampBorderedPanel(borderStyle.Width(leftW).Height(panelH).Render(listContent), panelH)

	detailTitleStyle := lipgloss.NewStyle().Bold(true).
		Foreground(ui.ColorWhite).Background(ui.ColorDanger).
		Width(innerRW).Padding(0, 1)
	detailTitle := detailTitleStyle.Render("Error Details")

	detailContent := ""
	if len(a.errlogItems) > 0 && a.errlogIdx < len(a.errlogItems) {
		entry := a.errlogItems[a.errlogIdx]
		detailContent = "\n" + components.RenderErrorLogDetail(entry, innerRW)
	}
	rightContent := detailTitle + detailContent
	if lines := strings.Split(rightContent, "\n"); len(lines) > innerH {
		rightContent = strings.Join(lines[:innerH], "\n")
	}
	rightPanel := clampBorderedPanel(borderStyle.Width(rightW).Height(panelH).Render(rightContent), panelH)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	upperView := tabBar + "\n\n" + panels
	upperLines := strings.Count(upperView, "\n")
	gap := a.height - upperLines - statusBarLines
	if gap < 0 {
		gap = 0
	}

	return upperView + strings.Repeat("\n", gap) + statusBarView
}

// clampBorderedPanel ensures a bordered panel has at most maxLines lines,
// preserving the bottom border when content wraps cause overflow.
func clampBorderedPanel(panel string, maxLines int) string {
	lines := strings.Split(panel, "\n")
	if len(lines) <= maxLines {
		return panel
	}
	result := make([]string, 0, maxLines)
	result = append(result, lines[:maxLines-1]...)
	result = append(result, lines[len(lines)-1])
	return strings.Join(result, "\n")
}

func (a App) renderInstallSettings() string {
	onStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
	offStyle := lipgloss.NewStyle().Foreground(ui.ColorDanger).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(ui.ColorSubtle)

	recState := offStyle.Render("OFF")
	if a.installRecommends {
		recState = onStyle.Render("ON")
	}
	sugState := offStyle.Render("OFF")
	if a.installSuggests {
		sugState = onStyle.Render("ON")
	}

	return " " + labelStyle.Render("Recommends: ") + recState + labelStyle.Render("  Suggests: ") + sugState
}

func (a App) renderSideBySide(w int, tabBar string) string {
	leftW := a.sideListWidth()
	rightW := a.sideDetailWidth()
	panelH := a.sideMainPanelHeight()
	innerLW := leftW - 2
	innerRW := rightW - 2
	innerH := panelH - 2

	// ── Row 1: Package List (left) + Package Detail (right) ──

	counterStyle := lipgloss.NewStyle().Foreground(ui.ColorSubtle)
	pos := a.selectedIdx + 1
	if len(a.filtered) == 0 {
		pos = 0
	}
	counterText := fmt.Sprintf("%d/%d", pos, len(a.filtered))

	var listContent string
	if a.loading {
		pad := innerH / 2
		loadingLine := fmt.Sprintf("Loading packages %s", a.spinner.View())
		centered := lipgloss.NewStyle().Width(innerLW).Align(lipgloss.Center).Render(loadingLine)
		listContent = strings.Repeat("\n", pad) + centered
	} else {
		maxVisible := a.packageListHeight()
		si := a.effectiveSortInfo()
		listContent = components.RenderPackageList(a.filtered, a.selectedIdx, a.scrollOffset, maxVisible, innerLW, a.selected, si)
	}
	leftPanel := renderTitledPanel("Package List", counterStyle.Render(counterText), listContent, leftW, panelH)

	rightTitle := "Package Detail"
	if a.fileListActive {
		rightTitle = fmt.Sprintf("Files: %s", a.fileListPkg)
	}

	var detailContent string
	if a.fileListActive {
		detailContent = a.renderPanelFileList(innerRW, innerH)
	} else if !a.loading && len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) && a.detailName != "" && a.detailInfo != "" {
		detailContent = scrollDetailView(components.RenderPackageDetail(enrichedDetailInfo(a.filtered[a.selectedIdx], a.detailInfo), innerRW, 0, 1), innerH, a.detailScrollOffset)
	} else if !a.loading && len(a.filtered) > 0 && a.selectedIdx < len(a.filtered) {
		detailContent = scrollDetailView(a.renderPanelBasicDetail(a.filtered[a.selectedIdx], innerRW), innerH, a.detailScrollOffset)
	}
	rightPanel := renderTitledPanel(rightTitle, "", detailContent, rightW, panelH)

	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// ── Row 2: Search/Filter (left) + Status (right) ──

	var searchContent string
	if a.importingPath {
		searchContent = " Import path: " + a.importInput.View()
	} else if a.searching {
		searchContent = " " + a.searchInput.View()
	} else {
		if a.filterQuery != "" {
			promptStyle := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
			queryStyle := lipgloss.NewStyle().Foreground(ui.ColorDetailValue)
			searchContent = " " + promptStyle.Render("❯ ") + queryStyle.Render(a.filterQuery)
		} else {
			searchContent = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(" Press / to search or filter...")
		}
	}
	searchPanel := renderTitledPanel("Search / Filter", "", searchContent, leftW, infoRowH)

	statusContent := a.status
	if statusContent == "" {
		statusContent = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(" Ready")
	} else {
		statusContent = " " + statusContent
	}
	settingsLine := a.renderInstallSettings()
	statusInner := statusContent + "\n" + settingsLine
	statusPanel := renderTitledPanel("Status", "", statusInner, rightW, infoRowH)

	infoRow := lipgloss.JoinHorizontal(lipgloss.Top, searchPanel, statusPanel)

	// ── Row 3: Keys (full width) ──

	keysH := a.keysRowH()
	helpText := a.help.View(a.keys)
	keysPanel := renderTitledPanel("Keys", "", helpText, w, keysH)

	// ── Assemble ──

	page := tabBar + "\n\n" + infoRow + "\n" + mainRow + "\n" + keysPanel

	// Apply modal overlays (import confirm, remove confirm)
	if a.importConfirm {
		page = a.applyImportConfirmOverlay(page, w)
	}
	if a.removeConfirm {
		page = a.applyRemoveConfirmOverlay(page, w)
	}

	return page
}

// renderTitledPanel renders a bordered panel with the title embedded in the
// top border line, lazygit-style: ╭─ Title ──── rightText ─╮
func renderTitledPanel(title string, rightText string, content string, width int, height int) string {
	bc := ui.ColorDim
	borderChar := lipgloss.NewStyle().Foreground(bc)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)

	innerW := width - 2

	// Build top border: ╭─ Title ──── rightText ─╮
	renderedTitle := titleStyle.Render(title)
	titleW := lipgloss.Width(renderedTitle)

	var topContent string
	if rightText != "" {
		rightW := lipgloss.Width(rightText)
		fillW := innerW - titleW - rightW - 5 // "─ "(2) + " "(1) + " "(1) + "─"(1) = 5
		if fillW < 1 {
			fillW = 1
		}
		topContent = borderChar.Render("─ ") + renderedTitle + borderChar.Render(" "+strings.Repeat("─", fillW)+" ") + rightText + borderChar.Render("─")
	} else {
		fillW := innerW - titleW - 3 // "─ " before + " " after + rest dashes
		if fillW < 1 {
			fillW = 1
		}
		topContent = borderChar.Render("─ ") + renderedTitle + borderChar.Render(" "+strings.Repeat("─", fillW))
	}

	topLine := borderChar.Render("╭") + topContent + borderChar.Render("╮")
	bottomLine := borderChar.Render("╰") + borderChar.Render(strings.Repeat("─", innerW)) + borderChar.Render("╯")

	// Build content lines with side borders
	contentLines := strings.Split(content, "\n")
	maxContentLines := height - 2 // minus top + bottom border
	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}

	var b strings.Builder
	b.WriteString(topLine + "\n")
	for i := 0; i < maxContentLines; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Truncate line if it exceeds inner width, then pad
		lineW := lipgloss.Width(line)
		if lineW > innerW {
			line = ansi.Truncate(line, innerW, "…")
			lineW = lipgloss.Width(line)
		}
		pad := innerW - lineW
		if pad < 0 {
			pad = 0
		}
		b.WriteString(borderChar.Render("│") + line + strings.Repeat(" ", pad) + borderChar.Render("│") + "\n")
	}
	b.WriteString(bottomLine)

	return b.String()
}

// renderPanelBasicDetail renders basic package detail for a bordered panel
// with narrower label width to fit the panel.
func (a App) renderPanelBasicDetail(pkg model.Package, maxW int) string {
	labelW := 12
	lbl := lipgloss.NewStyle().
		Foreground(ui.ColorWhite).Bold(true).Width(labelW).Align(lipgloss.Left)
	sepStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted)
	val := lipgloss.NewStyle().Foreground(ui.ColorWhite)

	// prefix: 1 space + label + space + colon + space = labelW + 4
	prefixW := labelW + 4
	maxValW := maxW - prefixW
	if maxValW < 10 {
		maxValW = 10
	}
	wrapVal := func(s string, style lipgloss.Style) string {
		if lipgloss.Width(s) <= maxValW {
			return style.Render(s)
		}
		// Word-wrap long values
		var lines []string
		for len(s) > 0 {
			if len(s) <= maxValW {
				lines = append(lines, s)
				break
			}
			cut := maxValW
			// Try to break at a space
			if idx := strings.LastIndex(s[:cut], " "); idx > 0 {
				cut = idx
			}
			lines = append(lines, s[:cut])
			s = strings.TrimLeft(s[cut:], " ")
		}
		indent := strings.Repeat(" ", prefixW)
		var result []string
		for i, l := range lines {
			if i == 0 {
				result = append(result, style.Render(l))
			} else {
				result = append(result, indent+style.Render(l))
			}
		}
		return strings.Join(result, "\n")
	}

	var b strings.Builder
	fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Name"), sepStyle.Render(":"), wrapVal(pkg.Name, val))
	fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Version"), sepStyle.Render(":"), wrapVal(pkg.Version, val))

	status := "Not installed"
	statusStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	if pkg.Held {
		status = "Held"
		statusStyle = lipgloss.NewStyle().Foreground(ui.ColorHeld).Bold(true)
	} else if pkg.Upgradable {
		status = "Upgrade available"
		statusStyle = lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true)
	} else if pkg.Installed {
		status = "Installed"
		statusStyle = lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
	}
	fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Status"), sepStyle.Render(":"), wrapVal(status, statusStyle))

	if pkg.NewVersion != "" {
		fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("New Version"), sepStyle.Render(":"),
			wrapVal(pkg.NewVersion, lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true)))
	}
	if pkg.Section != "" {
		fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Section"), sepStyle.Render(":"), wrapVal(pkg.Section, val))
	}
	if pkg.Architecture != "" {
		fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Architecture"), sepStyle.Render(":"), wrapVal(pkg.Architecture, val))
	}
	if pkg.Description != "" {
		fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Description"), sepStyle.Render(":"), wrapVal(pkg.Description, val))
	}

	return b.String()
}

// renderPanelFileList renders the file list for a bordered panel.
func (a App) renderPanelFileList(maxW int, maxLines int) string {
	end := a.fileListOffset + maxLines
	if end > len(a.fileListItems) {
		end = len(a.fileListItems)
	}
	visible := a.fileListItems[a.fileListOffset:end]

	selectedStyle := lipgloss.NewStyle().Background(ui.ColorSelectedBG).Foreground(ui.ColorWhite)
	normalStyle := lipgloss.NewStyle().Foreground(ui.ColorNormalText)

	var b strings.Builder
	for i, file := range visible {
		absIdx := a.fileListOffset + i
		line := fmt.Sprintf("  %s", file)
		if maxW > 5 && len(line) > maxW-2 {
			line = line[:maxW-5] + "..."
		}
		if absIdx == a.fileListIdx {
			b.WriteString(selectedStyle.Render(lipgloss.NewStyle().Width(maxW).Render(line)))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}
	return b.String()
}
