package disk

import "testing"

func TestPercentFree(t *testing.T) {
	tests := []struct {
		name  string
		usage Usage
		want  float64
	}{
		{"half free", Usage{TotalBytes: 100, FreeBytes: 50}, 50},
		{"full", Usage{TotalBytes: 100, FreeBytes: 0}, 0},
		{"empty", Usage{TotalBytes: 100, FreeBytes: 100}, 100},
		{"unknown total", Usage{TotalBytes: 0, FreeBytes: 50}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.usage.PercentFree(); got != tt.want {
				t.Errorf("PercentFree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThresholdIsLow(t *testing.T) {
	const gb = BytesPerGB

	tests := []struct {
		name      string
		threshold Threshold
		usage     Usage
		want      bool
	}{
		{
			name:      "plenty of space",
			threshold: DefaultThreshold(),
			usage:     Usage{TotalBytes: 500 * gb, FreeBytes: 250 * gb},
			want:      false,
		},
		{
			name:      "below percent bound only",
			threshold: Threshold{Percent: 10, Bytes: 1 * gb},
			usage:     Usage{TotalBytes: 1000 * gb, FreeBytes: 50 * gb},
			want:      true,
		},
		{
			name:      "below bytes bound only",
			threshold: Threshold{Percent: 10, Bytes: 10 * gb},
			usage:     Usage{TotalBytes: 20 * gb, FreeBytes: 5 * gb},
			want:      true,
		},
		{
			name:      "below both bounds",
			threshold: DefaultThreshold(),
			usage:     Usage{TotalBytes: 500 * gb, FreeBytes: 2 * gb},
			want:      true,
		},
		{
			name:      "exactly at percent bound is not low",
			threshold: Threshold{Percent: 10},
			usage:     Usage{TotalBytes: 100 * gb, FreeBytes: 10 * gb},
			want:      false,
		},
		{
			name:      "percent bound disabled",
			threshold: Threshold{Percent: 0, Bytes: 1 * gb},
			usage:     Usage{TotalBytes: 1000 * gb, FreeBytes: 50 * gb},
			want:      false,
		},
		{
			name:      "bytes bound disabled",
			threshold: Threshold{Percent: 10, Bytes: 0},
			usage:     Usage{TotalBytes: 1000 * gb, FreeBytes: 500 * gb},
			want:      false,
		},
		{
			name:      "both bounds disabled",
			threshold: Threshold{},
			usage:     Usage{TotalBytes: 100 * gb, FreeBytes: 0},
			want:      false,
		},
		{
			name:      "unknown total falls back to bytes bound",
			threshold: DefaultThreshold(),
			usage:     Usage{TotalBytes: 0, FreeBytes: 1 * gb},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.threshold.IsLow(tt.usage); got != tt.want {
				t.Errorf("IsLow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStat(t *testing.T) {
	usage, err := Stat(t.TempDir())
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if usage.TotalBytes == 0 {
		t.Error("TotalBytes = 0, want non-zero for a real filesystem")
	}
	if usage.FreeBytes > usage.TotalBytes {
		t.Errorf("FreeBytes (%d) > TotalBytes (%d)", usage.FreeBytes, usage.TotalBytes)
	}
}

func TestStatMissingPath(t *testing.T) {
	if _, err := Stat(t.TempDir() + "/does-not-exist"); err == nil {
		t.Error("expected an error for a nonexistent path")
	}
}
