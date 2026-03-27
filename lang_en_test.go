package nowandlater

import (
	"testing"
	"time"
)

// enNow is the fixed reference time for LangEn resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var enNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var englishCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors (DeltaSeconds, no time) ---
	// Returns now + delta (exact offset, no truncation).
	{"now", u(2026, 3, 22, 10, 0, 0)},
	{"today", u(2026, 3, 22, 10, 0, 0)},
	{"tomorrow", u(2026, 3, 23, 10, 0, 0)},
	{"yesterday", u(2026, 3, 21, 10, 0, 0)},

	// --- Pure relative delta ---
	{"3 days ago", u(2026, 3, 19, 10, 0, 0)},
	{"in 2 hours", u(2026, 3, 22, 12, 0, 0)},
	{"in 3 days", u(2026, 3, 25, 10, 0, 0)},
	{"a week ago", u(2026, 3, 15, 10, 0, 0)},
	{"2 weeks from now", u(2026, 4, 5, 10, 0, 0)},

	// --- Anchor + time-of-day ---
	{"today at 9:30", u(2026, 3, 22, 9, 30, 0)},
	{"today at 9:30 PM", u(2026, 3, 22, 21, 30, 0)},
	{"tomorrow at 9 PM", u(2026, 3, 23, 21, 0, 0)},
	{"yesterday at noon", u(2026, 3, 21, 12, 0, 0)},
	{"today at 3", u(2026, 3, 22, 3, 0, 0)},

	// --- Weekday only ---
	// today is Sunday (our weekday 6)
	{"next Monday", u(2026, 3, 23, 0, 0, 0)},    // 1 day ahead
	{"last Friday", u(2026, 3, 20, 0, 0, 0)},    // 2 days back
	{"this Sunday", u(2026, 3, 22, 0, 0, 0)},    // today matches → start of today
	{"this Wednesday", u(2026, 3, 25, 0, 0, 0)}, // 3 days ahead (vs 4 back)
	{"this Saturday", u(2026, 3, 21, 0, 0, 0)},  // 1 day back (vs 6 ahead)

	// --- Weekday + time ---
	{"next Monday at 9:30", u(2026, 3, 23, 9, 30, 0)},
	{"next Monday at 9 PM", u(2026, 3, 23, 21, 0, 0)},
	{"last Friday at 3 PM", u(2026, 3, 20, 15, 0, 0)},

	// --- Direction + unit ("next week", "last month", "this year") ---
	// today is Sunday 2026-03-22; current week's Monday is 2026-03-16
	{"next week", u(2026, 3, 23, 0, 0, 0)}, // Monday of next week
	{"last week", u(2026, 3, 9, 0, 0, 0)},  // Monday of last week
	{"this week", u(2026, 3, 16, 0, 0, 0)}, // Monday of current week
	{"next month", u(2026, 4, 1, 0, 0, 0)},
	{"last month", u(2026, 2, 1, 0, 0, 0)},
	{"this month", u(2026, 3, 1, 0, 0, 0)},
	{"next year", u(2027, 1, 1, 0, 0, 0)},
	{"last year", u(2025, 1, 1, 0, 0, 0)},
	{"this year", u(2026, 1, 1, 0, 0, 0)},
	{"next day", u(2026, 3, 23, 0, 0, 0)},
	{"last day", u(2026, 3, 21, 0, 0, 0)},

	// --- Full absolute date ---
	{"2026-12-04", u(2026, 12, 4, 0, 0, 0)},
	{"2026-12-04 09:30", u(2026, 12, 4, 9, 30, 0)},
	{"2026-12-04 9:30 PM", u(2026, 12, 4, 21, 30, 0)},
	{"Dec 3rd 2026", u(2026, 12, 3, 0, 0, 0)},
	{"Dec 3rd 2026 at 9 PM", u(2026, 12, 3, 21, 0, 0)},
	{"12/04/2026", u(2026, 12, 4, 0, 0, 0)},

	// --- Year + month ---
	{"December 2026", u(2026, 12, 1, 0, 0, 0)},
	{"2026 December", u(2026, 12, 1, 0, 0, 0)},

	// --- Month + day (year inferred) ---
	// today is 2026-03-22; March 5 has passed → next year
	{"March 5", u(2027, 3, 5, 0, 0, 0)},
	// April 10 is in the future this year
	{"April 10", u(2026, 4, 10, 0, 0, 0)},
	// March 22 at exactly enNow is not after now → next year
	{"March 22", u(2027, 3, 22, 0, 0, 0)},
	// April 10 with time
	{"April 10 at 3 PM", u(2026, 4, 10, 15, 0, 0)},

	// --- "second" as ordinal day-2 (unit/ordinal conflict) ---
	{"march second", u(2027, 3, 2, 0, 0, 0)},
	{"march second, 2010", u(2010, 3, 2, 0, 0, 0)},
	{"second of march", u(2027, 3, 2, 0, 0, 0)},
	{"second of march, 2010", u(2010, 3, 2, 0, 0, 0)},
	{"march second at 3pm", u(2027, 3, 2, 15, 0, 0)},

	// --- Number words (cardinal + ordinal) ---
	{"the first of March", u(2027, 3, 1, 0, 0, 0)}, // March 1 has passed → next year
	{"the twenty-first of April", u(2026, 4, 21, 0, 0, 0)},
	{"in three days", u(2026, 3, 25, 10, 0, 0)},
	{"twelve hours ago", u(2026, 3, 21, 22, 0, 0)}, // 10:00 - 12h → previous day 22:00
	{"thirty-first of January 2027", u(2027, 1, 31, 0, 0, 0)},
	{"twenty first of May", u(2026, 5, 21, 0, 0, 0)},

	// --- Year only ---
	{"2026", u(2026, 1, 1, 0, 0, 0)},
	{"2027", u(2027, 1, 1, 0, 0, 0)},

	// --- CLDR unit abbreviations ---
	{"in 2 wk", u(2026, 4, 5, 10, 0, 0)},    // wk = weeks
	{"3 mo ago", u(2025, 12, 22, 10, 0, 0)}, // mo = months (3×30 days)

	// --- 2-letter weekday abbreviations ---
	{"tu", u(2026, 3, 24, 0, 0, 0)}, // tu = Tuesday (nearest from Sunday)
	{"su", u(2026, 3, 22, 0, 0, 0)}, // su = Sunday (nearest = today)

	// --- Multi-word anchors ---
	{"day before yesterday", u(2026, 3, 20, 10, 0, 0)},
	{"day after tomorrow", u(2026, 3, 24, 10, 0, 0)},

	// --- Filler words ---
	{"about 3 days ago", u(2026, 3, 19, 10, 0, 0)}, // "about" as filler
	{"just now", u(2026, 3, 22, 10, 0, 0)},         // "just" as filler

	// --- Time-first formats (TIME AMPM MONTH INTEGER YEAR) ---
	{"8:25 a.m. Dec. 12, 2014", u(2014, 12, 12, 8, 25, 0)},
	{"2:21 p.m., December 11, 2014", u(2014, 12, 11, 14, 21, 0)},
	{"10:06am Dec 11, 2014", u(2014, 12, 11, 10, 6, 0)},

	// --- RFC 2822 with AM/PM (WEEKDAY INTEGER MONTH YEAR TIME AMPM) ---
	{"Wednesday, 22nd June, 2016, 12:16 pm", u(2016, 6, 22, 12, 16, 0)},

	// --- INTEGER MONTH YEAR TIME (day-monthname-year with time) ---
	{"21 January 2012 13:11:23", u(2012, 1, 21, 13, 11, 23)},
	{"4 July 2026 08:00:00", u(2026, 7, 4, 8, 0, 0)},
	{"21 January 2012 1:11:23 pm", u(2012, 1, 21, 13, 11, 23)},

	// --- INTEGER INTEGER YEAR TIME (numeric date + time, no preposition) ---
	{"29/02/2020 13:12", u(2020, 2, 29, 13, 12, 0)},
	{"12/25/2024 9:30", u(2024, 12, 25, 9, 30, 0)},
	{"29/02/2020 1:12 pm", u(2020, 2, 29, 13, 12, 0)},

	// --- INTEGER AMPM INTEGER INTEGER YEAR (time-first, numeric date, 4-digit year) ---
	{"1 a.m 20.07.2021", u(2021, 7, 20, 1, 0, 0)},
	{"3 pm 12/25/2024", u(2024, 12, 25, 15, 0, 0)},

	// --- INTEGER AMPM DATE_FRAGMENT (time-first, numeric date, 2-digit year) ---
	{"1 a.m 20.07.21", u(2021, 7, 20, 1, 0, 0)},
	{"3 pm 12/25/25", u(2025, 12, 25, 15, 0, 0)},

	// --- DATE_FRAGMENT standalone (2-digit year) ---
	{"29/02/20", u(2020, 2, 29, 0, 0, 0)},

	// --- TIME AMPM DATE_FRAGMENT (time-first, 2-digit year) ---
	{"1:30am 29/02/20", u(2020, 2, 29, 1, 30, 0)},
	{"1:30am at 29/02/20", u(2020, 2, 29, 1, 30, 0)},

	// --- DATE_FRAGMENT TIME / TIME AMPM / PREP TIME AMPM (2-digit year) ---
	{"1/1/16 9:02:43", u(2016, 1, 1, 9, 2, 43)},
	{"3/15/99 14:30:00", u(1999, 3, 15, 14, 30, 0)},
	{"12/25/25 0:00:00", u(2025, 12, 25, 0, 0, 0)},
	{"29/02/20 1:12 pm", u(2020, 2, 29, 13, 12, 0)},
	{"29/02/20 at 1:12 pm", u(2020, 2, 29, 13, 12, 0)},
	{"29/02/20 at 13:12", u(2020, 2, 29, 13, 12, 0)},

	// --- Time only (applied to today) ---
	{"at 9:30", u(2026, 3, 22, 9, 30, 0)},
	{"at 9:30 AM", u(2026, 3, 22, 9, 30, 0)},
	{"at 9:30 PM", u(2026, 3, 22, 21, 30, 0)},
	{"3pm", u(2026, 3, 22, 15, 0, 0)},
	{"noon", u(2026, 3, 22, 12, 0, 0)},
	{"midnight", u(2026, 3, 22, 0, 0, 0)},

	// --- applyAMPM edge cases: 12 AM = midnight, 12 PM = noon ---
	{"12 am", u(2026, 3, 22, 0, 0, 0)},
	{"12 pm", u(2026, 3, 22, 12, 0, 0)},

	// --- INTEGER INTEGER YEAR + PREP time variants (dateOrderHandler coverage) ---
	{"12/25/2020 at 13:12", u(2020, 12, 25, 13, 12, 0)},
	{"12/25/2020 at 1:12 pm", u(2020, 12, 25, 13, 12, 0)},
	{"12/25/2020 at 1 pm", u(2020, 12, 25, 13, 0, 0)},

	// --- DATE_FRAGMENT + PREP INTEGER AMPM (dateOrderHandler coverage) ---
	{"12/25/20 at 1 pm", u(2020, 12, 25, 13, 0, 0)},
}

func TestLangEn(t *testing.T) {
	for _, tc := range englishCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangEn.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, enNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}

// TestEnglishErrors verifies that out-of-range 12-hour clock values return errors
// across all handler paths that call validateAndApplyAMPM.
var englishErrorCases = []string{
	"0 pm",            // INTEGER AMPM: hour 0 out of range
	"13 am",           // INTEGER AMPM: hour 13 out of range
	"at 0 pm",         // PREP INTEGER AMPM: hour 0 out of range
	"at 13 am",        // PREP INTEGER AMPM: hour 13 out of range
	"0 am 20.07.21",   // INTEGER AMPM DATE_FRAGMENT: hour 0 out of range
	"0 am 20.07.2021", // INTEGER AMPM INTEGER INTEGER YEAR: hour 0 out of range
}

func TestEnglishErrors(t *testing.T) {
	for _, input := range englishErrorCases {
		t.Run(input, func(t *testing.T) {
			_, err := LangEn.Parse(input)
			if err == nil {
				t.Errorf("Parse(%q) expected error, got nil", input)
			}
		})
	}
}
