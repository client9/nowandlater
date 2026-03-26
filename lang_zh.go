package nowandlater

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Chinese is the built-in Simplified Chinese Lang.
//
// Like Japanese, Mandarin Chinese uses no word-separating spaces, so a
// custom character-level tokenizer is required. The tokenizer shares the
// same three-stage design (macros → words → digit+suffix), but with
// Simplified Chinese characters and time-unit vocabulary.
//
// Known limitations:
//   - Traditional Chinese characters are largely not supported (e.g. 時 vs 时);
//     個月 (traditional "month" measure word) is the only exception.
//   - Character ordinal day numbers (二十四日) are not supported; use digits (24日).
//   - Full-width digits are not normalised (use ASCII digits).
var Chinese = Lang{
	Words:         chineseWords,
	Handlers:      chineseHandlers,
	TokenizerFunc: chineseTokenize,
}

// chineseHandlers overrides the global dispatch map for signatures specific
// to Chinese/East-Asian word order (AMPM before TIME, as in 下午3点).
var chineseHandlers = map[string]Handler{
	"AMPM TIME":                    handleAMPMTime,
	"ANCHOR AMPM TIME":             handleAnchorAMPMTime,
	"YEAR MONTH INTEGER AMPM TIME": handleYearMonthIntegerAMPMTime,
}

// chineseWords is the word table for Simplified Chinese.
var chineseWords = map[string]WordEntry{
	// --- Anchors ---
	"现在":   {TokenAnchor, AnchorNow},
	"刚刚":   {TokenAnchor, AnchorNow}, // just now; supplementary data
	"此时":   {TokenAnchor, AnchorNow}, // at this moment; supplementary data
	"这一时间": {TokenAnchor, AnchorNow}, // at this time; supplementary data
	"今天":   {TokenAnchor, AnchorToday},
	"今日":   {TokenAnchor, AnchorToday},
	"明天":   {TokenAnchor, AnchorTomorrow},
	"明日":   {TokenAnchor, AnchorTomorrow},
	"昨天":   {TokenAnchor, AnchorYesterday},
	"昨日":   {TokenAnchor, AnchorYesterday},
	"后天":   {TokenAnchor, Anchor2DaysFromNow},
	"前天":   {TokenAnchor, Anchor2DaysAgo},

	// --- AM / PM ---
	"上午": {TokenAMPM, AMPMAm}, // morning
	"下午": {TokenAMPM, AMPMPm}, // afternoon
	"凌晨": {TokenAMPM, AMPMAm}, // early morning (before dawn)
	"晚上": {TokenAMPM, AMPMPm}, // evening

	// --- Time-word substitutes ---
	"正午": {TokenTime, "12:00"}, // formal noon
	"中午": {TokenTime, "12:00"}, // colloquial noon
	"午夜": {TokenTime, "0:00"},  // midnight

	// --- Weekdays (星期X full form) ---
	// Longer keys win over shorter prefix "星期" via longest-match.
	"星期一": {TokenWeekday, WeekdayMonday},
	"星期二": {TokenWeekday, WeekdayTuesday},
	"星期三": {TokenWeekday, WeekdayWednesday},
	"星期四": {TokenWeekday, WeekdayThursday},
	"星期五": {TokenWeekday, WeekdayFriday},
	"星期六": {TokenWeekday, WeekdaySaturday},
	"星期日": {TokenWeekday, WeekdaySunday},
	"星期天": {TokenWeekday, WeekdaySunday}, // colloquial Sunday

	// --- Weekdays (礼拜X colloquial form — supplementary data) ---
	"礼拜一": {TokenWeekday, WeekdayMonday},
	"礼拜二": {TokenWeekday, WeekdayTuesday},
	"礼拜三": {TokenWeekday, WeekdayWednesday},
	"礼拜四": {TokenWeekday, WeekdayThursday},
	"礼拜五": {TokenWeekday, WeekdayFriday},
	"礼拜六": {TokenWeekday, WeekdaySaturday},
	"礼拜日": {TokenWeekday, WeekdaySunday},
	"礼拜天": {TokenWeekday, WeekdaySunday},

	// --- Weekdays (周X abbreviated form) ---
	"周一": {TokenWeekday, WeekdayMonday},
	"周二": {TokenWeekday, WeekdayTuesday},
	"周三": {TokenWeekday, WeekdayWednesday},
	"周四": {TokenWeekday, WeekdayThursday},
	"周五": {TokenWeekday, WeekdayFriday},
	"周六": {TokenWeekday, WeekdaySaturday},
	"周日": {TokenWeekday, WeekdaySunday},
	"周天": {TokenWeekday, WeekdaySunday},

	// --- character month names ---
	// "十一月"/"十二月" (9 bytes) beat "十月" (6 bytes) via longest-match.
	"一月":  {TokenMonth, MonthJanuary},
	"二月":  {TokenMonth, MonthFebruary},
	"三月":  {TokenMonth, MonthMarch},
	"四月":  {TokenMonth, MonthApril},
	"五月":  {TokenMonth, MonthMay},
	"六月":  {TokenMonth, MonthJune},
	"七月":  {TokenMonth, MonthJuly},
	"八月":  {TokenMonth, MonthAugust},
	"九月":  {TokenMonth, MonthSeptember},
	"十月":  {TokenMonth, MonthOctober},
	"十一月": {TokenMonth, MonthNovember},
	"十二月": {TokenMonth, MonthDecember},

	// --- Relative direction modifiers ---
	// "前天"/"后天" (6 bytes) beat bare "前"/"后" (3 bytes) via longest-match.
	"后":  {TokenModifier, ModifierFuture}, // N后 = N later
	"以后": {TokenModifier, ModifierFuture},
	"前":  {TokenModifier, ModifierPast}, // N前 = N ago
	"以前": {TokenModifier, ModifierPast},

	// --- Filler particle ---
	"的": {TokenFiller, nil},
}

