package datadir

import (
	"path/filepath"
	"time"

	"github.com/mexirica/aptui/internal/model"
)

// AlreadyRunToday checks if the given command was marked as run today.
func AlreadyRunToday(cmd string) (bool, error) {
	var cmdRunned *model.CmdRunned
	if err := LoadJSON(filepath.Join(Dir(), "cmd_runned.json"), &cmdRunned); err != nil {
		return false, err
	}

	if cmdRunned == nil || cmdRunned.Cmd != cmd {
		return false, nil
	}
	today := time.Now().Truncate(24 * time.Hour)
	return cmdRunned.Time.Truncate(24 * time.Hour).Equal(today), nil
}

// MarkCmdRunned marks the given command as run at the current time.
func MarkCmdRunned(cmd string) error {
	cr := model.CmdRunned{
		Cmd:  cmd,
		Time: time.Now(),
	}
	return SaveJSON(filepath.Join(Dir(), "cmd_runned.json"), cr)
}
