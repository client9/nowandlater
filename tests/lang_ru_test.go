package tests

import (
	. "github.com/client9/nowandlater/internal/engine"
	. "github.com/client9/nowandlater/languages"
	"testing"
	"time"
)

// ruNow is the fixed reference time for LangRu resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var ruNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var russianCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"сейчас", u(2026, 3, 22, 10, 0, 0)},
	{"сегодня", u(2026, 3, 22, 10, 0, 0)},
	{"завтра", u(2026, 3, 23, 10, 0, 0)},
	{"вчера", u(2026, 3, 21, 10, 0, 0)},
	{"послезавтра", u(2026, 3, 24, 10, 0, 0)},
	{"позавчера", u(2026, 3, 20, 10, 0, 0)},

	// --- Time words ---
	{"полдень", u(2026, 3, 22, 12, 0, 0)},
	{"полночь", u(2026, 3, 22, 0, 0, 0)},

	// --- Absolute date: INTEGER MONTH YEAR (DMY order) ---
	{"24 марта 2026", u(2026, 3, 24, 0, 0, 0)},
	{"1 января 2027", u(2027, 1, 1, 0, 0, 0)},
	{"март 2026", u(2026, 3, 1, 0, 0, 0)}, // MONTH YEAR

	// --- Relative: future — PREP INTEGER UNIT ---
	{"через 3 дня", u(2026, 3, 25, 10, 0, 0)},
	{"через 2 недели", u(2026, 4, 5, 10, 0, 0)},

	// --- Relative: future — PREP UNIT ("через неделю" = in a week) ---
	{"через неделю", u(2026, 3, 29, 10, 0, 0)},
	{"через час", u(2026, 3, 22, 11, 0, 0)},

	// --- Relative: past — INTEGER UNIT MODIFIER ---
	{"3 дня назад", u(2026, 3, 19, 10, 0, 0)},
	{"2 недели назад", u(2026, 3, 8, 10, 0, 0)},

	// --- Direction + unit: inflected adjective forms ---
	{"следующая неделя", u(2026, 3, 23, 0, 0, 0)}, // nominative fem.
	{"следующей неделе", u(2026, 3, 23, 0, 0, 0)}, // prepositional fem.
	{"прошлая неделя", u(2026, 3, 9, 0, 0, 0)},    // last week (Mon March 9)
	{"эта неделя", u(2026, 3, 16, 0, 0, 0)},       // this week (Mon March 16)
	{"следующий месяц", u(2026, 4, 1, 0, 0, 0)},   // next month
	{"прошлый месяц", u(2026, 2, 1, 0, 0, 0)},     // last month
	{"следующий год", u(2027, 1, 1, 0, 0, 0)},     // next year

	// --- Direction + weekday: inflected adjective before weekday ---
	{"следующий понедельник", u(2026, 3, 23, 0, 0, 0)}, // next Monday (masc. acc.)
	{"прошлую пятницу", u(2026, 3, 20, 0, 0, 0)},       // last Friday (fem. acc.)

	// --- Weekday (standalone, nearest) ---
	{"понедельник", u(2026, 3, 23, 0, 0, 0)}, // nearest Monday (next, from Sunday)

	// --- Number words ---
	{"через три дня", u(2026, 3, 25, 10, 0, 0)},
	{"два дня назад", u(2026, 3, 20, 10, 0, 0)},

	// --- CLDR abbreviations ---
	{"15 февр 2026", u(2026, 2, 15, 0, 0, 0)}, // февр = февраль
	{"10 июн 2026", u(2026, 6, 10, 0, 0, 0)},  // июн = июнь
	{"5 июл 2026", u(2026, 7, 5, 0, 0, 0)},    // июл = июль
	{"3 дн назад", u(2026, 3, 19, 10, 0, 0)},  // дн = дней
	{"через 2 нед", u(2026, 4, 5, 10, 0, 0)},  // нед = недели
	{"через 1 мес", u(2026, 4, 21, 10, 0, 0)}, // мес = месяц (30 days)

	// --- Supplementary data ---
	{"послепослезавтра", u(2026, 3, 25, 10, 0, 0)},      // in 3 days
	{"спустя 2 часа", u(2026, 3, 22, 12, 0, 0)},         // "спустя" future prep
	{"около 3 дней назад", u(2026, 3, 19, 10, 0, 0)},    // "около" as filler
	{"примерно 3 дня назад", u(2026, 3, 19, 10, 0, 0)},  // "примерно" as filler
	{"во вторник", u(2026, 3, 24, 0, 0, 0)},             // "во" as filler → WEEKDAY
	{"2 суток назад", u(2026, 3, 20, 10, 0, 0)},         // "суток" = дней
	{"пнд", u(2026, 3, 23, 0, 0, 0)},                    // пнд = понедельник
	{"через двадцать минут", u(2026, 3, 22, 10, 20, 0)}, // number word 20
	{"через пятнадцать дней", u(2026, 4, 6, 10, 0, 0)},  // number word 15
}

func TestLangRu(t *testing.T) {
	for _, tc := range russianCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangRu.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, ruNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
