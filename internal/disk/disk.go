// Package disk reports filesystem space for a path and decides when free
// space is low enough to warrant warning the user.
package disk

// Usage describes space on the filesystem containing a path.
type Usage struct {
	TotalBytes uint64
	FreeBytes  uint64
}

// PercentFree returns free space as a percentage of total, or 0 if the
// total is unknown.
func (u Usage) PercentFree() float64 {
	if u.TotalBytes == 0 {
		return 0
	}
	return float64(u.FreeBytes) / float64(u.TotalBytes) * 100
}

// Default thresholds: warn when the filesystem is below either bound.
const (
	DefaultWarnPercent = 10
	DefaultWarnGB      = 10

	BytesPerGB uint64 = 1 << 30
)

// Threshold bounds below which free space is considered low. A zero or
// negative field disables that bound.
type Threshold struct {
	Percent float64
	Bytes   uint64
}

// DefaultThreshold returns the built-in warning thresholds.
func DefaultThreshold() Threshold {
	return Threshold{
		Percent: DefaultWarnPercent,
		Bytes:   DefaultWarnGB * BytesPerGB,
	}
}

// IsLow reports whether usage crosses either bound of the threshold.
func (t Threshold) IsLow(u Usage) bool {
	if t.Percent > 0 && u.TotalBytes > 0 && u.PercentFree() < t.Percent {
		return true
	}
	if t.Bytes > 0 && u.FreeBytes < t.Bytes {
		return true
	}
	return false
}

// Stat returns space usage for the filesystem containing path.
func Stat(path string) (Usage, error) {
	return stat(path)
}
