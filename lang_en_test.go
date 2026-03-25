package nowandlater

import (
	"testing"
	"time"
)

// enNow is the fixed reference time for English resolver tests.
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

	// --- Time only (applied to today) ---
	{"at 9:30", u(2026, 3, 22, 9, 30, 0)},
	{"at 9:30 AM", u(2026, 3, 22, 9, 30, 0)},
	{"at 9:30 PM", u(2026, 3, 22, 21, 30, 0)},
	{"3pm", u(2026, 3, 22, 15, 0, 0)},
	{"noon", u(2026, 3, 22, 12, 0, 0)},
	{"midnight", u(2026, 3, 22, 0, 0, 0)},
}

func TestEnglish(t *testing.T) {
	for _, tc := range englishCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := English.Parse(tc.input)
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
