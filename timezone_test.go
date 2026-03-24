package nowandlater

import (
	"testing"
	"time"
)

func TestParseNumericOffset(t *testing.T) {
	cases := []struct {
		input   string
		wantOff int // seconds east of UTC
		wantErr bool
	}{
		// Valid: +HH
		{"+00", 0, false},
		{"+05", 5 * 3600, false},
		{"-07", -7 * 3600, false},
		{"+14", 14 * 3600, false}, // boundary: max hours

		// Valid: +HHMM
		{"+0530", 5*3600 + 30*60, false},
		{"-0700", -7 * 3600, false},

		// Valid: +HH:MM
		{"+05:30", 5*3600 + 30*60, false},
		{"-07:00", -7 * 3600, false},
		{"+00:00", 0, false},

		// Too short (< 3 chars)
		{"+", 0, true},
		{"+5", 0, true},

		// No sign prefix
		{"0530", 0, true},
		{"UTC", 0, true},

		// Wrong length (3 chars with sign = 2 rest; only 2,4,5 are valid)
		{"+5:0", 0, true},    // len(rest)=3 → default branch
		{"+123456", 0, true}, // len(rest)=6 → default branch

		// +HH:MM with wrong separator
		{"+05X30", 0, true},

		// Hours out of range
		{"+15", 0, true},
		{"+15:00", 0, true},

		// Minutes out of range
		{"+05:60", 0, true},
		{"+0560", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			loc, err := parseNumericOffset(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("parseNumericOffset(%q) = %v, want error", tc.input, loc)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseNumericOffset(%q) error: %v", tc.input, err)
			}
			_, got := time.Unix(0, 0).In(loc).Zone()
			if got != tc.wantOff {
				t.Errorf("parseNumericOffset(%q) offset = %d, want %d", tc.input, got, tc.wantOff)
			}
		})
	}
}

func TestIsTimezoneOffset(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		// Valid forms
		{"+05", true},
		{"-07", true},
		{"+0530", true},
		{"-0700", true},
		{"+05:30", true},
		{"-07:00", true},

		// Too short
		{"+5", false},
		{"+", false},
		{"", false},

		// No sign
		{"0530", false},
		{"UTC", false},

		// Wrong length (rest=3 or 6, etc.)
		{"+5:0", false},
		{"+123456", false},

		// Non-digits in rest
		{"+HH", false},
		{"+05:XX", false},

		// +HH:MM with wrong separator char
		{"+05X30", false},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := isTimezoneOffset(tc.input)
			if got != tc.want {
				t.Errorf("isTimezoneOffset(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
