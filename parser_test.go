package nowandlater

import (
	"testing"
	"time"

	"github.com/client9/nowandlater/languages"
)

// parserNow is the fixed reference time for Parser tests.
var parserNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

func u(year, month, day, hour, min, sec int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}
func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestParserZeroValue(t *testing.T) {
	// Zero Parser must not panic and must return a plausible result.
	p := Parser{}
	got, err := p.Parse("today")
	if err != nil {
		t.Fatalf("Parse(\"today\") error: %v", err)
	}
	if got.IsZero() {
		t.Fatal("Parse(\"today\") returned zero time")
	}
}

func TestParserNowFunc(t *testing.T) {
	p := Parser{
		Lang: &languages.LangEn,
		Now:  fixedNow(parserNow),
	}
	got, err := p.Parse("tomorrow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := u(2026, 3, 23, 10, 0, 0)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParserLocation(t *testing.T) {
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York not available:", err)
	}
	p := Parser{
		Lang:     &languages.LangEn,
		Location: nyc,
		Now:      fixedNow(parserNow),
	}
	got, err := p.Parse("today")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Location() != nyc {
		t.Errorf("got location %v, want America/New_York", got.Location())
	}
}

func TestParserInputTzOverridesLocation(t *testing.T) {
	// Input-embedded timezone must win over Parser.Location.
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York not available:", err)
	}
	p := Parser{
		Lang:     &languages.LangEn,
		Location: nyc,
		Now:      fixedNow(parserNow),
	}
	got, err := p.Parse("3pm UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Location() != time.UTC {
		t.Errorf("got location %v, want UTC", got.Location())
	}
}

// spNow is the fixed reference time for LangEs resolver tests.
// Same date as resolveNow (2026-03-22 10:00:00 UTC, a Sunday) for easy comparison.
var spNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

func TestParserLangEs(t *testing.T) {
	p := Parser{
		Lang: &languages.LangEs,
		Now:  fixedNow(spNow),
	}
	got, err := p.Parse("mañana")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := u(2026, 3, 23, 10, 0, 0)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParserParseInterval(t *testing.T) {
	p := Parser{
		Lang: &languages.LangEn,
		Now:  fixedNow(parserNow), // 2026-03-22 10:00:00 UTC, Sunday
	}

	// "next week" → [Monday 2026-03-23, Monday 2026-03-30)
	start, end, err := p.ParseInterval("next week")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantStart := u(2026, 3, 23, 0, 0, 0)
	wantEnd := u(2026, 3, 30, 0, 0, 0)
	if !start.Equal(wantStart) {
		t.Errorf("start: got %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end:   got %v, want %v", end, wantEnd)
	}

	// "in a fortnight" → [2026-04-05, 2026-04-19) — covers EndOf PeriodFortnight
	start, end, err = p.ParseInterval("in a fortnight")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !start.Equal(u(2026, 4, 5, 0, 0, 0)) {
		t.Errorf("fortnight start: got %v, want %v", start, u(2026, 4, 5, 0, 0, 0))
	}
	if !end.Equal(u(2026, 4, 19, 0, 0, 0)) {
		t.Errorf("fortnight end:   got %v, want %v", end, u(2026, 4, 19, 0, 0, 0))
	}

	// "in 3 hours" → [13:00, 14:00) — covers startOfPeriod/EndOf PeriodHour
	start, end, err = p.ParseInterval("in 3 hours")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !start.Equal(u(2026, 3, 22, 13, 0, 0)) {
		t.Errorf("hours start: got %v, want %v", start, u(2026, 3, 22, 13, 0, 0))
	}
	if !end.Equal(u(2026, 3, 22, 14, 0, 0)) {
		t.Errorf("hours end:   got %v, want %v", end, u(2026, 3, 22, 14, 0, 0))
	}

	// "in 30 minutes" → [10:30, 10:31) — covers startOfPeriod/EndOf PeriodMinute
	start, end, err = p.ParseInterval("in 30 minutes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !start.Equal(u(2026, 3, 22, 10, 30, 0)) {
		t.Errorf("minutes start: got %v, want %v", start, u(2026, 3, 22, 10, 30, 0))
	}
	if !end.Equal(u(2026, 3, 22, 10, 31, 0)) {
		t.Errorf("minutes end:   got %v, want %v", end, u(2026, 3, 22, 10, 31, 0))
	}

	// "in 30 seconds" → [10:00:30, 10:00:31) — covers startOfPeriod/EndOf PeriodSecond
	start, end, err = p.ParseInterval("in 30 seconds")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !start.Equal(u(2026, 3, 22, 10, 0, 30)) {
		t.Errorf("seconds start: got %v, want %v", start, u(2026, 3, 22, 10, 0, 30))
	}
	if !end.Equal(u(2026, 3, 22, 10, 0, 31)) {
		t.Errorf("seconds end:   got %v, want %v", end, u(2026, 3, 22, 10, 0, 31))
	}

	// Error propagation: unknown input returns ErrUnknownSignature.
	_, _, err = p.ParseInterval("xyzzy frobozz")
	if err == nil {
		t.Error("expected error for unrecognised input, got nil")
	}
}
