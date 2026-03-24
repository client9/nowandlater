package nowandlater

import (
	"strconv"
	"strings"
)

// TokenType identifies the semantic category of a parsed token.
type TokenType int

const (
	// Word token types
	TokenWeekday   TokenType = iota // Monday, Mon, …
	TokenMonth                      // January, Jan, …
	TokenDirection                  // next, last, previous, this, coming
	TokenModifier                   // ago, hence, from, before, after
	TokenAnchor                     // now, today, tomorrow, yesterday
	TokenPrep                       // at, on, in, by
	TokenAMPM                       // AM, PM
	TokenUnit                       // second, minute, hour, day, week, month, year (and plurals)
	TokenFiller                     // the, of, a, and — kept in token list; dropped at signature step

	// Number token types
	TokenYear         // 4-digit year: 2026
	TokenTime         // H:MM, HH:MM, H:MM:SS, HH:MM:SS
	TokenDateFragment // ambiguous partial date with no year or month name: 12/03, 3-15
	TokenInteger      // bare 1–2 digit number: 3, 15
	TokenInteger2     // leading-zero 2-digit number: 03, 09
	TokenDecimal      // decimal number: 3.5, 1.5, 0.5 — Value is float64

	TokenTimezone // UTC, EST, +05:30, -07:00, Z

	TokenUnknown // unrecognised token
)

