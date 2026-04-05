package ui

import (
	"testing"
)

func TestApplyThemeDark(t *testing.T) {
	ApplyTheme(true)

	// Verify all styles render without panic after dark theme applied
	tests := []struct {
		name   string
		render func() string
	}{
		{name: "TitleStyle", render: func() string { return TitleStyle.Render("test") }},
		{name: "TabActiveStyle", render: func() string { return TabActiveStyle.Render("tab") }},
		{name: "TabInactiveStyle", render: func() string { return TabInactiveStyle.Render("tab") }},
		{name: "TabNotifyStyle", render: func() string { return TabNotifyStyle.Render("!") }},
		{name: "StatusBarStyle", render: func() string { return StatusBarStyle.Render("status") }},
		{name: "InstalledBadge", render: func() string { return InstalledBadge.String() }},
		{name: "NotInstalledBadge", render: func() string { return NotInstalledBadge.String() }},
		{name: "UpgradableBadge", render: func() string { return UpgradableBadge.String() }},
		{name: "PackageNameStyle", render: func() string { return PackageNameStyle.Render("pkg") }},
		{name: "PackageVersionStyle", render: func() string { return PackageVersionStyle.Render("1.0") }},
		{name: "PackageDescStyle", render: func() string { return PackageDescStyle.Render("desc") }},
		{name: "SelectedItemStyle", render: func() string { return SelectedItemStyle.Render("sel") }},
		{name: "DetailLabelStyle", render: func() string { return DetailLabelStyle.Render("label") }},
		{name: "DetailValueStyle", render: func() string { return DetailValueStyle.Render("value") }},
		{name: "HelpStyle", render: func() string { return HelpStyle.Render("help") }},
		{name: "BoxStyle", render: func() string { return BoxStyle.Render("box") }},
		{name: "ErrorStyle", render: func() string { return ErrorStyle.Render("err") }},
		{name: "SuccessStyle", render: func() string { return SuccessStyle.Render("ok") }},
		{name: "WarningStyle", render: func() string { return WarningStyle.Render("warn") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.render()
			if result == "" {
				t.Errorf("%s.Render() returned empty string", tt.name)
			}
		})
	}
}

func TestApplyThemeLight(t *testing.T) {
	ApplyTheme(false)

	tests := []struct {
		name   string
		render func() string
	}{
		{name: "TitleStyle", render: func() string { return TitleStyle.Render("test") }},
		{name: "TabActiveStyle", render: func() string { return TabActiveStyle.Render("tab") }},
		{name: "TabInactiveStyle", render: func() string { return TabInactiveStyle.Render("tab") }},
		{name: "StatusBarStyle", render: func() string { return StatusBarStyle.Render("status") }},
		{name: "ErrorStyle", render: func() string { return ErrorStyle.Render("err") }},
		{name: "SuccessStyle", render: func() string { return SuccessStyle.Render("ok") }},
		{name: "WarningStyle", render: func() string { return WarningStyle.Render("warn") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.render()
			if result == "" {
				t.Errorf("%s.Render() returned empty string after light theme", tt.name)
			}
		})
	}

	// Reset to dark for other tests
	ApplyTheme(true)
}

func TestApplyThemeBothThemesNotPanic(t *testing.T) {
	themes := []struct {
		name      string
		hasDarkBG bool
	}{
		{name: "dark theme", hasDarkBG: true},
		{name: "light theme", hasDarkBG: false},
	}
	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			ApplyTheme(tt.hasDarkBG)
			// Just verify no panic
		})
	}

	// Reset to dark
	ApplyTheme(true)
}
