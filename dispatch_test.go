package nowandlater

import (
	"errors"
	"testing"
)

// parseCase describes one input and its expected ParsedDateSlots fields.
// Only non-nil expected fields are checked; fields left as zero value are ignored.
type parseCase struct {
	input   string
	year    int
	month   int
	day     int
	weekday Weekday
	hour    *int
	minute  *int
	delta   *int
	dir     Direction
	period  Period
}

var parseCases = []parseCase{
	// --- Anchors ---
	{input: "now", delta: new(0), period: PeriodSecond},
	{input: "today", delta: new(0), period: PeriodDay},
	{input: "tomorrow", delta: new(86400), period: PeriodDay},
	{input: "yesterday", delta: new(-86400), period: PeriodDay},

	// --- Weekday ---
	{input: "Monday", weekday: WeekdayMonday, dir: DirectionNearest, period: PeriodDay},
	{input: "Sunday", weekday: WeekdaySunday, dir: DirectionNearest, period: PeriodDay},

	// --- Direction + weekday ---
	{input: "next Monday", weekday: WeekdayMonday, dir: DirectionFuture, period: PeriodDay},
	{input: "last Friday", weekday: WeekdayFriday, dir: DirectionPast, period: PeriodDay},
	{input: "this Wednesday", weekday: WeekdayWednesday, dir: DirectionNearest, period: PeriodDay},

	// --- Direction + weekday + time ---
	{input: "next Monday at 9:30", weekday: WeekdayMonday, dir: DirectionFuture, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "next Monday at 9:30 AM", weekday: WeekdayMonday, dir: DirectionFuture, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "next Monday at 9:30 PM", weekday: WeekdayMonday, dir: DirectionFuture, hour: new(21), minute: new(30), period: PeriodMinute},
	{input: "next Tuesday at 3", weekday: WeekdayTuesday, dir: DirectionFuture, hour: new(3), period: PeriodHour},
	{input: "last Friday at 5", weekday: WeekdayFriday, dir: DirectionPast, hour: new(5), period: PeriodHour},

	// --- Direction + unit ---
	{input: "next week", dir: DirectionFuture, period: PeriodWeek},
	{input: "last month", dir: DirectionPast, period: PeriodMonth},
	{input: "this year", dir: DirectionNearest, period: PeriodYear},
	{input: "next day", dir: DirectionFuture, period: PeriodDay},

	// --- Time of day ---
	{input: "3pm", hour: new(15), period: PeriodHour},
	{input: "12am", hour: new(0), period: PeriodHour},  // midnight
	{input: "12pm", hour: new(12), period: PeriodHour}, // noon
	{input: "at 09:30", hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "at 9:30 AM", hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "at 9:30 PM", hour: new(21), minute: new(30), period: PeriodMinute},
	{input: "at 09:30:45", hour: new(9), minute: new(30), period: PeriodSecond}, // with seconds
	{input: "7.15pm", hour: new(19), minute: new(15), period: PeriodMinute},     // dot notation
	{input: "7.15 PM", hour: new(19), minute: new(15), period: PeriodMinute},
	{input: "730pm", hour: new(19), minute: new(30), period: PeriodMinute}, // compact HHMM
	{input: "after 730pm", hour: new(19), minute: new(30), period: PeriodMinute},

	// --- Noon / midnight preprocessing ---
	{input: "noon", hour: new(12), minute: new(0), period: PeriodMinute},
	{input: "midnight", hour: new(0), minute: new(0), period: PeriodMinute},
	{input: "at noon", hour: new(12), minute: new(0), period: PeriodMinute},
	{input: "at midnight", hour: new(0), minute: new(0), period: PeriodMinute},

	// --- Anchor + time ---
	{input: "today at 9:30", delta: new(0), hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "today at 9:30 AM", delta: new(0), hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "today at 9:30 PM", delta: new(0), hour: new(21), minute: new(30), period: PeriodMinute},
	{input: "today at 3pm", delta: new(0), hour: new(15), period: PeriodHour},
	{input: "today at 3", delta: new(0), hour: new(3), period: PeriodHour},
	{input: "tomorrow at 9:30 AM", delta: new(86400), hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "yesterday at noon", delta: new(-86400), hour: new(12), minute: new(0), period: PeriodMinute},

	// --- Calendar: month name forms ---
	{input: "March 5", month: 3, day: 5, period: PeriodDay},
	{input: "January 21st", month: 1, day: 21, period: PeriodDay},
	{input: "3rd of January", month: 1, day: 3, period: PeriodDay},
	{input: "Dec 3rd 2026", year: 2026, month: 12, day: 3, period: PeriodDay},
	{input: "December 2026", year: 2026, month: 12, period: PeriodMonth},
	{input: "Dec 2026", year: 2026, month: 12, period: PeriodMonth},
	{input: "2026 December", year: 2026, month: 12, period: PeriodMonth},

	// --- Calendar: numeric compound (all separators → same result) ---
	{input: "2026-12-04", year: 2026, month: 12, day: 4, period: PeriodDay},
	{input: "2026/12/03", year: 2026, month: 12, day: 3, period: PeriodDay},
	{input: "2026.12.03", year: 2026, month: 12, day: 3, period: PeriodDay},

	// --- Calendar: letter month compound ---
	{input: "2026-dec-04", year: 2026, month: 12, day: 4, period: PeriodDay},
	{input: "04-dec-2026", year: 2026, month: 12, day: 4, period: PeriodDay},

	// --- Calendar: MM/DD/YYYY (English MDY default) ---
	{input: "12/04/2026", year: 2026, month: 12, day: 4, period: PeriodDay},
	{input: "1/5/2026", year: 2026, month: 1, day: 5, period: PeriodDay},
	{input: "02/03/2016", year: 2016, month: 2, day: 3, period: PeriodDay},  // MDY: Feb 3 (ambiguous)
	{input: "30-01-2016", year: 2016, month: 1, day: 30, period: PeriodDay}, // unambiguous: 30 can't be month
	{input: "13-02-2016", year: 2016, month: 2, day: 13, period: PeriodDay}, // unambiguous: 13 can't be month

	// --- Calendar: compound with leading prep ---
	{input: "on 2026-12-04", year: 2026, month: 12, day: 4, period: PeriodDay},
	{input: "on 2026-dec-04", year: 2026, month: 12, day: 4, period: PeriodDay},

	// --- Year only ---
	{input: "2026", year: 2026, period: PeriodYear},

	// --- Fractional unit durations ---
	{input: "3.5 days", delta: new(302400), period: PeriodDay},
	{input: "in 1.5 hours", delta: new(5400), period: PeriodHour},
	{input: "1.5 hours ago", delta: new(-5400), period: PeriodHour},
	{input: "2.5 weeks", delta: new(1512000), period: PeriodWeek},
	{input: "0.5 days", delta: new(43200), period: PeriodDay},

	// --- Relative deltas: bare INTEGER UNIT (implied future, GNU date style) ---
	{input: "4 hours", delta: new(14400), period: PeriodHour},
	{input: "3 days", delta: new(259200), period: PeriodDay},
	{input: "2 weeks", delta: new(1209600), period: PeriodWeek},
	{input: "1 minute", delta: new(60), period: PeriodMinute},

	// --- Relative deltas ---
	{input: "3 days ago", delta: new(-259200), period: PeriodDay},
	{input: "1 hour ago", delta: new(-3600), period: PeriodHour},
	{input: "2 weeks from now", delta: new(1209600), period: PeriodWeek},
	{input: "3 days before now", delta: new(-259200), period: PeriodDay},      // "before" modifier
	{input: "3 hours after now", delta: new(10800), period: PeriodHour},       // "after" modifier
	{input: "3 days before tomorrow", delta: new(-172800), period: PeriodDay}, // anchor shift
	{input: "2 hours after today", delta: new(7200), period: PeriodHour},
	{input: "in 2 days", delta: new(172800), period: PeriodDay},
	{input: "in next 2 days", delta: new(172800), period: PeriodDay},
	{input: "in 3 hours", delta: new(10800), period: PeriodHour},
	{input: "in a week", delta: new(604800), period: PeriodWeek}, // "a" is FILLER
	{input: "in an hour", delta: new(3600), period: PeriodHour},  // "an" is FILLER
	{input: "a week ago", delta: new(-604800), period: PeriodWeek},
	{input: "an hour ago", delta: new(-3600), period: PeriodHour},

	// --- Combined date + time ---
	{input: "2026-12-04 09:30", year: 2026, month: 12, day: 4, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "2026-12-04 9:30 PM", year: 2026, month: 12, day: 4, hour: new(21), minute: new(30), period: PeriodMinute},
	{input: "2026-12-04T09:30:00", year: 2026, month: 12, day: 4, hour: new(9), minute: new(30), period: PeriodSecond}, // ISO 8601 T
	{input: "2026-dec-04 09:30", year: 2026, month: 12, day: 4, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "2026-dec-04 9:30 AM", year: 2026, month: 12, day: 4, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "Dec 3 2026 09:30", year: 2026, month: 12, day: 3, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "Dec 3 2026 9:30 PM", year: 2026, month: 12, day: 3, hour: new(21), minute: new(30), period: PeriodMinute},

	// --- Time-only: PREP INTEGER AMPM ---
	{input: "at 3 PM", hour: new(15), period: PeriodHour},
	{input: "at 9 AM", hour: new(9), period: PeriodHour},
	{input: "at 12 AM", hour: new(0), period: PeriodHour}, // midnight via 12-hour clock

	// --- Weekday + time (withPrepTime) ---
	{input: "Monday at 9:30", weekday: WeekdayMonday, dir: DirectionNearest, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "Friday at 9:30 PM", weekday: WeekdayFriday, dir: DirectionNearest, hour: new(21), minute: new(30), period: PeriodMinute},
	{input: "Saturday at 3 PM", weekday: WeekdaySaturday, dir: DirectionNearest, hour: new(15), period: PeriodHour},

	// --- Direction + weekday + integer + AMPM ---
	{input: "next Monday at 9 PM", weekday: WeekdayMonday, dir: DirectionFuture, hour: new(21), period: PeriodHour},
	{input: "last Friday at 8 AM", weekday: WeekdayFriday, dir: DirectionPast, hour: new(8), period: PeriodHour},

	// --- Month + day + time (withPrepTime) ---
	{input: "March 5 at 9:30", month: 3, day: 5, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "March 5 at 9:30 PM", month: 3, day: 5, hour: new(21), minute: new(30), period: PeriodMinute},
	{input: "March 5 at 3 PM", month: 3, day: 5, hour: new(15), period: PeriodHour},

	// --- Month + day + year + time (withPrepTime) ---
	{input: "Dec 3rd 2026 at 9:30", year: 2026, month: 12, day: 3, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "Dec 3rd 2026 at 9:30 AM", year: 2026, month: 12, day: 3, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "Dec 3rd 2026 at 3 PM", year: 2026, month: 12, day: 3, hour: new(15), period: PeriodHour},

	// --- YEAR INTEGER INTEGER + PREP + time (withPrepTime) ---
	{input: "2026-12-04 at 9:30", year: 2026, month: 12, day: 4, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "2026-12-04 at 9:30 AM", year: 2026, month: 12, day: 4, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "2026-12-04 at 3 PM", year: 2026, month: 12, day: 4, hour: new(15), period: PeriodHour},

	// --- YEAR MONTH INTEGER + PREP + time (withPrepTime) ---
	{input: "2026-dec-04 at 9:30", year: 2026, month: 12, day: 4, hour: new(9), minute: new(30), period: PeriodMinute},
	{input: "2026-dec-04 at 9:30 PM", year: 2026, month: 12, day: 4, hour: new(21), minute: new(30), period: PeriodMinute},
	{input: "2026-dec-04 at 3 PM", year: 2026, month: 12, day: 4, hour: new(15), period: PeriodHour},

	// --- RFC 2822 (WEEKDAY INTEGER MONTH YEAR TIME ±TZ) ---
	// Go reference time: Mon, 02 Jan 2006 15:04:05 -0700
	{input: "Mon, 02 Jan 2006 15:04:05 -0700", year: 2006, month: 1, day: 2, hour: new(15), minute: new(4), period: PeriodSecond},
	{input: "Mon, 2 Jan 2006 15:04:05 +0000", year: 2006, month: 1, day: 2, hour: new(15), minute: new(4), period: PeriodSecond},
	{input: "Tue, 10 Mar 2026 09:30:00 +0530", year: 2026, month: 3, day: 10, hour: new(9), minute: new(30), period: PeriodSecond},
	// Date-only (no time component)
	{input: "Mon, 02 Jan 2006", year: 2006, month: 1, day: 2, period: PeriodDay},
	// Leading-zero bare day (INTEGER2 folded into INTEGER by Signature)
	{input: "02 Jan 2026", year: 2026, month: 1, day: 2, period: PeriodDay},
}

