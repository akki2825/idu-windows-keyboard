package main

// ── Combining accent codepoints ──

const (
	accentNone   rune = 0
	accentGrave  rune = '\u0300' // ̀
	accentAcute  rune = '\u0301' // ́
	accentTilde  rune = '\u0303' // ̃
	accentMacron rune = '\u0304' // ̄
)

// ── Direct AltGr mappings (no dead key needed) ──

// directAltGrKeys produce output immediately when AltGr+key is pressed.
// [0] = AltGr (lowercase), [1] = AltGr+Shift (uppercase)
var directAltGrKeys = map[uint32][2]string{
	VK_E: {"\u0259", "\u018F"},       // ə, Ə
	VK_R: {"\u0259\u0331", "\u018F\u0331"}, // ə̱, Ə̱
	VK_O: {"o\u0331", "O\u0331"},             // o̱, O̱
	VK_U: {"u\u0331", "U\u0331"},             // u̱, U̱
}

// lookupDirect returns the output for AltGr+key (no dead key).
// Returns empty string if not a direct mapping.
func lookupDirect(vkCode uint32, shift bool) string {
	if m, ok := directAltGrKeys[vkCode]; ok {
		if shift {
			return m[1]
		}
		return m[0]
	}
	return ""
}

// ── Dead key detection ──

// deadKeyAccents maps AltGr+key to the combining accent it activates.
// Shift state selects between two accents on the same key.
// [0] = AltGr (no shift), [1] = AltGr+Shift
var deadKeyAccents = map[uint32][2]rune{
	VK_OEM_3:     {accentGrave, accentTilde}, // ` → grave, ~ → tilde
	VK_OEM_7:     {accentAcute, 0},           // ' → acute
	VK_OEM_MINUS: {accentMacron, 0},          // - → macron
}

// getDeadKeyAccent returns the combining accent rune if AltGr+key triggers
// a dead key. Returns accentNone (0) if not a dead key.
func getDeadKeyAccent(vkCode uint32, shift bool) rune {
	if accents, ok := deadKeyAccents[vkCode]; ok {
		if shift && accents[1] != 0 {
			return accents[1]
		}
		return accents[0]
	}
	return accentNone
}

// ── Dead key resolution: accent + next keystroke ──

// Precomposed accented vowels for standard Latin letters.
// Maps combining accent → base letter → precomposed string.
var precomposed = map[rune]map[rune]string{
	accentGrave: {
		'a': "\u00E0", 'A': "\u00C0", // à À
		'e': "\u00E8", 'E': "\u00C8", // è È
		'i': "\u00EC", 'I': "\u00CC", // ì Ì
		'o': "\u00F2", 'O': "\u00D2", // ò Ò
		'u': "\u00F9", 'U': "\u00D9", // ù Ù
	},
	accentAcute: {
		'a': "\u00E1", 'A': "\u00C1", // á Á
		'e': "\u00E9", 'E': "\u00C9", // é É
		'i': "\u00ED", 'I': "\u00CD", // í Í
		'o': "\u00F3", 'O': "\u00D3", // ó Ó
		'u': "\u00FA", 'U': "\u00DA", // ú Ú
	},
	accentTilde: {
		'a': "\u00E3", 'A': "\u00C3", // ã Ã
		'e': "\u1EBD", 'E': "\u1EBC", // ẽ Ẽ
		'i': "\u0129", 'I': "\u0128", // ĩ Ĩ
		'o': "\u00F5", 'O': "\u00D5", // õ Õ
		'u': "\u0169", 'U': "\u0168", // ũ Ũ
	},
	accentMacron: {
		'a': "\u0101", 'A': "\u0100", // ā Ā
		'e': "\u0113", 'E': "\u0112", // ē Ē
		'i': "\u012B", 'I': "\u012A", // ī Ī
		'o': "\u014D", 'O': "\u014C", // ō Ō
		'u': "\u016B", 'U': "\u016A", // ū Ū
	},
}

// vowelForKey maps virtual key codes to their base vowel rune.
var vowelForKey = map[uint32][2]rune{
	VK_A: {'a', 'A'},
	VK_E: {'e', 'E'},
	VK_I: {'i', 'I'},
	VK_O: {'o', 'O'},
	VK_U: {'u', 'U'},
}

// resolveDeadKey resolves a pending accent + the next keystroke.
// It handles:
//   - Regular vowel keys → precomposed accented vowel
//   - AltGr+E (ə/Ə) → schwa with accent
//   - AltGr+R (ə̱/Ə̱) → retracted schwa with accent
//   - AltGr+O (o̱/O̱) → retracted o with accent
//   - AltGr+U (u̱/U̱) → retracted u with accent
//
// Returns (output, handled). If handled is false, the dead key should be
// cancelled and the keystroke processed normally.
func resolveDeadKey(accent rune, vkCode uint32, altGr, shift bool) (string, bool) {
	// AltGr+key: check direct mappings (ə, ə̱, o̱, u̱)
	if altGr {
		if direct := lookupDirect(vkCode, shift); direct != "" {
			return applyAccent(accent, direct), true
		}
	}

	// Regular vowel key (works whether or not AltGr is still held)
	if vowels, ok := vowelForKey[vkCode]; ok {
		v := vowels[0]
		if shift {
			v = vowels[1]
		}
		if accentMap, ok := precomposed[accent]; ok {
			if result, ok := accentMap[v]; ok {
				return result, true
			}
		}
		// Fallback: base + combining
		return string(v) + string(accent), true
	}

	return "", false
}

// applyAccent appends a combining accent to a string.
// For "ə̱" + grave → "ə̱̀" (base + U+0331 + accent, canonical order).
// For "ə" + grave → "ə̀".
func applyAccent(accent rune, s string) string {
	if len([]rune(s)) == 0 {
		return string(accent)
	}
	return s + string(accent)
}
