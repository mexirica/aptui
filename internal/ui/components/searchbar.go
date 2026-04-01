package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/mexirica/aptui/internal/ui"
)

// RenderQueryPrompt renders the unified search/filter prompt.
func RenderQueryPrompt(query string, focused bool) string {
	queryPromptStyle := lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).
		Bold(true)
	queryTextStyle := lipgloss.NewStyle().
		Foreground(ui.ColorDetailValue)
	cursor := ""
	if focused {
		cursor = "█"
	}
	prompt := queryPromptStyle.Render("❯ ")
	q := queryTextStyle.Render(query + cursor)
	return fmt.Sprintf("  %s%s", prompt, q)
}
