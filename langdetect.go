package langdetect

import (
	"strings"
	"unicode"
)

// Detect returns a 2-character language code based on Unicode scripts in the input string.
// It scans the entire string once to detect which scripts are present, then applies
// priority order to determine the language. Priority order (first match wins):
//   - empty or whitespace-only => "en"
//   - contains Hiragana OR Katakana => "ja"
//   - contains Han => "cn"
//   - contains Hangul => "ko"
//   - contains Arabic => "ar"
//   - contains Cyrillic => "ru"
//   - contains Thai => "th"
//   - default => "en"
func Detect(s string) string {
	// Check for empty or whitespace-only string
	if strings.TrimSpace(s) == "" {
		return "en"
	}

	// Scan the string once to detect which scripts are present
	var (
		hasHiraKata bool
		hasHan      bool
		hasHangul   bool
		hasArabic   bool
		hasCyrillic bool
		hasThai     bool
	)

	for _, r := range s {
		if unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
			hasHiraKata = true
		}
		if unicode.Is(unicode.Han, r) {
			hasHan = true
		}
		if unicode.Is(unicode.Hangul, r) {
			hasHangul = true
		}
		if unicode.Is(unicode.Arabic, r) {
			hasArabic = true
		}
		if unicode.Is(unicode.Cyrillic, r) {
			hasCyrillic = true
		}
		if unicode.Is(unicode.Thai, r) {
			hasThai = true
		}
	}

	// Apply priority cascade (first match wins)
	if hasHiraKata {
		return "ja"
	}
	if hasHan {
		return "cn"
	}
	if hasHangul {
		return "ko"
	}
	if hasArabic {
		return "ar"
	}
	if hasCyrillic {
		return "ru"
	}
	if hasThai {
		return "th"
	}

	// Default to English
	return "en"
}
