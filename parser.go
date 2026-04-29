package nowandlater

import (
	"time"

	"github.com/client9/nowandlater/internal/engine"
	"github.com/client9/nowandlater/languages"
)

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

	// Ambiguity controls how underspecified inputs like "5 hours", "monday",
	// or "March 5" are resolved. The zero value defaults to scheduling-oriented
	// behaviour.
	Ambiguity AmbiguityConfig

	// WeekStartSunday changes week-based boundaries like "this week" and
	// "next week" to use Sunday instead of Monday. The zero value keeps the
	// default Monday-start behavior.
	WeekStartSunday bool
}

// Parse converts a natural-language date/time string into a time.Time.
// It calls Lang.Parse followed by Resolve, using the configured defaults.
func (p Parser) Parse(input string) (time.Time, error) {
	lang := p.Lang
	if lang == nil {
		lang = &languages.LangEn
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
	policy := p.Ambiguity.resolvePolicy()
	policy.WeekStartSunday = p.WeekStartSunday
	return engine.ResolveWithPolicy(slots, now, policy)
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
		lang = &languages.LangEn
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
	policy := p.Ambiguity.resolvePolicy()
	policy.WeekStartSunday = p.WeekStartSunday
	return engine.ResolveIntervalWithPolicy(slots, now, policy)
}
