package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mexirica/aptui/internal/errlog"
	"github.com/mexirica/aptui/internal/ui"
)

func RenderErrorLogList(entries []errlog.Entry, selected int, offset int, maxVisible int, width int) string {
	errIDStyle := lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true)
	errSrcStyle := lipgloss.NewStyle().Foreground(ui.ColorInfo).Bold(true)
	errDateStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	errMsgStyle := lipgloss.NewStyle().Foreground(ui.ColorWhite)
	errDimStyle := lipgloss.NewStyle().Foreground(ui.ColorSecondary)
	errHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)
	cursorSt := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)

	if len(entries) == 0 {
		return lipgloss.NewStyle().Foreground(ui.ColorSecondary).
			Render("\n  No errors logged.\n")
	}

	colID := 6
	colSrc := 16
	colDate := 21
	prefixW := 4
	colMsg := width - prefixW - colID - colSrc - colDate - 8
	if colMsg < 1 {
		colMsg = 1
	}

	var b strings.Builder

	header := fmt.Sprintf("%s%s  %s%s  %s%s  %s",
		strings.Repeat(" ", prefixW),
		errHeaderStyle.Render("ID"), strings.Repeat(" ", colID-2),
		errHeaderStyle.Render("Source"), strings.Repeat(" ", colSrc-6),
		errHeaderStyle.Render("Date"),
		strings.Repeat(" ", colDate-4)+errHeaderStyle.Render("Message"))
	b.WriteString(header + "\n")

	end := offset + maxVisible
	if end > len(entries) {
		end = len(entries)
	}

	for i := offset; i < end; i++ {
		e := entries[i]

		idStr := fmt.Sprintf("#%-4d", e.ID)

		srcStr := e.Source
		if len(srcStr) > colSrc {
			srcStr = srcStr[:colSrc-1] + "…"
		}

		dateStr := errlog.FormatTimestamp(e.Timestamp)

		msgStr := e.Message
		if len(msgStr) > colMsg {
			msgStr = msgStr[:colMsg-1] + "…"
		}

		srcPad := colSrc - len(srcStr)
		if srcPad < 0 {
			srcPad = 0
		}
		datePad := colDate - len(dateStr)
		if datePad < 0 {
			datePad = 0
		}

		if i == selected {
			cursor := cursorSt.Render(" ▌")
			row := fmt.Sprintf("%s %s %s%s  %s%s  %s\n",
				cursor,
				errIDStyle.Render(idStr),
				errSrcStyle.Render(srcStr), strings.Repeat(" ", srcPad),
				errMsgStyle.Render(dateStr), strings.Repeat(" ", datePad),
				errMsgStyle.Render(msgStr))
			b.WriteString(row)
		} else {
			row := fmt.Sprintf("    %s %s%s  %s%s  %s\n",
				errDimStyle.Render(idStr),
				errSrcStyle.Render(srcStr), strings.Repeat(" ", srcPad),
				errDateStyle.Render(dateStr), strings.Repeat(" ", datePad),
				errDimStyle.Render(msgStr))
			b.WriteString(row)
		}
	}

	return b.String()
}

func RenderErrorLogDetail(entry errlog.Entry, width int) string {
	errSrcStyle := lipgloss.NewStyle().Foreground(ui.ColorInfo).Bold(true)
	lbl := lipgloss.NewStyle().
		Foreground(ui.ColorWhite).Bold(true)
	sep := lipgloss.NewStyle().Foreground(ui.ColorMuted)
	val := lipgloss.NewStyle().Foreground(ui.ColorWhite)

	var b strings.Builder
	fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("ID"), sep.Render(":"), val.Render(fmt.Sprintf("#%d", entry.ID)))
	fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Source"), sep.Render(":"), errSrcStyle.Render(entry.Source))
	fmt.Fprintf(&b, " %s %s %s\n", lbl.Render("Date"), sep.Render(":"), val.Render(errlog.FormatTimestamp(entry.Timestamp)))

	// Wrap message
	msgLabel := "Message"
	prefix := fmt.Sprintf(" %s %s ", lbl.Render(msgLabel), sep.Render(":"))
	indent := " " + strings.Repeat(" ", len(msgLabel)+3)
	avail := width - len(msgLabel) - 5
	if avail < 20 {
		avail = 20
	}

	msgLines := wrapText(entry.Message, avail)
	for idx, line := range msgLines {
		if idx == 0 {
			b.WriteString(prefix + val.Render(line) + "\n")
		} else {
			b.WriteString(indent + val.Render(line) + "\n")
		}
	}

	return b.String()
}

func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	runes := []rune(text)
	var lines []string
	for len(runes) > maxWidth {
		lines = append(lines, string(runes[:maxWidth]))
		runes = runes[maxWidth:]
	}
	if len(runes) > 0 {
		lines = append(lines, string(runes))
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}
