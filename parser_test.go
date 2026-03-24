package nowandlater

import (
	"testing"
	"time"
)

// parserNow is the fixed reference time for Parser tests.
var parserNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

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
		Lang: &English,
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
		Lang:     &English,
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
		Lang:     &English,
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

func TestParserSpanish(t *testing.T) {
	p := Parser{
		Lang: &Spanish,
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
		Lang: &English,
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

	// Error propagation: unknown input returns ErrUnknownSignature.
	_, _, err = p.ParseInterval("xyzzy frobozz")
	if err == nil {
		t.Error("expected error for unrecognised input, got nil")
	}
}
