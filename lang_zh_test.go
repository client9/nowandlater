package nowandlater

import (
	"testing"
	"time"
)

// zhNow is the fixed reference time for Chinese resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday).
var zhNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var chineseCases = []struct {
	input string
	want  time.Time
}{
	// --- Anchors ---
	{"现在", u(2026, 3, 22, 10, 0, 0)},
	{"今天", u(2026, 3, 22, 10, 0, 0)},
	{"今日", u(2026, 3, 22, 10, 0, 0)},
	{"明天", u(2026, 3, 23, 10, 0, 0)},
	{"昨天", u(2026, 3, 21, 10, 0, 0)},
	{"后天", u(2026, 3, 24, 10, 0, 0)},
	{"前天", u(2026, 3, 20, 10, 0, 0)},

	// --- Time words ---
	{"中午", u(2026, 3, 22, 12, 0, 0)},
	{"正午", u(2026, 3, 22, 12, 0, 0)},
	{"午夜", u(2026, 3, 22, 0, 0, 0)},

	// --- Absolute date: YEAR MONTH INTEGER (年月日) ---
	{"2026年3月24日", u(2026, 3, 24, 0, 0, 0)},
	{"2026年3月1日", u(2026, 3, 1, 0, 0, 0)},
	{"2026年3月", u(2026, 3, 1, 0, 0, 0)}, // YEAR MONTH

	// --- AMPM TIME ---
	{"下午3点", u(2026, 3, 22, 15, 0, 0)},
	{"上午9点30分", u(2026, 3, 22, 9, 30, 0)},
	{"下午3点30分", u(2026, 3, 22, 15, 30, 0)},

	// --- Anchor + AMPM + TIME ---
	{"明天下午3点", u(2026, 3, 23, 15, 0, 0)},
	{"明天上午9点30分", u(2026, 3, 23, 9, 30, 0)},

	// --- Combined date + AMPM + TIME ---
	{"2026年3月24日下午3点", u(2026, 3, 24, 15, 0, 0)},

	// --- Relative: N天后 / N天前 (INTEGER UNIT MODIFIER) ---
	{"3天后", u(2026, 3, 25, 10, 0, 0)},
	{"2天前", u(2026, 3, 20, 10, 0, 0)},
	{"30分钟前", u(2026, 3, 22, 9, 30, 0)},
	{"2小时后", u(2026, 3, 22, 12, 0, 0)},

	// --- Direction + unit macros ---
	{"下周", u(2026, 3, 23, 0, 0, 0)}, // next week (Monday March 23)
	{"上周", u(2026, 3, 9, 0, 0, 0)},  // last week (Monday March 9)
	{"本周", u(2026, 3, 16, 0, 0, 0)}, // this week (Monday March 16)
	{"下个月", u(2026, 4, 1, 0, 0, 0)}, // next month
	{"上个月", u(2026, 2, 1, 0, 0, 0)}, // last month
	{"本月", u(2026, 3, 1, 0, 0, 0)},  // this month
	{"明年", u(2027, 1, 1, 0, 0, 0)},  // next year
	{"去年", u(2025, 1, 1, 0, 0, 0)},  // last year

	// --- Direction + weekday macros ---
	{"下周一", u(2026, 3, 23, 0, 0, 0)}, // next Monday
	{"上周五", u(2026, 3, 20, 0, 0, 0)}, // last Friday

	// --- Weekdays (standalone) ---
	{"星期一", u(2026, 3, 23, 0, 0, 0)}, // nearest Monday (forward from Sunday)
	{"周五", u(2026, 3, 20, 0, 0, 0)},  // nearest Friday (March 20, 2 days before Sunday)

	// --- CLDR patterns: N年前/后 (year unit, bug fix) ---
	{"2年前", u(2024, 3, 22, 10, 0, 0)}, // 2 years ago
	{"3年后", u(2029, 3, 21, 10, 0, 0)}, // in 3 years (3×365 days, crosses 2028 leap year)

	// --- CLDR patterns: N秒钟前/后 (emphatic second form) ---
	{"30秒钟前", u(2026, 3, 22, 9, 59, 30)}, // 30 seconds ago
	{"10秒钟后", u(2026, 3, 22, 10, 0, 10)}, // in 10 seconds

	// --- Supplementary data ---
	{"礼拜一", u(2026, 3, 23, 0, 0, 0)},   // colloquial Monday (nearest, from Sunday)
	{"礼拜天", u(2026, 3, 22, 0, 0, 0)},   // colloquial Sunday (nearest = today, reference is Sunday)
	{"刚刚", u(2026, 3, 22, 10, 0, 0)},   // just now → AnchorNow
	{"此时", u(2026, 3, 22, 10, 0, 0)},   // at this moment → AnchorNow
	{"2個月后", u(2026, 5, 21, 10, 0, 0)}, // traditional 個月 (2×30 days = 60 days)
}

func TestChinese(t *testing.T) {
	for _, tc := range chineseCases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := Chinese.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, zhNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q)\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
		})
	}
}
