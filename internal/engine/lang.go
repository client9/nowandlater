package engine

import (
	"time"
)

// DateOrder specifies how to interpret ambiguous all-numeric dates like
// "02/03/2016" where month and day cannot be inferred from the tokens alone.
// It only affects the INTEGER INTEGER YEAR signature; dates containing a
// month name (e.g. "04-dec-2026") or an ISO year-first form ("2026-12-04")
// are always unambiguous.
//
// The zero value is MDY (month-day-year), matching US English convention
// and preserving backwards compatibility.
type DateOrder int

const (
	MDY DateOrder = iota // month-day-year  (US English default, e.g. 12/04/2026 = Dec 4)
	DMY                  // day-month-year  (most of Europe, Latin America, e.g. 04/12/2026 = Dec 4)
	YMD                  // year-month-day  (ISO 8601 — only meaningful when year is two digits)
)

// MaxPhraseWords is the maximum number of space-separated words in any phrase
// key across all built-in languages. The tokenizer tries phrase matches up to
// this span before falling back to single-word lookup.
//
// Verified by TestMaxPhraseWords: English "day after tomorrow", Spanish
// "de la manana", French "vingt et une", and Portuguese "trinta e uma" all
// reach 3 words; no built-in language exceeds that.
// Update this constant and rerun TestMaxPhraseWords when adding a new language
// whose phrases exceed 3 words.
const MaxPhraseWords = 3

// WordEntry is the classification for a single recognized word.
// Value is the typed semantic value stored in Token.Value.
//
// Semantic token types carry typed values:
//
//	TokenWeekday   → Weekday  (WeekdayMonday … WeekdaySunday)
//	TokenMonth     → Month    (MonthJanuary … MonthDecember)
//	TokenDirection → Direction (DirectionFuture, DirectionPast, DirectionNearest)
//	TokenModifier  → Modifier  (ModifierFuture, ModifierPast)
//	TokenAnchor    → Anchor    (AnchorNow, AnchorToday, …)
//	TokenUnit      → Period    (PeriodSecond … PeriodYear)
//	TokenAMPM      → AMPM      (AMPMAm, AMPMPm)
//	TokenPrep      → nil       (value never consumed; token used structurally only)
//	TokenFiller    → nil       (filtered out before handler dispatch)
//
// Raw/numeric tokens (TokenYear, TokenTime, TokenInteger, TokenTimezone, etc.)
// carry string values and are created directly by the tokenizer, not via WordEntry.
type WordEntry struct {
	Type  TokenType
	Value any // typed semantic value — see above
}

// Lang holds all language-specific tokenizer rules.
// Handlers, signatures, the dispatch map, and the resolver are language-neutral;
// only this layer varies by language.
type Lang struct {
	// Words maps normalized lowercase input words (and multi-word phrases) to
	// their token classification. Token.Value is set to WordEntry.Value.
	//
	// Multi-word keys (containing a space) are treated as phrases and matched
	// longest-first before single-word lookup. Hyphenated forms ("avant-hier")
	// are single chunks after whitespace splitting and belong here too.
	//
	// This map covers everything:
	//   - regular words:          "monday":       {TokenWeekday, WeekdayMonday}
	//   - multi-word phrases:     "il y a":       {TokenModifier, ModifierPast}
	//   - time-word substitutes:  "noon":         {TokenTime, "12:00"}
	//   - number words:           "five":         {TokenInteger, "5"}
	//   - multi-word numbers:     "twenty first": {TokenInteger, "21"}
	Words map[string]WordEntry

	// OrdinalSuffixes lists suffixes to strip from trailing digit sequences
	// during normalization. All entries must be lowercase.
	// Example English: ["st", "nd", "rd", "th"] strips "1st"→"1", "3rd"→"3".
	// Example French:  ["er", "re", "me", "ème"] — supports multibyte suffixes.
	OrdinalSuffixes []string

	// TokenizerFunc is an optional custom tokenizer. When non-nil it replaces the
	// default whitespace-splitting tokenizer entirely. Use for languages (e.g.
	// Japanese) that do not use spaces as word delimiters. The function receives
	// the raw input string and the Lang, and must return the full token slice.
	TokenizerFunc func(input string, lang *Lang) []Token

	// Timezones maps lowercase timezone token values to *time.Location.
	// Checked before the built-in defaultTimezones table in timezone.go.
	// nil means use only the built-in table.
	// Use this to override ambiguous abbreviations (e.g. IST) or add custom zones.
	Timezones map[string]*time.Location

	// DateOrder controls how ambiguous all-numeric dates are interpreted.
	// It applies to the INTEGER INTEGER YEAR signature ("02/03/2016" etc.).
	// The zero value MDY matches US English convention.
	// Set to DMY for European/Latin-American languages, YMD for ISO-style input.
	DateOrder DateOrder

	// Handlers provides language-specific handler overrides and additions.
	// Parse checks Handlers first; if the signature is not found here, it falls
	// back to the global handlers map (which covers language-neutral patterns:
	// ISO dates, numeric compounds, relative deltas, etc.).
	//
	// Use this for signatures whose token order differs from English, e.g.:
	//   "WEEKDAY DIRECTION" → handleWeekdayDirection  (French "lundi prochain")
	// Leave nil to use only the global handlers (correct for English).
	Handlers map[string]Handler

}
