package languages

import (
	"fmt"
	. "github.com/client9/nowandlater/internal/engine"
	"strings"
	"unicode/utf8"
)

// LangJa is the built-in Japanese Lang.
//
// Known limitations:
//   - Kanji day numbers (二十四日) are not supported; use Arabic numerals (24日).
//   - Kanji year numbers (二〇二六年) are not supported; use Arabic numerals (2026年).
//   - Irregular day-of-month readings (ついたち, ふつか, etc.) are not supported.
//   - Full-width numerals (３月２４日) are normalized to ASCII automatically.
var LangJa = Lang{
	Words:         japaneseWords,
	Handlers:      japaneseHandlers,
	TokenizerFunc: japaneseTokenize,
}

// japaneseHandlers overrides the global dispatch map for signatures that are
// specific to Japanese word order (AMPM before TIME, as in "午後3時30分").
var japaneseHandlers = map[string]Handler{
	"AMPM TIME":                    handleAMPMTime,
	"ANCHOR AMPM TIME":             handleAnchorAMPMTime,
	"YEAR MONTH INTEGER AMPM TIME": handleYearMonthIntegerAMPMTime,
}

// japaneseWords is the word table for Japanese.
// All entries are looked up by longest-match during character-level scanning.
// Longer keys always win over shorter ones sharing the same prefix (e.g.
// "月曜日" beats bare "月"; "十一月" beats "十月").
var japaneseWords = map[string]WordEntry{
	// --- Anchors ---
	"今":   {Type: TokenAnchor, Value: AnchorNow},
	"現在":  {Type: TokenAnchor, Value: AnchorNow}, // formal synonym for 今
	"今日":  {Type: TokenAnchor, Value: AnchorToday},
	"明日":  {Type: TokenAnchor, Value: AnchorTomorrow},
	"昨日":  {Type: TokenAnchor, Value: AnchorYesterday},
	"一昨日": {Type: TokenAnchor, Value: Anchor2DaysAgo},
	"明後日": {Type: TokenAnchor, Value: Anchor2DaysFromNow},

	// --- AM / PM ---
	"午前": {Type: TokenAMPM, Value: AMPMAm},
	"午後": {Type: TokenAMPM, Value: AMPMPm},

	// --- Time-word substitutes ---
	"正午": {Type: TokenTime, Value: "12:00"}, // noon

	// --- Weekdays: full form (月曜日) and single-kanji abbreviation (月) ---
	// Full forms win via longest-match when both are present.
	"月曜日": {Type: TokenWeekday, Value: WeekdayMonday},
	"火曜日": {Type: TokenWeekday, Value: WeekdayTuesday},
	"水曜日": {Type: TokenWeekday, Value: WeekdayWednesday},
	"木曜日": {Type: TokenWeekday, Value: WeekdayThursday},
	"金曜日": {Type: TokenWeekday, Value: WeekdayFriday},
	"土曜日": {Type: TokenWeekday, Value: WeekdaySaturday},
	"日曜日": {Type: TokenWeekday, Value: WeekdaySunday},
	"月":   {Type: TokenWeekday, Value: WeekdayMonday},    // abbreviation (also in 月曜日 — longest match wins)
	"火":   {Type: TokenWeekday, Value: WeekdayTuesday},   // abbreviation
	"水":   {Type: TokenWeekday, Value: WeekdayWednesday}, // abbreviation
	"木":   {Type: TokenWeekday, Value: WeekdayThursday},  // abbreviation
	"金":   {Type: TokenWeekday, Value: WeekdayFriday},    // abbreviation
	"土":   {Type: TokenWeekday, Value: WeekdaySaturday},  // abbreviation
	"日":   {Type: TokenWeekday, Value: WeekdaySunday},    // abbreviation (also in 日曜日 and 今日 — longest match wins)

	// --- Kanji month names (一月–十二月) ---
	// Longer keys (十一月, 十二月) beat shorter (十月) when both share a prefix.
	"一月":  {Type: TokenMonth, Value: MonthJanuary},
	"二月":  {Type: TokenMonth, Value: MonthFebruary},
	"三月":  {Type: TokenMonth, Value: MonthMarch},
	"四月":  {Type: TokenMonth, Value: MonthApril},
	"五月":  {Type: TokenMonth, Value: MonthMay},
	"六月":  {Type: TokenMonth, Value: MonthJune},
	"七月":  {Type: TokenMonth, Value: MonthJuly},
	"八月":  {Type: TokenMonth, Value: MonthAugust},
	"九月":  {Type: TokenMonth, Value: MonthSeptember},
	"十月":  {Type: TokenMonth, Value: MonthOctober},
	"十一月": {Type: TokenMonth, Value: MonthNovember},
	"十二月": {Type: TokenMonth, Value: MonthDecember},

	// --- Temporal direction modifiers ---
	"後": {Type: TokenModifier, Value: ModifierFuture}, // 3日後 = 3 days later
	"前": {Type: TokenModifier, Value: ModifierPast},   // 2週間前 = 2 weeks ago

	// --- Filler particle ---
	"の": {Type: TokenFiller, Value: nil},
}

