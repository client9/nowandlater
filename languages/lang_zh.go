package languages

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	. "github.com/client9/nowandlater/internal/engine"
)

// LangZh is the built-in Simplified Chinese Lang.
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
var LangZh = Lang{
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
	"现在":   {Type: TokenAnchor, Value: AnchorNow},
	"刚刚":   {Type: TokenAnchor, Value: AnchorNow}, // just now; supplementary data
	"此时":   {Type: TokenAnchor, Value: AnchorNow}, // at this moment; supplementary data
	"这一时间": {Type: TokenAnchor, Value: AnchorNow}, // at this time; supplementary data
	"今天":   {Type: TokenAnchor, Value: AnchorToday},
	"今日":   {Type: TokenAnchor, Value: AnchorToday},
	"明天":   {Type: TokenAnchor, Value: AnchorTomorrow},
	"明日":   {Type: TokenAnchor, Value: AnchorTomorrow},
	"昨天":   {Type: TokenAnchor, Value: AnchorYesterday},
	"昨日":   {Type: TokenAnchor, Value: AnchorYesterday},
	"后天":   {Type: TokenAnchor, Value: Anchor2DaysFromNow},
	"前天":   {Type: TokenAnchor, Value: Anchor2DaysAgo},

	// --- AM / PM ---
	"上午": {Type: TokenAMPM, Value: AMPMAm}, // morning
	"下午": {Type: TokenAMPM, Value: AMPMPm}, // afternoon
	"凌晨": {Type: TokenAMPM, Value: AMPMAm}, // early morning (before dawn)
	"晚上": {Type: TokenAMPM, Value: AMPMPm}, // evening

	// --- Time-word substitutes ---
	"正午": {Type: TokenTime, Value: "12:00"}, // formal noon
	"中午": {Type: TokenTime, Value: "12:00"}, // colloquial noon
	"午夜": {Type: TokenTime, Value: "0:00"},  // midnight

	// --- Weekdays (星期X full form) ---
	// Longer keys win over shorter prefix "星期" via longest-match.
	"星期一": {Type: TokenWeekday, Value: WeekdayMonday},
	"星期二": {Type: TokenWeekday, Value: WeekdayTuesday},
	"星期三": {Type: TokenWeekday, Value: WeekdayWednesday},
	"星期四": {Type: TokenWeekday, Value: WeekdayThursday},
	"星期五": {Type: TokenWeekday, Value: WeekdayFriday},
	"星期六": {Type: TokenWeekday, Value: WeekdaySaturday},
	"星期日": {Type: TokenWeekday, Value: WeekdaySunday},
	"星期天": {Type: TokenWeekday, Value: WeekdaySunday}, // colloquial Sunday

	// --- Weekdays (礼拜X colloquial form — supplementary data) ---
	"礼拜一": {Type: TokenWeekday, Value: WeekdayMonday},
	"礼拜二": {Type: TokenWeekday, Value: WeekdayTuesday},
	"礼拜三": {Type: TokenWeekday, Value: WeekdayWednesday},
	"礼拜四": {Type: TokenWeekday, Value: WeekdayThursday},
	"礼拜五": {Type: TokenWeekday, Value: WeekdayFriday},
	"礼拜六": {Type: TokenWeekday, Value: WeekdaySaturday},
	"礼拜日": {Type: TokenWeekday, Value: WeekdaySunday},
	"礼拜天": {Type: TokenWeekday, Value: WeekdaySunday},

	// --- Weekdays (周X abbreviated form) ---
	"周一": {Type: TokenWeekday, Value: WeekdayMonday},
	"周二": {Type: TokenWeekday, Value: WeekdayTuesday},
	"周三": {Type: TokenWeekday, Value: WeekdayWednesday},
	"周四": {Type: TokenWeekday, Value: WeekdayThursday},
	"周五": {Type: TokenWeekday, Value: WeekdayFriday},
	"周六": {Type: TokenWeekday, Value: WeekdaySaturday},
	"周日": {Type: TokenWeekday, Value: WeekdaySunday},
	"周天": {Type: TokenWeekday, Value: WeekdaySunday},

	// --- character month names ---
	// "十一月"/"十二月" (9 bytes) beat "十月" (6 bytes) via longest-match.
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

	// --- Relative direction modifiers ---
	// "前天"/"后天" (6 bytes) beat bare "前"/"后" (3 bytes) via longest-match.
	"后":  {Type: TokenModifier, Value: ModifierFuture}, // N后 = N later
	"以后": {Type: TokenModifier, Value: ModifierFuture},
	"前":  {Type: TokenModifier, Value: ModifierPast}, // N前 = N ago
	"以前": {Type: TokenModifier, Value: ModifierPast},

	// --- Filler particle ---
	"的": {Type: TokenFiller, Value: nil},
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
		if IsDigitByte(input[i]) {
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
	for i < len(input) && IsDigitByte(input[i]) {
		i++
	}
	digits := input[start:i]
	n, err := strconv.Atoi(digits)
	if err != nil {
		return nil, i - start // skip the digit run, emit no token
	}
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
		return []Token{ClassifyBareInteger(digits)}, i - start
	}
}

// zhParseTime handles a time expression starting with a digit run at position
// start, followed by 点 or 时 (both meaning "o'clock"). It greedily consumes
// an optional digit+分 and digit+秒, then emits a single TokenTime.
func zhParseTime(input string, start, hour int) ([]Token, int) {
	i := start
	for i < len(input) && IsDigitByte(input[i]) {
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
	if i < len(input) && IsDigitByte(input[i]) {
		j := i
		for j < len(input) && IsDigitByte(input[j]) {
			j++
		}
		if strings.HasPrefix(input[j:], "分") {
			if m, err := strconv.Atoi(input[i:j]); err == nil {
				minute = m
				i = j + len("分")

				// Optional: digit run + 秒
				if i < len(input) && IsDigitByte(input[i]) {
					k := i
					for k < len(input) && IsDigitByte(input[k]) {
						k++
					}
					if strings.HasPrefix(input[k:], "秒") {
						if sec, err := strconv.Atoi(input[i:k]); err == nil {
							second = sec
							i = k + len("秒")
						}
					}
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
