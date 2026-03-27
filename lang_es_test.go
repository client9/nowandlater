package nowandlater

import (
	"errors"
	"testing"
	"time"
)

// spNow is the fixed reference time for LangEs resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday) for easy comparison.
var spNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var spanishCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"hoy", u(2026, 3, 22, 10, 0, 0)},
	{"ahora", u(2026, 3, 22, 10, 0, 0)},
	{"mañana", u(2026, 3, 23, 10, 0, 0)},
	{"ayer", u(2026, 3, 21, 10, 0, 0)},
	{"anteayer", u(2026, 3, 20, 10, 0, 0)},
	{"antier", u(2026, 3, 20, 10, 0, 0)},

	// --- Phrases: pasado mañana ---
	{"pasado mañana", u(2026, 3, 24, 10, 0, 0)},

	// --- Fractional unit durations ---
	{"3.5 días", u(2026, 3, 25, 22, 0, 0)},
	{"en 1.5 horas", u(2026, 3, 22, 11, 30, 0)},
	{"hace 2.5 horas", u(2026, 3, 22, 7, 30, 0)},

	// --- Relative delta: bare INTEGER UNIT (implied future, no prep/modifier) ---
	{"3 días", u(2026, 3, 25, 10, 0, 0)},
	{"2 horas", u(2026, 3, 22, 12, 0, 0)},
	{"1 semana", u(2026, 3, 29, 10, 0, 0)},
	{"1 quincena", u(2026, 4, 5, 10, 0, 0)},   // +14 days
	{"2 quincenas", u(2026, 4, 19, 10, 0, 0)}, // +28 days

	// --- Relative delta: modifier-first word order ---
	{"hace 3 días", u(2026, 3, 19, 10, 0, 0)},
	{"hace una semana", u(2026, 3, 15, 10, 0, 0)},
	{"hace 2 horas", u(2026, 3, 22, 8, 0, 0)},
	{"hace 1 quincena", u(2026, 3, 8, 10, 0, 0)}, // −14 days

	// --- Compound durations ---
	{"en 1 hora y 10 minutos", u(2026, 3, 22, 11, 10, 0)},
	{"1 hora y 10 minutos", u(2026, 3, 22, 11, 10, 0)},
	{"1 hora y 10 minutos atrás", u(2026, 3, 22, 8, 50, 0)},
	{"hace 1 hora y 10 minutos", u(2026, 3, 22, 8, 50, 0)},
	{"dentro de 2 días y 3 horas", u(2026, 3, 24, 13, 0, 0)},

	// --- Relative delta: prep-first word order ---
	{"en 3 días", u(2026, 3, 25, 10, 0, 0)},
	{"en 2 horas", u(2026, 3, 22, 12, 0, 0)},
	{"dentro de 5 días", u(2026, 3, 27, 10, 0, 0)},

	// --- Direction + weekday ---
	{"el próximo lunes", u(2026, 3, 23, 0, 0, 0)},
	{"el lunes pasado", u(2026, 3, 16, 0, 0, 0)},

	// --- Weekday + direction (LangEs word order) ---
	{"lunes próximo", u(2026, 3, 23, 0, 0, 0)},
	{"lunes pasado", u(2026, 3, 16, 0, 0, 0)},

	// --- Direction + unit ---
	{"la próxima semana", u(2026, 3, 23, 0, 0, 0)},
	{"el mes pasado", u(2026, 2, 1, 0, 0, 0)},
	{"este año", u(2026, 1, 1, 0, 0, 0)},

	// --- Absolute date: day of month ---
	{"el 5 de marzo", u(2027, 3, 5, 0, 0, 0)}, // March 5 has passed → next year
	{"el 10 de abril", u(2026, 4, 10, 0, 0, 0)},
	{"el primero de mayo", u(2026, 5, 1, 0, 0, 0)},
	{"el segundo de marzo", u(2027, 3, 2, 0, 0, 0)}, // "segundo" as ordinal
	{"el segunda de marzo", u(2027, 3, 2, 0, 0, 0)}, // feminine form

	// --- Absolute date with year ---
	{"el 5 de marzo de 2027", u(2027, 3, 5, 0, 0, 0)},
	{"15 de agosto de 2026", u(2026, 8, 15, 0, 0, 0)},

	// --- Month + year ---
	{"enero de 2027", u(2027, 1, 1, 0, 0, 0)},

	// --- Time of day ---
	{"a las 9:30", u(2026, 3, 22, 9, 30, 0)},
	{"a las 3 de la tarde", u(2026, 3, 22, 15, 0, 0)},
	{"a las 9 de la mañana", u(2026, 3, 22, 9, 0, 0)},
	{"a las 10 de la noche", u(2026, 3, 22, 22, 0, 0)},
	{"a medianoche", u(2026, 3, 22, 0, 0, 0)},
	{"a mediodía", u(2026, 3, 22, 12, 0, 0)},

	// --- Anchor + time ---
	{"mañana a las 9:30", u(2026, 3, 23, 9, 30, 0)},
	{"hoy a las 3 de la tarde", u(2026, 3, 22, 15, 0, 0)},

	// --- Number words ---
	{"hace doce horas", u(2026, 3, 21, 22, 0, 0)},
	{"en veinte días", u(2026, 4, 11, 10, 0, 0)},
	{"el veintiuno de abril", u(2026, 4, 21, 0, 0, 0)},

	// --- CLDR abbreviations ---
	{"15 may 2026", u(2026, 5, 15, 0, 0, 0)},        // may = mayo
	{"10 sept 2026", u(2026, 9, 10, 0, 0, 0)},       // sept = septiembre (base es)
	{"hace 2 sem", u(2026, 3, 8, 10, 0, 0)},         // sem = semanas
	{"dentro de 30 min", u(2026, 3, 22, 10, 30, 0)}, // min = minutos
	{"hace 45 seg", u(2026, 3, 22, 9, 59, 15)},      // seg = segundos (es-AR)

	// --- Supplementary data ---
	{"hace cerca de 3 días", u(2026, 3, 19, 10, 0, 0)}, // "cerca" as filler
	{"lu próximo", u(2026, 3, 23, 0, 0, 0)},            // "lu" = lunes
	{"vi pasado", u(2026, 3, 20, 0, 0, 0)},             // "vi" = viernes (last Friday)

	// --- "después" as ModifierFuture (INTEGER UNIT MODIFIER) ---
	{"2 horas después", u(2026, 3, 22, 12, 0, 0)}, // 2 hours later
	{"3 días después", u(2026, 3, 25, 10, 0, 0)},  // 3 days later
	{"2 horas despues", u(2026, 3, 22, 12, 0, 0)}, // unaccented variant
}

// spanishAmbiguousCases are inputs that are recognisably date-like but cannot be
// resolved because "mar" abbreviates both martes (Tuesday) and marzo (March).
var spanishAmbiguousCases = []string{
	"mar 5",
	"5 de mar",
	"mar 5 2027",
	"5 de mar 2027",
}

func TestLangEsAmbiguous(t *testing.T) {
	for _, input := range spanishAmbiguousCases {
		t.Run(input, func(t *testing.T) {
			_, err := LangEs.Parse(input)
			if !errors.Is(err, ErrAmbiguous) {
				t.Errorf("Parse(%q) error = %v, want ErrAmbiguous", input, err)
			}
		})
	}
}

func TestLangEs(t *testing.T) {
	for _, tc := range spanishCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangEs.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, spNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