// japaneseMacros maps compound words to their multi-token expansion.
// These cannot be expressed as a single WordEntry because they emit more than one token.
// Matched longest-first before lang.Words.
var japaneseMacros = map[string][]Token{
	// Two-token: direction + unit
	"来週": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenUnit, Value: PeriodWeek}},
	"先週": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodWeek}},
	"今週": {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodWeek}},
	"来月": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenUnit, Value: PeriodMonth}},
	"先月": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodMonth}},
	"今月": {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodMonth}},
	"来年": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenUnit, Value: PeriodYear}},
	"去年": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodYear}},
	"昨年": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodYear}},
	"今年": {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodYear}},

	// Three-token: integer + unit + modifier (relative anchors with no Arabic numeral)
	"先々週": {{Type: TokenInteger, Value: 2}, {Type: TokenUnit, Value: PeriodWeek}, {Type: TokenModifier, Value: ModifierPast}},   // week before last
	"再来週": {{Type: TokenInteger, Value: 2}, {Type: TokenUnit, Value: PeriodWeek}, {Type: TokenModifier, Value: ModifierFuture}}, // week after next
}

// japaneseTokenize is the custom tokenizer for Japanese. Japanese text contains
// no word-separating spaces, so tokenization is done by character-level scanning
// rather than whitespace splitting.
//
// At each position the scanner tries (in order):
//  1. japaneseMacros longest-match (direction+unit compounds like 来週, 先月)
//  2. lang.Words longest-match (anchors, weekdays, AM/PM, particles, etc.)
//  3. Digit run followed by a kanji unit suffix (年, 月, 日, 時, 時間, etc.)
//  4. Skip unknown character silently
func japaneseTokenize(input string, lang *Lang) []Token {
	input = jaFullWidthNorm(input)
	tokens := make([]Token, 0, 8)
	i := 0
	for i < len(input) {
		// 0. Imperial era prefix (令和7年, 平成元年, etc.) — must precede digit check
		if toks, n, ok := jaEraMatch(input[i:]); ok {
			tokens = append(tokens, toks...)
			i += n
			continue
		}
		// 1. Macro longest-match
		if toks, n, ok := jaMacroMatch(input[i:]); ok {
			tokens = append(tokens, toks...)
			i += n
			continue
		}
		// 2. Words longest-match
		if entry, n, ok := jaWordMatch(input[i:], lang.Words); ok {
			tokens = append(tokens, Token(entry))
			i += n
			continue
		}
		// 3. Digit run
		if IsDigitByte(input[i]) {
			toks, n := jaParseNumber(input, i)
			tokens = append(tokens, toks...)
			i += n
			continue
		}
		// 4. Skip unknown rune
		_, size := utf8.DecodeRuneInString(input[i:])
		i += size
	}
	return tokens
}

// jaMacroMatch returns the longest matching token slice from japaneseMacros.
func jaMacroMatch(s string) ([]Token, int, bool) {
	bestLen := 0
	var bestToks []Token
	for key, toks := range japaneseMacros {
		if len(key) > bestLen && strings.HasPrefix(s, key) {
			bestLen = len(key)
			bestToks = toks
		}
	}
	if bestLen > 0 {
		return bestToks, bestLen, true
	}
	return nil, 0, false
}

// jaWordMatch returns the longest matching WordEntry from words.
func jaWordMatch(s string, words map[string]WordEntry) (WordEntry, int, bool) {
	bestLen := 0
	var best WordEntry
	for key, entry := range words {
		if len(key) > bestLen && strings.HasPrefix(s, key) {
			bestLen = len(key)
			best = entry
		}
	}
	if bestLen > 0 {
		return best, bestLen, true
	}
	return WordEntry{}, 0, false
}

