package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mexirica/aptui/internal/apt"
	"github.com/mexirica/aptui/internal/ui"
)

// RenderVersionList renders the version selection overlay for a package.
func RenderVersionList(pkgName string, versions []apt.VersionInfo, selected int, offset int, maxVisible int, width int) string {
	if width < 20 {
		width = 20
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorColumnHeader)
	verStyle := lipgloss.NewStyle().Foreground(ui.ColorWhite)
	installedStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
	originStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	cursorSt := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)

	if len(versions) == 0 {
		return lipgloss.NewStyle().Foreground(ui.ColorSecondary).
			Render("\n  No versions available in configured repositories.\n" +
				"  Tip: add PPAs or use Ubuntu Snapshot Service for older versions.\n")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s %s\n",
		headerStyle.Render("Select version for:"),
		lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render(pkgName)))
	b.WriteString(fmt.Sprintf("  %s\n\n",
		dimStyle.Render("(enter to install, esc to cancel)")))

	// Column widths - responsive
	showSource := width > 60
	colVer := width - 6 // cursor(4) + padding(2)
	if showSource {
		colVer = width * 55 / 100 // 55% for version
		if colVer < 25 {
			colVer = 25
		}
	}

	// Header
	if showSource {
		padding := colVer - 7
		if padding < 1 {
			padding = 1
		}
		header := fmt.Sprintf("    %s%s%s",
			headerStyle.Render("Version"),
			strings.Repeat(" ", padding),
			headerStyle.Render("Source"))
		b.WriteString(header + "\n")
	} else {
		b.WriteString("    " + headerStyle.Render("Version") + "\n")
	}

	end := offset + maxVisible
	if end > len(versions) {
		end = len(versions)
	}

	for i := offset; i < end; i++ {
		v := versions[i]

		verStr := v.Version
		maxVer := colVer - 4
		if maxVer < 10 {
			maxVer = 10
		}
		if len(verStr) > maxVer {
			verStr = verStr[:maxVer-1] + "…"
		}

		style := verStyle
		marker := "  "
		if v.Installed {
			style = installedStyle
			marker = "● "
		}

		if i == selected {
			cursor := cursorSt.Render("▌ ")
			if showSource {
				originStr := v.Origin
				maxOrigin := width - colVer - 8
				if maxOrigin < 1 {
					maxOrigin = 1
				}
				if len(originStr) > maxOrigin {
					originStr = originStr[:maxOrigin-1] + "…"
				}
				verPad := colVer - len(verStr) - len(marker)
				if verPad < 0 {
					verPad = 0
				}
				row := fmt.Sprintf("  %s%s%s%s%s\n",
					cursor,
					style.Render(marker),
					style.Render(verStr),
					strings.Repeat(" ", verPad),
					originStyle.Render(originStr))
				b.WriteString(row)
			} else {
				row := fmt.Sprintf("  %s%s%s\n",
					cursor,
					style.Render(marker),
					style.Render(verStr))
				b.WriteString(row)
			}
		} else {
			if showSource {
				originStr := v.Origin
				maxOrigin := width - colVer - 8
				if maxOrigin < 1 {
					maxOrigin = 1
				}
				if len(originStr) > maxOrigin {
					originStr = originStr[:maxOrigin-1] + "…"
				}
				verPad := colVer - len(verStr) - len(marker)
				if verPad < 0 {
					verPad = 0
				}
				row := fmt.Sprintf("    %s%s%s%s\n",
					dimStyle.Render(marker),
					dimStyle.Render(verStr),
					strings.Repeat(" ", verPad),
					dimStyle.Render(originStr))
				b.WriteString(row)
			} else {
				row := fmt.Sprintf("    %s%s\n",
					dimStyle.Render(marker),
					dimStyle.Render(verStr))
				b.WriteString(row)
			}
		}
	}

	return b.String()
}
