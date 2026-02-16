//go:build windows

package main

// isKeyDown checks if a virtual key is currently pressed.
func isKeyDown(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

// isKeyToggledOn checks if a toggle key (CapsLock, ScrollLock, etc.) is toggled on.
func isKeyToggledOn(vk int) bool {
	ret, _, _ := procGetKeyState.Call(uintptr(vk))
	return ret&0x0001 != 0
}

// isTrueShift returns true if either physical Shift key is held.
func isTrueShift() bool {
	return isKeyDown(VK_LSHIFT) || isKeyDown(VK_RSHIFT)
}

// isLogicalShift returns the effective shift state (Shift XOR CapsLock)
// for determining whether to output uppercase or lowercase.
func isLogicalShift() bool {
	shift := isTrueShift()
	caps := isKeyToggledOn(VK_CAPITAL)
	return shift != caps // XOR
}

// isAltGr returns true if the Right Alt (AltGr) key is held.
// On Windows, AltGr sends both LCtrl+RAlt, but we detect via RAlt alone.
func isAltGr() bool {
	return isKeyDown(VK_RMENU)
}

// isCtrlDown returns true if either Ctrl key is held.
func isCtrlDown() bool {
	return isKeyDown(VK_LCONTROL) || isKeyDown(VK_RCONTROL)
}

// isOnlyCtrl returns true if Ctrl is held WITHOUT AltGr.
// Used to detect Ctrl+C, Ctrl+V, etc. which should pass through.
// When AltGr is pressed, Windows also sets LCtrl, so we exclude that case.
func isOnlyCtrl() bool {
	return isCtrlDown() && !isAltGr()
}

// isLeftAltDown returns true if the Left Alt key is held (not AltGr).
func isLeftAltDown() bool {
	return isKeyDown(VK_LMENU)
}

// isWinKeyDown returns true if either Win key is held.
func isWinKeyDown() bool {
	return isKeyDown(VK_LWIN) || isKeyDown(VK_RWIN)
}