// jaParseNumber parses a digit run beginning at start in input, then inspects
// the following kanji unit suffix to classify the token(s).
// Returns the produced tokens and the total bytes consumed.
func jaParseNumber(input string, start int) ([]Token, int) {
	// Collect digit run
	i := start
	for i < len(input) && IsDigitByte(input[i]) {
		i++
	}
	digits := input[start:i]
	n := MustAtoi(digits)
	rest := input[i:]

	switch {
	case strings.HasPrefix(rest, "時間"): // hour duration — must check before bare "時"
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodHour},
		}, (i - start) + len("時間")

	case strings.HasPrefix(rest, "時"): // time component — build H:MM[:SS]
		return jaParseTime(input, start, n)

	case strings.HasPrefix(rest, "週間"): // week duration
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodWeek},
		}, (i - start) + len("週間")

	case strings.HasPrefix(rest, "ヶ月"), strings.HasPrefix(rest, "か月"),
		strings.HasPrefix(rest, "カ月"): // month duration (ヶ/か/カ variants)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodMonth},
		}, (i - start) + len("ヶ月") // all three variants are 6 bytes in UTF-8

	case strings.HasPrefix(rest, "年"): // year
		if len(digits) == 4 {
			return []Token{{Type: TokenYear, Value: n}}, (i - start) + len("年")
		}
		return []Token{{Type: TokenInteger, Value: n}}, (i - start) + len("年")

	case strings.HasPrefix(rest, "月"): // month (1–12)
		if n >= 1 && n <= 12 {
			return []Token{{Type: TokenMonth, Value: Month(n)}}, (i - start) + len("月")
		}
		return []Token{{Type: TokenInteger, Value: n}}, (i - start) + len("月")

	case strings.HasPrefix(rest, "日間"): // day duration with explicit 間 suffix
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodDay},
		}, (i - start) + len("日間")

	case strings.HasPrefix(rest, "日"): // day-of-month or day duration (bare 日)
		// Disambiguate: if immediately followed by 後 or 前, treat as a duration unit.
		afterDay := rest[len("日"):]
		if strings.HasPrefix(afterDay, "後") || strings.HasPrefix(afterDay, "前") {
			return []Token{
				{Type: TokenInteger, Value: n},
				{Type: TokenUnit, Value: PeriodDay},
			}, (i - start) + len("日")
		}
		return []Token{{Type: TokenInteger, Value: n}}, (i - start) + len("日")

	case strings.HasPrefix(rest, "分間"): // minute duration with explicit 間 suffix
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodMinute},
		}, (i - start) + len("分間")

	case strings.HasPrefix(rest, "分"): // minute duration (bare 分, standalone)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodMinute},
		}, (i - start) + len("分")

	case strings.HasPrefix(rest, "秒間"): // second duration with explicit 間 suffix
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodSecond},
		}, (i - start) + len("秒間")

	case strings.HasPrefix(rest, "秒"): // second duration (bare 秒)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodSecond},
		}, (i - start) + len("秒")

	default: // no unit suffix
		return []Token{ClassifyBareInteger(digits)}, i - start
	}
}

// jaParseTime handles a time expression that starts with a digit run at position
// start in input, followed by "時". It greedily consumes the optional following
// digit+"分" and digit+"秒" components, then emits a single TokenTime.
func jaParseTime(input string, start, hour int) ([]Token, int) {
	// Advance past the digit run
	i := start
	for i < len(input) && IsDigitByte(input[i]) {
		i++
	}
	i += len("時") // consume 時 (3 bytes in UTF-8)

	minute, second := 0, 0

	// Optional: digit run + 分
	if i < len(input) && IsDigitByte(input[i]) {
		j := i
		for j < len(input) && IsDigitByte(input[j]) {
			j++
		}
		if strings.HasPrefix(input[j:], "分") {
			minute = MustAtoi(input[i:j])
			i = j + len("分")

			// Optional: digit run + 秒
			if i < len(input) && IsDigitByte(input[i]) {
				k := i
				for k < len(input) && IsDigitByte(input[k]) {
					k++
				}
				if strings.HasPrefix(input[k:], "秒") {
					second = MustAtoi(input[i:k])
					i = k + len("秒")
				}
			}
		}
	}

	var timeStr string
	if second != 0 {
		timeStr = fmt.Sprintf("%d:%02d:%02d", hour, minute, second)
	} else {
		timeStr = fmt.Sprintf("%d:%02d", hour, minute)
	}
	return []Token{{Type: TokenTime, Value: timeStr}}, i - start
}

