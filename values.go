package nowandlater

import (
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Semantic lookup tables
// ---------------------------------------------------------------------------

// periodToSeconds maps a Period constant to its approximate number of seconds.
// Month and year values are approximations (30-day month, 365-day year);
// the resolver layer is responsible for calendar-correct arithmetic.
var periodToSeconds = map[Period]int{
	PeriodSecond:    1,
	PeriodMinute:    60,
	PeriodHour:      3600,
	PeriodDay:       86400,
	PeriodFortnight: 86400 * 14, // 1209600
	PeriodWeek:      604800,
	PeriodMonth:     2592000,  // 30 days
	PeriodYear:      31536000, // 365 days
}

// anchorToSeconds maps an Anchor constant to its offset from "now" in seconds.
// Used by handleRelativeDeltaAnchor to compute combined deltas like
// "3 days before tomorrow" = anchorToSeconds[AnchorTomorrow] + (-3 * 86400).
var anchorToSeconds = map[Anchor]int{
	AnchorNow:          0,
	AnchorToday:        0,
	AnchorTomorrow:     86400,
	AnchorYesterday:    -86400,
	Anchor2DaysAgo:     -172800,
	Anchor2DaysFromNow: 172800,
	Anchor3DaysAgo:     -259200,
	Anchor3DaysFromNow: 259200,
}

// ---------------------------------------------------------------------------
// Extraction helpers
// ---------------------------------------------------------------------------

// filterFillers returns a new slice containing only non-FILLER tokens.
// Handlers call this first so positional indices align with the signature.
func filterFillers(tokens []Token) []Token {
	out := make([]Token, 0, len(tokens))
	for _, t := range tokens {
		if t.Type != TokenFiller {
			out = append(out, t)
		}
	}
	return out
}

// mustAtoi converts a token value string to an int, panicking if it fails.
// Token values for INTEGER and YEAR tokens are guaranteed digit-only by the
// tokenizer, so Atoi failure here indicates a programming error.
func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic("nowandlater: mustAtoi: invalid token value " + s)
	}
	return n
}

// mustParseTime parses "H:MM", "HH:MM", "H:MM:SS", or "HH:MM:SS", panicking
// on failure. TIME token values are guaranteed well-formed by the tokenizer.
func mustParseTime(s string) (hour, minute, second int) {
	colon := strings.IndexByte(s, ':')
	if colon < 1 {
		panic("nowandlater: mustParseTime: invalid time token " + s)
	}
	h := mustAtoi(s[:colon])
	rest := s[colon+1:]
	var m, sec int
	switch len(rest) {
	case 2:
		m = mustAtoi(rest)
	case 5:
		if rest[2] != ':' {
			panic("nowandlater: mustParseTime: invalid time token " + s)
		}
		m = mustAtoi(rest[:2])
		sec = mustAtoi(rest[3:])
	default:
		panic("nowandlater: mustParseTime: invalid time token " + s)
	}
	return h, m, sec
}

// timePeriod returns PeriodSecond if the TIME token value includes seconds, PeriodMinute otherwise.
func timePeriod(timeValue string) Period {
	if strings.Count(timeValue, ":") == 2 {
		return PeriodSecond
	}
	return PeriodMinute
}

// parseAMPM converts a raw "am"/"pm" string (from the tokenizer's number sub-parser)
// to a typed AMPM constant. Called only at tokenization time for glued suffixes like
// "3pm" or "7:15am"; all other AMPM tokens come from the Words map.
func parseAMPM(s string) AMPM {
	if s == "am" {
		return AMPMAm
	}
	return AMPMPm
}

// expand2DigitYear converts a 2-digit year to a 4-digit year using the
// RFC 2822 rule: 00–49 → 2000+y, 50–99 → 1900+y.
// Values ≥ 100 are returned unchanged (already a full year).
func expand2DigitYear(y int) int {
	if y >= 100 {
		return y
	}
	if y < 50 {
		return 2000 + y
	}
	return 1900 + y
}

// applyAMPM converts a 12-hour clock value to 24-hour using an AMPM token value.
// Handles the edge cases: 12 AM = 0 (midnight), 12 PM = 12 (noon).
func applyAMPM(hour int, ampm AMPM) int {
	switch ampm {
	case AMPMAm:
		if hour == 12 {
			return 0
		}
		return hour
	case AMPMPm:
		if hour == 12 {
			return 12
		}
		return hour + 12
	}
	return hour
}
