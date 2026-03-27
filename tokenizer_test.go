package nowandlater

import (
	"testing"
)

// tt is a shorthand for building expected token lists.
// Each pair is (TokenType, value) where value is any typed constant or string.
func tt(pairs ...any) []Token {
	if len(pairs)%2 != 0 {
		panic("tt: must pass pairs of (TokenType, value)")
	}
	tokens := make([]Token, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		tokens = append(tokens, Token{
			Type:  pairs[i].(TokenType),
			Value: pairs[i+1],
		})
	}
	return tokens
}

var tokenizerCases = []struct {
	input  string
	tokens []Token
	sig    string
}{
	// --- Relative: direction + unit ---
	{
		"in 2 days",
		tt(TokenPrep, nil, TokenInteger, 2, TokenUnit, PeriodDay),
		"PREP INTEGER UNIT",
	},
	{
		"in next 2 days",
		tt(TokenPrep, nil, TokenDirection, DirectionFuture, TokenInteger, 2, TokenUnit, PeriodDay),
		"PREP DIRECTION INTEGER UNIT",
	},
	{
		"3 days ago",
		tt(TokenInteger, 3, TokenUnit, PeriodDay, TokenModifier, ModifierPast),
		"INTEGER UNIT MODIFIER",
	},
	{
		"2 weeks from now",
		tt(TokenInteger, 2, TokenUnit, PeriodWeek, TokenModifier, ModifierFuture, TokenAnchor, AnchorNow),
		"INTEGER UNIT MODIFIER ANCHOR",
	},

	// --- Anchors ---
	{"today", tt(TokenAnchor, AnchorToday), "ANCHOR"},
	{"tomorrow", tt(TokenAnchor, AnchorTomorrow), "ANCHOR"},
	{"yesterday", tt(TokenAnchor, AnchorYesterday), "ANCHOR"},
	{"now", tt(TokenAnchor, AnchorNow), "ANCHOR"},

	// --- Weekday with direction ---
	{
		"next Monday",
		tt(TokenDirection, DirectionFuture, TokenWeekday, WeekdayMonday),
		"DIRECTION WEEKDAY",
	},
	{
		"last Friday",
		tt(TokenDirection, DirectionPast, TokenWeekday, WeekdayFriday),
		"DIRECTION WEEKDAY",
	},

	// --- Weekday + time ---
	{
		"next Monday at 9:30 AM",
		tt(TokenDirection, DirectionFuture, TokenWeekday, WeekdayMonday, TokenPrep, nil, TokenTime, "9:30", TokenAMPM, AMPMAm),
		"DIRECTION WEEKDAY PREP TIME AMPM",
	},

	// --- Glued AM/PM ---
	{
		"3pm",
		tt(TokenInteger, 3, TokenAMPM, AMPMPm),
		"INTEGER AMPM",
	},
	{
		"11am",
		tt(TokenInteger, 11, TokenAMPM, AMPMAm),
		"INTEGER AMPM",
	},

	// --- Compact time notation (3-4 digit + glued AM/PM) ---
	{"730pm", tt(TokenTime, "7:30", TokenAMPM, AMPMPm), "TIME AMPM"},
	{"930pm", tt(TokenTime, "9:30", TokenAMPM, AMPMPm), "TIME AMPM"},
	{"1230am", tt(TokenTime, "12:30", TokenAMPM, AMPMAm), "TIME AMPM"},

	// --- Month + day ---
	{
		"March 5",
		tt(TokenMonth, MonthMarch, TokenInteger, 5),
		"MONTH INTEGER",
	},
	{
		"January 21st",
		tt(TokenMonth, MonthJanuary, TokenInteger, 21),
		"MONTH INTEGER",
	},
	{
		"Dec 3rd 2026",
		tt(TokenMonth, MonthDecember, TokenInteger, 3, TokenYear, 2026),
		"MONTH INTEGER YEAR",
	},

	// --- Compound date: numeric separators ---
	{
		"2026-12-04",
		tt(TokenYear, 2026, TokenInteger, 12, TokenInteger, 4),
		"YEAR INTEGER INTEGER",
	},
	{
		"2026/12/03",
		tt(TokenYear, 2026, TokenInteger, 12, TokenInteger, 3),
		"YEAR INTEGER INTEGER",
	},
	{
		"2026.12.03",
		tt(TokenYear, 2026, TokenInteger, 12, TokenInteger, 3),
		"YEAR INTEGER INTEGER",
	},
	{
		"on 2026-12-04",
		tt(TokenPrep, nil, TokenYear, 2026, TokenInteger, 12, TokenInteger, 4),
		"PREP YEAR INTEGER INTEGER",
	},

	// --- Compound date: letter month ---
	{
		"2026-dec-04",
		tt(TokenYear, 2026, TokenMonth, MonthDecember, TokenInteger, 4),
		"YEAR MONTH INTEGER",
	},
	{
		"04-dec-2026",
		tt(TokenInteger, 4, TokenMonth, MonthDecember, TokenYear, 2026),
		"INTEGER MONTH YEAR",
	},

	// --- Ambiguous fragment (no YEAR or MONTH) → DATE_FRAGMENT ---
	{
		"12-03",
		tt(TokenDateFragment, "12-03"),
		"DATE_FRAGMENT",
	},

	// --- Dotted abbreviations (normalization) ---
	{
		"next Mon. at 9:30 A.M.",
		tt(TokenDirection, DirectionFuture, TokenWeekday, WeekdayMonday, TokenPrep, nil, TokenTime, "9:30", TokenAMPM, AMPMAm),
		"DIRECTION WEEKDAY PREP TIME AMPM",
	},

	// --- Leading-zero INTEGER2 ---
	{
		"at 09:30",
		tt(TokenPrep, nil, TokenTime, "09:30"),
		"PREP TIME",
	},

	// --- Filler words ---
	{
		"the 3rd of January",
		tt(TokenFiller, nil, TokenInteger, 3, TokenFiller, nil, TokenMonth, MonthJanuary),
		"INTEGER MONTH",
	},

	// --- Number words (cardinal + ordinal) ---
	{
		"the first of March",
		tt(TokenFiller, nil, TokenInteger, 1, TokenFiller, nil, TokenMonth, MonthMarch),
		"INTEGER MONTH",
	},
	{
		"in three days",
		tt(TokenPrep, nil, TokenInteger, 3, TokenUnit, PeriodDay),
		"PREP INTEGER UNIT",
	},
	{
		"twenty-first of April",
		tt(TokenInteger, 21, TokenFiller, nil, TokenMonth, MonthApril),
		"INTEGER MONTH",
	},
	{
		"twelve hours ago",
		tt(TokenInteger, 12, TokenUnit, PeriodHour, TokenModifier, ModifierPast),
		"INTEGER UNIT MODIFIER",
	},

	// --- Year alone ---
	{
		"2026",
		tt(TokenYear, 2026),
		"YEAR",
	},

	// --- Preprocessing: noon / midnight → time tokens ---
	{"noon", tt(TokenTime, "12:00"), "TIME"},
	{"midnight", tt(TokenTime, "0:00"), "TIME"},
	{"at noon", tt(TokenPrep, nil, TokenTime, "12:00"), "PREP TIME"},
	{"next Monday at noon", tt(TokenDirection, DirectionFuture, TokenWeekday, WeekdayMonday, TokenPrep, nil, TokenTime, "12:00"), "DIRECTION WEEKDAY PREP TIME"},

	// --- Preprocessing: time with dot notation (only when followed by AM/PM) ---
	{"7.15pm", tt(TokenTime, "7:15", TokenAMPM, AMPMPm), "TIME AMPM"},
	{"7.15 PM", tt(TokenTime, "7:15", TokenAMPM, AMPMPm), "TIME AMPM"},
	{"7.15", tt(TokenDecimal, 7.15), "DECIMAL"}, // bare dot — decimal number (no AM/PM)

	// --- Decimal numbers ---
	{"3.5", tt(TokenDecimal, 3.5), "DECIMAL"},
	{"1.5", tt(TokenDecimal, 1.5), "DECIMAL"},
	{"0.5", tt(TokenDecimal, 0.5), "DECIMAL"},
	{"1.25", tt(TokenDecimal, 1.25), "DECIMAL"},

	// --- Preprocessing: ISO 8601 T separator ---
	{"2026-12-04T09:30:00", tt(TokenYear, 2026, TokenInteger, 12, TokenInteger, 4, TokenTime, "09:30:00"), "YEAR INTEGER INTEGER TIME"},

	// --- HH:MM:SS time token ---
	{"at 09:30:45", tt(TokenPrep, nil, TokenTime, "09:30:45"), "PREP TIME"},

	// --- "an" as FILLER ---
	{"in an hour", tt(TokenPrep, nil, TokenFiller, nil, TokenUnit, PeriodHour), "PREP UNIT"},

	// --- MM/DD/YYYY numeric compound ---
	{"12/04/2026", tt(TokenInteger, 12, TokenInteger, 4, TokenYear, 2026), "INTEGER INTEGER YEAR"},

	// --- Timezone: abbreviations (globally recognised, not per-language) ---
	{"9:30 EST", tt(TokenTime, "9:30", TokenTimezone, "est"), "TIME TIMEZONE"},
	{"9:30 BST", tt(TokenTime, "9:30", TokenTimezone, "bst"), "TIME TIMEZONE"},
	{"9:30 NZST", tt(TokenTime, "9:30", TokenTimezone, "nzst"), "TIME TIMEZONE"},
	{"9:30 AM UTC", tt(TokenTime, "9:30", TokenAMPM, AMPMAm, TokenTimezone, "utc"), "TIME AMPM TIMEZONE"},
	{"3pm PST", tt(TokenInteger, 3, TokenAMPM, AMPMPm, TokenTimezone, "pst"), "INTEGER AMPM TIMEZONE"},

	// --- Timezone: standalone numeric offsets ---
	{"+05:30", tt(TokenTimezone, "+05:30"), "TIMEZONE"},
	{"-07:00", tt(TokenTimezone, "-07:00"), "TIMEZONE"},
	{"+0530", tt(TokenTimezone, "+0530"), "TIMEZONE"},
	{"-07", tt(TokenTimezone, "-07"), "TIMEZONE"},

	// --- Timezone: glued numeric offset on time ---
	{"09:30:00-07:00", tt(TokenTime, "09:30:00", TokenTimezone, "-07:00"), "TIME TIMEZONE"},
	{"9:30+05:30", tt(TokenTime, "9:30", TokenTimezone, "+05:30"), "TIME TIMEZONE"},

	// --- Timezone: ISO 8601 with Z suffix ---
	{
		"2026-12-04T09:30:00Z",
		tt(TokenYear, 2026, TokenInteger, 12, TokenInteger, 4, TokenTime, "09:30:00", TokenTimezone, "z"),
		"YEAR INTEGER INTEGER TIME TIMEZONE",
	},

	// --- Overflow guards: integers longer than 10 digits must not panic ---
	// classifyBareInteger rejects len>10 → TokenUnknown; no Atoi is called.
	{"100000000000000000000", tt(TokenUnknown, "100000000000000000000"), "UNKNOWN"},
	// splitCompoundDate rejects any part with len>10 → TokenDateFragment.
	{"0.1000000000000000000", tt(TokenDateFragment, "0.1000000000000000000"), "DATE_FRAGMENT"},

	// --- Non-ASCII digit guard: must not panic ---
	// unicode.IsDigit accepts Arabic-Indic, Devanagari, etc.; allDigits must not.
	// These must produce TokenUnknown without reaching mustAtoi/strconv.Atoi.
	{"0٣", tt(TokenUnknown, "0٣"), "UNKNOWN"}, // mixed ASCII + Arabic-Indic
	{"١٢", tt(TokenUnknown, "١٢"), "UNKNOWN"}, // all Arabic-Indic digits
}

func TestTokenize(t *testing.T) {
	for _, tc := range tokenizerCases {
		t.Run(tc.input, func(t *testing.T) {
			got := LangEn.Tokenize(tc.input)
			if len(got) != len(tc.tokens) {
				t.Fatalf("Tokenize(%q)\n  got  %v\n  want %v", tc.input, got, tc.tokens)
			}
			for i, tok := range got {
				want := tc.tokens[i]
				if tok != want {
					t.Errorf("Tokenize(%q) token[%d]\n  got  %+v\n  want %+v", tc.input, i, tok, want)
				}
			}
		})
	}
}

func TestSignature(t *testing.T) {
	for _, tc := range tokenizerCases {
		if tc.sig == "" {
			continue
		}
		t.Run(tc.input, func(t *testing.T) {
			tokens := LangEn.Tokenize(tc.input)
			got := Signature(tokens)
			if got != tc.sig {
				t.Errorf("Signature(%q)\n  got  %q\n  want %q", tc.input, got, tc.sig)
			}
		})
	}
}
