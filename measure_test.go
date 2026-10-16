package langdetect

import (
	"math"
	"testing"
)

// eps is the float comparison tolerance. Every expectation below is an exact
// ratio of small integers, so anything beyond binary rounding is a real defect.
const eps = 1e-9

func closeEnough(got, want float64) bool { return math.Abs(got-want) < eps }

func TestMeasure(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		letters    int
		han        float64
		kana       float64
		hangul     float64
		wantDetect string // the Detect answer this measurement must not contradict
	}{
		{
			name:       "empty string is the zero profile",
			in:         "",
			letters:    0,
			wantDetect: "en",
		},
		{
			name:       "whitespace only",
			in:         " \t\n  ",
			letters:    0,
			wantDetect: "en",
		},
		{
			name: "digits and punctuation only — nothing is a letter",
			// The denominator would be 14 if non-letters counted, and every
			// fraction would still be 0; Letters is what proves they did not.
			in:         "1234567890 !?.,-",
			letters:    0,
			wantDetect: "en",
		},
		{
			name:       "pure ASCII prose",
			in:         "Hello",
			letters:    5,
			wantDetect: "en",
		},
		{
			name: "pure Han",
			// 9 Han characters, no punctuation.
			in:         "他們在城市裡的故事",
			letters:    9,
			han:        1,
			wantDetect: "cn",
		},
		{
			name: "mixed Han and ASCII with a known fraction",
			// 6 Han + 6 Latin letters = 12 letters, so Han is exactly half. The
			// space and the full stop are excluded, which is what makes the
			// answer a clean 0.5 rather than 6/13 or 6/14.
			in:         "中文字中文字 abcdef.",
			letters:    12,
			han:        0.5,
			wantDetect: "cn",
		},
		{
			name: "Japanese — Kana is counted and Han is NOT inflated",
			// 日本語 (3 Han) + のテキストです (7 Kana) = 10 letters. The Kana
			// share is what stops this reading as Chinese; Han is a genuine 0.3
			// because Japanese really does use Han characters, and Detect says
			// "ja" regardless.
			in:         "日本語のテキストです",
			letters:    10,
			han:        0.3,
			kana:       0.7,
			wantDetect: "ja",
		},
		{
			name:       "Japanese Kana only",
			in:         "ひらがなカタカナ",
			letters:    8,
			kana:       1,
			wantDetect: "ja",
		},
		{
			name: "Korean",
			// 한국어 (3) + 이야기 (3) = 6 Hangul syllables; the space is excluded.
			in:         "한국어 이야기",
			letters:    6,
			hangul:     1,
			wantDetect: "ko",
		},
		{
			name: "non-letters would change the answer if they were counted",
			// 2 Han letters against a pile of markup, digits, emoji and
			// punctuation. Han is 1.0 because none of the rest is a letter — with
			// a rune-based denominator it would be about 0.07 and this text would
			// read as "barely Chinese".
			in:         "中文 —— 12345!!! 😀😀😀 ()[]{} …",
			letters:    2,
			han:        1,
			wantDetect: "cn",
		},
		{
			name: "markup letters DO dilute — the same two Han characters in a tag",
			// <p></p> is 2 Latin letters, so 2 Han of 4 letters is 0.5. Recorded
			// as the deliberate counterpart to the case above: this helper does
			// not strip HTML, because the classifiers were not handed stripped
			// text either (README § Frozen behaviour).
			in:         "<p>中文</p>",
			letters:    4,
			han:        0.5,
			wantDetect: "cn",
		},
		{
			name: "Cyrillic counts toward the denominator and toward nothing else",
			// 6 Cyrillic letters, so all three CJK fractions are 0 while Letters
			// is 6. A caller reading Han == 0 alone cannot tell this from the
			// empty string; Letters is what separates them.
			in:         "Привет",
			letters:    6,
			wantDetect: "ru",
		},
		{
			name: "one Han character in an English sentence is a small fraction",
			// The false-positive shape that matters in reverse: a Latin post with
			// a single decorative Han glyph must NOT read as Chinese content,
			// even though Detect answers "cn" for it (Detect flags presence).
			// 1 Han of 18 letters.
			in:         "Buy cheap followers 中",
			letters:    18,
			han:        1.0 / 18.0,
			wantDetect: "cn",
		},
		{
			name: "a Symbol-category Kana code point still flips Han-dominant to Kana-bearing",
			// 9 Han (他們在城市裡的故事) plus U+1F200 SQUARE HIRAGANA HOKO, which is
			// General_Category=Symbol (unicode.IsLetter is false for it) but
			// Script=Hiragana. Detect tests script membership on every rune with
			// no IsLetter gate, so it already says "ja" for this string; before
			// the fix, Measure's IsLetter pre-filter skipped the rune entirely and
			// answered Han:1, Kana:0 — Han-dominant with zero Kana on text Detect
			// itself calls Japanese, exactly the shape README § The
			// Chinese-content signal forbids.
			in:         "他們在城市裡的故事" + string(rune(0x1F200)),
			letters:    10,
			han:        0.9,
			kana:       0.1,
			wantDetect: "ja",
		},
		{
			name: "circled Katakana is Kana even though it fails IsLetter",
			// ㋐ (U+32D0) is one of the circled-Katakana symbols ㋐–㋾:
			// General_Category=Symbol, Script=Katakana. Detect's
			// unicode.Is(unicode.Katakana, …) does not check IsLetter and answers
			// "ja"; Measure must count all three as Kana, not skip them.
			in:         "㋐㋐㋐",
			letters:    3,
			kana:       1,
			wantDetect: "ja",
		},
		{
			name: "circled Hangul is Hangul even though it fails IsLetter",
			// ㈀ (U+3200) is one of the parenthesized/circled-Hangul symbols
			// ㈀–㉾: General_Category=Symbol, Script=Hangul. Same shape as the
			// circled-Katakana case above, for the other script Detect tests
			// without an IsLetter gate.
			in:         "㈀㈀㈀",
			letters:    3,
			hangul:     1,
			wantDetect: "ko",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Measure(tt.in)
			if got.Letters != tt.letters {
				t.Errorf("Letters = %d, want %d", got.Letters, tt.letters)
			}
			if !closeEnough(got.Han, tt.han) {
				t.Errorf("Han = %v, want %v", got.Han, tt.han)
			}
			if !closeEnough(got.Kana, tt.kana) {
				t.Errorf("Kana = %v, want %v", got.Kana, tt.kana)
			}
			if !closeEnough(got.Hangul, tt.hangul) {
				t.Errorf("Hangul = %v, want %v", got.Hangul, tt.hangul)
			}
			// Measure is a refinement of Detect and must never contradict it.
			if d := Detect(tt.in); d != tt.wantDetect {
				t.Errorf("Detect = %q, want %q (the measurement is only valid as a refinement of it)", d, tt.wantDetect)
			}
		})
	}
}

