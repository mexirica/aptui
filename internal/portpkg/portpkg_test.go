package portpkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExportAndImport(t *testing.T) {
	t.Run("export and import sorted", func(t *testing.T) {
		tmpDir := t.TempDir()
		orig := DefaultPath
		DefaultPath = func() string { return filepath.Join(tmpDir, "test-packages.json") }
		defer func() { DefaultPath = orig }()

		packages := []PackageEntry{
			{Name: "zsh"},
			{Name: "curl"},
			{Name: "git"},
		}

		path, err := Export(packages)
		if err != nil {
			t.Fatalf("export: %v", err)
		}

		entries, gotPath, err := Import("")
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if gotPath != path {
			t.Errorf("path = %q, want %q", gotPath, path)
		}
		if len(entries) != 3 {
			t.Fatalf("expected 3 packages, got %d", len(entries))
		}
		want := []string{"curl", "git", "zsh"}
		for i, e := range entries {
			if e.Name != want[i] {
				t.Errorf("entries[%d].Name = %q, want %q", i, e.Name, want[i])
			}
		}
	})
}

func TestExportSortsPackages(t *testing.T) {
	t.Run("export sorts packages", func(t *testing.T) {
		tmpDir := t.TempDir()
		orig := DefaultPath
		DefaultPath = func() string { return filepath.Join(tmpDir, "sorted.json") }
		defer func() { DefaultPath = orig }()

		packages := []PackageEntry{
			{Name: "zsh"},
			{Name: "apt"},
			{Name: "curl"},
		}
		path, err := Export(packages)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var ef ExportFile
		if err := json.Unmarshal(data, &ef); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if ef.Packages[0].Name != "apt" || ef.Packages[1].Name != "curl" || ef.Packages[2].Name != "zsh" {
			t.Errorf("packages not sorted: %+v", ef.Packages)
		}
	})
}

func TestImportMissingFile(t *testing.T) {
	t.Run("import missing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		orig := DefaultPath
		DefaultPath = func() string { return filepath.Join(tmpDir, "nonexistent.json") }
		defer func() { DefaultPath = orig }()

		_, _, err := Import("")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestImportInvalidJSON(t *testing.T) {
	t.Run("import invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "bad.json")
		if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
			t.Fatal(err)
		}
		orig := DefaultPath
		DefaultPath = func() string { return path }
		defer func() { DefaultPath = orig }()

		_, _, err := Import("")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestImportValidFile(t *testing.T) {
	t.Run("import valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "valid.json")
		ef := ExportFile{Packages: []PackageEntry{
			{Name: "vim"},
			{Name: "tmux"},
		}}
		data, _ := json.MarshalIndent(ef, "", "  ")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		orig := DefaultPath
		DefaultPath = func() string { return path }
		defer func() { DefaultPath = orig }()

		entries, gotPath, err := Import("")
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if gotPath != path {
			t.Errorf("path = %q, want %q", gotPath, path)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Name != "vim" || entries[1].Name != "tmux" {
			t.Errorf("unexpected entries: %+v", entries)
		}
	})
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name       string
		createFile bool
		expect     bool
	}{
		{name: "file exists", createFile: true, expect: true},
		{name: "file not exists", createFile: false, expect: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.json")
			orig := DefaultPath
			DefaultPath = func() string { return path }
			defer func() { DefaultPath = orig }()

			if tt.createFile {
				if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			got := FileExists()
			if got != tt.expect {
				t.Errorf("FileExists() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestImportWithTildePath(t *testing.T) {
	tmpDir := t.TempDir()
	ef := ExportFile{Packages: []PackageEntry{{Name: "curl"}}}
	data, _ := json.MarshalIndent(ef, "", "  ")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	path := filepath.Join(home, "aptui-test-import.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	defer os.Remove(path)

	_ = tmpDir

	entries, _, err := Import("~/aptui-test-import.json")
	if err != nil {
		t.Fatalf("import with tilde: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "curl" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestImportExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "explicit.json")
	ef := ExportFile{Packages: []PackageEntry{{Name: "nano"}}}
	data, _ := json.MarshalIndent(ef, "", "  ")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	entries, gotPath, err := Import(path)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if gotPath != path {
		t.Errorf("path = %q, want %q", gotPath, path)
	}
	if len(entries) != 1 || entries[0].Name != "nano" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}
