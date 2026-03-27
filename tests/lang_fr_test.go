package tests

import (
	"errors"
	. "github.com/client9/nowandlater/internal/engine"
	. "github.com/client9/nowandlater/languages"
	"testing"
	"time"
)

// frNow is the fixed reference time for LangFr resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var frNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var frenchCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"maintenant", u(2026, 3, 22, 10, 0, 0)},
	{"aujourd'hui", u(2026, 3, 22, 10, 0, 0)},
	{"demain", u(2026, 3, 23, 10, 0, 0)},
	{"hier", u(2026, 3, 21, 10, 0, 0)},
	{"avant-hier", u(2026, 3, 20, 10, 0, 0)},
	{"après-demain", u(2026, 3, 24, 10, 0, 0)},
	{"apres-demain", u(2026, 3, 24, 10, 0, 0)}, // unaccented variant

	// --- Time words ---
	{"midi", u(2026, 3, 22, 12, 0, 0)},
	{"minuit", u(2026, 3, 22, 0, 0, 0)},

	// --- Absolute date: INTEGER MONTH YEAR (DMY order) ---
	{"24 mars 2026", u(2026, 3, 24, 0, 0, 0)},
	{"1er mars 2026", u(2026, 3, 1, 0, 0, 0)}, // ordinal suffix stripped: "1er"→"1"
	{"mars 2026", u(2026, 3, 1, 0, 0, 0)},     // MONTH YEAR

	// --- "second"/"seconde" as ordinal day-2 ---
	{"le second mars", u(2027, 3, 2, 0, 0, 0)},       // masculine form
	{"la seconde mars", u(2027, 3, 2, 0, 0, 0)},      // feminine form
	{"second de mars", u(2027, 3, 2, 0, 0, 0)},       // UNIT MONTH
	{"seconde de mars 2026", u(2026, 3, 2, 0, 0, 0)}, // UNIT MONTH YEAR

	// --- Relative: future (PREP INTEGER UNIT) ---
	{"dans 3 jours", u(2026, 3, 25, 10, 0, 0)},
	{"dans une semaine", u(2026, 3, 29, 10, 0, 0)},
	{"dans deux heures", u(2026, 3, 22, 12, 0, 0)},

	// --- Relative: past — 3-word modifier "il y a" ---
	{"il y a 3 jours", u(2026, 3, 19, 10, 0, 0)},
	{"il y a 2 semaines", u(2026, 3, 8, 10, 0, 0)},

	// --- Direction + unit: DIRECTION UNIT and UNIT DIRECTION ---
	{"cette semaine", u(2026, 3, 16, 0, 0, 0)},        // DIRECTION UNIT (nearest)
	{"la semaine prochaine", u(2026, 3, 23, 0, 0, 0)}, // UNIT DIRECTION (future)
	{"le mois dernier", u(2026, 2, 1, 0, 0, 0)},       // UNIT DIRECTION (past)
	{"l'année prochaine", u(2027, 1, 1, 0, 0, 0)},     // "l'année" elided unit
	{"l'an prochain", u(2027, 1, 1, 0, 0, 0)},         // "l'an" via single-char unit "an"... wait

	// --- Direction + weekday: WEEKDAY DIRECTION ---
	{"lundi prochain", u(2026, 3, 23, 0, 0, 0)},   // next Monday
	{"vendredi dernier", u(2026, 3, 20, 0, 0, 0)}, // last Friday

	// --- Anchor + time ---
	{"demain à midi", u(2026, 3, 23, 12, 0, 0)},
	{"demain à 9:30", u(2026, 3, 23, 9, 30, 0)},
	{"demain à 9:30 du matin", u(2026, 3, 23, 9, 30, 0)},

	// --- Number word (cardinal) ---
	{"dans trois jours", u(2026, 3, 25, 10, 0, 0)},

	// --- 3-word number phrase: "vingt et un" (primary test of span-3 phrase match) ---
	{"il y a vingt et un jours", u(2026, 3, 1, 10, 0, 0)}, // 21 days ago = March 1

	// --- CLDR abbreviations ---
	{"1 janv 2026", u(2026, 1, 1, 0, 0, 0)},   // janv = janvier
	{"15 févr 2026", u(2026, 2, 15, 0, 0, 0)}, // févr = février
	{"il y a 2 sem", u(2026, 3, 8, 10, 0, 0)}, // sem = semaine
	{"dans 3 sem", u(2026, 4, 12, 10, 0, 0)},  // sem = semaines

	// --- Supplementary data ---
	{"15 jul 2026", u(2026, 7, 15, 0, 0, 0)},             // jul = juillet
	{"10 aoû 2026", u(2026, 8, 10, 0, 0, 0)},             // aoû = août
	{"il ya 3 jours", u(2026, 3, 19, 10, 0, 0)},          // no-space "il ya" variant
	{"il y a environ 3 jours", u(2026, 3, 19, 10, 0, 0)}, // "environ" as filler
	{"après 3 jours", u(2026, 3, 25, 10, 0, 0)},          // "après" as future prep
	{"lu prochain", u(2026, 3, 23, 0, 0, 0)},             // "lu" = lundi (Monday)
}

// frenchAmbiguousCases are inputs that are recognisably date-like but cannot be
// resolved because "mar" abbreviates both mardi (Tuesday) and mars (March).
var frenchAmbiguousCases = []string{
	"mar 5",
	"5 mar",
	"mar 5 2026",
	"5 mar 2026",
}

func TestLangFrAmbiguous(t *testing.T) {
	for _, input := range frenchAmbiguousCases {
		t.Run(input, func(t *testing.T) {
			_, err := LangFr.Parse(input)
			if !errors.Is(err, ErrAmbiguous) {
				t.Errorf("Parse(%q) error = %v, want ErrAmbiguous", input, err)
			}
		})
	}
}

func TestLangFr(t *testing.T) {
	for _, tc := range frenchCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangFr.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, frNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
