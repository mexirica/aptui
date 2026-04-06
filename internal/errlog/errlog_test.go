package errlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)


func TestLoadEmpty(t *testing.T) {
       t.Run("load empty store", func(t *testing.T) {
	       tmp := t.TempDir()
	       origPath := logPath
	       logPath = func() string { return filepath.Join(tmp, "errors.json") }
	       defer func() { logPath = origPath }()

	       s := Load()
	       if s == nil {
		       t.Fatal("expected non-nil store")
	       }
	       if s.NextID != 1 {
		       t.Errorf("expected NextID=1, got %d", s.NextID)
	       }
	       if len(s.Entries) != 0 {
		       t.Errorf("expected 0 entries, got %d", len(s.Entries))
	       }
       })
}


func TestLogAndAll(t *testing.T) {
       t.Run("log and all", func(t *testing.T) {
	       tmp := t.TempDir()
	       origPath := logPath
	       logPath = func() string { return filepath.Join(tmp, "errors.json") }
	       defer func() { logPath = origPath }()

	       s := Load()
	       s.Log("test-source", "first error")
	       s.Log("apt", "second error")

	       if s.Count() != 2 {
		       t.Fatalf("expected 2 entries, got %d", s.Count())
	       }

	       all := s.All()
	       if len(all) != 2 {
		       t.Fatalf("expected 2 entries from All(), got %d", len(all))
	       }
	       // Newest first
	       if all[0].Message != "second error" {
		       t.Errorf("expected newest first, got %q", all[0].Message)
	       }
	       if all[1].Message != "first error" {
		       t.Errorf("expected oldest second, got %q", all[1].Message)
	       }
	       if all[0].Source != "apt" {
		       t.Errorf("expected source 'apt', got %q", all[0].Source)
	       }
       })
}


func TestPersistence(t *testing.T) {
       t.Run("persistence", func(t *testing.T) {
	       tmp := t.TempDir()
	       origPath := logPath
	       logPath = func() string { return filepath.Join(tmp, "errors.json") }
	       defer func() { logPath = origPath }()

	       s := Load()
	       s.Log("test", "persisted error")

	       // Reload from disk
	       s2 := Load()
	       if s2.Count() != 1 {
		       t.Fatalf("expected 1 entry after reload, got %d", s2.Count())
	       }
	       all := s2.All()
	       if all[0].Message != "persisted error" {
		       t.Errorf("unexpected message: %q", all[0].Message)
	       }
       })
}


func TestClear(t *testing.T) {
       t.Run("clear", func(t *testing.T) {
	       tmp := t.TempDir()
	       origPath := logPath
	       logPath = func() string { return filepath.Join(tmp, "errors.json") }
	       defer func() { logPath = origPath }()

	       s := Load()
	       s.Log("test", "error1")
	       s.Log("test", "error2")
	       s.Clear()

	       if s.Count() != 0 {
		       t.Fatalf("expected 0 entries after clear, got %d", s.Count())
	       }

	       // Verify persisted
	       data, _ := os.ReadFile(filepath.Join(tmp, "errors.json"))
	       var loaded Store
	       _ = json.Unmarshal(data, &loaded)
	       if len(loaded.Entries) != 0 {
		       t.Errorf("expected 0 entries on disk after clear, got %d", len(loaded.Entries))
	       }
       })
}


func TestFormatTimestamp(t *testing.T) {
       t.Run("format timestamp", func(t *testing.T) {
	       tmp := t.TempDir()
	       origPath := logPath
	       logPath = func() string { return filepath.Join(tmp, "errors.json") }
	       defer func() { logPath = origPath }()
	       s := Load()
	       e := s.Log("test", "msg")
	       ts := FormatTimestamp(e.Timestamp)
	       if len(ts) != 19 { // "2006-01-02 15:04:05"
		       t.Errorf("unexpected timestamp format: %q", ts)
	       }
       })
}
