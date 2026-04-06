package pin

import (
	"os"
	"path/filepath"
	"testing"
)


func TestToggleAndSet(t *testing.T) {
       dir := t.TempDir()
       orig := pinPath
       pinPath = func() string { return filepath.Join(dir, "pins.json") }
       defer func() { pinPath = orig }()

       t.Run("initially empty", func(t *testing.T) {
	       s := Load()
	       set := s.Set()
	       if len(set) != 0 {
		       t.Fatalf("expected empty set, got %d", len(set))
	       }
       })

       t.Run("pin, unpin, set", func(t *testing.T) {
	       s := Load()
	       if !s.Toggle("vim") {
		       t.Fatal("expected Toggle to return true (pinned)")
	       }
	       if !s.IsPinned("vim") {
		       t.Fatal("expected vim to be pinned")
	       }
	       s.Toggle("git")
	       set := s.Set()
	       if len(set) != 2 {
		       t.Fatalf("expected 2 pinned, got %d", len(set))
	       }
	       if s.Toggle("vim") {
		       t.Fatal("expected Toggle to return false (unpinned)")
	       }
	       if s.IsPinned("vim") {
		       t.Fatal("expected vim to not be pinned")
	       }
	       set = s.Set()
	       if len(set) != 1 || !set["git"] {
		       t.Fatal("expected only git to be pinned")
	       }
       })
}


func TestPersistence(t *testing.T) {
       dir := t.TempDir()
       orig := pinPath
       pinPath = func() string { return filepath.Join(dir, "pins.json") }
       defer func() { pinPath = orig }()

       t.Run("persist pins", func(t *testing.T) {
	       s := Load()
	       s.Toggle("curl")
	       s.Toggle("wget")
	       s2 := Load()
	       set := s2.Set()
	       if !set["curl"] || !set["wget"] {
		       t.Fatal("expected pinned packages to persist")
	       }
       })
}


func TestLoadMalformedJSON(t *testing.T) {
       dir := t.TempDir()
       orig := pinPath
       pinPath = func() string { return filepath.Join(dir, "pins.json") }
       defer func() { pinPath = orig }()

       t.Run("malformed json", func(t *testing.T) {
	       if err := os.WriteFile(filepath.Join(dir, "pins.json"), []byte("{bad json"), 0o644); err != nil {
		       t.Fatal(err)
	       }
	       s := Load()
	       if len(s.Packages) != 0 {
		       t.Fatalf("expected empty packages for malformed file, got %d", len(s.Packages))
	       }
       })
}


func TestLoadMissingFile(t *testing.T) {
       dir := t.TempDir()
       orig := pinPath
       pinPath = func() string { return filepath.Join(dir, "nonexistent", "pins.json") }
       defer func() { pinPath = orig }()

       t.Run("missing file", func(t *testing.T) {
	       s := Load()
	       if len(s.Packages) != 0 {
		       t.Fatal("expected empty packages for missing file")
	       }
	       // Ensure directory is created on save
	       s.Toggle("test")
	       if _, err := os.Stat(filepath.Join(dir, "nonexistent", "pins.json")); os.IsNotExist(err) {
		       t.Fatal("expected pins.json to be created")
	       }
       })
}
