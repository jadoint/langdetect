package langdetect

import (
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty and whitespace cases
		{
			name:     "empty string",
			input:    "",
			expected: "en",
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: "en",
		},

		// English cases
		{
			name:     "pure English",
			input:    "Hello world",
			expected: "en",
		},
		{
			name:     "English with numbers and punctuation",
			input:    "Chapter 1: The End!",
			expected: "en",
		},

		// Japanese cases (Hiragana and Katakana)
		{
			name:     "Japanese hiragana",
			input:    "こんにちは",
			expected: "ja",
		},
		{
			name:     "Japanese katakana",
			input:    "カタカナ",
			expected: "ja",
		},
		{
			name:     "Japanese hiragana with English",
			input:    "Hello こんにちは",
			expected: "ja",
		},

		// Chinese (Han) case
		{
			name:     "Chinese Han only",
			input:    "你好世界",
			expected: "cn",
		},

		// Priority test: Han + Hiragana should return Japanese
		{
			name:     "mixed Han and Hiragana returns Japanese",
			input:    "你好こ",
			expected: "ja",
		},

		// Korean (Hangul) case
		{
			name:     "Korean Hangul",
			input:    "안녕하세요",
			expected: "ko",
		},

		// Arabic case
		{
			name:     "Arabic",
			input:    "مرحبا",
			expected: "ar",
		},

		// Russian (Cyrillic) case
		{
			name:     "Russian Cyrillic",
			input:    "Привет",
			expected: "ru",
		},

		// Thai case
		{
			name:     "Thai",
			input:    "สวัสดี",
			expected: "th",
		},

		// Mixed cases testing priority
		{
			name:     "Cyrillic with Arabic returns Arabic",
			input:    "Привет مرحبا",
			expected: "ar",
		},
		{
			name:     "Hangul with Thai returns Korean",
			input:    "안녕하세요 สวัสดี",
			expected: "ko",
		},
		{
			name:     "Han with Arabic returns Chinese",
			input:    "你好 مرحبا",
			expected: "cn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Detect(tt.input)
			if result != tt.expected {
				t.Errorf("Detect(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
