//go:build windows

package main

import "unsafe"

// sendUnicodeString injects a Unicode string as keyboard input.
// Each rune in the string generates a key-down + key-up pair using KEYEVENTF_UNICODE.
//
// Held Ctrl/Alt keys are released before the characters and re-pressed after:
// a character delivered while Alt is down arrives as WM_SYSCHAR (an Alt+key
// accelerator) and while Ctrl is down as a control character, so the target
// app discards it instead of inserting text. A trailing Ctrl tap masks the
// re-pressed Alt so the physical Alt release doesn't focus the menu bar.
func sendUnicodeString(s string) {
	runes := []rune(s)
	if len(runes) == 0 {
		return
	}

	type modKey struct {
		vk       uint16
		extended uint32
	}
	var held []modKey
	for _, m := range []modKey{
		{VK_LCONTROL, 0},
		{VK_RCONTROL, KEYEVENTF_EXTENDEDKEY},
		{VK_LMENU, 0},
		{VK_RMENU, KEYEVENTF_EXTENDEDKEY},
	} {
		if isKeyDown(int(m.vk)) {
			held = append(held, m)
		}
	}

	altHeld := false
	for _, m := range held {
		if m.vk == VK_LMENU || m.vk == VK_RMENU {
			altHeld = true
		}
	}
	ctrlTap := []INPUT{
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_LCONTROL}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_LCONTROL, DwFlags: KEYEVENTF_KEYUP}},
	}

	inputs := make([]INPUT, 0, len(held)*2+len(runes)*2+4)

	// Mask the upcoming Alt release: without a key event between the physical
	// Alt-down and the injected Alt-up, the target treats it as a bare Alt tap
	// and moves focus to its menu bar, swallowing the first character.
	if altHeld {
		inputs = append(inputs, ctrlTap...)
	}

	// Release held modifiers so the characters arrive as plain text.
	for _, m := range held {
		inputs = append(inputs, INPUT{
			Type: INPUT_KEYBOARD,
			Ki:   KEYBDINPUT{WVk: m.vk, DwFlags: m.extended | KEYEVENTF_KEYUP},
		})
	}

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

	// Re-press the released modifiers so their physical state stays consistent.
	for _, m := range held {
		inputs = append(inputs, INPUT{
			Type: INPUT_KEYBOARD,
			Ki:   KEYBDINPUT{WVk: m.vk, DwFlags: m.extended},
		})
	}
	if altHeld {
		// Ctrl tap so the later physical Alt key-up isn't a bare Alt tap.
		inputs = append(inputs, ctrlTap...)
	}

	logf("sendinput: %d modifiers released around injection", len(held))
	n, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(sizeofInput()),
	)
	logf("sendinput: injected %d/%d events (err=%v)", n, len(inputs), err)
}
