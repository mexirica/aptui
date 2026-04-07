// Package format provides shared formatting utilities.
package format

import "fmt"

const (
	gb = 1048576 // kB threshold for GB
	mb = 1024    // kB threshold for MB
)

// Size converts a size in kB (as reported by dpkg/apt) to a human-friendly string.
// The input is an int64 in kB units.
func Size(kB int64) string {
	if kB <= 0 {
		return "-"
	}
	switch {
	case kB >= gb:
		return fmt.Sprintf("%.1f GB", float64(kB)/gb)
	case kB >= mb:
		return fmt.Sprintf("%.1f MB", float64(kB)/mb)
	default:
		return fmt.Sprintf("%d kB", kB)
	}
}
