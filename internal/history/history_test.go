package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "history.json")
	return &Store{path: p, NextID: 1}
}

func TestRecordAndAll(t *testing.T) {
	s := tempStore(t)
	s.Record(OpInstall, []string{"vim"}, true)
	s.Record(OpRemove, []string{"nano"}, false)
	s.Record(OpUpgrade, []string{"git", "curl"}, true)

	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(all))
	}
	// All() returns newest first
	if all[0].ID != 3 {
		t.Errorf("expected newest first (ID=3), got ID=%d", all[0].ID)
	}
	if all[2].ID != 1 {
		t.Errorf("expected oldest last (ID=1), got ID=%d", all[2].ID)
	}
}

func TestRecordIncrementsID(t *testing.T) {
	s := tempStore(t)
	tx1 := s.Record(OpInstall, []string{"a"}, true)
	tx2 := s.Record(OpInstall, []string{"b"}, true)
	if tx1.ID != 1 || tx2.ID != 2 {
		t.Errorf("expected IDs 1,2 got %d,%d", tx1.ID, tx2.ID)
	}
	if s.NextID != 3 {
		t.Errorf("expected NextID=3, got %d", s.NextID)
	}
}

func TestGet(t *testing.T) {
	s := tempStore(t)
	s.Record(OpInstall, []string{"vim"}, true)
	s.Record(OpRemove, []string{"nano"}, false)

	tx, ok := s.Get(1)
	if !ok {
		t.Fatal("expected to find transaction with ID 1")
	}
	if tx.Operation != OpInstall {
		t.Errorf("expected install, got %s", tx.Operation)
	}

	tx, ok = s.Get(2)
	if !ok {
		t.Fatal("expected to find transaction with ID 2")
	}
	if tx.Operation != OpRemove {
		t.Errorf("expected remove, got %s", tx.Operation)
	}

	_, ok = s.Get(99)
	if ok {
		t.Error("expected not found for ID 99")
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "history.json")
	s := &Store{path: p, NextID: 1}

	s.Record(OpInstall, []string{"vim", "git"}, true)

	// Read file and unmarshal
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("failed to read history file: %v", err)
	}
	var loaded Store
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(loaded.Transactions) != 1 {
		t.Fatalf("expected 1 transaction on disk, got %d", len(loaded.Transactions))
	}
	if loaded.Transactions[0].Packages[0] != "vim" {
		t.Errorf("expected 'vim', got '%s'", loaded.Transactions[0].Packages[0])
	}
}

func TestLoadEmptyPath(t *testing.T) {
	// Load with non-existent path should return empty store
	original := historyPath
	defer func() { historyPath = original }()
	dir := t.TempDir()
	historyPath = func() string {
		return filepath.Join(dir, "nonexistent", "history.json")
	}

	s := Load()
	if s == nil {
		t.Fatal("Load() should never return nil")
	}
	if len(s.Transactions) != 0 {
		t.Errorf("expected empty transactions, got %d", len(s.Transactions))
	}
	if s.NextID != 1 {
		t.Errorf("expected NextID=1, got %d", s.NextID)
	}
}

func TestUndoOperation(t *testing.T) {
	tests := []struct {
		input    Operation
		expected Operation
	}{
		{OpInstall, OpRemove},
		{OpRemove, OpInstall},
		{OpUpgrade, OpInstall},
		{OpUpgradeAll, OpInstall},
	}
	for _, tt := range tests {
		got := UndoOperation(tt.input)
		if got != tt.expected {
			t.Errorf("UndoOperation(%s) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestTransactionSummary(t *testing.T) {
	tx := Transaction{
		ID:        1,
		Operation: OpInstall,
		Packages:  []string{"vim"},
		Success:   true,
	}
	s := tx.Summary()
	if s == "" {
		t.Error("summary should not be empty")
	}
	if !contains(s, "✔") {
		t.Error("successful transaction should contain ✔")
	}
	if !contains(s, "install") {
		t.Error("summary should contain operation name")
	}
	if !contains(s, "vim") {
		t.Error("single-package summary should contain package name")
	}

	tx2 := Transaction{
		ID:        2,
		Operation: OpRemove,
		Packages:  []string{"a", "b", "c"},
		Success:   false,
	}
	s2 := tx2.Summary()
	if !contains(s2, "✘") {
		t.Error("failed transaction should contain ✘")
	}
	if !contains(s2, "3 packages") {
		t.Errorf("multi-package summary should say '3 packages', got: %s", s2)
	}
}

func TestFormatTimestamp(t *testing.T) {
	s := tempStore(t)
	tx := s.Record(OpInstall, []string{"test"}, true)
	ts := FormatTimestamp(tx.Timestamp)
	if len(ts) < 10 {
		t.Errorf("timestamp too short: %s", ts)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