func FuzzParse(f *testing.F) {
	for _, tc := range parseCases {
		f.Add(tc.input)
	}
	f.Fuzz(func(t *testing.T, orig string) {
		English.Parse(orig)
	})
}

func TestParse(t *testing.T) {
	for _, tc := range parseCases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := English.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			if tc.year != 0 && got.Year != tc.year {
				t.Errorf("Parse(%q).Year = %d, want %d", tc.input, got.Year, tc.year)
			}
			if tc.month != 0 && got.Month != tc.month {
				t.Errorf("Parse(%q).Month = %d, want %d", tc.input, got.Month, tc.month)
			}
			if tc.day != 0 && got.Day != tc.day {
				t.Errorf("Parse(%q).Day = %d, want %d", tc.input, got.Day, tc.day)
			}
			if tc.weekday != 0 && got.Weekday != tc.weekday {
				t.Errorf("Parse(%q).Weekday = %v, want %v", tc.input, got.Weekday, tc.weekday)
			}
			checkInt(t, tc.input, "Hour", &got.Hour, tc.hour)
			checkInt(t, tc.input, "Minute", &got.Minute, tc.minute)
			checkInt(t, tc.input, "DeltaSeconds", got.DeltaSeconds, tc.delta)
			if tc.dir != 0 && got.Direction != tc.dir {
				t.Errorf("Parse(%q).Direction = %v, want %v", tc.input, got.Direction, tc.dir)
			}
			if tc.period != 0 && got.Period != tc.period {
				t.Errorf("Parse(%q).Period = %q, want %q", tc.input, got.Period, tc.period)
			}
		})
	}
}

