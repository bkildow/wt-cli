package ui

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"kilobytes", 4 * 1024, "4.0 KB"},
		{"large kilobytes", 900 * 1024, "900 KB"},
		{"megabytes", 820 * 1024 * 1024, "820 MB"},
		{"single digit gigabytes", 9_878_424_780, "9.2 GB"},
		{"gigabytes", 460 * 1024 * 1024 * 1024, "460 GB"},
		{"terabytes", 2 * 1024 * 1024 * 1024 * 1024, "2.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBytes(tt.bytes); got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}
