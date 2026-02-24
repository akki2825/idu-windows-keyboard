//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

// logFile is the debug log written next to the .exe.
var logFile *os.File

func initLog() {
	exePath, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(exePath), "idu-keyboard.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	logFile = f
}

func logf(format string, args ...any) {
	if logFile != nil {
		fmt.Fprintf(logFile, format+"\n", args...)
		logFile.Sync()
	}
}

func main() {
	// Pin to OS thread -- required for Windows message loop and hooks.
	runtime.LockOSThread()

	initLog()
	logf("Idu Mishmi Keyboard starting")

	// Catch panics so they go to the log file.
	defer func() {
		if r := recover(); r != nil {
			logf("PANIC: %v", r)
		}
	}()

	// Single-instance check via named mutex.
	mutexName := utf16Ptr("Global\\IduMishmiKeyboard")
	handle, _, _ := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(mutexName)))
	if handle == 0 {
		logf("FATAL: CreateMutexW returned 0")
		os.Exit(1)
	}
	lastErr, _, _ := procGetLastError.Call()
	if lastErr == ERROR_ALREADY_EXISTS {
		logf("Another instance already running, exiting")
		os.Exit(0)
	}
	logf("Single-instance mutex acquired")

	ensureNotoSansInstalled()

	// Initialize system tray.
	logf("calling initTray...")
	err := initTray()
	logf("initTray returned: err=%v", err)
	if err != nil {
		logf("FATAL: initTray failed: %v", err)
		os.Exit(1)
	}
	logf("System tray initialized")

	// Install keyboard hook.
	logf("calling installHook...")
	err = installHook()
	logf("installHook returned: err=%v", err)
	if err != nil {
		logf("FATAL: installHook failed: %v", err)
		removeTrayIcon()
		os.Exit(1)
	}
	logf("Keyboard hook installed (handle=%d)", hookHandle)
	defer uninstallHook()

	logf("Entering message loop")

	// Run message loop.
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || ret == ^uintptr(0) { // 0 = WM_QUIT, -1 = error
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	logf("Message loop ended, shutting down")
}