// TestDateOrder verifies that Lang.DateOrder controls MM/DD vs DD/MM interpretation
// for the ambiguous INTEGER INTEGER YEAR signature.
func TestDateOrder(t *testing.T) {
	dmyLang := Lang{Words: englishWords, OrdinalSuffixes: English.OrdinalSuffixes, DateOrder: DMY}

	cases := []struct {
		lang  *Lang
		input string
		month int
		day   int
	}{
		// MDY (English default): 02/03/2016 → Feb 3
		{&English, "02/03/2016", 2, 3},
		{&English, "12-04-2026", 12, 4},
		// DMY: 02/03/2016 → Mar 2
		{&dmyLang, "02/03/2016", 3, 2},
		{&dmyLang, "04-12-2026", 12, 4},
		// DMY: Spanish
		{&Spanish, "02/03/2016", 3, 2},
		{&Spanish, "04/12/2026", 12, 4},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := tc.lang.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			if slots.Month != tc.month {
				t.Errorf("Parse(%q).Month = %d, want %d", tc.input, slots.Month, tc.month)
			}
			if slots.Day != tc.day {
				t.Errorf("Parse(%q).Day = %d, want %d", tc.input, slots.Day, tc.day)
			}
		})
	}
}

func TestParseInvalidDate(t *testing.T) {
	cases := []string{
		"30-31-2016", // both values > 12 — no valid month/day assignment
		"00-01-2016", // day/month of zero
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := English.Parse(input)
			if err == nil {
				t.Errorf("Parse(%q) expected error, got nil", input)
			}
		})
	}
}

func TestParseUnknownSignature(t *testing.T) {
	cases := []string{
		"",
		"purple monkey dishwasher",
		"12-03",                 // DATE_FRAGMENT — no handler
		"100000000000000000000", // 21-digit integer — classifyBareInteger len>10 guard → UNKNOWN
		"0.1000000000000000000", // long decimal part — splitCompoundDate len>10 guard → DATE_FRAGMENT
		"0٣",                    // non-ASCII digit — allDigits ASCII guard → UNKNOWN, no mustAtoi panic
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := English.Parse(input)
			if !errors.Is(err, ErrUnknownSignature) {
				t.Errorf("Parse(%q) error = %v, want ErrUnknownSignature", input, err)
			}
		})
	}
}

func checkInt(t *testing.T, input, field string, got, want *int) {
	t.Helper()
	if want == nil {
		return // not checked
	}
	if got == nil {
		t.Errorf("Parse(%q).%s = nil, want %d", input, field, *want)
		return
	}
	if *got != *want {
		t.Errorf("Parse(%q).%s = %d, want %d", input, field, *got, *want)
	}
}
