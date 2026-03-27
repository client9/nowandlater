package tests

import (
	. "github.com/client9/nowandlater/internal/engine"
	. "github.com/client9/nowandlater/languages"
	"testing"
	"time"
)

// jaNow is the fixed reference time for LangJa resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var jaNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var japaneseCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"今", u(2026, 3, 22, 10, 0, 0)},
	{"今日", u(2026, 3, 22, 10, 0, 0)},
	{"明日", u(2026, 3, 23, 10, 0, 0)},
	{"昨日", u(2026, 3, 21, 10, 0, 0)},
	{"一昨日", u(2026, 3, 20, 10, 0, 0)},
	{"明後日", u(2026, 3, 24, 10, 0, 0)},

	// --- Absolute date ---
	{"2026年3月24日", u(2026, 3, 24, 0, 0, 0)},
	{"2026年3月", u(2026, 3, 1, 0, 0, 0)},

	// --- Full-width digit normalization ---
	{"２０２６年３月２４日", u(2026, 3, 24, 0, 0, 0)},

	// --- Time of day (AMPM TIME) ---
	{"午後3時30分", u(2026, 3, 22, 15, 30, 0)},
	{"午前10時", u(2026, 3, 22, 10, 0, 0)},
	{"午後3時30分15秒", u(2026, 3, 22, 15, 30, 15)},

	// --- Relative deltas: future ---
	{"3日後", u(2026, 3, 25, 10, 0, 0)},
	{"1時間後", u(2026, 3, 22, 11, 0, 0)},
	{"2週間後", u(2026, 4, 5, 10, 0, 0)},

	// --- Relative deltas: past ---
	{"2週間前", u(2026, 3, 8, 10, 0, 0)},
	{"30分前", u(2026, 3, 22, 9, 30, 0)},

	// --- Direction + unit (来週, 先月, etc.) ---
	{"来週", u(2026, 3, 23, 0, 0, 0)},
	{"先月", u(2026, 2, 1, 0, 0, 0)},
	{"来年", u(2027, 1, 1, 0, 0, 0)},
	{"今週", u(2026, 3, 16, 0, 0, 0)},

	// --- Direction + unit + weekday ---
	{"来週の月曜日", u(2026, 3, 23, 0, 0, 0)},
	{"先週の月曜日", u(2026, 3, 16, 0, 0, 0)},

	// --- Anchor + AMPM time ---
	{"明日の午後3時", u(2026, 3, 23, 15, 0, 0)},
	{"今日の午前10時30分", u(2026, 3, 22, 10, 30, 0)},

	// --- 正午 (noon) and 現在 (now) ---
	{"正午", u(2026, 3, 22, 12, 0, 0)},
	{"現在", u(2026, 3, 22, 10, 0, 0)},

	// --- Kanji month names ---
	{"2026年一月", u(2026, 1, 1, 0, 0, 0)},
	{"2026年三月", u(2026, 3, 1, 0, 0, 0)},
	{"2026年十月", u(2026, 10, 1, 0, 0, 0)},
	{"2026年十一月", u(2026, 11, 1, 0, 0, 0)},
	{"2026年十二月", u(2026, 12, 1, 0, 0, 0)},

	// --- Single-kanji weekday abbreviations ---
	{"来週の月", u(2026, 3, 23, 0, 0, 0)}, // 月 = Monday
	{"来週の金", u(2026, 3, 27, 0, 0, 0)}, // 金 = Friday

	// --- 先々週 / 再来週 ---
	{"先々週", u(2026, 3, 8, 10, 0, 0)}, // 2 weeks ago
	{"再来週", u(2026, 4, 5, 10, 0, 0)}, // 2 weeks from now

	// --- 日間 / 分間 / 秒間 duration variants ---
	{"3日間後", u(2026, 3, 25, 10, 0, 0)},
	{"30分間前", u(2026, 3, 22, 9, 30, 0)},

	// --- カ月 (katakana variant) ---
	// Month deltas use the 30-day approximation from periodToSeconds (3×30d = 90d).
	{"3カ月後", u(2026, 6, 20, 10, 0, 0)},

	// --- Combined absolute date + AMPM time ---
	{"2026年3月24日の午後3時", u(2026, 3, 24, 15, 0, 0)},
	{"2026年3月24日の午後3時30分", u(2026, 3, 24, 15, 30, 0)},
	{"2026年3月24日の午前10時", u(2026, 3, 24, 10, 0, 0)},

	// --- Imperial era years ---
	{"令和7年3月", u(2025, 3, 1, 0, 0, 0)},     // 2019 + 7 - 1 = 2025
	{"令和元年", u(2019, 1, 1, 0, 0, 0)},       // era year 1 = base year
	{"平成31年4月", u(2019, 4, 1, 0, 0, 0)},    // 1989 + 31 - 1 = 2019
	{"昭和64年1月", u(1989, 1, 1, 0, 0, 0)},    // 1926 + 64 - 1 = 1989
	{"令和7年3月24日", u(2025, 3, 24, 0, 0, 0)}, // era + full date
}

func TestLangJa(t *testing.T) {
	for _, tc := range japaneseCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangJa.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, jaNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
