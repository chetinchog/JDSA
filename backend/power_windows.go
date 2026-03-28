//go:build windows
// +build windows

package backend

import (
	"log"
	"syscall"
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

const (
	ES_CONTINUOUS       = 0x80000000
	ES_SYSTEM_REQUIRED  = 0x00000001
	ES_DISPLAY_REQUIRED = 0x00000002
)

// PreventSleep prevents the computer from sleeping
func PreventSleep() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Warning: PreventSleep panic: %v", r)
		}
	}()
	ret, _, err := procSetThreadExecutionState.Call(uintptr(ES_CONTINUOUS | ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED))
	if ret == 0 {
		log.Printf("Warning: SetThreadExecutionState(PreventSleep) failed: %v", err)
	}
}

// AllowSleep allows the computer to sleep normally
func AllowSleep() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Warning: AllowSleep panic: %v", r)
		}
	}()
	ret, _, err := procSetThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
	if ret == 0 {
		log.Printf("Warning: SetThreadExecutionState(AllowSleep) failed: %v", err)
	}
}
