// Package datadir centralizes data directory resolution and file persistence for aptui.
// It handles the SUDO_USER case so files always go to the real user's home,
// and fixes ownership after writing when running under sudo.
package datadir

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// Dir returns the aptui data directory (~/.local/share/aptui) for the real user.
func Dir() string {
	return filepath.Join(RealUserHome(), ".local", "share", "aptui")
}

// RealUserHome returns the home directory of the real user,
// even when running under sudo.
func RealUserHome() string {
	if u := os.Getenv("SUDO_USER"); u != "" {
		if lu, err := user.Lookup(u); err == nil {
			return lu.HomeDir
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return home
}

// SaveJSON marshals v as indented JSON and writes it to path,
// creating parent directories as needed. When running under sudo,
// it chowns the directory and file to the real user.
func SaveJSON(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	fixOwnership(dir, path)
	return nil
}

// AppendJSON reads path as a JSON array, appends v to it, and writes it back.
func AppendJSON(path string, v any) error {
	var arr []any
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
	}
	arr = append(arr, v)
	return SaveJSON(path, arr)
}

// ReplaceJSONByKey substitutes an item in the JSON array at path by matching keyField, or appends it if not found.
func ReplaceJSONByKey(path string, keyField string, v any) error {
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var arr []any
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		vMap, ok := v.(map[string]any)
		if !ok {
			return SaveJSON(path, arr) 
		}
		vKey, ok := vMap[keyField]
		if !ok {
			return SaveJSON(path, arr)
		}
		for i, item := range arr {
			if m, ok := item.(map[string]any); ok {
				if mKey, ok := m[keyField]; ok && mKey == vKey {
					arr[i] = v
					return SaveJSON(path, arr)
				}
			}
		}
		arr = append(arr, v)
		return SaveJSON(path, arr)
	}
	return SaveJSON(path, []any{v})
}

// LoadJSON reads path and unmarshals the JSON into v.
func LoadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func fixOwnership(paths ...string) {
	u := os.Getenv("SUDO_USER")
	if u == "" {
		return
	}
	lu, err := user.Lookup(u)
	if err != nil {
		return
	}
	uid, uidErr := strconv.Atoi(lu.Uid)
	gid, gidErr := strconv.Atoi(lu.Gid)
	if uidErr != nil || gidErr != nil {
		return
	}
	for _, p := range paths {
		_ = os.Chown(p, uid, gid)
	}
}