// chineseMacros maps compound words to their multi-token expansion.
// Checked before lang.Words so that longer compound forms win over shared prefixes.
var chineseMacros = map[string][]Token{
	// --- Next/last/this week ---
	"下周":   {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenUnit, Value: PeriodWeek}},
	"下个星期": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenUnit, Value: PeriodWeek}},
	"上周":   {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodWeek}},
	"上个星期": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodWeek}},
	"本周":   {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodWeek}},
	"这周":   {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodWeek}},
	"本星期":  {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodWeek}},

	// --- Next/last/this month ---
	"下个月": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenUnit, Value: PeriodMonth}},
	"上个月": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodMonth}},
	"本月":  {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodMonth}},
	"这个月": {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodMonth}},

	// --- Next/last/this year ---
	"明年": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenUnit, Value: PeriodYear}},
	"去年": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenUnit, Value: PeriodYear}},
	"今年": {{Type: TokenDirection, Value: DirectionNearest}, {Type: TokenUnit, Value: PeriodYear}},

	// --- Next week + weekday: 下周X → [DIRECTION(Future), WEEKDAY(X)] ---
	"下周一": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdayMonday}},
	"下周二": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdayTuesday}},
	"下周三": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdayWednesday}},
	"下周四": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdayThursday}},
	"下周五": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdayFriday}},
	"下周六": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdaySaturday}},
	"下周日": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdaySunday}},
	"下周天": {{Type: TokenDirection, Value: DirectionFuture}, {Type: TokenWeekday, Value: WeekdaySunday}},

	// --- Last week + weekday: 上周X → [DIRECTION(Past), WEEKDAY(X)] ---
	"上周一": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdayMonday}},
	"上周二": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdayTuesday}},
	"上周三": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdayWednesday}},
	"上周四": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdayThursday}},
	"上周五": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdayFriday}},
	"上周六": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdaySaturday}},
	"上周日": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdaySunday}},
	"上周天": {{Type: TokenDirection, Value: DirectionPast}, {Type: TokenWeekday, Value: WeekdaySunday}},
}

