package tests

import (
	. "github.com/client9/nowandlater/internal/engine"
	. "github.com/client9/nowandlater/languages"
	"testing"
	"time"
)

// resolveNow is the fixed reference time used across all resolver tests.
// 2026-03-22 10:00:00 UTC — a Sunday (Go: Weekday()=0, our system: weekday=6).
var resolveNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

func u(year, month, day, hour, min, sec int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}

// TestResolveTimezone verifies that an explicit timezone in the input overrides
// now.Location() in the returned time.Time.
func TestResolveTimezone(t *testing.T) {
	est := time.FixedZone("EST", -5*3600)
	pst := time.FixedZone("PST", -8*3600)
	mdt := time.FixedZone("MDT", -6*3600)

	cases := []struct {
		input string
		want  time.Time
	}{
		// Time-only in a named zone
		{"9:30 EST", time.Date(2026, 3, 22, 9, 30, 0, 0, est)},
		{"9:30 AM UTC", time.Date(2026, 3, 22, 9, 30, 0, 0, time.UTC)},
		// Weekday + time in a named zone
		{"next Monday at 9 AM PST", time.Date(2026, 3, 23, 9, 0, 0, 0, pst)},
		// Full date + time with Z suffix (ISO 8601 dashed)
		{"2026-12-04T09:30:00Z", time.Date(2026, 12, 4, 9, 30, 0, 0, time.UTC)},
		// Compact ISO 8601 UTC (YYYYMMDDThhmmssZ)
		{"20260429T030444Z", time.Date(2026, 4, 29, 3, 4, 44, 0, time.UTC)},
		{"20260429t030444z", time.Date(2026, 4, 29, 3, 4, 44, 0, time.UTC)},
		// Full date + time with numeric offset glued
		{"2026-12-04T09:30:00-07:00",
			time.Date(2026, 12, 4, 9, 30, 0, 0, time.FixedZone("-07:00", -7*3600))},
		// Full date + time + separate named zone
		{"2026-12-04 09:30 EST", time.Date(2026, 12, 4, 9, 30, 0, 0, est)},
		// Numeric offset as separate token
		{"at 3 PM -07:00", time.Date(2026, 3, 22, 15, 0, 0, 0, time.FixedZone("-07:00", -7*3600))},
		// MDT
		{"tomorrow at noon MDT", time.Date(2026, 3, 23, 12, 0, 0, 0, mdt)},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangEn.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			got, err := Resolve(slots, resolveNow)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			// Compare as UTC instants (Equal) and also check the zone name/offset.
			if !got.Equal(tc.want) {
				t.Errorf("Resolve(%q) time\n  got  %v\n  want %v", tc.input, got, tc.want)
			}
			_, gotOffset := got.Zone()
			_, wantOffset := tc.want.Zone()
			if gotOffset != wantOffset {
				t.Errorf("Resolve(%q) zone offset got %d, want %d", tc.input, gotOffset, wantOffset)
			}
		})
	}
}

// TestResolveWeekdayArithmetic stress-tests ResolveWeekday from every starting
// weekday to every target weekday for all three directions.
func TestResolveWeekdayArithmetic(t *testing.T) {
	allWeekdays := []Weekday{
		WeekdayMonday, WeekdayTuesday, WeekdayWednesday, WeekdayThursday,
		WeekdayFriday, WeekdaySaturday, WeekdaySunday,
	}

	// For each starting weekday, verify future/past/nearest produce the correct
	// day-of-week and that future/past never return today.
	for startIdx, startDay := range allWeekdays {
		// Build a now that falls on startDay. resolveNow is Sunday (our wd=6).
		// Shift by (startIdx - 6) days.
		now := resolveNow.AddDate(0, 0, startIdx-6)
		gotWd := (int(now.Weekday()) + 6) % 7
		if gotWd != startIdx {
			t.Fatalf("test setup error: startIdx=%d but now.Weekday()=%d", startIdx, gotWd)
		}

		for _, targetWd := range allWeekdays {
			name := startDay.String() + "→" + targetWd.String()
			// 0-based index of target for post-call weekday verification.
			wantIdx := int(targetWd) - 1

			// future: must be strictly after now, must be correct weekday
			t.Run("future/"+name, func(t *testing.T) {
				got := ResolveWeekday(targetWd, DirectionFuture, now)
				if !got.After(now) {
					t.Errorf("future: %s got %v, not after now %v", name, got, now)
				}
				gotWd := (int(got.Weekday()) + 6) % 7
				if gotWd != wantIdx {
					t.Errorf("future: %s weekday = %d, want %d", name, gotWd, wantIdx)
				}
				if got.Sub(now) >= 8*24*time.Hour {
					t.Errorf("future: %s 8+ days ahead: %v", name, got.Sub(now))
				}
			})

			// past: must be strictly before now, must be correct weekday
			t.Run("past/"+name, func(t *testing.T) {
				got := ResolveWeekday(targetWd, DirectionPast, now)
				if !got.Before(now) {
					t.Errorf("past: %s got %v, not before now %v", name, got, now)
				}
				gotWd := (int(got.Weekday()) + 6) % 7
				if gotWd != wantIdx {
					t.Errorf("past: %s weekday = %d, want %d", name, gotWd, wantIdx)
				}
				if now.Sub(got) >= 8*24*time.Hour {
					t.Errorf("past: %s 8+ days back: %v", name, now.Sub(got))
				}
			})

			// nearest: must be correct weekday; if today matches, must equal StartOfDay(now)
			t.Run("nearest/"+name, func(t *testing.T) {
				got := ResolveWeekday(targetWd, DirectionNearest, now)
				gotWd := (int(got.Weekday()) + 6) % 7
				if gotWd != wantIdx {
					t.Errorf("nearest: %s weekday = %d, want %d", name, gotWd, wantIdx)
				}
				diff := got.Sub(StartOfDay(now))
				if diff < 0 {
					diff = -diff
				}
				if diff > 3*24*time.Hour+12*time.Hour {
					t.Errorf("nearest: %s more than 3.5 days from today: %v", name, diff)
				}
			})
		}
	}
}

