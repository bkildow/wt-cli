//go:build windows

package project

// IsProcessAlive on Windows always returns false. This is conservative: any
// "running" setup state will be treated as stale, triggering the safe
// reconciliation path that marks it as failed.
func IsProcessAlive(_ int) bool {
	return false
}
