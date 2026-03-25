package nowandlater

import (
	"testing"
	"time"
)

// deNow is the fixed reference time for German resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var deNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var germanCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"jetzt", u(2026, 3, 22, 10, 0, 0)},
	{"heute", u(2026, 3, 22, 10, 0, 0)},
	{"morgen", u(2026, 3, 23, 10, 0, 0)},
	{"gestern", u(2026, 3, 21, 10, 0, 0)},
	{"vorgestern", u(2026, 3, 20, 10, 0, 0)},
	{"übermorgen", u(2026, 3, 24, 10, 0, 0)},
	{"ubermorgen", u(2026, 3, 24, 10, 0, 0)}, // unaccented variant

	// --- Time words ---
	{"mittags", u(2026, 3, 22, 12, 0, 0)},
	{"mitternacht", u(2026, 3, 22, 0, 0, 0)},

	// --- Absolute date: INTEGER MONTH YEAR (DMY order) ---
	{"24 März 2026", u(2026, 3, 24, 0, 0, 0)},
	{"24 märz 2026", u(2026, 3, 24, 0, 0, 0)}, // lowercase
	{"1. März 2026", u(2026, 3, 1, 0, 0, 0)},  // ordinal dot → "1" INTEGER
	{"01.03.2026", u(2026, 3, 1, 0, 0, 0)},    // compound numeric DMY
	{"März 2026", u(2026, 3, 1, 0, 0, 0)},     // MONTH YEAR

	// --- Relative: future (PREP INTEGER UNIT) ---
	{"in 3 Tagen", u(2026, 3, 25, 10, 0, 0)},
	{"in einer Woche", u(2026, 3, 29, 10, 0, 0)},
	{"in zwei Stunden", u(2026, 3, 22, 12, 0, 0)},

	// --- Relative: past — "vor" as modifier ---
	{"vor 3 Tagen", u(2026, 3, 19, 10, 0, 0)},
	{"vor 2 Wochen", u(2026, 3, 8, 10, 0, 0)},

	// --- Direction + unit: inflected adjective forms ---
	{"diese Woche", u(2026, 3, 16, 0, 0, 0)},   // DIRECTION UNIT (nearest)
	{"nächste Woche", u(2026, 3, 23, 0, 0, 0)}, // DIRECTION UNIT (future, fem.)
	{"letzte Woche", u(2026, 3, 9, 0, 0, 0)},   // last week (Mon March 9 – Sun March 15)

	// --- Direction + weekday: inflected adjective before weekday ---
	{"nächsten Montag", u(2026, 3, 23, 0, 0, 0)}, // DIRECTION WEEKDAY (future, masc. acc.)
	{"letzten Freitag", u(2026, 3, 20, 0, 0, 0)}, // DIRECTION WEEKDAY (past, masc. acc.)

	// --- Anchor + time ---
	{"morgen um 9:30", u(2026, 3, 23, 9, 30, 0)},
	{"morgen um 15:00 Uhr", u(2026, 3, 23, 15, 0, 0)}, // "Uhr" as filler

	// --- Number word ---
	{"in drei Tagen", u(2026, 3, 25, 10, 0, 0)},

	// --- Compound numeric date ---
	{"24.03.2026", u(2026, 3, 24, 0, 0, 0)},
}

func TestGerman(t *testing.T) {
	for _, tc := range germanCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := German.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, deNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