func TestEndOf(t *testing.T) {
	cases := []struct {
		start  time.Time
		period Period
		want   time.Time
	}{
		// Sub-day periods
		{u(2026, 3, 22, 10, 0, 0), PeriodSecond, u(2026, 3, 22, 10, 0, 1)},
		{u(2026, 3, 22, 10, 0, 0), PeriodMinute, u(2026, 3, 22, 10, 1, 0)},
		{u(2026, 3, 22, 10, 0, 0), PeriodHour, u(2026, 3, 22, 11, 0, 0)},
		// Day and fortnight
		{u(2026, 3, 22, 0, 0, 0), PeriodDay, u(2026, 3, 23, 0, 0, 0)},
		{u(2026, 3, 22, 0, 0, 0), PeriodFortnight, u(2026, 4, 5, 0, 0, 0)},
		// Week (Monday to following Monday)
		{u(2026, 3, 16, 0, 0, 0), PeriodWeek, u(2026, 3, 23, 0, 0, 0)},
		// Month (handles Dec→Jan and leap year via time.Date overflow)
		{u(2026, 3, 1, 0, 0, 0), PeriodMonth, u(2026, 4, 1, 0, 0, 0)},
		{u(2026, 12, 1, 0, 0, 0), PeriodMonth, u(2027, 1, 1, 0, 0, 0)},
		// Year
		{u(2026, 1, 1, 0, 0, 0), PeriodYear, u(2027, 1, 1, 0, 0, 0)},
		// Jan 31 + 1 day → Feb 1 (not Feb 31)
		{u(2026, 1, 31, 0, 0, 0), PeriodDay, u(2026, 2, 1, 0, 0, 0)},
		// Leap year: Feb 28 → Feb 29
		{u(2024, 2, 28, 0, 0, 0), PeriodDay, u(2024, 2, 29, 0, 0, 0)},
	}
	for _, c := range cases {
		got := EndOf(c.start, c.period)
		if !got.Equal(c.want) {
			t.Errorf("EndOf(%v, %v) = %v, want %v", c.start, c.period, got, c.want)
		}
	}
}