// chineseTokenize is the custom tokenizer for Simplified Chinese.
// At each position the scanner tries (in order):
//  1. chineseMacros longest-match (direction+unit/weekday compounds like 下周, 上周五)
//  2. lang.Words longest-match (anchors, AM/PM, weekdays, particles, etc.)
//  3. Digit run followed by a Chinese unit suffix (年, 月, 日, 天, 点, 时, 分, etc.)
//  4. Skip unknown character silently
func chineseTokenize(input string, lang *Lang) []Token {
	tokens := make([]Token, 0, 8)
	i := 0
	for i < len(input) {
		// 1. Macro longest-match
		if toks, n, ok := zhMacroMatch(input[i:]); ok {
			tokens = append(tokens, toks...)
			i += n
			continue
		}
		// 2. Words longest-match
		if entry, n, ok := zhWordMatch(input[i:], lang.Words); ok {
			tokens = append(tokens, Token(entry))
			i += n
			continue
		}
		// 3. Digit run
		if isDigitByte(input[i]) {
			toks, n := zhParseNumber(input, i)
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

// zhMacroMatch returns the longest matching token slice from chineseMacros.
func zhMacroMatch(s string) ([]Token, int, bool) {
	bestLen := 0
	var bestToks []Token
	for key, toks := range chineseMacros {
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

// zhWordMatch returns the longest matching WordEntry from words.
func zhWordMatch(s string, words map[string]WordEntry) (WordEntry, int, bool) {
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

// zhParseNumber parses a digit run beginning at start in input, then inspects
// the following Chinese unit suffix to classify the token(s).
func zhParseNumber(input string, start int) ([]Token, int) {
	i := start
	for i < len(input) && isDigitByte(input[i]) {
		i++
	}
	digits := input[start:i]
	n := mustAtoi(digits)
	rest := input[i:]

	switch {
	case strings.HasPrefix(rest, "小时"): // hour duration (must check before bare 时)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodHour},
		}, (i - start) + len("小时")

	case strings.HasPrefix(rest, "分钟"): // minute (must check before bare 分)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodMinute},
		}, (i - start) + len("分钟")

	case strings.HasPrefix(rest, "个月") || strings.HasPrefix(rest, "個月"): // month (个月 simplified / 個月 traditional)
		suffixLen := len("个月") // both are 6 bytes
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodMonth},
		}, (i - start) + suffixLen

	case strings.HasPrefix(rest, "星期"): // week (must check before 期 alone)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodWeek},
		}, (i - start) + len("星期")

	case strings.HasPrefix(rest, "年"): // year
		if len(digits) == 4 {
			return []Token{{Type: TokenYear, Value: n}}, (i - start) + len("年")
		}
		// Short count (e.g. "2年前" = 2 years ago): emit INTEGER + UNIT so that
		// INTEGER UNIT MODIFIER and PREP INTEGER UNIT handlers apply.
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodYear},
		}, (i - start) + len("年")

	case strings.HasPrefix(rest, "月"): // month (1–12)
		if n >= 1 && n <= 12 {
			return []Token{{Type: TokenMonth, Value: Month(n)}}, (i - start) + len("月")
		}
		return []Token{{Type: TokenInteger, Value: n}}, (i - start) + len("月")

	case strings.HasPrefix(rest, "日"): // day-of-month or day duration
		// If immediately followed by 后 or 前, treat as duration.
		afterDay := rest[len("日"):]
		if strings.HasPrefix(afterDay, "后") || strings.HasPrefix(afterDay, "前") {
			return []Token{
				{Type: TokenInteger, Value: n},
				{Type: TokenUnit, Value: PeriodDay},
			}, (i - start) + len("日")
		}
		return []Token{{Type: TokenInteger, Value: n}}, (i - start) + len("日")

	case strings.HasPrefix(rest, "天"): // day as duration (always a unit, unlike 日)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodDay},
		}, (i - start) + len("天")

	case strings.HasPrefix(rest, "点") || strings.HasPrefix(rest, "时"): // o'clock → build H:MM
		return zhParseTime(input, start, n)

	case strings.HasPrefix(rest, "分"): // minute
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodMinute},
		}, (i - start) + len("分")

	case strings.HasPrefix(rest, "秒钟"): // second, emphatic form (must check before bare 秒)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodSecond},
		}, (i - start) + len("秒钟")

	case strings.HasPrefix(rest, "秒"): // second
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodSecond},
		}, (i - start) + len("秒")

	case strings.HasPrefix(rest, "周"): // week (1周, 2周)
		return []Token{
			{Type: TokenInteger, Value: n},
			{Type: TokenUnit, Value: PeriodWeek},
		}, (i - start) + len("周")

	default:
		return []Token{classifyBareInteger(digits)}, i - start
	}
}

// zhParseTime handles a time expression starting with a digit run at position
// start, followed by 点 or 时 (both meaning "o'clock"). It greedily consumes
// an optional digit+分 and digit+秒, then emits a single TokenTime.
func zhParseTime(input string, start, hour int) ([]Token, int) {
	i := start
	for i < len(input) && isDigitByte(input[i]) {
		i++
	}
	// Consume 点 or 时 (both are 3-byte UTF-8 characters)
	rest := input[i:]
	if strings.HasPrefix(rest, "点") {
		i += len("点")
	} else if strings.HasPrefix(rest, "时") {
		i += len("时")
	}

	minute, second := 0, 0

	// Optional: digit run + 分
	if i < len(input) && isDigitByte(input[i]) {
		j := i
		for j < len(input) && isDigitByte(input[j]) {
			j++
		}
		if strings.HasPrefix(input[j:], "分") {
			minute = mustAtoi(input[i:j])
			i = j + len("分")

			// Optional: digit run + 秒
			if i < len(input) && isDigitByte(input[i]) {
				k := i
				for k < len(input) && isDigitByte(input[k]) {
					k++
				}
				if strings.HasPrefix(input[k:], "秒") {
					second = mustAtoi(input[i:k])
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
