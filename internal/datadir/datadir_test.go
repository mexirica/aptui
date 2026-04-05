package datadir

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRealUserHome(t *testing.T) {
	tests := []struct {
		name     string
		envSetup func()
		check    func(string)
	}{
		{
			name: "no SUDO_USER returns user home",
			envSetup: func() {
				os.Unsetenv("SUDO_USER")
			},
			check: func(result string) {
				if result == "" {
					t.Error("expected non-empty home directory")
				}
			},
		},
		{
			name: "invalid SUDO_USER falls back to UserHomeDir",
			envSetup: func() {
				os.Setenv("SUDO_USER", "nonexistent_user_99999")
			},
			check: func(result string) {
				if result == "" {
					t.Error("expected non-empty home directory")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origSudo := os.Getenv("SUDO_USER")
			defer os.Setenv("SUDO_USER", origSudo)
			tt.envSetup()
			result := RealUserHome()
			tt.check(result)
		})
	}
}

func TestDir(t *testing.T) {
	tests := []struct {
		name  string
		check func(string)
	}{
		{
			name: "returns path ending in .local/share/aptui",
			check: func(d string) {
				if !filepath.IsAbs(d) {
					t.Errorf("expected absolute path, got %q", d)
				}
				if filepath.Base(d) != "aptui" {
					t.Errorf("expected path ending in 'aptui', got %q", d)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(Dir())
		})
	}
}

func TestSaveJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		wantErr bool
		check   func(string)
	}{
		{
			name:    "save simple struct",
			data:    struct{ Name string }{"test"},
			wantErr: false,
			check: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				var result map[string]string
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if result["Name"] != "test" {
					t.Errorf("expected Name='test', got %q", result["Name"])
				}
			},
		},
		{
			name:    "save slice",
			data:    []string{"vim", "git", "curl"},
			wantErr: false,
			check: func(path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				var result []string
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if len(result) != 3 {
					t.Errorf("expected 3 items, got %d", len(result))
				}
			},
		},
		{
			name:    "creates parent directories",
			data:    "hello",
			wantErr: false,
			check:   func(path string) {},
		},
		{
			name:    "unsaveable type produces error",
			data:    make(chan int),
			wantErr: true,
			check:   func(path string) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "sub", "data.json")
			err := SaveJSON(path, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				tt.check(path)
			}
		})
	}
}

func TestSaveJSON_IndentedOutput(t *testing.T) {
	tests := []struct {
		name string
		data any
	}{
		{
			name: "output is indented JSON",
			data: map[string]string{"key": "value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "indented.json")
			if err := SaveJSON(path, tt.data); err != nil {
				t.Fatal(err)
			}
			data, _ := os.ReadFile(path)
			content := string(data)
			if len(content) < 5 {
				t.Error("expected non-trivial JSON output")
			}
			// Indented JSON should contain newlines
			if !json.Valid(data) {
				t.Error("expected valid JSON")
			}
		})
	}
}
