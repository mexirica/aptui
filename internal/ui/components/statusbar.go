package components

import (
	"charm.land/lipgloss/v2"
	"github.com/mexirica/aptui/internal/ui"
)

func RenderStatusBar(status string, width int) string {
	statusBarStyle := lipgloss.NewStyle().
		Foreground(ui.ColorSecondary).
		Padding(0, 1)
	return statusBarStyle.Width(width).Render(status)
}
