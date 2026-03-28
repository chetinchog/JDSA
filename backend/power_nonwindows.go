//go:build !windows
// +build !windows

package backend

// PreventSleep is a no-op on non-Windows
func PreventSleep() {
}

// AllowSleep is a no-op on non-Windows
func AllowSleep() {
}
