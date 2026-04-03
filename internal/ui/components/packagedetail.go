// Package components provides UI components for the package manager.
package components

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mexirica/aptui/internal/ui"
)

var displayFields = []string{
	"Package",
	"Status",
	"Version",
	"Section",
	"Installed-Size",
	"Maintainer",
	"Architecture",
	"Depends",
	"Description",
	"Homepage",
}

func extractFirstEntry(info string) string {
	lines := strings.Split(info, "\n")
	var result []string
	for _, line := range lines {
		if line == "" && len(result) > 0 {
			break
		}
		if line != "" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func parseFields(info string) map[string]string {
	first := extractFirstEntry(info)
	fields := make(map[string]string)
	lines := strings.Split(first, "\n")
	var lastKey string
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if lastKey != "" {
				fields[lastKey] += " " + strings.TrimSpace(line)
			}
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if strings.HasPrefix(key, "Description-") && key != "Description-md5" {
				if _, exists := fields["Description"]; !exists {
					key = "Description"
				}
			}
			fields[key] = val
			lastKey = key
		}
	}
	return fields
}

func RenderPackageDetail(info string, width int, maxLines int, pageNum int) string {
	labelW := 15
	detailLabel := lipgloss.NewStyle().
		Foreground(ui.ColorDetailLabel).
		Bold(true).
		Width(labelW).
		Align(lipgloss.Left)

	detailSep := lipgloss.NewStyle().
		Foreground(ui.ColorDetailSep)

	detailValue := lipgloss.NewStyle().
		Foreground(ui.ColorDetailValue)

	detailMuted := lipgloss.NewStyle().
		Foreground(ui.ColorDim)

	if info == "" {
		return detailMuted.Render(" No package selected.")
	}

	fields := parseFields(info)

	// prefix: 1 space + label + space + colon + space = labelW + 4
	prefixW := labelW + 4
	maxValW := width - prefixW
	if maxValW < 20 {
		maxValW = 20
	}

	var rendered []string

	for _, key := range displayFields {
		val, ok := fields[key]
		if !ok || val == "" {
			val = "N/A"
		}

		// Wrap long values instead of truncating
		wrapValue := func(display string, style lipgloss.Style) []string {
			if len(display) <= maxValW {
				return []string{style.Render(display)}
			}
			var lines []string
			for len(display) > 0 {
				if len(display) <= maxValW {
					lines = append(lines, display)
					break
				}
				cut := maxValW
				if idx := strings.LastIndex(display[:cut], " "); idx > 0 {
					cut = idx
				}
				lines = append(lines, display[:cut])
				display = strings.TrimLeft(display[cut:], " ")
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
			return result
		}

		var valStyle lipgloss.Style
		switch key {
		case "Installed-Size":
			// apt-cache reports Installed-Size in kB; convert to human-friendly.
			if val != "N/A" && val != "" {
				if n, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64); err == nil && n > 0 {
					switch {
					case n >= 1048576:
						val = fmt.Sprintf("%.1f GB", float64(n)/1048576)
					case n >= 1024:
						val = fmt.Sprintf("%.1f MB", float64(n)/1024)
					default:
						val = fmt.Sprintf("%d kB", n)
					}
				} else {
					val = val + " kB"
				}
			}
			valStyle = detailValue
		case "Homepage":
			if val == "N/A" {
				valStyle = detailMuted
			} else {
				valStyle = lipgloss.NewStyle().Foreground(ui.ColorInfo)
			}
		case "Status":
			statusColor := ui.ColorSecondary
			if strings.Contains(val, "Upgrade") {
				statusColor = ui.ColorWarning
			} else if strings.Contains(val, "Installed") {
				statusColor = ui.ColorSuccess
			}
			valStyle = lipgloss.NewStyle().Foreground(statusColor).Bold(true)
		default:
			valStyle = detailValue
		}

		wrappedLines := wrapValue(val, valStyle)
		firstLine := fmt.Sprintf(" %s %s %s",
			detailLabel.Render(key),
			detailSep.Render(":"),
			wrappedLines[0])
		rendered = append(rendered, firstLine)
		for _, extra := range wrappedLines[1:] {
			rendered = append(rendered, extra)
		}
	}

	if len(rendered) == 0 {
		return detailMuted.Render("  No information available.") + "\n"
	}

	if maxLines > 0 && len(rendered) > maxLines {
		rendered = rendered[:maxLines]
	}

	var b strings.Builder
	for _, l := range rendered {
		b.WriteString(l + "\n")
	}

	return b.String()
}
