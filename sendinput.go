//go:build windows

package main

import "unsafe"

// sendUnicodeString injects a Unicode string as keyboard input.
// Each rune in the string generates a key-down + key-up pair using KEYEVENTF_UNICODE.
func sendUnicodeString(s string) {
	runes := []rune(s)
	if len(runes) == 0 {
		return
	}

	// Each rune needs a down + up event
	inputs := make([]INPUT, 0, len(runes)*2)

	for _, r := range runes {
		// Key down
		inputs = append(inputs, INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WScan:   uint16(r),
				DwFlags: KEYEVENTF_UNICODE,
			},
		})
		// Key up
		inputs = append(inputs, INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WScan:   uint16(r),
				DwFlags: KEYEVENTF_UNICODE | KEYEVENTF_KEYUP,
			},
		})
	}

	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(sizeofInput()),
	)
}
