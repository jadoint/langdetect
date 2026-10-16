package langdetect

import (
	"math"
	"testing"
)

func TestNonLatinFraction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"empty", "", 0},
		{"whitespace only", " \t\n ", 0},
		{"digits and punctuation only", "12345 !?.,;:-()[] 2026-10-05", 0},
		{"emoji only", "🙂🔥✨", 0},
		{"pure English", "Hello world, this is a story.", 0},
		{"Latin with diacritics", "Él está aquí. Tiếng Việt có dấu.", 0},
		{"pure Simplified Chinese", "这是一个风雨交加的夜晚。", 1},
		{"pure Traditional Chinese", "這是一個風雨交加的夜晚！", 1},
		{"Chinese with digits and punctuation", "第1章：2026年的夏天……", 1},
		{"Japanese kana and kanji", "これは日本語の文章です", 1},
		{"Korean", "안녕하세요 여러분", 1},
		{"Cyrillic", "Привет мир", 1},
		{"full-width Latin is Latin", "ｈｅｌｌｏ", 0},
		{"boundary exactly 0.20", "abcd中", 0.20},
		{"just under 0.20", "abcde中", 1.0 / 6},
		{"half and half", "ab中文", 0.5},
		{"one sprinkled Han in English", "Buy cheap followers now 中", 1.0 / 21},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NonLatinFraction(tt.input)
			if math.Abs(got-tt.want) > 1e-12 {
				t.Errorf("NonLatinFraction(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}

	// The D-2 gate compares with >=, so the boundary must land exactly on the
	// default threshold, not a hair below it.
	if got := NonLatinFraction("abcd中"); !(got >= 0.20) {
		t.Errorf("boundary: NonLatinFraction = %v, want >= 0.20", got)
	}
	if got := NonLatinFraction("abcde中"); got >= 0.20 {
		t.Errorf("below boundary: NonLatinFraction = %v, want < 0.20", got)
	}
}

func TestNonLatinFractionDoesNotAllocate(t *testing.T) {
	s := "风雨欲来 Summer Love 张艺兴 exo angst"
	if n := testing.AllocsPerRun(100, func() { _ = NonLatinFraction(s) }); n != 0 {
		t.Errorf("NonLatinFraction allocates %v times per call, want 0", n)
	}
}
