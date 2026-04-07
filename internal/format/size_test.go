package format

import "testing"

func TestSize(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "-"},
		{-1, "-"},
		{500, "500 kB"},
		{1024, "1.0 MB"},
		{2048, "2.0 MB"},
		{1048576, "1.0 GB"},
		{2097152, "2.0 GB"},
	}
	for _, tt := range tests {
		if got := Size(tt.input); got != tt.want {
			t.Errorf("Size(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
