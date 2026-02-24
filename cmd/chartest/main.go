package main

import "fmt"

func main() {
	fmt.Println("=== Idu Mishmi Keyboard Character Test ===")
	fmt.Println()

	fmt.Println("DIRECT KEYS:")
	fmt.Println()
	fmt.Printf("  AltGr+E          →  ə   (U+0259)          schwa\n")
	fmt.Printf("  AltGr+Shift+E    →  Ə   (U+018F)          capital schwa\n")
	fmt.Printf("  AltGr+R          →  \u0259\u0331   (U+0259 U+0331)  retracted schwa\n")
	fmt.Printf("  AltGr+Shift+R    →  \u018F\u0331   (U+018F U+0331)  capital retracted schwa\n")
	fmt.Printf("  AltGr+O          →  o\u0331   (o + U+0331)     retracted o\n")
	fmt.Printf("  AltGr+Shift+O    →  O\u0331   (O + U+0331)     capital retracted O\n")
	fmt.Printf("  AltGr+U          →  u\u0331   (u + U+0331)     retracted u\n")
	fmt.Printf("  AltGr+Shift+U    →  U\u0331   (U + U+0331)     capital retracted U\n")
	fmt.Println()

	fmt.Println("ACCENTED VOWELS:")
	fmt.Println()
	fmt.Printf("  Grave:   à è ì ò ù\n")
	fmt.Printf("  Acute:   á é í ó ú\n")
	fmt.Printf("  Tilde:   ã ẽ ĩ õ ũ\n")
	fmt.Printf("  Macron:  ā ē ī ō ū\n")
	fmt.Println()

	fmt.Println("ACCENTED SCHWAS (ə + accent):")
	fmt.Println()
	fmt.Printf("  ə̀  ə́  ə̃  ə̄\n")
	fmt.Printf("  Ə̀  Ə́  Ə̃  Ə̄\n")
	fmt.Println()

	fmt.Println("ACCENTED RETRACTED SCHWAS (ə̱ + accent):")
	fmt.Println()
	fmt.Printf("  \u0259\u0331\u0300  \u0259\u0331\u0301  \u0259\u0331\u0303  \u0259\u0331\u0304\n")
	fmt.Printf("  \u018F\u0331\u0300  \u018F\u0331\u0301  \u018F\u0331\u0303  \u018F\u0331\u0304\n")
	fmt.Println()

	fmt.Println("ACCENTED RETRACTED O (o̱ + accent):")
	fmt.Println()
	fmt.Printf("  o\u0331\u0300  o\u0331\u0301  o\u0331\u0303  o\u0331\u0304\n")
	fmt.Printf("  O\u0331\u0300  O\u0331\u0301  O\u0331\u0303  O\u0331\u0304\n")
	fmt.Println()

	fmt.Println("ACCENTED RETRACTED U (u̱ + accent):")
	fmt.Println()
	fmt.Printf("  u\u0331\u0300  u\u0331\u0301  u\u0331\u0303  u\u0331\u0304\n")
	fmt.Printf("  U\u0331\u0300  U\u0331\u0301  U\u0331\u0303  U\u0331\u0304\n")
}