// TestMeasureKanaBeatsHan pins the one ordering decision in Measure: a Japanese
// string must not read as Han-dominant. Both Japan and Korea are on the spam
// scorer's spam-country list, so a Chinese-only leniency that widened to all CJK
// would exempt Japanese and Korean spam.
func TestMeasureKanaBeatsHan(t *testing.T) {
	// Deliberately Han-heavy Japanese: 4 Han (日本語勉強) against 3 Kana (します).
	p := Measure("日本語勉強します")
	if p.Letters != 8 {
		t.Fatalf("Letters = %d, want 8", p.Letters)
	}
	if p.Kana == 0 {
		t.Error("Kana = 0 on Japanese text; the Kana share is the only thing separating it from Chinese")
	}
	if p.Han >= 1 {
		t.Errorf("Han = %v on Japanese text; Kana runes must not be counted as Han", p.Han)
	}
	if got := p.Han + p.Kana; !closeEnough(got, 1) {
		t.Errorf("Han+Kana = %v, want 1 — every letter here is one or the other, and none may be counted twice", got)
	}
	if d := Detect("日本語勉強します"); d != "ja" {
		t.Errorf("Detect = %q, want \"ja\"", d)
	}
}

// TestMeasureSquaredHanSymbolIsScriptCommon records a corrected assumption
// rather than a fix. The squared-ideograph "Enclosed CJK Letters and Months"
// symbols — 🈁 🈚 🈯 🈲 🈳 🈴 🈵 🈶 🈷 🈸 🈹 🈺 🉐 🉑 — visually enclose a Han
// character, but Unicode's own Scripts.txt gives every one of them
// Script=Common, not Script=Han; Go's unicode.Han table agrees (verified
// directly against unicode.Is(unicode.Han, '🈚') on this Go toolchain). Detect
// backs that up: Detect("🈚") is "en", not "cn" — there is no script cascade
// match at all. So unlike U+1F200 SQUARE HIRAGANA HOKO and the circled
// Katakana/Hangul symbols above, a squared-Han-look-alike was never a case of
// Measure disagreeing with Detect: both the pre-fix IsLetter-gated Measure and
// the post-fix script-gated one return the zero Profile for it, because it
// matches neither IsLetter nor any of the four unicode.Is checks. It is
// recorded here, rather than folded into the table above as a Han:1 case,
// precisely because that shape does not hold.
func TestMeasureSquaredHanSymbolIsScriptCommon(t *testing.T) {
	const in = "🈚🈚🈚"
	p := Measure(in)
	if p.Letters != 0 {
		t.Fatalf("Letters = %d, want 0 — these squared symbols are Script=Common, not Han", p.Letters)
	}
	if d := Detect(in); d != "en" {
		t.Errorf("Detect = %q, want \"en\" — confirms Measure's zero Profile is not a disagreement", d)
	}
}

// BenchmarkMeasure pins the "does not allocate" half of Measure's doc comment.
// It runs on the spam scorer's hot path (same comment), so a future change
// that reaches for a []rune conversion or a map instead of the plain switch
// would be a real regression, not just a style nit — `go test .
// -run XXX -bench . -benchmem` must keep reading 0 allocs/op.
func BenchmarkMeasure(b *testing.B) {
	const in = "他們在城市裡的故事 mixed with のテキストです and 한국어 and Hello" + string(rune(0x1F200))
	b.ReportAllocs()
	for range b.N {
		_ = Measure(in)
	}
}

// TestMeasureDoesNotModifyInput is a guard on the README's frozen-behaviour rule:
// the Han-fraction measurement reads the same raw string the classifiers were
// given and must not touch it. Go strings are immutable, so this pins the
// weaker-but-real property that Measure returns nothing derived from a rewritten
// copy — a future "strip HTML first" would have to change the call, not hide
// inside here.
func TestMeasureDoesNotModifyInput(t *testing.T) {
	const in = "  <p>中文</p> 123  "
	before := in
	_ = Measure(in)
	if in != before {
		t.Errorf("input changed: %q", in)
	}
	// Trimming or stripping would change the denominator, which is observable.
	if got := Measure(in); got.Letters != 4 {
		t.Errorf("Letters = %d, want 4 — leading/trailing whitespace and markup must not be pre-processed away", got.Letters)
	}
}
