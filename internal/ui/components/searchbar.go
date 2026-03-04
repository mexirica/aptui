package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var searchPromptStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#7D56F4")).
	Bold(true)

var searchQueryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#D0D0E0"))

func RenderSearchPrompt(query string, focused bool) string {
	cursor := ""
	if focused {
		cursor = "█"
	}
	prompt := searchPromptStyle.Render("❯ ")
	q := searchQueryStyle.Render(query + cursor)
	return fmt.Sprintf("  %s%s", prompt, q)
}
