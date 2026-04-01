// Package ui contains UI styles and components for the application.
package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorAccent    = lipgloss.Color("#A78BFA") // lighter purple for secondary accents
	ColorSecondary = lipgloss.Color("#6C6C6C")
	ColorSuccess   = lipgloss.Color("#04B575")
	ColorDanger    = lipgloss.Color("#FF4672")
	ColorWarning   = lipgloss.Color("#FFC107")
	ColorInfo      = lipgloss.Color("#00BCD4")
	ColorWhite     = lipgloss.Color("#FAFAFA")
	ColorDark      = lipgloss.Color("#1A1A2E")
	ColorMuted     = lipgloss.Color("#4A4A4A")
	ColorDim       = lipgloss.Color("#3A3A4A") // very dim for N/A, separators
	ColorSubtle    = lipgloss.Color("#8888AA") // soft blue-gray for values

	// Extended semantic colors for themed components
	ColorNormalText    = lipgloss.Color("#B0B0C0") // standard body text
	ColorDetailLabel   = lipgloss.Color("#f2edff") // detail panel label text
	ColorDetailSep     = lipgloss.Color("#5B3FC4") // detail panel separator
	ColorDetailValue   = lipgloss.Color("#D0D0E0") // detail panel value text
	ColorSizeText      = lipgloss.Color("#6C6C8A") // package size text
	ColorUncheck       = lipgloss.Color("#4A4A5A") // unchecked checkbox indicator
	ColorHeld          = lipgloss.Color("#FF8C00") // held package indicator
	ColorTabInactiveBG = lipgloss.Color("#1E1E2E") // tab inactive background
	ColorStatusBarBG   = lipgloss.Color("#333346") // status bar background
	ColorSelectedBG    = lipgloss.Color("#2A2A5E") // selected item background
	ColorHelpSep       = lipgloss.Color("#555555") // help widget separator

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorSubtle).
				Background(ColorTabInactiveBG).
				Padding(0, 2)

	TabNotifyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWarning).
			Background(ColorTabInactiveBG).
			Padding(0, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorStatusBarBG).
			Padding(0, 1)

	InstalledBadge = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true).
			SetString("●")

	NotInstalledBadge = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				SetString("○")

	UpgradableBadge = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true).
			SetString("↑")

	PackageNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite)

	PackageVersionStyle = lipgloss.NewStyle().
				Foreground(ColorInfo)

	PackageDescStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	SelectedItemStyle = lipgloss.NewStyle().
				Background(ColorSelectedBG).
				Foreground(ColorWhite)

	DetailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(ColorWhite)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)
)

// ApplyTheme updates all color and style variables based on the terminal
// background color. Call this when a tea.BackgroundColorMsg is received.
func ApplyTheme(hasDarkBG bool) {
	pick := func(light, dark string) color.Color {
		if hasDarkBG {
			return lipgloss.Color(dark)
		}
		return lipgloss.Color(light)
	}

	ColorPrimary = pick("#5B3FC4", "#7D56F4")
	ColorAccent = pick("#7C5FCF", "#A78BFA")
	ColorSecondary = pick("#888888", "#6C6C6C")
	ColorSuccess = pick("#038A59", "#04B575")
	ColorDanger = pick("#D63A5E", "#FF4672")
	ColorWarning = pick("#CC9A06", "#FFC107")
	ColorInfo = pick("#0097A7", "#00BCD4")
	ColorWhite = pick("#1A1A2E", "#FAFAFA")
	ColorDark = pick("#F0F0F5", "#1A1A2E")
	ColorMuted = pick("#999999", "#4A4A4A")
	ColorDim = pick("#AAAABC", "#3A3A4A")
	ColorSubtle = pick("#555577", "#8888AA")

	ColorNormalText = pick("#3A3A4A", "#B0B0C0")
	ColorDetailLabel = pick("#2E1F6F", "#f2edff")
	ColorDetailSep = pick("#8B6FD4", "#5B3FC4")
	ColorDetailValue = pick("#2A2A3A", "#D0D0E0")
	ColorSizeText = pick("#777799", "#6C6C8A")
	ColorUncheck = pick("#AAAABC", "#4A4A5A")
	ColorHeld = pick("#CC7000", "#FF8C00")
	ColorTabInactiveBG = pick("#E8E8F0", "#1E1E2E")
	ColorStatusBarBG = pick("#D8D8E8", "#333346")
	ColorSelectedBG = pick("#D0D0E8", "#2A2A5E")
	ColorHelpSep = pick("#AAAAAA", "#555555")

	TitleStyle = lipgloss.NewStyle().
		Bold(true).Foreground(ColorWhite).Background(ColorPrimary).
		Padding(0, 2).MarginBottom(1)

	TabActiveStyle = lipgloss.NewStyle().
		Bold(true).Foreground(ColorWhite).Background(ColorPrimary).
		Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(ColorSubtle).Background(ColorTabInactiveBG).
		Padding(0, 2)

	TabNotifyStyle = lipgloss.NewStyle().
		Bold(true).Foreground(ColorWarning).Background(ColorTabInactiveBG).
		Padding(0, 2)

	StatusBarStyle = lipgloss.NewStyle().
		Foreground(ColorWhite).Background(ColorStatusBarBG).
		Padding(0, 1)

	InstalledBadge = lipgloss.NewStyle().
		Foreground(ColorSuccess).Bold(true).SetString("●")

	NotInstalledBadge = lipgloss.NewStyle().
		Foreground(ColorSecondary).SetString("○")

	UpgradableBadge = lipgloss.NewStyle().
		Foreground(ColorWarning).Bold(true).SetString("↑")

	PackageNameStyle = lipgloss.NewStyle().
		Bold(true).Foreground(ColorWhite)

	PackageVersionStyle = lipgloss.NewStyle().
		Foreground(ColorInfo)

	PackageDescStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	SelectedItemStyle = lipgloss.NewStyle().
		Background(ColorSelectedBG).Foreground(ColorWhite)

	DetailLabelStyle = lipgloss.NewStyle().
		Bold(true).Foreground(ColorPrimary)

	DetailValueStyle = lipgloss.NewStyle().
		Foreground(ColorWhite)

	HelpStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(ColorPrimary).
		Padding(1, 2)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorDanger).Bold(true)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).Bold(true)
}
