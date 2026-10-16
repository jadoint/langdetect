# langdetect

`github.com/jadoint/langdetect` — script-based language detection and script mix
measurement, used by Asianfanfics goth and `spam-service`.

## Vocabulary

`Detect(s string) string` returns a 2-character language code:
- `en` — English (default; also empty/whitespace-only, and Latin-script text)
- `ja` — Japanese (Hiragana or Katakana)
- `cn` — Chinese (Han script); `cn` is the historical code for Chinese in this codebase and is not changed
- `ko` — Korean (Hangul)
- `ar` — Arabic
- `ru` — Russian (Cyrillic)
- `th` — Thai

Detection applies a first-match-wins cascade in the priority order:
Hiragana/Katakana (`ja`) → Han (`cn`) → Hangul (`ko`) → Arabic (`ar`) → Cyrillic (`ru`) → Thai (`th`) → fallback (`en`).

## Exported API

- `Detect(s string) string`: returns the 2-character language code from the vocabulary above.
- `Measure(s string) Profile`: returns the script mix (`Letters` denominator, and `Han`, `Kana`, `Hangul` fractions) counting letters and CJK-script symbols. Does not allocate.
- `NonLatinFraction(s string) float64`: returns the fraction of letters that are not in Unicode's `Latin` script (0.0 when no letters are present). Does not allocate.
- `Profile`: struct holding `Letters int`, `Han float64`, `Kana float64`, `Hangul float64` with JSON tags.
