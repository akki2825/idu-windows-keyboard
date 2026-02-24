//go:build windows

package main

import (
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

var (
	hookHandle uintptr
	active     = true
	activeMu   sync.Mutex

	// Dead key state: stores the pending combining accent, or accentNone if none.
	pendingAccent rune = accentNone

	// Track VK codes of keys we blocked on key-down, so we also suppress their key-up.
	blockedKeys   = make(map[uint32]bool)
	blockedKeysMu sync.Mutex

	// Log only the first N key events to avoid huge log files.
	keyLogCount atomic.Int32
)

const maxKeyLogs = 50

// isActive returns the current enabled state.
func isActive() bool {
	activeMu.Lock()
	defer activeMu.Unlock()
	return active
}

// setActive sets the enabled state and updates the tray icon.
func setActive(state bool) {
	activeMu.Lock()
	active = state
	activeMu.Unlock()
	updateTrayIcon()
}

// toggleActive flips the enabled state.
func toggleActive() {
	activeMu.Lock()
	active = !active
	activeMu.Unlock()
	updateTrayIcon()
}

// installHook sets up the low-level keyboard hook.
func installHook() error {
	cb := syscall.NewCallback(hookCallback)

	// WH_KEYBOARD_LL requires the module handle of the executable on Windows 11.
	hMod, _, _ := procGetModuleHandleW.Call(0)
	logf("hModule=%d callback=%d", hMod, cb)

	handle, _, err := procSetWindowsHookExW.Call(
		WH_KEYBOARD_LL,
		cb,
		hMod,
		0,
	)
	if handle == 0 {
		return err
	}
	hookHandle = handle
	return nil
}

// uninstallHook removes the keyboard hook.
func uninstallHook() {
	if hookHandle != 0 {
		procUnhookWindowsHookEx.Call(hookHandle)
		hookHandle = 0
	}
}

// hookCallback is the WH_KEYBOARD_LL callback function.
func hookCallback(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode < 0 {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	kb := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))

	isDown := wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN

	// Log first N key-down events so we know the hook is alive.
	if isDown && keyLogCount.Load() < int32(maxKeyLogs) {
		keyLogCount.Add(1)
		logf("hook: vk=0x%X flags=0x%X injected=%v down=%v",
			kb.VkCode, kb.Flags, kb.Flags&LLKHF_INJECTED != 0, isDown)
	}

	// Skip injected keys (our own SendInput output).
	if kb.Flags&LLKHF_INJECTED != 0 {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	isUp := wParam == WM_KEYUP || wParam == WM_SYSKEYUP

	// Always pass through modifier keys.
	if isIgnorableModifier(kb.VkCode) {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Handle toggle hotkey (Scroll Lock) on key-down only.
	if isDown && kb.VkCode == VK_SCROLL {
		toggleActive()
		logf("toggled active=%v", isActive())
		// Pass through so Scroll Lock LED still toggles.
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// On key-up, check if we blocked this key's down event.
	if isUp {
		blockedKeysMu.Lock()
		wasBlocked := blockedKeys[kb.VkCode]
		if wasBlocked {
			delete(blockedKeys, kb.VkCode)
		}
		blockedKeysMu.Unlock()
		if wasBlocked {
			return 1 // Suppress the key-up
		}
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Only process key-down events from here.
	if !isDown {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// If not active, pass everything through.
	if !isActive() {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Pass through Ctrl+key shortcuts (Ctrl+C, Ctrl+V, etc.)
	// but not when AltGr is pressed (AltGr synthesizes LCtrl).
	if isOnlyCtrl() {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Pass through Alt+key shortcuts (Alt+Tab, Alt+F4, etc.)
	if isLeftAltDown() {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Pass through Win+key shortcuts (Win+D, Win+R, etc.)
	if isWinKeyDown() {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	ag := isAltGr()
	sh := isLogicalShift()

	// ── Dead key triggers (AltGr + `, ~, ', -) ──
	if ag {
		if accent := getDeadKeyAccent(kb.VkCode, sh); accent != accentNone {
			pendingAccent = accent
			blockedKeysMu.Lock()
			blockedKeys[kb.VkCode] = true
			blockedKeysMu.Unlock()
			logf("dead key set: accent=U+%04X vk=0x%X", accent, kb.VkCode)
			return 1 // Suppress the dead key itself
		}
	}

	// ── Resolve pending dead key ──
	if pendingAccent != accentNone {
		accent := pendingAccent
		pendingAccent = accentNone

		if output, ok := resolveDeadKey(accent, kb.VkCode, ag, sh); ok {
			blockedKeysMu.Lock()
			blockedKeys[kb.VkCode] = true
			blockedKeysMu.Unlock()
			logf("dead key resolved: accent=U+%04X vk=0x%X → %q", accent, kb.VkCode, output)
			sendUnicodeString(output)
			return 1
		}
		logf("dead key cancelled: accent=U+%04X vk=0x%X", accent, kb.VkCode)
		// Dead key not resolved — cancel and process key normally (fall through)
	}

	// ── Direct AltGr mappings (ə, ə̱, o̱, u̱) ──
	if ag {
		if output := lookupDirect(kb.VkCode, sh); output != "" {
			blockedKeysMu.Lock()
			blockedKeys[kb.VkCode] = true
			blockedKeysMu.Unlock()
			logf("direct AltGr: vk=0x%X shift=%v → %q", kb.VkCode, sh, output)
			sendUnicodeString(output)
			return 1
		}
	}

	// No mapping — pass through.
	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}
