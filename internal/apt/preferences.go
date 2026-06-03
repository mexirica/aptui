package apt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ListSystemPinned returns package names pinned by /etc/apt/preferences files.
// It supports exact names and shell-style wildcards (*, ?, [..]) in Package stanzas.
func ListSystemPinned(knownNames []string) (map[string]bool, error) {
	patterns, err := readPreferencesPackagePatterns("/etc/apt/preferences", "/etc/apt/preferences.d")
	if err != nil {
		return nil, err
	}

	pinned := make(map[string]bool)
	if len(patterns) == 0 {
		return pinned, nil
	}

	seenKnown := make(map[string]bool, len(knownNames))
	for _, name := range knownNames {
		if name != "" {
			seenKnown[name] = true
		}
	}

	for _, pat := range patterns {
		if pat == "" {
			continue
		}
		if !hasGlobPattern(pat) {
			pinned[pat] = true
			continue
		}
		for name := range seenKnown {
			matched, matchErr := path.Match(pat, name)
			if matchErr != nil {
				continue
			}
			if matched {
				pinned[name] = true
			}
		}
	}

	return pinned, nil
}

func hasGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func readPreferencesPackagePatterns(mainPath, dirPath string) ([]string, error) {
	paths := []string{mainPath}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			entries = nil
		} else {
			return nil, fmt.Errorf("read %s: %w", dirPath, err)
		}
	}
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		paths = append(paths, filepath.Join(dirPath, ent.Name()))
	}

	var patterns []string
	var errs []string
	for _, p := range paths {
		parts, parseErr := parsePreferencesFilePatterns(p)
		if parseErr != nil {
			errs = append(errs, parseErr.Error())
			continue
		}
		patterns = append(patterns, parts...)
	}

	if len(errs) > 0 {
		return patterns, errors.New(strings.Join(errs, "; "))
	}
	return patterns, nil
}

func parsePreferencesFilePatterns(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var patterns []string
	currentField := ""
	packageValue := ""

	flush := func() {
		if packageValue == "" {
			return
		}
		patterns = append(patterns, splitPackagePatterns(packageValue)...)
		packageValue = ""
		currentField = ""
	}

	for s.Scan() {
		line := s.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flush()
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if startsWithWhitespace(line) {
			if currentField == "package" {
				if packageValue == "" {
					packageValue = trimmed
				} else {
					packageValue += " " + trimmed
				}
			}
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:idx]))
		val := strings.TrimSpace(line[idx+1:])
		currentField = key
		if key == "package" {
			packageValue = val
		}
	}
	flush()

	if scanErr := s.Err(); scanErr != nil {
		return patterns, fmt.Errorf("scan %s: %w", path, scanErr)
	}
	return patterns, nil
}

func splitPackagePatterns(v string) []string {
	v = strings.ReplaceAll(v, ",", " ")
	parts := strings.Fields(v)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func startsWithWhitespace(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == ' ' || s[0] == '\t'
}