func TestResolveInterval(t *testing.T) {
	now := resolveNow // 2026-03-22 10:00:00 UTC, Sunday

	cases := []struct {
		input     string
		wantStart time.Time
		wantEnd   time.Time
	}{
		// --- Calendar path: start == Resolve result ---
		{"next week", u(2026, 3, 23, 0, 0, 0), u(2026, 3, 30, 0, 0, 0)},
		{"last month", u(2026, 2, 1, 0, 0, 0), u(2026, 3, 1, 0, 0, 0)},
		{"next year", u(2027, 1, 1, 0, 0, 0), u(2028, 1, 1, 0, 0, 0)},
		{"next Monday", u(2026, 3, 23, 0, 0, 0), u(2026, 3, 24, 0, 0, 0)},
		{"March 15, 2026", u(2026, 3, 15, 0, 0, 0), u(2026, 3, 16, 0, 0, 0)},

		// --- Delta path: start is calendar-aligned, may differ from Resolve ---
		// "tomorrow" → Resolve=2026-03-23 10:00; interval=[03-23 00:00, 03-24 00:00)
		{"tomorrow", u(2026, 3, 23, 0, 0, 0), u(2026, 3, 24, 0, 0, 0)},
		// "today" → Resolve=2026-03-22 10:00; interval=[03-22 00:00, 03-23 00:00)
		{"today", u(2026, 3, 22, 0, 0, 0), u(2026, 3, 23, 0, 0, 0)},
		// "in 3 days" → Resolve=2026-03-25 10:00; interval=[03-25 00:00, 03-26 00:00)
		{"in 3 days", u(2026, 3, 25, 0, 0, 0), u(2026, 3, 26, 0, 0, 0)},

		// --- Sub-day: Period=minute, truncated to minute boundary ---
		{"at 9:30", u(2026, 3, 22, 9, 30, 0), u(2026, 3, 22, 9, 31, 0)},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			slots, err := LangEn.Parse(c.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			start, end, err := ResolveInterval(slots, now)
			if err != nil {
				t.Fatalf("ResolveInterval error: %v", err)
			}
			if !start.Equal(c.wantStart) {
				t.Errorf("start: got %v, want %v", start, c.wantStart)
			}
			if !end.Equal(c.wantEnd) {
				t.Errorf("end:   got %v, want %v", end, c.wantEnd)
			}
		})
	}
}

func TestResolveWeekStartSunday(t *testing.T) {
	now := resolveNow // 2026-03-22 10:00:00 UTC, Sunday
	policy := ResolvePolicy{
		ImplicitDurationDirection: DirectionFuture,
		CalendarDirection:         DirectionNearest,
		MonthDayDirection:         DirectionFuture,
		WeekStartSunday:           true,
	}

	cases := []struct {
		input string
		want  time.Time
	}{
		{"this week", u(2026, 3, 22, 0, 0, 0)},
		{"next week", u(2026, 3, 29, 0, 0, 0)},
		{"last week", u(2026, 3, 15, 0, 0, 0)},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangEn.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			got, err := ResolveWithPolicy(slots, now, policy)
			if err != nil {
				t.Fatalf("ResolveWithPolicy error: %v", err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("ResolveWithPolicy(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}

	start, end, err := ResolveIntervalWithPolicy(mustParse(t, "this week"), now, policy)
	if err != nil {
		t.Fatalf("ResolveIntervalWithPolicy error: %v", err)
	}
	if !start.Equal(u(2026, 3, 22, 0, 0, 0)) {
		t.Errorf("interval start: got %v, want %v", start, u(2026, 3, 22, 0, 0, 0))
	}
	if !end.Equal(u(2026, 3, 29, 0, 0, 0)) {
		t.Errorf("interval end: got %v, want %v", end, u(2026, 3, 29, 0, 0, 0))
	}
}

func TestResolveWeekStartSundayMidWeek(t *testing.T) {
	// Wednesday 2026-03-25: exercises the non-zero offset path for both formulas.
	// Sunday-start offset = int(Wednesday) = 3 → back to 2026-03-22.
	// Monday-start offset = (3+6)%7 = 2 → back to 2026-03-23 (verified by existing tests).
	now := time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)
	policy := ResolvePolicy{
		ImplicitDurationDirection: DirectionFuture,
		CalendarDirection:         DirectionNearest,
		MonthDayDirection:         DirectionFuture,
		WeekStartSunday:           true,
	}

	cases := []struct {
		input string
		want  time.Time
	}{
		{"this week", u(2026, 3, 22, 0, 0, 0)}, // Sun 2026-03-22
		{"next week", u(2026, 3, 29, 0, 0, 0)}, // Sun 2026-03-29
		{"last week", u(2026, 3, 15, 0, 0, 0)}, // Sun 2026-03-15
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := LangEn.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			got, err := ResolveWithPolicy(slots, now, policy)
			if err != nil {
				t.Fatalf("ResolveWithPolicy error: %v", err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("ResolveWithPolicy(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}

	start, end, err := ResolveIntervalWithPolicy(mustParse(t, "this week"), now, policy)
	if err != nil {
		t.Fatalf("ResolveIntervalWithPolicy error: %v", err)
	}
	if !start.Equal(u(2026, 3, 22, 0, 0, 0)) {
		t.Errorf("interval start: got %v, want %v", start, u(2026, 3, 22, 0, 0, 0))
	}
	if !end.Equal(u(2026, 3, 29, 0, 0, 0)) {
		t.Errorf("interval end: got %v, want %v", end, u(2026, 3, 29, 0, 0, 0))
	}
}

func mustParse(t *testing.T, input string) *ParsedDateSlots {
	t.Helper()
	slots, err := LangEn.Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) error: %v", input, err)
	}
	return slots
}
