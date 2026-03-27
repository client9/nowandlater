package nowandlater

import (
	"errors"
	"testing"
	"time"
)

// itNow is the fixed reference time for LangIt resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var itNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var italianCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"adesso", u(2026, 3, 22, 10, 0, 0)},
	{"oggi", u(2026, 3, 22, 10, 0, 0)},
	{"domani", u(2026, 3, 23, 10, 0, 0)},
	{"ieri", u(2026, 3, 21, 10, 0, 0)},
	{"dopodomani", u(2026, 3, 24, 10, 0, 0)},
	{"altroieri", u(2026, 3, 20, 10, 0, 0)},
	{"l'altroieri", u(2026, 3, 20, 10, 0, 0)},  // elided single token
	{"l'altro ieri", u(2026, 3, 20, 10, 0, 0)}, // multi-word phrase

	// --- Time words ---
	{"mezzogiorno", u(2026, 3, 22, 12, 0, 0)},
	{"mezzanotte", u(2026, 3, 22, 0, 0, 0)},

	// --- Absolute date: INTEGER MONTH YEAR (DMY order) ---
	{"24 marzo 2026", u(2026, 3, 24, 0, 0, 0)},
	{"1° marzo 2026", u(2026, 3, 1, 0, 0, 0)},    // ordinal ° stripped
	{"primo marzo 2026", u(2026, 3, 1, 0, 0, 0)}, // ordinal word
	{"marzo 2026", u(2026, 3, 1, 0, 0, 0)},       // MONTH YEAR

	// --- "secondo"/"seconda" as ordinal day-2 ---
	{"il secondo marzo", u(2027, 3, 2, 0, 0, 0)},   // masculine form
	{"la seconda marzo", u(2027, 3, 2, 0, 0, 0)},   // feminine form (was missing from Words)
	{"secondo marzo 2026", u(2026, 3, 2, 0, 0, 0)}, // UNIT MONTH YEAR
	{"seconda di marzo", u(2027, 3, 2, 0, 0, 0)},   // UNIT MONTH

	// --- Relative: future (PREP INTEGER UNIT) ---
	{"fra 3 giorni", u(2026, 3, 25, 10, 0, 0)},
	{"tra 2 settimane", u(2026, 4, 5, 10, 0, 0)},
	{"in 2 ore", u(2026, 3, 22, 12, 0, 0)},

	// --- Relative: future (PREP UNIT — elided "un'ora") ---
	{"fra un'ora", u(2026, 3, 22, 11, 0, 0)},

	// --- Relative: past — INTEGER UNIT MODIFIER ("3 giorni fa") ---
	{"3 giorni fa", u(2026, 3, 19, 10, 0, 0)},
	{"2 settimane fa", u(2026, 3, 8, 10, 0, 0)},

	// --- Direction + unit ---
	{"prossima settimana", u(2026, 3, 23, 0, 0, 0)}, // DIRECTION UNIT
	{"settimana scorsa", u(2026, 3, 9, 0, 0, 0)},    // UNIT DIRECTION
	{"questa settimana", u(2026, 3, 16, 0, 0, 0)},   // DIRECTION UNIT (nearest)
	{"prossimo mese", u(2026, 4, 1, 0, 0, 0)},
	{"mese scorso", u(2026, 2, 1, 0, 0, 0)}, // UNIT DIRECTION
	{"prossimo anno", u(2027, 1, 1, 0, 0, 0)},
	{"l'anno prossimo", u(2027, 1, 1, 0, 0, 0)}, // UNIT DIRECTION via elided "l'anno"

	// --- Direction + weekday ---
	{"prossimo lunedì", u(2026, 3, 23, 0, 0, 0)}, // DIRECTION WEEKDAY
	{"lunedì prossimo", u(2026, 3, 23, 0, 0, 0)}, // WEEKDAY DIRECTION
	{"venerdì scorso", u(2026, 3, 20, 0, 0, 0)},  // WEEKDAY DIRECTION

	// --- Anchor + time ---
	{"domani alle 9:30", u(2026, 3, 23, 9, 30, 0)},
	{"domani a mezzogiorno", u(2026, 3, 23, 12, 0, 0)},

	// --- Number words ---
	{"fra tre giorni", u(2026, 3, 25, 10, 0, 0)},
	{"due settimane fa", u(2026, 3, 8, 10, 0, 0)},

	// --- Unit abbreviations ---
	{"3 gg fa", u(2026, 3, 19, 10, 0, 0)},    // gg = giorni
	{"tra 2 sett", u(2026, 4, 5, 10, 0, 0)},  // sett = settimane
	{"30 sec fa", u(2026, 3, 22, 9, 59, 30)}, // sec = secondi

	// --- Supplementary data ---
	{"altro ieri", u(2026, 3, 20, 10, 0, 0)},        // bare form (no elided article)
	{"circa 3 giorni fa", u(2026, 3, 19, 10, 0, 0)}, // "circa" as filler

	// --- "dopo" as ModifierFuture (INTEGER UNIT MODIFIER) ---
	{"3 ore dopo", u(2026, 3, 22, 13, 0, 0)},    // 3 hours later
	{"2 giorni dopo", u(2026, 3, 24, 10, 0, 0)}, // 2 days later
}

// italianAmbiguousCases are inputs that are recognisably date-like but cannot be
// resolved because "mar" abbreviates both martedì (Tuesday) and marzo (March).
var italianAmbiguousCases = []string{
	"mar 5",
	"5 mar",
	"mar 5 2026",
	"5 mar 2026",
}

func TestLangItAmbiguous(t *testing.T) {
	for _, input := range italianAmbiguousCases {
		t.Run(input, func(t *testing.T) {
			_, err := LangIt.Parse(input)
			if !errors.Is(err, ErrAmbiguous) {
				t.Errorf("Parse(%q) error = %v, want ErrAmbiguous", input, err)
			}
		})
	}
}

func TestLangIt(t *testing.T) {
	for _, tc := range italianCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangIt.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, itNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
