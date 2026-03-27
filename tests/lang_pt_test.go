package tests

import (
	"errors"
	. "github.com/client9/nowandlater/internal/engine"
	. "github.com/client9/nowandlater/languages"
	"testing"
	"time"
)

// ptNow is the fixed reference time for LangPt resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var ptNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var portugueseCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"agora", u(2026, 3, 22, 10, 0, 0)},
	{"hoje", u(2026, 3, 22, 10, 0, 0)},
	{"amanhã", u(2026, 3, 23, 10, 0, 0)},
	{"amanha", u(2026, 3, 23, 10, 0, 0)}, // unaccented variant
	{"ontem", u(2026, 3, 21, 10, 0, 0)},
	{"anteontem", u(2026, 3, 20, 10, 0, 0)},
	{"depois de amanhã", u(2026, 3, 24, 10, 0, 0)},
	{"depois de amanha", u(2026, 3, 24, 10, 0, 0)},

	// --- Time words ---
	{"meio-dia", u(2026, 3, 22, 12, 0, 0)},
	{"meia-noite", u(2026, 3, 22, 0, 0, 0)},

	// --- Absolute date: INTEGER MONTH YEAR (DMY order, "de" as filler) ---
	{"24 de março de 2026", u(2026, 3, 24, 0, 0, 0)},
	{"1º de março de 2026", u(2026, 3, 1, 0, 0, 0)}, // ordinal º stripped
	{"março de 2026", u(2026, 3, 1, 0, 0, 0)},       // MONTH YEAR

	// --- "segundo" as ordinal day-2 (replaceSecondUnit) ---
	{"o segundo de março", u(2027, 3, 2, 0, 0, 0)},
	{"segundo de março 2026", u(2026, 3, 2, 0, 0, 0)},

	// --- Relative: future (PREP INTEGER UNIT) ---
	{"em 3 dias", u(2026, 3, 25, 10, 0, 0)},
	{"daqui a 3 dias", u(2026, 3, 25, 10, 0, 0)},
	{"dentro de 2 semanas", u(2026, 4, 5, 10, 0, 0)},

	// --- Relative: future (PREP UNIT) ---
	{"em uma semana", u(2026, 3, 29, 10, 0, 0)},

	// --- Relative: past — MODIFIER INTEGER UNIT ("há 3 dias") ---
	{"há 3 dias", u(2026, 3, 19, 10, 0, 0)},
	{"faz 2 semanas", u(2026, 3, 8, 10, 0, 0)}, // Brazilian

	// --- Relative: past — INTEGER UNIT MODIFIER ("3 dias atrás") ---
	{"3 dias atrás", u(2026, 3, 19, 10, 0, 0)},
	{"2 semanas atrás", u(2026, 3, 8, 10, 0, 0)},

	// --- Direction + unit ---
	{"próxima semana", u(2026, 3, 23, 0, 0, 0)}, // DIRECTION UNIT
	{"semana passada", u(2026, 3, 9, 0, 0, 0)},  // UNIT DIRECTION
	{"esta semana", u(2026, 3, 16, 0, 0, 0)},    // DIRECTION UNIT (nearest)
	{"próximo mês", u(2026, 4, 1, 0, 0, 0)},
	{"mês passado", u(2026, 2, 1, 0, 0, 0)}, // UNIT DIRECTION
	{"próximo ano", u(2027, 1, 1, 0, 0, 0)},

	// --- Direction + weekday ---
	{"próxima segunda", u(2026, 3, 23, 0, 0, 0)}, // DIRECTION WEEKDAY
	{"segunda passada", u(2026, 3, 16, 0, 0, 0)}, // WEEKDAY DIRECTION — last Monday

	// --- Weekday + direction (reversed order) ---
	{"sexta passada", u(2026, 3, 20, 0, 0, 0)}, // WEEKDAY DIRECTION — last Friday

	// --- Anchor + time ---
	{"amanhã às 9:30", u(2026, 3, 23, 9, 30, 0)},
	{"amanhã ao meio-dia", u(2026, 3, 23, 12, 0, 0)},

	// --- Number words ---
	{"em três dias", u(2026, 3, 25, 10, 0, 0)},
	{"há duas semanas", u(2026, 3, 8, 10, 0, 0)},

	// --- CLDR abbreviations ---
	{"15 mai 2026", u(2026, 5, 15, 0, 0, 0)}, // mai = maio
	{"há 2 sem", u(2026, 3, 8, 10, 0, 0)},    // sem = semanas
	{"em 30 min", u(2026, 3, 22, 10, 30, 0)}, // min = minutos

	// --- Supplementary data ---
	{"10 septembro 2026", u(2026, 9, 10, 0, 0, 0)},   // alternate September spelling
	{"há cerca de 3 dias", u(2026, 3, 19, 10, 0, 0)}, // "cerca" as filler

	// --- "depois" as ModifierFuture (INTEGER UNIT MODIFIER) ---
	{"2 horas depois", u(2026, 3, 22, 12, 0, 0)}, // 2 hours later
	{"3 dias depois", u(2026, 3, 25, 10, 0, 0)},  // 3 days later
}

// portugueseAmbiguousCases are inputs that are recognisably date-like but cannot
// be resolved because weekday names double as feminine ordinals in LangPt:
// "segunda" (Monday / 2nd), "quarta" (Wednesday / 4th), "quinta" (Thursday / 5th),
// "sexta" (Friday / 6th).
var portugueseAmbiguousCases = []string{
	"segunda de março",
	"quarta de março",
	"quinta de março",
	"sexta de março",
	"quarta de março 2026",
	"quinta de março 2026",
}

func TestLangPtAmbiguous(t *testing.T) {
	for _, input := range portugueseAmbiguousCases {
		t.Run(input, func(t *testing.T) {
			_, err := LangPt.Parse(input)
			if !errors.Is(err, ErrAmbiguous) {
				t.Errorf("Parse(%q) error = %v, want ErrAmbiguous", input, err)
			}
		})
	}
}

func TestLangPt(t *testing.T) {
	for _, tc := range portugueseCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangPt.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, ptNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
