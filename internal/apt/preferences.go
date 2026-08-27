package apt

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// ListPolicyPinned returns package names that are targeted by package-specific
// rules from /etc/apt/preferences and /etc/apt/preferences.d/*. Global rules
// (Package: *) are ignored to avoid marking all packages in the UI.
func ListPolicyPinned(knownNames []string) (map[string]bool, error) {
	patterns, err := readPolicyPinnedPatterns("/etc/apt/preferences", "/etc/apt/preferences.d")
	if err != nil {
		return nil, err
	}

	pinned := make(map[string]bool)
	if len(patterns) == 0 {
		return pinned, nil
	}

	known := make(map[string]bool, len(knownNames))
	for _, name := range knownNames {
		if name != "" {
			known[name] = true
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
		for name := range known {
			ok, mErr := path.Match(pat, name)
			if mErr == nil && ok {
				pinned[name] = true
			}
		}
	}

	return pinned, nil
}

func readPolicyPinnedPatterns(mainPath, dirPath string) ([]string, error) {
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
		name := ent.Name()
		if strings.HasSuffix(name, "~") || strings.Contains(name, ".save") || strings.Contains(name, ".disabled") {
			continue
		}
		paths = append(paths, filepath.Join(dirPath, name))
	}

	var patterns []string
	var errs []string
	for _, p := range paths {
		parts, parseErr := parsePolicyPinnedPatternsFile(p)
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

func parsePolicyPinnedPatternsFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var patterns []string
	currentField := ""
	packageValue := ""
	priorityValue := ""

	flush := func() {
		if packageValue != "" && hasNumericPinPriority(priorityValue) {
			for _, p := range splitPackagePatterns(packageValue) {
				n := normalizePackageSelector(p)
				if n == "" || isGlobalPackageSelector(n) {
					continue
				}
				patterns = append(patterns, n)
			}
		}
		currentField = ""
		packageValue = ""
		priorityValue = ""
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
		idxTrimmed := strings.Index(trimmed, ":")
		if startsWithWhitespace(line) && idxTrimmed <= 0 {
			switch currentField {
			case "package":
				if packageValue == "" {
					packageValue = trimmed
				} else {
					packageValue += " " + trimmed
				}
			case "pin-priority":
				if priorityValue == "" {
					priorityValue = trimmed
				} else {
					priorityValue += " " + trimmed
				}
			}
			continue
		}
		if idxTrimmed <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(trimmed[:idxTrimmed]))
		val := strings.TrimSpace(trimmed[idxTrimmed+1:])
		currentField = key
		switch key {
		case "package":
			packageValue = val
		case "pin-priority":
			priorityValue = val
		}
	}
	flush()

	if err := s.Err(); err != nil {
		return patterns, fmt.Errorf("scan %s: %w", path, err)
	}
	return patterns, nil
}

func hasNumericPinPriority(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	fields := strings.Fields(v)
	if len(fields) == 0 {
		return false
	}
	_, err := strconv.Atoi(fields[0])
	return err == nil
}

func splitPackagePatterns(v string) []string {
	v = strings.ReplaceAll(v, ",", " ")
	return strings.Fields(v)
}

func normalizePackageSelector(selector string) string {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return ""
	}
	if strings.HasPrefix(selector, "src:") {
		return ""
	}
	if idx := strings.LastIndex(selector, ":"); idx > 0 {
		selector = selector[:idx]
	}
	return selector
}

func isGlobalPackageSelector(selector string) bool {
	return strings.TrimSpace(selector) == "*"
}

func hasGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func startsWithWhitespace(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == ' ' || s[0] == '\t'
}