// jaFullWidthNorm replaces full-width ASCII digits (０-９, U+FF10–U+FF19)
// with their ASCII equivalents. Japanese input occasionally uses these.
func jaFullWidthNorm(s string) string {
	// Quick check: all full-width digits share the first byte 0xEF.
	hasFW := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0xEF {
			hasFW = true
			break
		}
	}
	if !hasFW {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '０' && r <= '９' {
			b.WriteByte(byte('0' + (r - '０')))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Imperial era support
// ---------------------------------------------------------------------------

// japaneseEras maps era name prefixes to their Gregorian start year.
// Era year 1 equals the base year (e.g. 令和1年 = 2019, 令和7年 = 2025).
var japaneseEras = []struct {
	name     string
	baseYear int
}{
	{"令和", 2019}, // Reiwa: 2019–
	{"平成", 1989}, // Heisei: 1989–2019
	{"昭和", 1926}, // Showa: 1926–1989
	{"大正", 1912}, // Taisho: 1912–1926
	{"明治", 1868}, // Meiji: 1868–1912
}

// jaEraMatch checks whether s begins with a known era name followed by either
// "元年" (year 1) or ASCII digits + "年". On success it returns a single
// TokenYear and the number of bytes consumed.
func jaEraMatch(s string) ([]Token, int, bool) {
	for _, era := range japaneseEras {
		if !strings.HasPrefix(s, era.name) {
			continue
		}
		rest := s[len(era.name):]

		// 元年 = era year 1 (= base year)
		if strings.HasPrefix(rest, "元年") {
			return []Token{{Type: TokenYear, Value: era.baseYear}},
				len(era.name) + len("元年"), true
		}

		// digit run + 年
		if len(rest) > 0 && IsDigitByte(rest[0]) {
			j := 0
			for j < len(rest) && IsDigitByte(rest[j]) {
				j++
			}
			if strings.HasPrefix(rest[j:], "年") {
				eraYear := MustAtoi(rest[:j])
				return []Token{{Type: TokenYear, Value: era.baseYear + eraYear - 1}},
					len(era.name) + j + len("年"), true
			}
		}
	}
	return nil, 0, false
}

// ---------------------------------------------------------------------------
// Handlers for Japanese (and other AMPM-prefix) token orders
// ---------------------------------------------------------------------------

// handleAMPMTime handles: AMPM TIME
// Example: 午後3時30分 → 15:30
func handleAMPMTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens) // [AMPM, TIME]
	timeVal := toks[1].Value.(string)
	h, m, sec := MustParseTime(timeVal)
	h = ApplyAMPM(h, toks[0].Value.(AMPM))
	return &ParsedDateSlots{Hour: h, Minute: m, Second: sec, Period: TimePeriod(timeVal)}, nil
}

// handleAnchorAMPMTime handles: ANCHOR AMPM TIME
// Example: 明日の午後3時30分 → tomorrow at 15:30
func handleAnchorAMPMTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens) // [ANCHOR, AMPM, TIME]
	aDelta := AnchorToSeconds[toks[0].Value.(Anchor)]
	timeVal := toks[2].Value.(string)
	h, m, sec := MustParseTime(timeVal)
	h = ApplyAMPM(h, toks[1].Value.(AMPM))
	slots := &ParsedDateSlots{
		DeltaSeconds: new(aDelta),
		Hour:         h,
		Minute:       m,
		Period:       TimePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleYearMonthIntegerAMPMTime handles: YEAR MONTH INTEGER AMPM TIME
// Example: 2026年3月24日の午後3時 → 2026-03-24 15:00
// Japanese word order places AMPM before TIME (cf. English "3:00 PM").
func handleYearMonthIntegerAMPMTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens) // [YEAR, MONTH, INTEGER, AMPM, TIME]
	y := toks[0].Value.(int)
	d := toks[2].Value.(int)
	timeVal := toks[4].Value.(string)
	h, m, sec := MustParseTime(timeVal)
	h = ApplyAMPM(h, toks[3].Value.(AMPM))
	slots := &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[1].Value.(Month)),
		Day:    d,
		Hour:   h,
		Minute: m,
		Period: TimePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}
