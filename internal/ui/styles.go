// Package ui contains UI styles and components for the application.
package ui

import (
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
//
// NOTE: This function mutates package-level variables and must only be called
// from within the Bubbletea event loop (i.e. from an Update handler) to avoid
// data races with concurrent goroutines.
func ApplyTheme(hasDarkBG bool) {
	pick := lipgloss.LightDark(hasDarkBG)

	ColorPrimary = pick(lipgloss.Color("#5B3FC4"), lipgloss.Color("#7D56F4"))
	ColorAccent = pick(lipgloss.Color("#7C5FCF"), lipgloss.Color("#A78BFA"))
	ColorSecondary = pick(lipgloss.Color("#888888"), lipgloss.Color("#6C6C6C"))
	ColorSuccess = pick(lipgloss.Color("#038A59"), lipgloss.Color("#04B575"))
	ColorDanger = pick(lipgloss.Color("#D63A5E"), lipgloss.Color("#FF4672"))
	ColorWarning = pick(lipgloss.Color("#CC9A06"), lipgloss.Color("#FFC107"))
	ColorInfo = pick(lipgloss.Color("#0097A7"), lipgloss.Color("#00BCD4"))
	ColorWhite = pick(lipgloss.Color("#1A1A2E"), lipgloss.Color("#FAFAFA"))
	ColorDark = pick(lipgloss.Color("#F0F0F5"), lipgloss.Color("#1A1A2E"))
	ColorMuted = pick(lipgloss.Color("#999999"), lipgloss.Color("#4A4A4A"))
	ColorDim = pick(lipgloss.Color("#AAAABC"), lipgloss.Color("#3A3A4A"))
	ColorSubtle = pick(lipgloss.Color("#555577"), lipgloss.Color("#8888AA"))

	ColorNormalText = pick(lipgloss.Color("#3A3A4A"), lipgloss.Color("#B0B0C0"))
	ColorDetailLabel = pick(lipgloss.Color("#2E1F6F"), lipgloss.Color("#f2edff"))
	ColorDetailSep = pick(lipgloss.Color("#8B6FD4"), lipgloss.Color("#5B3FC4"))
	ColorDetailValue = pick(lipgloss.Color("#2A2A3A"), lipgloss.Color("#D0D0E0"))
	ColorSizeText = pick(lipgloss.Color("#777799"), lipgloss.Color("#6C6C8A"))
	ColorUncheck = pick(lipgloss.Color("#AAAABC"), lipgloss.Color("#4A4A5A"))
	ColorHeld = pick(lipgloss.Color("#CC7000"), lipgloss.Color("#FF8C00"))
	ColorTabInactiveBG = pick(lipgloss.Color("#E8E8F0"), lipgloss.Color("#1E1E2E"))
	ColorStatusBarBG = pick(lipgloss.Color("#D8D8E8"), lipgloss.Color("#333346"))
	ColorSelectedBG = pick(lipgloss.Color("#D0D0E8"), lipgloss.Color("#2A2A5E"))
	ColorHelpSep = pick(lipgloss.Color("#AAAAAA"), lipgloss.Color("#555555"))

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
