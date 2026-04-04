package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mexirica/aptui/internal/apt"
	"github.com/mexirica/aptui/internal/ui"
)

func RenderPPAList(ppas []apt.PPA, selected int, offset int, maxVisible int, width int) string {
	ppaNameStyle := lipgloss.NewStyle().Foreground(ui.ColorInfo).Bold(true)
	ppaURLStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	ppaEnabledStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
	ppaDisabledStyle := lipgloss.NewStyle().Foreground(ui.ColorDanger).Bold(true)
	ppaHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorColumnHeader)
	ppaDimStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	cursorSt := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)

	if len(ppas) == 0 {
		return lipgloss.NewStyle().Foreground(ui.ColorSecondary).
			Render("\n  No repositories found.\n  Press 'a' to add a PPA.\n")
	}

	// prefix: cursor(3) + space(1) = 4
	prefixW := 4
	colGap := 2
	colType := 6
	// badge(2) + space(1) = 3 extra before name
	badgeW := 3
	available := width - prefixW - colType - badgeW - colGap*2
	if available < 40 {
		available = 40
	}
	// Proportional: Name ~40%, URL ~60%
	colName := available * 40 / 100
	colURL := available - colName
	if colName < 20 {
		colName = 20
	}
	if colURL < 20 {
		colURL = 20
	}

	var b strings.Builder

	padType := colType - 4 // "Type" = 4 chars
	padName := colName - 4 // "Name" = 4 chars
	if padType < 0 {
		padType = 0
	}
	if padName < 0 {
		padName = 0
	}
	header := fmt.Sprintf("%s%s%s%s%s%s",
		strings.Repeat(" ", prefixW),
		ppaHeaderStyle.Render("Type"), strings.Repeat(" ", padType+colGap),
		strings.Repeat(" ", badgeW),
		ppaHeaderStyle.Render("Name"), strings.Repeat(" ", padName+colGap))
	header += ppaHeaderStyle.Render("URL")
	b.WriteString(header + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorPrimary).Render(strings.Repeat("─", width)) + "\n")

	end := offset + maxVisible
	if end > len(ppas) {
		end = len(ppas)
	}

	for i := offset; i < end; i++ {
		p := ppas[i]

		badge := "✔"
		badgeStyle := ppaEnabledStyle
		if !p.Enabled {
			badge = "✘"
			badgeStyle = ppaDisabledStyle
		}
		renderedBadge := badgeStyle.Render(badge)

		typeStr := "repo"
		if p.IsPPA {
			typeStr = "PPA"
		}

		nameStr := p.Name
		nameRunes := []rune(nameStr)
		if len(nameRunes) > colName {
			nameStr = string(nameRunes[:colName-1]) + "…"
		}

		urlStr := p.URL
		urlRunes := []rune(urlStr)
		if len(urlRunes) > colURL {
			urlStr = string(urlRunes[:colURL-1]) + "…"
		}

		typePad := colType - lipgloss.Width(typeStr)
		if typePad < 0 {
			typePad = 0
		}
		namePad := colName - lipgloss.Width(nameStr)
		if namePad < 0 {
			namePad = 0
		}

		if i == selected {
			cursor := cursorSt.Render("  ▌")
			row := fmt.Sprintf("%s %s%s%s %s%s%s\n",
				cursor,
				ppaNameStyle.Render(typeStr), strings.Repeat(" ", typePad+colGap),
				renderedBadge,
				ppaNameStyle.Render(nameStr), strings.Repeat(" ", namePad+colGap),
				ppaURLStyle.Render(urlStr))
			b.WriteString(row)
		} else {
			row := fmt.Sprintf("    %s%s%s %s%s%s\n",
				ppaDimStyle.Render(typeStr), strings.Repeat(" ", typePad+colGap),
				renderedBadge,
				ppaDimStyle.Render(nameStr), strings.Repeat(" ", namePad+colGap),
				ppaDimStyle.Render(urlStr))
			b.WriteString(row)
		}
	}

	return b.String()
}

func RenderPPAHelp() string {
	return "a: add PPA • r: remove PPA • e: enable/disable • esc: back • q: quit"
}
