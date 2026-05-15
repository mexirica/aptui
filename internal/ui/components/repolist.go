package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mexirica/aptui/internal/ui"
)

// RenderRepoList renders the repository filter overlay.
func RenderRepoList(items []string, selected int, offset int, maxVisible int, width int, activeFilter string) string {
	if width < 20 {
		width = 20
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorColumnHeader)
	cursorSt := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	activeStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	itemStyle := lipgloss.NewStyle().Foreground(ui.ColorWhite)
	clearStyle := lipgloss.NewStyle().Foreground(ui.ColorWarning)

	if len(items) == 0 {
		return lipgloss.NewStyle().Foreground(ui.ColorSecondary).
			Render("\n  No repository metadata available.\n" +
				"  Tip: run apt update to refresh package lists.\n")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s\n",
		headerStyle.Render("Filter by repository")))
	b.WriteString(fmt.Sprintf("  %s\n\n",
		dimStyle.Render("(enter to filter, esc to cancel)")))

	b.WriteString("    " + headerStyle.Render("Repository") + "\n")

	end := offset + maxVisible
	if end > len(items) {
		end = len(items)
	}

	for i := offset; i < end; i++ {
		label := items[i]
		isClear := label == ""

		// Truncate if too long
		maxLen := width - 8
		if maxLen < 10 {
			maxLen = 10
		}
		display := label
		if isClear {
			display = "  [Clear repo filter]"
		} else if len(display) > maxLen {
			display = display[:maxLen-1] + "…"
		}

		isActive := !isClear && label == activeFilter

		if i == selected {
			cursor := cursorSt.Render("▌ ")
			var rendered string
			if isClear {
				rendered = clearStyle.Render(display)
			} else if isActive {
				rendered = activeStyle.Render("● " + display)
			} else {
				rendered = itemStyle.Render("  " + display)
			}
			b.WriteString(fmt.Sprintf("  %s%s\n", cursor, rendered))
		} else {
			var rendered string
			if isClear {
				rendered = dimStyle.Render(display)
			} else if isActive {
				rendered = activeStyle.Render("● " + display)
			} else {
				rendered = dimStyle.Render("  " + display)
			}
			b.WriteString(fmt.Sprintf("    %s\n", rendered))
		}
	}

	return b.String()
}
