package langdetect

import "unicode"

// NonLatinFraction is the share of s's letters that are not Latin script: of
// the runes unicode.IsLetter accepts, the fraction not in unicode.Latin. It is 0
// when s has no letters, so digits, punctuation, emoji and an empty string read
// as Latin. Latin-script languages (Spanish, Indonesian, Vietnamese…) score 0.
//
// It is the spam second opinion's routing gate
// (asianfanfics/docs/spam-ai-review D-2): a candidate at or above
// SPAM_AI_MIN_NON_LATIN goes to the AI. A share, not Detect's any-rune-wins
// cascade, so one sprinkled CJK character does not route an English spam item.
// It leaves Detect, Measure and Profile alone (their JSON is pinned by the spam
// scorer's tests).
func NonLatinFraction(s string) float64 {
	var letters, nonLatin int
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		letters++
		if !unicode.Is(unicode.Latin, r) {
			nonLatin++
		}
	}
	if letters == 0 {
		return 0
	}
	return float64(nonLatin) / float64(letters)
}
