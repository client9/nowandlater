package nowandlater

import "time"

// Parser wraps a Lang with runtime defaults, providing a single-call
// Parse method that combines tokenization and resolution.
//
// The zero value is valid: it uses LangEn, time.Local, and time.Now.
type Parser struct {
	// Lang is the language configuration to use for parsing.
	// If nil, LangEn is used.
	Lang *Lang

	// Location is the default timezone for the returned time when the
	// input string contains no explicit timezone. If nil, time.Local is used.
	// An explicit timezone in the input (e.g. "3pm EST") always takes priority.
	Location *time.Location

	// Now is the time source used as the reference point for relative
	// expressions ("tomorrow", "next week", etc.). If nil, time.Now is used.
	// Set this to a fixed function for deterministic tests.
	Now func() time.Time
}

// Parse converts a natural-language date/time string into a time.Time.
// It calls Lang.Parse followed by Resolve, using the configured defaults.
func (p Parser) Parse(input string) (time.Time, error) {
	lang := p.Lang
	if lang == nil {
		lang = &LangEn
	}
	now := time.Now()
	if p.Now != nil {
		now = p.Now()
	}
	if p.Location != nil {
		now = now.In(p.Location)
	}
	slots, err := lang.Parse(input)
	if err != nil {
		return time.Time{}, err
	}
	return Resolve(slots, now)
}

// ParseInterval converts a natural-language date/time string into a
// half-open calendar interval [start, end). It calls Lang.Parse followed by
// ResolveInterval, using the configured defaults.
//
// See ResolveInterval for semantics, including the calendar-alignment
// behaviour for delta-based expressions.
func (p Parser) ParseInterval(input string) (start, end time.Time, err error) {
	lang := p.Lang
	if lang == nil {
		lang = &LangEn
	}
	now := time.Now()
	if p.Now != nil {
		now = p.Now()
	}
	if p.Location != nil {
		now = now.In(p.Location)
	}
	slots, err := lang.Parse(input)
	if err != nil {
		return
	}
	return ResolveInterval(slots, now)
}
