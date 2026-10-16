package langdetect

import "unicode"

// This file is a REFINEMENT of Detect, never a second opinion on it.
//
// Detect answers "which language is this", as a first-match-wins cascade over
// the scripts that are PRESENT. The spam-trust gate
// (asianfanfics/docs/spam-trust-gate/README.md § The Chinese-content signal)
// needs a finer answer for its evidence log — "how MUCH of this is Han", not "is
// there any Han" — so Measure counts instead of flagging.
//
// 🛑 The two must never disagree, which is why the classification below is a
// SINGLE switch in Detect's own priority order: Kana beats Han, then Hangul.
// Japanese is written with Han characters as well as Kana, and both Japan and
// Korea are on the spam scorer's own spam-country list
// (asianfanfics-goth/internal/service/spamscore_tree.go), so a Japanese string
// reading as Han-dominant here would quietly widen a Chinese-only leniency into
// an all-CJK one on a K-pop fanfiction site. A Kana rune is counted as Kana and
// nothing else; a Han rune in a Japanese sentence still counts as Han, but the
// Kana fraction beside it is what says the text is not Chinese.
//
// Detect itself is deliberately untouched. Goth's story creation depends on it
// (asianfanfics-goth/internal/web/storyedit/story_create_edit.go), and this
// module is not the place to change it.

// Profile is a script mix: how much of one piece of text is written in each of
// the three CJK scripts the spam-trust gate cares about.
//
// The fractions are shares of LETTERS PLUS CJK-SCRIPT SYMBOLS, not of runes and
// not of bytes. Digits, punctuation, whitespace, emoji and stray markup are
// excluded from the denominator, because none of them says anything about which
// language the text is in — a Chinese sentence wrapped in HTML tags and a bare
// one are the same text as far as this measurement is concerned, and the
// classifiers are handed the raw string either way (README § Frozen behaviour:
// raw text is assembled with no sanitising).
//
// "CJK-script symbols" is not a hand-wave: Unicode gives a handful of code
// points — the circled Katakana ㋐–㋾, the circled/parenthesized Hangul ㈀–㉾,
// and U+1F200 SQUARE HIRAGANA HOKO — a General_Category of Symbol rather than
// Letter while still carrying a Script of Kana or Hangul. Detect tests script
// membership directly and does not care about General_Category, so a rune like
// that flips Detect to "ja" or "ko" while an IsLetter-gated denominator would
// have skipped it entirely — Han-dominant with zero Kana on text Detect itself
// calls Japanese, which is precisely the shape README § The Chinese-content
// signal forbids. Counting it here, in the same rune loop that decides Han vs.
// Kana vs. Hangul, is what keeps that from happening.
//
// The three fractions do not have to sum to 1: Latin, Cyrillic, Arabic, Thai and
// every other letter counts toward Letters and toward none of them. That is the
// point — Han at 0.95 with 0.05 of Latin is Chinese content with a stray English
// word in it, and the missing 0.05 is exactly what says so.
//
// The zero Profile is the honest answer for text with no letters at all. Letters
// is 0 and every fraction is 0, which reads as "no evidence of any script"
// rather than as "definitely not Chinese" — a caller must not treat Han == 0 as
// a positive claim without looking at Letters.
//
// The json tags are load-bearing rather than decorative: a whole Profile is
// recorded inside the `signals` blob of every spam_trust_hold row
// (asianfanfics-goth/internal/service/spamscore_tree.go), because the hold
// table's own han_fraction column carries the Han share and nothing else — and a
// stored 0.60 with no Kana beside it cannot afterwards be told apart from
// Japanese, which is the one distinction this whole measurement exists to make.
// Renaming a tag orphans every row already written.
type Profile struct {
	// Letters is the denominator: the count of letter runes PLUS the
	// Symbol-category CJK-script code points described above (they are not
	// letters by General_Category, but Detect counts them toward a script all
	// the same, so Letters must too).
	Letters int `json:"letters"`
	// Han is the fraction of those letters that are Han (Chinese characters,
	// Simplified and Traditional alike, plus the Kanji inside Japanese text).
	Han float64 `json:"han"`
	// Kana is the fraction that are Hiragana or Katakana — the two scripts that
	// make a string Japanese rather than Chinese.
	//
	// ⚠️ It reads slightly LOW on real Japanese. The prolonged sound mark ー
	// (U+30FC) and the iteration marks are Script=Common in Unicode, so
	// unicode.Is(unicode.Katakana, …) answers false for them and they land in
	// the denominator without landing here. That is the same blind spot Detect
	// has, deliberately shared: any Japanese string long enough to matter carries
	// ordinary Kana too, so the direction of the error is a slightly smaller
	// non-zero Kana fraction, never a zero one.
	Kana float64 `json:"kana"`
	// Hangul is the fraction that are Hangul (Korean).
	Hangul float64 `json:"hangul"`
}

// Measure counts the script mix of s. It does not modify s and does not
// allocate.
//
// It is called on the spam scorer's hot path, on the SAME raw string the two
// classifiers were given, so that the han_fraction stored beside a held decision
// describes the text that was actually judged.
func Measure(s string) Profile {
	var letters, han, kana, hangul int

	for _, r := range s {
		// Script membership is decided FIRST, before anything asks whether r is a
		// letter — the reverse order is the bug this file was written to fix. A
		// rune must count toward Letters (and toward its bucket) the moment
		// Detect would have counted it toward a script, or the two functions can
		// disagree on the same string: Detect scans every rune for script
		// membership regardless of General_Category, so an IsLetter pre-filter
		// here would silently drop the Symbol-category CJK code points documented
		// on Profile above — a squared-Kana symbol or U+1F200 SQUARE HIRAGANA HOKO
		// makes Detect say "ja", but an IsLetter-gated Measure would miss the
		// whole Kana bucket that is supposed to say the text isn't Chinese, and
		// hand back Han-dominant-with-zero-Kana on text Detect itself calls
		// Japanese.
		//
		// Classified ONCE, in Detect's priority order — Kana beats Han, then
		// Hangul — with a final fallback arm for every OTHER letter (Latin,
		// Cyrillic, Arabic, Thai, …), which counts toward Letters and no bucket.
		// Unicode assigns each code point exactly one script, so the three CJK
		// arms do not in fact overlap; the switch is still written in Detect's
		// own order, rather than as three independent ifs, so a reader sees the
		// same Kana-beats-Han rule spelled out identically in both files.
		switch {
		case unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r):
			letters++
			kana++
		case unicode.Is(unicode.Han, r):
			letters++
			han++
		case unicode.Is(unicode.Hangul, r):
			letters++
			hangul++
		case unicode.IsLetter(r):
			letters++
		}
	}

	if letters == 0 {
		// No division, and the zero value rather than a struct full of NaN.
		return Profile{}
	}

	denom := float64(letters)
	return Profile{
		Letters: letters,
		Han:     float64(han) / denom,
		Kana:    float64(kana) / denom,
		Hangul:  float64(hangul) / denom,
	}
}