func (t TokenType) String() string {
	switch t {
	case TokenWeekday:
		return "WEEKDAY"
	case TokenMonth:
		return "MONTH"
	case TokenDirection:
		return "DIRECTION"
	case TokenModifier:
		return "MODIFIER"
	case TokenAnchor:
		return "ANCHOR"
	case TokenPrep:
		return "PREP"
	case TokenAMPM:
		return "AMPM"
	case TokenUnit:
		return "UNIT"
	case TokenFiller:
		return "FILLER"
	case TokenYear:
		return "YEAR"
	case TokenTime:
		return "TIME"
	case TokenDateFragment:
		return "DATE_FRAGMENT"
	case TokenInteger:
		return "INTEGER"
	case TokenInteger2:
		return "INTEGER2"
	case TokenDecimal:
		return "DECIMAL"
	case TokenTimezone:
		return "TIMEZONE"
	case TokenUnknown:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

// Token is a single classified unit from the input string.
// Value holds a typed semantic value for word tokens, or a raw string for
// numeric/timezone tokens. See WordEntry for the type convention per TokenType.
type Token struct {
	Type  TokenType
	Value any
}

// Tokenize preprocesses and normalizes input, splits on whitespace, classifies
// each chunk, and returns the resulting token slice.
// FILLER tokens are included; callers that build signatures should skip them.
func (lang Lang) Tokenize(input string) []Token {
	normalized := normalize(preprocess(input, lang), lang)
	chunks := strings.Fields(normalized)

	// Compute the maximum phrase length (in words) by scanning Words for
	// space-containing keys. Multi-word keys are matched longest-first.
	maxPhraseWords := 0
	for key := range lang.Words {
		if n := strings.Count(key, " ") + 1; n > maxPhraseWords {
			maxPhraseWords = n
		}
	}

	var tokens []Token
	i := 0
	for i < len(chunks) {
		// Phrase match: try longest possible span first, down to 2 words.
		if maxPhraseWords >= 2 {
			matched := false
			for span := maxPhraseWords; span >= 2; span-- {
				if i+span > len(chunks) {
					continue
				}
				candidate := strings.Join(chunks[i:i+span], " ")
				if entry, ok := lang.Words[candidate]; ok {
					tokens = append(tokens, Token(entry))
					i += span
					matched = true
					break
				}
			}
			if matched {
				continue
			}
		}

		chunk := chunks[i]
		i++
		// Numeric timezone offset: +05:30, -07:00, +0530, -0700, +05, -07
		if len(chunk) > 0 && (chunk[0] == '+' || chunk[0] == '-') && isTimezoneOffset(chunk) {
			tokens = append(tokens, Token{Type: TokenTimezone, Value: chunk})
			continue
		}
		// Number sub-parser (handles glued suffix like "3pm")
		if len(chunk) > 0 && isDigit(rune(chunk[0])) {
			tokens = append(tokens, classifyNumber(chunk, lang)...)
			continue
		}
		tokens = append(tokens, classifyWord(chunk, lang))
	}
	return tokens
}

// Signature returns the space-joined token type names, skipping FILLER tokens.
// This is the key used for hash-dispatch to a handler function.
// INTEGER2 is folded into INTEGER: the leading-zero distinction is not relevant
// to handler logic (handlers read the pre-parsed int value, not the raw string).
func Signature(tokens []Token) string {
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t.Type == TokenFiller {
			continue
		}
		name := t.Type.String()
		if t.Type == TokenInteger2 {
			name = "INTEGER"
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, " ")
}

// ---------------------------------------------------------------------------
// Normalization
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Preprocessing (runs before normalize)
// ---------------------------------------------------------------------------

// preprocess performs surface-level substitutions that simplify the tokenizer:
//   - lowercases the input
//   - replaces the ISO 8601 T date/time separator (digit T digit) with a space
//   - replaces time-dot notation when followed by AM/PM: "7.15pm" → "7:15pm"
//
// Time-word and number-word substitutions are no longer done here; they are
// expressed as Words entries and handled by the tokenizer's phrase/word lookup.
func preprocess(s string, lang Lang) string {
	s = strings.ToLower(s)

	// ISO 8601 T separator: "2026-12-04T09:30:00" → "2026-12-04 09:30:00".
	// Replace any 't' that is immediately preceded and followed by a digit.
	if strings.ContainsRune(s, 't') {
		out := make([]byte, 0, len(s))
		for i := 0; i < len(s); i++ {
			if s[i] == 't' && i > 0 && isDigitByte(s[i-1]) && i+1 < len(s) && isDigitByte(s[i+1]) {
				out = append(out, ' ')
			} else {
				out = append(out, s[i])
			}
		}
		s = string(out)
	}

	// Time with dot separator followed by AM/PM: "7.15pm" → "7:15pm".
	// Only triggers when pattern is: 1–2 digits, '.', exactly 2 digits,
	// optional space, then "am" or "pm". Bare "7.15" is left unchanged
	// (ambiguous — could be a date fragment).
	if strings.ContainsRune(s, '.') {
		s = preprocessTimeDot(s)
	}

	return s
}

// preprocessTimeDot converts dot-separated time notation to colon notation
// when the dot is unambiguously marking a time (followed by am/pm).
// "7.15pm" → "7:15pm",  "7.15 pm" → "7:15 pm".
func preprocessTimeDot(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if !isDigitByte(s[i]) {
			b.WriteByte(s[i])
			i++
			continue
		}
		start := i
		for i < len(s) && isDigitByte(s[i]) {
			i++
		}
		digitLen := i - start
		// Candidate: 1–2 digits followed by '.', then exactly 2 more digits
		if digitLen <= 2 && i < len(s) && s[i] == '.' && i+3 <= len(s) &&
			isDigitByte(s[i+1]) && isDigitByte(s[i+2]) &&
			(i+3 == len(s) || !isDigitByte(s[i+3])) {
			// Check what follows the 2 digits: optional space, then am/pm
			k := i + 3
			if k < len(s) && s[k] == ' ' {
				k++
			}
			if k+2 <= len(s) && (s[k:k+2] == "am" || s[k:k+2] == "pm") {
				b.WriteString(s[start:i]) // write the digit run
				b.WriteByte(':')          // colon instead of dot
				i++                       // skip the dot
				continue
			}
		}
		b.WriteString(s[start:i])
	}
	return b.String()
}

// normalize prepares the raw string for tokenization:
//   - strips ordinal suffixes defined by lang.OrdinalSuffixes (e.g. 3rd → 3)
//   - expands dotted abbreviations (A.M. → AM, Mon. → Mon)
//   - collapses whitespace
//   - lowercases (word lookup is case-insensitive; we lowercase here once)
func normalize(s string, lang Lang) string {
	// Work through the string byte by byte building the output.
	// We lowercase everything; word table keys are already lowercase.
	s = strings.ToLower(s)

	var b strings.Builder
	b.Grow(len(s))

	i := 0
	for i < len(s) {
		c := s[i]

		// Ordinal suffix: digit run followed by a language-defined suffix.
		if isDigitByte(c) {
			b.WriteByte(c)
			i++
			// Consume digit run
			for i < len(s) && isDigitByte(s[i]) {
				b.WriteByte(s[i])
				i++
			}
			// Try each ordinal suffix from the language definition.
			for _, sfx := range lang.OrdinalSuffixes {
				n := len(sfx)
				if i+n <= len(s) && s[i:i+n] == sfx {
					// Only strip if followed by non-letter (end, space, punctuation)
					if i+n == len(s) || !isLetterByte(s[i+n]) {
						i += n
						break
					}
				}
			}
			continue
		}

		// Dotted abbreviations: a sequence of letter + "." pairs → strip dots
		// e.g. a.m. → am,  p.m. → pm,  mon. → mon
		if isLetterByte(c) {
			start := i
			// Collect letter(s) then dot
			for i < len(s) && isLetterByte(s[i]) {
				b.WriteByte(s[i])
				i++
			}
			if i < len(s) && s[i] == '.' {
				// Check if this looks like a dotted abbreviation:
				// letter(s).letter(s). ...
				// Just strip the trailing dot (and any following letter.dot groups).
				// We already wrote the letters; just skip dots between letter groups.
				for i < len(s) && s[i] == '.' {
					i++ // skip dot
					// If next char is a letter, continue collecting
					inner := i
					for i < len(s) && isLetterByte(s[i]) {
						b.WriteByte(s[i])
						i++
					}
					if i == inner {
						// dot wasn't followed by a letter — put a space and move on
						b.WriteByte(' ')
						break
					}
					// If the letter group ends with another dot, loop continues.
				}
				_ = start
				continue
			}
			continue
		}

		// Whitespace or comma → single space.
		// Comma is a word separator in RFC 2822 ("Mon, 02 Jan 2006") and in
		// natural language ("January 5, 2026"); it carries no semantic value.
		if isSpaceByte(c) || c == ',' {
			b.WriteByte(' ')
			i++
			for i < len(s) && (isSpaceByte(s[i]) || s[i] == ',') {
				i++
			}
			continue
		}

		// Everything else (colon, slash, dash, etc.) pass through
		b.WriteByte(c)
		i++
	}

	return strings.TrimSpace(b.String())
}

// ---------------------------------------------------------------------------
// Word classification
// ---------------------------------------------------------------------------

func classifyWord(w string, lang Lang) Token {
	if entry, ok := lang.Words[w]; ok {
		return Token(entry)
	}
	// Global timezone fallback: abbreviations are internationally recognised,
	// not language-specific. defaultTimezones is the single source of truth for
	// both tokenisation and resolution. lang.Words overrides take priority above.
	if _, ok := defaultTimezones[w]; ok {
		return Token{Type: TokenTimezone, Value: w}
	}
	return Token{Type: TokenUnknown, Value: w}
}

// ---------------------------------------------------------------------------
// Number sub-parser
// ---------------------------------------------------------------------------

// classifyNumber handles a chunk that begins with a digit.
// It may return more than one token (e.g. "3pm" → INTEGER + AMPM,
// "2026-12-04" → YEAR + INTEGER + INTEGER, "7:15pm" → TIME + AMPM).
func classifyNumber(chunk string, lang Lang) []Token {
	// Time with glued numeric timezone: "09:30:00-07:00", "9:30+05:30"
	if toks := splitTimeAndZone(chunk); toks != nil {
		return toks
	}

	// Time with glued Z (UTC/Zulu): "09:30:00z", "9:30z" (already lowercased)
	if len(chunk) > 1 && chunk[len(chunk)-1] == 'z' {
		if base := chunk[:len(chunk)-1]; isTime(base) {
			return []Token{
				{Type: TokenTime, Value: base},
				{Type: TokenTimezone, Value: "z"},
			}
		}
	}

	// Time with glued AM/PM: "7:15pm", "09:30am" — check before bare isTime.
	if len(chunk) >= 5 {
		suffix := chunk[len(chunk)-2:]
		if suffix == "am" || suffix == "pm" {
			if base := chunk[:len(chunk)-2]; isTime(base) {
				return []Token{
					{Type: TokenTime, Value: base},
					{Type: TokenAMPM, Value: parseAMPM(suffix)},
				}
			}
		}
	}

	// Time: H:MM, HH:MM, H:MM:SS, HH:MM:SS
	if isTime(chunk) {
		return []Token{{Type: TokenTime, Value: chunk}}
	}

	// Glued AM/PM suffix: "3pm", "11am", "730pm", "1230am"
	if idx := gluedAMPM(chunk); idx > 0 {
		base := chunk[:idx]
		ampm := chunk[idx:]
		// 3–4 digit run: compact HHMM time — "730pm" → TIME "7:30" + AMPM pm.
		// Validate hours ≤ 23 and minutes ≤ 59; fall through on invalid input.
		if len(base) == 3 || len(base) == 4 {
			h := mustAtoi(base[:len(base)-2])
			m := mustAtoi(base[len(base)-2:])
			if h <= 23 && m <= 59 {
				timeStr := base[:len(base)-2] + ":" + base[len(base)-2:]
				return []Token{
					{Type: TokenTime, Value: timeStr},
					{Type: TokenAMPM, Value: parseAMPM(ampm)},
				}
			}
		}
		return []Token{
			classifyBareInteger(base),
			{Type: TokenAMPM, Value: parseAMPM(ampm)},
		}
	}

	// Compound date chunk: contains a consistent separator (-, /, .)
	// Returns decomposed tokens if YEAR or MONTH is present, DATE_FRAGMENT otherwise.
	// Returns nil if no separator found.
	if toks := splitCompoundDate(chunk, lang); toks != nil {
		return toks
	}

	// Bare integer
	return []Token{classifyBareInteger(chunk)}
}

// classifyBareInteger returns YEAR (4 digits), INTEGER2 (leading zero ≤2 digits), or INTEGER.
// Values are stored as int; the token type alone distinguishes leading-zero forms.
func classifyBareInteger(s string) Token {
	// len(s) > 10 can cause overflow in Atoi, panic
	if !allDigits(s) || len(s) > 10 {
		return Token{Type: TokenUnknown, Value: s}
	}
	n := mustAtoi(s)
	switch {
	case len(s) == 4:
		return Token{Type: TokenYear, Value: n}
	case len(s) == 2 && s[0] == '0':
		return Token{Type: TokenInteger2, Value: n}
	default:
		return Token{Type: TokenInteger, Value: n}
	}
}

// splitCompoundDate splits a chunk that contains a date separator (-, /, .)
// into component tokens. Returns nil if no separator is found.
//
// Rules:
//   - All separators in the chunk must be the same character (no mixing).
//   - Each piece is classified: 4 digits → YEAR, 1–3 digits → INTEGER,
//     letters → word-table lookup (e.g. "dec" → MONTH).
//   - Within a compound date, leading zeros on numeric pieces are ignored;
//     all 1–3 digit numeric pieces become INTEGER (not INTEGER2).
//   - If no YEAR or MONTH token is found among the pieces, the result is
//     ambiguous (unknown day/month order) → returns DATE_FRAGMENT.
func splitCompoundDate(chunk string, lang Lang) []Token {
	sep := byte(0)
	for i := 0; i < len(chunk); i++ {
		c := chunk[i]
		if c == '-' || c == '/' || c == '.' {
			if sep == 0 {
				sep = c
			} else if sep != c {
				// Mixed separators — unrecognised
				return []Token{{Type: TokenDateFragment, Value: chunk}}
			}
		}
	}
	if sep == 0 {
		return nil // no separator; let caller handle as bare integer
	}

	parts := strings.Split(chunk, string(sep))
	if len(parts) < 2 || len(parts) > 3 {
		return []Token{{Type: TokenDateFragment, Value: chunk}}
	}

	tokens := make([]Token, 0, len(parts))
	hasYearOrMonth := false
	for _, part := range parts {
		if len(part) == 0 {
			return []Token{{Type: TokenDateFragment, Value: chunk}}
		}
		var tok Token
		switch {
		case isDigitByte(part[0]):
			if !allDigits(part) || len(part) > 10 {
				return []Token{{Type: TokenDateFragment, Value: chunk}}
			}
			if len(part) == 4 {
				tok = Token{Type: TokenYear, Value: mustAtoi(part)}
				hasYearOrMonth = true
			} else {
				tok = Token{Type: TokenInteger, Value: mustAtoi(part)} // leading zeros ignored in date context
			}
		case isLetterByte(part[0]):
			tok = classifyWord(part, lang)
			if tok.Type == TokenMonth {
				hasYearOrMonth = true
			}
		default:
			return []Token{{Type: TokenDateFragment, Value: chunk}}
		}
		tokens = append(tokens, tok)
	}

	if !hasYearOrMonth {
		// Dot-separated all-digit pairs are decimal numbers, not date fragments.
		// "3.5" → TokenDecimal 3.5;  "12-03" (hyphen) stays TokenDateFragment.
		if sep == '.' && len(parts) == 2 {
			if f, err := strconv.ParseFloat(chunk, 64); err == nil {
				return []Token{{Type: TokenDecimal, Value: f}}
			}
		}
		return []Token{{Type: TokenDateFragment, Value: chunk}}
	}
	return tokens
}

// isTime matches H:MM, HH:MM, H:MM:SS, or HH:MM:SS.
func isTime(s string) bool {
	// lengths: 4 (H:MM), 5 (HH:MM), 7 (H:MM:SS), 8 (HH:MM:SS)
	if len(s) < 4 || len(s) > 8 {
		return false
	}
	colon := strings.IndexByte(s, ':')
	if colon < 1 || colon > 2 {
		return false
	}
	if !allDigits(s[:colon]) {
		return false
	}
	rest := s[colon+1:]
	if len(rest) == 2 {
		return allDigits(rest) // H:MM or HH:MM
	}
	if len(rest) == 5 && rest[2] == ':' {
		return allDigits(rest[:2]) && allDigits(rest[3:]) // H:MM:SS or HH:MM:SS
	}
	return false
}

// gluedAMPM returns the index where "am" or "pm" starts, or 0 if not found.
func gluedAMPM(s string) int {
	if len(s) < 3 {
		return 0
	}
	suffix := s[len(s)-2:]
	if (suffix == "am" || suffix == "pm") && allDigits(s[:len(s)-2]) {
		return len(s) - 2
	}
	return 0
}

// splitTimeAndZone splits a chunk that is a TIME immediately followed by a
// numeric timezone offset, e.g. "09:30:00-07:00" or "9:30+05:30".
// Returns [TIME, TIMEZONE] or nil if the chunk doesn't match that pattern.
func splitTimeAndZone(chunk string) []Token {
	// Scan for a + or - after position 0 that could start a timezone offset.
	for i := 1; i < len(chunk); i++ {
		if chunk[i] == '+' || chunk[i] == '-' {
			timePart := chunk[:i]
			zonePart := chunk[i:]
			if isTime(timePart) && isTimezoneOffset(zonePart) {
				return []Token{
					{Type: TokenTime, Value: timePart},
					{Type: TokenTimezone, Value: zonePart},
				}
			}
		}
	}
	return nil
}

// isTimezoneOffset reports whether s is a numeric timezone offset.
// Valid forms: +HH, -HH, +HHMM, -HHMM, +HH:MM, -HH:MM.
func isTimezoneOffset(s string) bool {
	if len(s) < 3 {
		return false
	}
	if s[0] != '+' && s[0] != '-' {
		return false
	}
	rest := s[1:]
	switch len(rest) {
	case 2: // +HH
		return allDigits(rest)
	case 4: // +HHMM
		return allDigits(rest)
	case 5: // +HH:MM
		return rest[2] == ':' && allDigits(rest[:2]) && allDigits(rest[3:])
	}
	return false
}

// ---------------------------------------------------------------------------
// Character helpers
// ---------------------------------------------------------------------------

func isDigit(r rune) bool      { return r >= '0' && r <= '9' }
func isDigitByte(b byte) bool  { return b >= '0' && b <= '9' }
func isLetterByte(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }
func isSpaceByte(b byte) bool  { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }

func allDigits(s string) bool {
	for _, c := range s {
		if !isDigit(c) {
			return false
		}
	}
	return len(s) > 0
}
