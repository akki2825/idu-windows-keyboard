//go:build windows

package main

import (
	"runtime"
	"runtime/debug"
	"syscall"
	"unsafe"
)

func main() {
	// Pin to OS thread -- required for Windows message loop and hooks.
	runtime.LockOSThread()

	initLog()
	logf("Idu Mishmi Keyboard starting, version 1.0.1-debug1")

	// Catch panics so they go to the log file and the user sees an error
	// instead of a silent crash.
	defer func() {
		if r := recover(); r != nil {
			logf("PANIC: %v\n%s", r, debug.Stack())
			messageBox("Idu Mishmi Keyboard hit an internal error and will close.",
				"Idu Mishmi Keyboard", MB_OK|MB_ICONERROR)
		}
	}()

	// Single-instance check via named mutex. Per-session (Local\) namespace:
	// a keyboard hook is per-session anyway, and Global\ can fail with
	// ACCESS_DENIED when the mutex exists under another security context.
	mutexName := utf16Ptr("Local\\IduMishmiKeyboard")
	handle, _, err := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(mutexName)))
	if errno, ok := err.(syscall.Errno); ok && errno == ERROR_ALREADY_EXISTS {
		logf("another instance is already running; exiting")
		messageBox("Idu Mishmi Keyboard is already running.\nLook for its icon in the system tray.",
			"Idu Mishmi Keyboard", MB_OK|MB_ICONINFORMATION)
		return
	}
	if handle == 0 {
		// Not fatal -- continue without single-instance protection.
		logf("CreateMutexW failed (%v); continuing without single-instance check", err)
	} else {
		logf("single-instance mutex acquired")
	}

	ensureNotoSansInstalled()
	logf("font install step done")

	// Initialize system tray. Failure is not fatal: the keyboard still works,
	// and the icon is re-added when the shell broadcasts TaskbarCreated.
	if err := initTray(); err != nil {
		logf("initTray failed: %v; continuing without tray", err)
	} else {
		logf("system tray initialized")
	}

	// Settings popup, shown once at launch so there is visible UI.
	if err := initPopup(); err != nil {
		logf("initPopup failed: %v", err)
	} else {
		showPopup()
	}

	// Floating indicator badge.
	if err := initIndicator(); err != nil {
		logf("initIndicator failed: %v", err)
	}

	// Install keyboard hook. This is the one genuinely fatal failure.
	if err := installHook(); err != nil {
		logf("FATAL: installHook failed: %v", err)
		messageBox("Could not install the keyboard hook ("+err.Error()+").\nIdu Mishmi Keyboard will now close.",
			"Idu Mishmi Keyboard", MB_OK|MB_ICONERROR)
		removeTrayIcon()
		return
	}
	logf("keyboard hook installed")
	defer uninstallHook()

	logf("entering message loop")

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

	logf("message loop ended, shutting down")
}
