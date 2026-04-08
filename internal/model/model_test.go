package model

import (
	"testing"
)

func TestPackageFields(t *testing.T) {
	tests := []struct {
		name string
		pkg  Package
		want struct {
			installed  bool
			upgradable bool
			held       bool
			essential  bool
		}
	}{
		{
			name: "installed package",
			pkg:  Package{Name: "vim", Version: "8.2", Installed: true},
			want: struct {
				installed  bool
				upgradable bool
				held       bool
				essential  bool
			}{true, false, false, false},
		},
		{
			name: "upgradable package",
			pkg:  Package{Name: "git", Version: "2.34", Installed: true, Upgradable: true, NewVersion: "2.40"},
			want: struct {
				installed  bool
				upgradable bool
				held       bool
				essential  bool
			}{true, true, false, false},
		},
		{
			name: "held package",
			pkg:  Package{Name: "curl", Installed: true, Held: true},
			want: struct {
				installed  bool
				upgradable bool
				held       bool
				essential  bool
			}{true, false, true, false},
		},
		{
			name: "essential package",
			pkg:  Package{Name: "base-files", Installed: true, Essential: true},
			want: struct {
				installed  bool
				upgradable bool
				held       bool
				essential  bool
			}{true, false, false, true},
		},
		{
			name: "not installed package",
			pkg:  Package{Name: "nano"},
			want: struct {
				installed  bool
				upgradable bool
				held       bool
				essential  bool
			}{false, false, false, false},
		},
		{
			name: "pinned manually installed package",
			pkg:  Package{Name: "htop", Installed: true, Pinned: true, ManuallyInstalled: true},
			want: struct {
				installed  bool
				upgradable bool
				held       bool
				essential  bool
			}{true, false, false, false},
		},
		{
			name: "package with all metadata",
			pkg: Package{
				Name:         "vim",
				Version:      "8.2.4919",
				Size:         "9.8 MB",
				Description:  "Vi IMproved - enhanced vi editor",
				Installed:    true,
				Upgradable:   true,
				NewVersion:   "9.1.0",
				Section:      "editors",
				Architecture: "amd64",
			},
			want: struct {
				installed  bool
				upgradable bool
				held       bool
				essential  bool
			}{true, true, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pkg.Installed != tt.want.installed {
				t.Errorf("Installed = %v, want %v", tt.pkg.Installed, tt.want.installed)
			}
			if tt.pkg.Upgradable != tt.want.upgradable {
				t.Errorf("Upgradable = %v, want %v", tt.pkg.Upgradable, tt.want.upgradable)
			}
			if tt.pkg.Held != tt.want.held {
				t.Errorf("Held = %v, want %v", tt.pkg.Held, tt.want.held)
			}
			if tt.pkg.Essential != tt.want.essential {
				t.Errorf("Essential = %v, want %v", tt.pkg.Essential, tt.want.essential)
			}
		})
	}
}

func TestKeyMapShortHelp(t *testing.T) {
	tests := []struct {
		name     string
		wantLen  int
		wantKeys []string
	}{
		{
			name:     "returns expected number of bindings",
			wantLen:  5,
			wantKeys: []string{"space", "i", "r", "h", "q"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bindings := Keys.ShortHelp()
			if len(bindings) != tt.wantLen {
				t.Errorf("ShortHelp() returned %d bindings, want %d", len(bindings), tt.wantLen)
			}
		})
	}
}

func TestKeyMapFullHelp(t *testing.T) {
	tests := []struct {
		name        string
		wantGroups  int
		wantNonZero bool
	}{
		{
			name:        "returns multiple groups",
			wantGroups:  6,
			wantNonZero: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := Keys.FullHelp()
			if len(groups) != tt.wantGroups {
				t.Errorf("FullHelp() returned %d groups, want %d", len(groups), tt.wantGroups)
			}
			for i, g := range groups {
				if tt.wantNonZero && len(g) == 0 {
					t.Errorf("group %d is empty", i)
				}
			}
		})
	}
}

func TestKeyBindings(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string
		helpKey  string
		helpDesc string
	}{
		{name: "Quit", keys: []string{"q", "ctrl+c"}, helpKey: "q", helpDesc: "quit"},
		{name: "Help", keys: []string{"h"}, helpKey: "h", helpDesc: "help"},
		{name: "Enter", keys: []string{"enter"}, helpKey: "enter", helpDesc: "confirm"},
		{name: "Search", keys: []string{"/"}, helpKey: "/", helpDesc: "search/filter"},
		{name: "Install", keys: []string{"i"}, helpKey: "i", helpDesc: "install"},
		{name: "Remove", keys: []string{"r"}, helpKey: "r", helpDesc: "remove"},
		{name: "Upgrade", keys: []string{"u"}, helpKey: "u", helpDesc: "upgrade"},
		{name: "Select", keys: []string{" "}, helpKey: "space", helpDesc: "select"},
		{name: "SelectAll", keys: []string{"a"}, helpKey: "a", helpDesc: "select all"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify help text is set
			help := Keys.ShortHelp()
			found := false
			for _, b := range help {
				h := b.Help()
				if h.Key == tt.helpKey && h.Desc == tt.helpDesc {
					found = true
					break
				}
			}
			// Also check FullHelp
			if !found {
				for _, group := range Keys.FullHelp() {
					for _, b := range group {
						h := b.Help()
						if h.Key == tt.helpKey && h.Desc == tt.helpDesc {
							found = true
							break
						}
					}
				}
			}
			if !found {
				t.Errorf("binding %q (key=%q, desc=%q) not found in help", tt.name, tt.helpKey, tt.helpDesc)
			}
		})
	}
}
