package main

import (
	"testing"
)

// ── Direct AltGr mappings ──

func TestDirectSchwa(t *testing.T) {
	tests := []struct {
		name  string
		vk    uint32
		shift bool
		want  string
	}{
		{"AltGr+E", VK_E, false, "\u0259"},
		{"AltGr+Shift+E", VK_E, true, "\u018F"},
		{"AltGr+R", VK_R, false, "\u0259\u0331"},
		{"AltGr+Shift+R", VK_R, true, "\u018F\u0331"},
		{"AltGr+O", VK_O, false, "o\u0331"},
		{"AltGr+Shift+O", VK_O, true, "O\u0331"},
		{"AltGr+U", VK_U, false, "u\u0331"},
		{"AltGr+Shift+U", VK_U, true, "U\u0331"},
	}
	for _, tc := range tests {
		got := lookupDirect(tc.vk, tc.shift)
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestDirectNonMapped(t *testing.T) {
	for _, vk := range []uint32{VK_A, VK_W, VK_M, VK_OEM_1} {
		if got := lookupDirect(vk, false); got != "" {
			t.Errorf("VK 0x%X should not be a direct mapping, got %q", vk, got)
		}
	}
}

// ── Dead key detection ──

func TestDeadKeys(t *testing.T) {
	tests := []struct {
		name  string
		vk    uint32
		shift bool
		want  rune
	}{
		{"AltGr+`", VK_OEM_3, false, accentGrave},
		{"AltGr+~", VK_OEM_3, true, accentTilde},
		{"AltGr+'", VK_OEM_7, false, accentAcute},
		{"AltGr+-", VK_OEM_MINUS, false, accentMacron},
	}
	for _, tc := range tests {
		got := getDeadKeyAccent(tc.vk, tc.shift)
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestNonDeadKeys(t *testing.T) {
	for _, vk := range []uint32{VK_E, VK_A, VK_R} {
		if got := getDeadKeyAccent(vk, false); got != accentNone {
			t.Errorf("VK 0x%X should not be a dead key, got %q", vk, got)
		}
	}
}

// ── Dead key + vowel resolution ──

func TestResolveDeadKeyVowels(t *testing.T) {
	tests := []struct {
		name   string
		accent rune
		vk     uint32
		shift  bool
		want   string
	}{
		// Grave
		{"grave+a", accentGrave, VK_A, false, "\u00E0"},
		{"grave+E", accentGrave, VK_E, true, "\u00C8"},
		{"grave+i", accentGrave, VK_I, false, "\u00EC"},
		{"grave+o", accentGrave, VK_O, false, "\u00F2"},
		{"grave+u", accentGrave, VK_U, false, "\u00F9"},
		// Acute
		{"acute+a", accentAcute, VK_A, false, "\u00E1"},
		{"acute+e", accentAcute, VK_E, false, "\u00E9"},
		// Tilde
		{"tilde+a", accentTilde, VK_A, false, "\u00E3"},
		{"tilde+e", accentTilde, VK_E, false, "\u1EBD"},
		{"tilde+i", accentTilde, VK_I, false, "\u0129"},
		{"tilde+o", accentTilde, VK_O, false, "\u00F5"},
		{"tilde+u", accentTilde, VK_U, false, "\u0169"},
		// Macron
		{"macron+a", accentMacron, VK_A, false, "\u0101"},
		{"macron+e", accentMacron, VK_E, false, "\u0113"},
		{"macron+U", accentMacron, VK_U, true, "\u016A"},
	}
	for _, tc := range tests {
		got, ok := resolveDeadKey(tc.accent, tc.vk, false, tc.shift)
		if !ok {
			t.Errorf("%s: not handled", tc.name)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// ── Dead key + schwa resolution ──

func TestResolveDeadKeySchwa(t *testing.T) {
	tests := []struct {
		name   string
		accent rune
		vk     uint32
		shift  bool
		want   string
	}{
		// accent + ə
		{"grave+ə", accentGrave, VK_E, false, "\u0259\u0300"},
		{"acute+ə", accentAcute, VK_E, false, "\u0259\u0301"},
		{"tilde+ə", accentTilde, VK_E, false, "\u0259\u0303"},
		{"macron+ə", accentMacron, VK_E, false, "\u0259\u0304"},
		// accent + Ə (capital)
		{"grave+Ə", accentGrave, VK_E, true, "\u018F\u0300"},
		{"tilde+Ə", accentTilde, VK_E, true, "\u018F\u0303"},
		// accent + ə̱ (retracted)
		{"grave+ə̱", accentGrave, VK_R, false, "\u0259\u0331\u0300"},
		{"acute+ə̱", accentAcute, VK_R, false, "\u0259\u0331\u0301"},
		{"tilde+ə̱", accentTilde, VK_R, false, "\u0259\u0331\u0303"},
		{"macron+ə̱", accentMacron, VK_R, false, "\u0259\u0331\u0304"},
		// accent + Ə̱ (capital retracted)
		{"grave+Ə̱", accentGrave, VK_R, true, "\u018F\u0331\u0300"},
		{"tilde+Ə̱", accentTilde, VK_R, true, "\u018F\u0331\u0303"},
		// accent + o̱ (retracted o)
		{"grave+o̱", accentGrave, VK_O, false, "o\u0331\u0300"},
		{"acute+o̱", accentAcute, VK_O, false, "o\u0331\u0301"},
		{"tilde+o̱", accentTilde, VK_O, false, "o\u0331\u0303"},
		{"macron+o̱", accentMacron, VK_O, false, "o\u0331\u0304"},
		// accent + u̱ (retracted u)
		{"grave+u̱", accentGrave, VK_U, false, "u\u0331\u0300"},
		{"acute+U̱", accentAcute, VK_U, true, "U\u0331\u0301"},
	}
	for _, tc := range tests {
		got, ok := resolveDeadKey(tc.accent, tc.vk, true, tc.shift) // altGr=true
		if !ok {
			t.Errorf("%s: not handled", tc.name)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// ── Dead key + vowel with AltGr still held ──

func TestResolveDeadKeyVowelWithAltGr(t *testing.T) {
	tests := []struct {
		name   string
		accent rune
		vk     uint32
		shift  bool
		want   string
	}{
		{"grave+a (AltGr held)", accentGrave, VK_A, false, "\u00E0"},
		{"acute+o̱ (AltGr held)", accentAcute, VK_O, false, "o\u0331\u0301"},
		{"tilde+i (AltGr held)", accentTilde, VK_I, false, "\u0129"},
		{"macron+U̱ (AltGr held)", accentMacron, VK_U, true, "U\u0331\u0304"},
	}
	for _, tc := range tests {
		got, ok := resolveDeadKey(tc.accent, tc.vk, true, tc.shift) // altGr=true
		if !ok {
			t.Errorf("%s: not handled", tc.name)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// ── Dead key + non-vowel should not resolve ──

func TestResolveDeadKeyNonVowel(t *testing.T) {
	_, ok := resolveDeadKey(accentGrave, VK_W, false, false)
	if ok {
		t.Error("grave+W (no AltGr) should not resolve")
	}
}

// ── applyAccent ──

func TestApplyAccent(t *testing.T) {
	tests := []struct {
		name   string
		accent rune
		input  string
		want   string
	}{
		{"grave on ə", accentGrave, "\u0259", "\u0259\u0300"},
		{"tilde on ə̱", accentTilde, "\u0259\u0331", "\u0259\u0331\u0303"},
		{"acute on Ə̱", accentAcute, "\u018F\u0331", "\u018F\u0331\u0301"},
		{"grave on o̱", accentGrave, "o\u0331", "o\u0331\u0300"},
		{"macron on u̱", accentMacron, "u\u0331", "u\u0331\u0304"},
		{"macron on ə", accentMacron, "\u0259", "\u0259\u0304"},
	}
	for _, tc := range tests {
		got := applyAccent(tc.accent, tc.input)
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// ── Total character coverage ──

func TestAllMobileKeyboardSchwasCovered(t *testing.T) {
	// All 20 schwa characters from the mobile keyboard must be producible.
	// 10 lowercase + 10 uppercase.
	schwas := []struct {
		name string
		want string
	}{
		// Lowercase
		{"ə", "\u0259"},
		{"ə̃", "\u0259\u0303"},
		{"ə̀", "\u0259\u0300"},
		{"ə́", "\u0259\u0301"},
		{"ə̄", "\u0259\u0304"},
		{"ə̱", "\u0259\u0331"},
		{"ə̱̃", "\u0259\u0331\u0303"},
		{"ə̱̀", "\u0259\u0331\u0300"},
		{"ə̱́", "\u0259\u0331\u0301"},
		{"ə̱̄", "\u0259\u0331\u0304"},
		// Uppercase
		{"Ə", "\u018F"},
		{"Ə̃", "\u018F\u0303"},
		{"Ə̀", "\u018F\u0300"},
		{"Ə́", "\u018F\u0301"},
		{"Ə̄", "\u018F\u0304"},
		{"Ə̱", "\u018F\u0331"},
		{"Ə̱̃", "\u018F\u0331\u0303"},
		{"Ə̱̀", "\u018F\u0331\u0300"},
		{"Ə̱́", "\u018F\u0331\u0301"},
		{"Ə̱̄", "\u018F\u0331\u0304"},
	}

	producible := make(map[string]bool)

	// Direct: ə, Ə, ə̱, Ə̱, o̱, O̱, u̱, U̱
	for _, shift := range []bool{false, true} {
		for _, vk := range []uint32{VK_E, VK_R, VK_O, VK_U} {
			if s := lookupDirect(vk, shift); s != "" {
				producible[s] = true
			}
		}
	}

	// Dead key + AltGr+E/R: accented schwas
	for _, accent := range []rune{accentGrave, accentAcute, accentTilde, accentMacron} {
		for _, shift := range []bool{false, true} {
			for _, vk := range []uint32{VK_E, VK_R} {
				if s, ok := resolveDeadKey(accent, vk, true, shift); ok {
					producible[s] = true
				}
			}
		}
	}

	for _, tc := range schwas {
		if !producible[tc.want] {
			t.Errorf("schwa %s (%q) not producible via keyboard", tc.name, tc.want)
		}
	}
}

func TestIsIgnorableModifier(t *testing.T) {
	modifiers := []uint32{
		VK_SHIFT, VK_LSHIFT, VK_RSHIFT,
		VK_CONTROL, VK_LCONTROL, VK_RCONTROL,
		VK_MENU, VK_LMENU, VK_RMENU,
		VK_CAPITAL,
	}
	for _, vk := range modifiers {
		if !isIgnorableModifier(vk) {
			t.Errorf("VK 0x%X should be ignorable modifier", vk)
		}
	}
}
