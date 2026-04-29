package engine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Handlers
// Each handler receives the full token list (including FILLER) and returns
// a populated ParsedDateSlots. All handlers call FilterFillers first so that
// positional indices align exactly with the signature string.
// ---------------------------------------------------------------------------

// HandleAmbiguous is registered for signatures that are recognisable as
// date-like but cannot be resolved to a single meaning without more context.
// It returns ErrAmbiguous so callers can distinguish "I don't understand this"
// from "this could mean multiple things".
func HandleAmbiguous(_ []Token) (*ParsedDateSlots, error) {
	return nil, ErrAmbiguous
}

// handleAnchor handles: ANCHOR
// Examples: "today", "tomorrow", "yesterday", "now"
func handleAnchor(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	anchor := toks[0].Value.(Anchor)
	period := PeriodDay
	if anchor == AnchorNow {
		period = PeriodSecond
	}
	return &ParsedDateSlots{DeltaSeconds: new(AnchorToSeconds[anchor]), Period: period}, nil
}

// handleWeekday handles: WEEKDAY
// Example: "Monday"
func handleWeekday(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	return &ParsedDateSlots{
		Weekday:       toks[0].Value.(Weekday),
		Period:        PeriodDay,
		AmbiguousForm: AmbiguousBareWeekday,
	}, nil
}

// handleDirectionWeekday handles: DIRECTION WEEKDAY
// Examples: "next Monday", "last Friday"
func handleDirectionWeekday(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	return &ParsedDateSlots{
		Weekday:   toks[1].Value.(Weekday),
		Direction: toks[0].Value.(Direction),
		Period:    PeriodDay,
	}, nil
}

// validateAndApplyAMPM validates that h is in the 12-hour range and applies ampm.
func validateAndApplyAMPM(h int, ampm AMPM) (int, error) {
	if h < 1 || h > 12 {
		return 0, fmt.Errorf("nowandlater: hour %d out of range for 12-hour clock", h)
	}
	return ApplyAMPM(h, ampm), nil
}

// handleIntegerAMPM handles: INTEGER AMPM
// Examples: "3pm", "11am"
func handleIntegerAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	h, err := validateAndApplyAMPM(toks[0].Value.(int), toks[1].Value.(AMPM))
	if err != nil {
		return nil, err
	}
	return &ParsedDateSlots{Hour: h, Period: PeriodHour}, nil
}

// handlePrepTime handles: PREP TIME
// Example: "at 09:30", "at 09:30:45"
func handlePrepTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	timeVal := toks[1].Value.(string)
	h, m, sec := MustParseTime(timeVal)
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: TimePeriod(timeVal)}
	slots.Second = sec
	return slots, nil
}

// handlePrepTimeAMPM handles: PREP TIME AMPM
// Example: "at 9:30 AM"
func handlePrepTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	timeVal := toks[1].Value.(string)
	h, m, sec := MustParseTime(timeVal)
	h = ApplyAMPM(h, toks[2].Value.(AMPM))
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: TimePeriod(timeVal)}
	slots.Second = sec
	return slots, nil
}

// handleMonthDay handles: MONTH INTEGER
// Example: "March 5"
func handleMonthDay(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	d := toks[1].Value.(int)
	return &ParsedDateSlots{
		Month:         int(toks[0].Value.(Month)),
		Day:           d,
		Period:        PeriodDay,
		AmbiguousForm: AmbiguousMonthDay,
	}, nil
}

// handleMonthIntegerTime handles: MONTH INTEGER TIME
// Example: "Mar  2 15:04:05" (Go Stamp format)
// No year present; resolver uses the current year.
// handleDayMonth handles: INTEGER MONTH
// Example: "3rd of January" (fillers already dropped in signature)
func handleDayMonth(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	d := toks[0].Value.(int)
	return &ParsedDateSlots{
		Month:         int(toks[1].Value.(Month)),
		Day:           d,
		Period:        PeriodDay,
		AmbiguousForm: AmbiguousMonthDay,
	}, nil
}

// handleMonth handles: MONTH
// Example: "October"
func handleMonth(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	return &ParsedDateSlots{
		Month:         int(toks[0].Value.(Month)),
		Period:        PeriodMonth,
		AmbiguousForm: AmbiguousBareMonth,
	}, nil
}

// handleMonthDayYear handles: MONTH INTEGER YEAR
// Example: "Dec 3rd 2026"
func handleMonthDayYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	d := toks[1].Value.(int)
	y := toks[2].Value.(int)
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[0].Value.(Month)),
		Day:    d,
		Period: PeriodDay,
	}, nil
}

// handleRelativeDelta handles: INTEGER UNIT MODIFIER
// Example: "3 days ago"
func handleRelativeDelta(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[0].Value.(int)
	period := toks[1].Value.(Period)
	secs := periodToSeconds[period]
	sign := int(toks[2].Value.(Modifier))
	return &ParsedDateSlots{
		DeltaSeconds: new(n * secs * sign),
		Period:       period,
	}, nil
}

// handleModifierIntegerUnit handles: MODIFIER INTEGER UNIT
// Example: "hace 3 días" (Spanish: 3 days ago), "il y a 3 jours" (French, after phrase substitution)
// This is the modifier-first word order used by many non-English languages.
func handleModifierIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	sign := int(toks[0].Value.(Modifier))
	n := toks[1].Value.(int)
	period := toks[2].Value.(Period)
	secs := periodToSeconds[period]
	return &ParsedDateSlots{
		DeltaSeconds: new(n * secs * sign),
		Period:       period,
	}, nil
}

// handleModifierIntegerUnitIntegerUnit handles: MODIFIER INTEGER UNIT INTEGER UNIT
// Example: "hace 1 hora y 10 minutos" (Spanish: 1 hour and 10 minutes ago)
// This is the modifier-first compound form used by many non-English languages.
func handleModifierIntegerUnitIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	sign := int(toks[0].Value.(Modifier))
	secs, period := sumTwoUnits(toks, 1)
	return &ParsedDateSlots{DeltaSeconds: new(secs * sign), Period: period}, nil
}

// handleWeekdayDirection handles: WEEKDAY DIRECTION
// Example: "lunes próximo" (Spanish: next Monday), "lundi prochain" (French: next Monday)
// This is the weekday-first word order used by many non-English languages.
func handleWeekdayDirection(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	return &ParsedDateSlots{
		Weekday:   toks[0].Value.(Weekday),
		Direction: toks[1].Value.(Direction),
		Period:    PeriodDay,
	}, nil
}

// handleRelativeDeltaAnchor handles: INTEGER UNIT MODIFIER ANCHOR
// Examples: "2 weeks from now", "3 days before tomorrow", "1 hour after today"
// The anchor shifts the reference point: "3 days before tomorrow" = -2 days from now.
func handleRelativeDeltaAnchor(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[0].Value.(int)
	period := toks[1].Value.(Period)
	secs := periodToSeconds[period]
	sign := int(toks[2].Value.(Modifier))
	aDelta := AnchorToSeconds[toks[3].Value.(Anchor)]
	return &ParsedDateSlots{
		DeltaSeconds: new(n*secs*sign + aDelta),
		Period:       period,
	}, nil
}

// handleIntegerUnit handles: INTEGER UNIT
// Examples: "4 hours", "3 días" — bare count with no preposition or modifier.
// Implied direction is future (positive delta), matching GNU date behaviour.
func handleIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[0].Value.(int)
	period := toks[1].Value.(Period)
	secs := periodToSeconds[period]
	return &ParsedDateSlots{
		DeltaSeconds:  new(n * secs),
		Period:        period,
		AmbiguousForm: AmbiguousImplicitDuration,
	}, nil
}

// decimalUnitSeconds converts a fractional count and period to whole seconds.
func decimalUnitSeconds(n float64, period Period) int {
	return int(math.Round(n * float64(periodToSeconds[period])))
}

// handleDecimalUnit handles: DECIMAL UNIT
// Examples: "3.5 days", "1.5 horas" — implied future.
func handleDecimalUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[0].Value.(float64)
	period := toks[1].Value.(Period)
	return &ParsedDateSlots{
		DeltaSeconds:  new(decimalUnitSeconds(n, period)),
		Period:        period,
		AmbiguousForm: AmbiguousImplicitDuration,
	}, nil
}

// handlePrepDecimalUnit handles: PREP DECIMAL UNIT
// Examples: "in 1.5 hours", "en 2.5 días"
func handlePrepDecimalUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[1].Value.(float64)
	period := toks[2].Value.(Period)
	return &ParsedDateSlots{DeltaSeconds: new(decimalUnitSeconds(n, period)), Period: period}, nil
}

// handleDecimalUnitModifier handles: DECIMAL UNIT MODIFIER
// Examples: "1.5 hours ago", "2.5 días después"
func handleDecimalUnitModifier(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[0].Value.(float64)
	period := toks[1].Value.(Period)
	sign := int(toks[2].Value.(Modifier))
	return &ParsedDateSlots{DeltaSeconds: new(decimalUnitSeconds(n, period) * sign), Period: period}, nil
}

// handleModifierDecimalUnit handles: MODIFIER DECIMAL UNIT
// Examples: "hace 1.5 días" (Spanish modifier-first word order)
func handleModifierDecimalUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	sign := int(toks[0].Value.(Modifier))
	n := toks[1].Value.(float64)
	period := toks[2].Value.(Period)
	return &ParsedDateSlots{DeltaSeconds: new(decimalUnitSeconds(n, period) * sign), Period: period}, nil
}

// handlePrepIntegerUnit handles: PREP INTEGER UNIT
// Example: "in 2 days"
func handlePrepIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[1].Value.(int)
	period := toks[2].Value.(Period)
	secs := periodToSeconds[period]
	return &ParsedDateSlots{
		DeltaSeconds: new(n * secs),
		Period:       period,
	}, nil
}

// sumTwoUnits sums two adjacent INTEGER UNIT pairs starting at toks[idx]
// and returns the total seconds and the finer of the two periods.
// toks must be already filtered (no FILLER tokens).
func sumTwoUnits(toks []Token, idx int) (secs int, period Period) {
	n1 := toks[idx].Value.(int)
	p1 := toks[idx+1].Value.(Period)
	n2 := toks[idx+2].Value.(int)
	p2 := toks[idx+3].Value.(Period)
	secs = n1*periodToSeconds[p1] + n2*periodToSeconds[p2]
	if periodToSeconds[p1] < periodToSeconds[p2] {
		period = p1
	} else {
		period = p2
	}
	return
}

// handlePrepIntegerUnitIntegerUnit handles: PREP INTEGER UNIT INTEGER UNIT
// Example: "in 1 hour and 10 minutes"
func handlePrepIntegerUnitIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	secs, period := sumTwoUnits(toks, 1)
	return &ParsedDateSlots{DeltaSeconds: new(secs), Period: period}, nil
}

// handleIntegerUnitIntegerUnit handles: INTEGER UNIT INTEGER UNIT
// Example: "1 hour and 10 minutes" — implied future.
func handleIntegerUnitIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	secs, period := sumTwoUnits(toks, 0)
	return &ParsedDateSlots{
		DeltaSeconds:  new(secs),
		Period:        period,
		AmbiguousForm: AmbiguousImplicitDuration,
	}, nil
}

// handleIntegerUnitIntegerUnitModifier handles: INTEGER UNIT INTEGER UNIT MODIFIER
// Example: "1 hour and 10 minutes ago"
func handleIntegerUnitIntegerUnitModifier(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	secs, period := sumTwoUnits(toks, 0)
	sign := int(toks[4].Value.(Modifier))
	return &ParsedDateSlots{DeltaSeconds: new(secs * sign), Period: period}, nil
}

// handlePrepDirectionIntegerUnit handles: PREP DIRECTION INTEGER UNIT
// Example: "in next 2 days"
func handlePrepDirectionIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	n := toks[2].Value.(int)
	period := toks[3].Value.(Period)
	secs := periodToSeconds[period]
	return &ParsedDateSlots{
		DeltaSeconds: new(n * secs),
		Period:       period,
	}, nil
}

// handleYearIntegerInteger handles: YEAR INTEGER INTEGER
// Examples: "2026-12-04", "2026/12/03", "2026.12.03" (all decomposed to same token sequence)
func handleYearIntegerInteger(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	y := toks[0].Value.(int)
	m := toks[1].Value.(int)
	d := toks[2].Value.(int)
	return &ParsedDateSlots{Year: y, Month: m, Day: d, Period: PeriodDay}, nil
}

// handleYearMonthInteger handles: YEAR MONTH INTEGER
// Example: "2026-dec-04"
func handleYearMonthInteger(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	y := toks[0].Value.(int)
	d := toks[2].Value.(int)
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[1].Value.(Month)),
		Day:    d,
		Period: PeriodDay,
	}, nil
}

// handleIntegerMonthYear handles: INTEGER MONTH YEAR
// Example: "04-dec-2026"
func handleIntegerMonthYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	d := toks[0].Value.(int)
	y := toks[2].Value.(int)
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[1].Value.(Month)),
		Day:    d,
		Period: PeriodDay,
	}, nil
}

// parseDateFragment splits a DATE_FRAGMENT raw string (e.g. "29/02/20") into
// month, day, and year applying DateOrder and Expand2DigitYear.
func parseDateFragment(raw string, order DateOrder) (mo, d, y int, err error) {
	var sep byte
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c == '-' || c == '/' || c == '.' {
			sep = c
			break
		}
	}
	if sep == 0 {
		return 0, 0, 0, fmt.Errorf("nowandlater: DATE_FRAGMENT has no separator: %q", raw)
	}
	parts := strings.Split(raw, string(sep))
	if len(parts) != 3 {
		return 0, 0, 0, ErrUnknownSignature
	}
	a, aerr := strconv.Atoi(parts[0])
	b, berr := strconv.Atoi(parts[1])
	c, cerr := strconv.Atoi(parts[2])
	if aerr != nil || berr != nil || cerr != nil {
		return 0, 0, 0, fmt.Errorf("nowandlater: DATE_FRAGMENT has non-numeric part: %q", raw)
	}
	mo, d = swapDateOrder(a, b, order)
	y = Expand2DigitYear(c)
	if mo < 1 || mo > 12 || d < 1 || d > 31 {
		return 0, 0, 0, fmt.Errorf("nowandlater: invalid date in %q", raw)
	}
	return mo, d, y, nil
}

// makeTimeAMPMDateFragmentHandler returns a Handler for TIME AMPM DATE_FRAGMENT
// and TIME AMPM PREP DATE_FRAGMENT (time precedes date, 2-digit year).
// DATE_FRAGMENT is always the last token; any intervening PREP is ignored.
// Examples: "1:30am 29/02/20", "1:30am at 29/02/20"
func makeTimeAMPMDateFragmentHandler(order DateOrder) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := FilterFillers(tokens)
		// [0]=TIME, [1]=AMPM, [last]=DATE_FRAGMENT (PREP, if present, is in between)
		timeVal := toks[0].Value.(string)
		h, m, s := MustParseTime(timeVal)
		h = ApplyAMPM(h, toks[1].Value.(AMPM))
		mo, d, y, err := parseDateFragment(toks[len(toks)-1].Value.(string), order)
		if err != nil {
			return nil, err
		}
		return &ParsedDateSlots{
			Year: y, Month: mo, Day: d,
			Hour: h, Minute: m, Second: s,
			Period: TimePeriod(timeVal),
		}, nil
	}
}

// makeDateFragmentHandler returns a Handler for DATE_FRAGMENT (date only).
// Used as the base for withPrepTime variants.
func makeDateFragmentHandler(order DateOrder) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := FilterFillers(tokens)
		mo, d, y, err := parseDateFragment(toks[0].Value.(string), order)
		if err != nil {
			return nil, err
		}
		return &ParsedDateSlots{Year: y, Month: mo, Day: d, Period: PeriodDay}, nil
	}
}

// handleWeekdayIntegerMonthYear handles: WEEKDAY INTEGER MONTH YEAR
// The weekday is informational (derivable from the date) and is ignored.
// Example: "Mon, 02 Jan 2006" (RFC 2822 date portion)
func handleWeekdayIntegerMonthYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	// [0]=WEEKDAY (ignored), [1]=INTEGER, [2]=MONTH, [3]=YEAR
	d := toks[1].Value.(int)
	y := toks[3].Value.(int)
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[2].Value.(Month)),
		Day:    d,
		Period: PeriodDay,
	}, nil
}

// handleIntegerMonthIntegerTime handles: INTEGER MONTH INTEGER TIME
// Examples: "02 Jan 06 15:04 MST" (RFC822), "22 Mar 26 10:04 -0700" (RFC822Z)
// The second INTEGER is a 2-digit year expanded via the RFC 2822 rule.
func handleIntegerMonthIntegerTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	// [0]=INTEGER (day), [1]=MONTH, [2]=INTEGER (2-digit year), [3]=TIME
	h, m, s := MustParseTime(toks[3].Value.(string))
	return &ParsedDateSlots{
		Year:   Expand2DigitYear(toks[2].Value.(int)),
		Month:  int(toks[1].Value.(Month)),
		Day:    toks[0].Value.(int),
		Hour:   h,
		Minute: m,
		Second: s,
		Period: TimePeriod(toks[3].Value.(string)),
	}, nil
}

// handleWeekdayIntegerMonthIntegerTime handles: WEEKDAY INTEGER MONTH INTEGER TIME
// Example: "Monday, 02-Jan-06 15:04:05 MST" (RFC850)
// The weekday is informational and is ignored. Second INTEGER is a 2-digit year.
func handleWeekdayIntegerMonthIntegerTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	// [0]=WEEKDAY (ignored), [1]=INTEGER (day), [2]=MONTH, [3]=INTEGER (2-digit year), [4]=TIME
	h, m, s := MustParseTime(toks[4].Value.(string))
	return &ParsedDateSlots{
		Year:   Expand2DigitYear(toks[3].Value.(int)),
		Month:  int(toks[2].Value.(Month)),
		Day:    toks[1].Value.(int),
		Hour:   h,
		Minute: m,
		Second: s,
		Period: TimePeriod(toks[4].Value.(string)),
	}, nil
}

// handleWeekdayMonthIntegerYear handles: WEEKDAY MONTH INTEGER YEAR
// The weekday is informational and is ignored.
// Example: "Mon Jan 2 2026"
func handleWeekdayMonthIntegerYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	// [0]=WEEKDAY (ignored), [1]=MONTH, [2]=INTEGER (day), [3]=YEAR
	return &ParsedDateSlots{
		Year:   toks[3].Value.(int),
		Month:  int(toks[1].Value.(Month)),
		Day:    toks[2].Value.(int),
		Period: PeriodDay,
	}, nil
}

// handleWeekdayMonthIntegerTimeYear handles: WEEKDAY MONTH INTEGER TIME YEAR
// The weekday is informational and is ignored.
// Examples: "Mon Jan  2 15:04:05 2026" (ANSIC), "Mon Jan  2 15:04:05 MST 2026" (UnixDate),
//
//	"Mon Jan 02 15:04:05 -0700 2026" (RubyDate)
func handleWeekdayMonthIntegerTimeYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	// [0]=WEEKDAY (ignored), [1]=MONTH, [2]=INTEGER (day), [3]=TIME, [4]=YEAR
	h, m, s := MustParseTime(toks[3].Value.(string))
	return &ParsedDateSlots{
		Year:   toks[4].Value.(int),
		Month:  int(toks[1].Value.(Month)),
		Day:    toks[2].Value.(int),
		Hour:   h,
		Minute: m,
		Second: s,
		Period: TimePeriod(toks[3].Value.(string)),
	}, nil
}

// handleYear handles: YEAR
// Example: "2026"
func handleYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	y := toks[0].Value.(int)
	return &ParsedDateSlots{Year: y, Period: PeriodYear}, nil
}

// handleUnixTimestamp handles a bare integer as a Unix timestamp (seconds since
// 1970-01-01 00:00:00 UTC). Only integers with 5+ digits (≥ 10000) are accepted;
// 1–3 digit integers are rejected as ambiguous, and 4-digit integers are already
// routed to handleYear via TokenYear.
func handleUnixTimestamp(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	ts := int64(toks[0].Value.(int))
	if ts < 10000 {
		return nil, ErrUnknownSignature
	}
	return &ParsedDateSlots{UnixTime: ts, Period: PeriodSecond}, nil
}

// prepDelegate strips a leading PREP token and delegates to the base handler.
// Used for PREP YEAR ... signatures that are identical to their YEAR ... counterparts.
func prepDelegate(h Handler) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		withoutPrep := make([]Token, 0, len(tokens))
		skipped := false
		for _, t := range tokens {
			if !skipped && t.Type == TokenPrep {
				skipped = true
				continue
			}
			withoutPrep = append(withoutPrep, t)
		}
		return h(withoutPrep)
	}
}

// ---------------------------------------------------------------------------
// withPrepTime combinator and helpers
// ---------------------------------------------------------------------------

// applyTimeToks merges a trailing time sequence into an existing slots.
// timeToks must be one of: TIME, TIME AMPM, or INTEGER AMPM.
func applyTimeToks(slots *ParsedDateSlots, timeToks []Token) (*ParsedDateSlots, error) {
	if len(timeToks) == 0 {
		return nil, fmt.Errorf("nowandlater: applyTimeToks: no time tokens")
	}
	switch timeToks[0].Type {
	case TokenTime:
		timeVal := timeToks[0].Value.(string)
		h, m, sec := MustParseTime(timeVal)
		if len(timeToks) > 1 && timeToks[1].Type == TokenAMPM {
			h = ApplyAMPM(h, timeToks[1].Value.(AMPM))
		}
		slots.Hour = h
		slots.Minute = m
		slots.Period = TimePeriod(timeVal)
		slots.Second = sec
	case TokenInteger:
		h := timeToks[0].Value.(int)
		if len(timeToks) > 1 && timeToks[1].Type == TokenAMPM {
			h = ApplyAMPM(h, timeToks[1].Value.(AMPM))
		}
		slots.Hour = h
		slots.Period = PeriodHour
	default:
		return nil, fmt.Errorf("nowandlater: applyTimeToks: unexpected token type %v", timeToks[0].Type)
	}
	return slots, nil
}

// withPrepTime wraps a date-only handler to also accept a trailing PREP time suffix.
// It finds the last PREP token, passes everything before it to dateHandler,
// then merges the time extracted from tokens after it.
func withPrepTime(dateHandler Handler) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := FilterFillers(tokens)
		prepIdx := -1
		for i := len(toks) - 1; i >= 0; i-- {
			if toks[i].Type == TokenPrep {
				prepIdx = i
				break
			}
		}
		if prepIdx < 0 {
			return nil, fmt.Errorf("nowandlater: withPrepTime: no PREP token found")
		}
		slots, err := dateHandler(toks[:prepIdx])
		if err != nil {
			return nil, err
		}
		return applyTimeToks(slots, toks[prepIdx+1:])
	}
}

// withTrailingTime wraps a date-only handler to also accept a trailing TIME
// or TIME AMPM suffix (no preposition required). Mirrors withPrepTime.
func withTrailingTime(dateHandler Handler) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := FilterFillers(tokens)
		timeIdx := -1
		for i, t := range toks {
			if t.Type == TokenTime {
				timeIdx = i
				break
			}
		}
		if timeIdx < 0 {
			return nil, fmt.Errorf("nowandlater: withTrailingTime: no TIME token found")
		}
		slots, err := dateHandler(toks[:timeIdx])
		if err != nil {
			return nil, err
		}
		return applyTimeToks(slots, toks[timeIdx:])
	}
}

// handlePrepIntegerAMPM handles: PREP INTEGER AMPM
// Example: "at 3 PM"
func handlePrepIntegerAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	h, err := validateAndApplyAMPM(toks[1].Value.(int), toks[2].Value.(AMPM))
	if err != nil {
		return nil, err
	}
	return &ParsedDateSlots{Hour: h, Period: PeriodHour}, nil
}

// ---------------------------------------------------------------------------
// New handlers added in TODO pass
// ---------------------------------------------------------------------------

// handleTime handles: TIME
// Example: "noon" (preprocessed to "12:00"), "midnight" (preprocessed to "0:00")
func handleTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	return parseTimeSlots(toks[0].Value.(string))
}

// handleTimeAMPM handles: TIME AMPM
// Example: "7.15pm" (dot notation preprocessed to "7:15pm" → TIME + AMPM)
func handleTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	timeVal := toks[0].Value.(string)
	h, m, sec := MustParseTime(timeVal)
	h = ApplyAMPM(h, toks[1].Value.(AMPM))
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: TimePeriod(timeVal)}
	slots.Second = sec
	return slots, nil
}

// parseTimeSlots is a shared helper that builds slots from a TIME token value.
func parseTimeSlots(timeValue string) (*ParsedDateSlots, error) {
	h, m, sec := MustParseTime(timeValue)
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: TimePeriod(timeValue)}
	slots.Second = sec
	return slots, nil
}

// handleDirectionUnit handles: DIRECTION UNIT
// Examples: "next week", "last month", "this year"
// Sets Direction and Anchor; the resolver computes the calendar boundary.
func handleDirectionUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	period := toks[1].Value.(Period)
	return &ParsedDateSlots{
		Direction: toks[0].Value.(Direction),
		Anchor:    period,
		Period:    period,
	}, nil
}

// handleUnitDirection handles: UNIT DIRECTION
// Examples: "mes pasado" (last month), "semana próxima" (next week) — unit-first word order.
func handleUnitDirection(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	period := toks[0].Value.(Period)
	return &ParsedDateSlots{
		Direction: toks[1].Value.(Direction),
		Anchor:    period,
		Period:    period,
	}, nil
}

// handleMonthYear handles: MONTH YEAR
// Examples: "December 2026", "Dec 2026"
func handleMonthYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	y := toks[1].Value.(int)
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[0].Value.(Month)),
		Period: PeriodMonth,
	}, nil
}

// handleYearMonth handles: YEAR MONTH
// Example: "2026 December"
func handleYearMonth(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	y := toks[0].Value.(int)
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[1].Value.(Month)),
		Period: PeriodMonth,
	}, nil
}

// handlePrepUnit handles: PREP UNIT
// Examples: "in a week", "in an hour" ("a"/"an" are FILLER, already dropped)
// Implied quantity is 1.
func handlePrepUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	period := toks[1].Value.(Period)
	secs := periodToSeconds[period]
	return &ParsedDateSlots{DeltaSeconds: new(secs), Period: period}, nil
}

// handleUnitModifier handles: UNIT MODIFIER
// Examples: "a week ago", "an hour hence" ("a"/"an" are FILLER, already dropped)
// Implied quantity is 1.
func handleUnitModifier(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	period := toks[0].Value.(Period)
	secs := periodToSeconds[period]
	sign := int(toks[1].Value.(Modifier))
	return &ParsedDateSlots{DeltaSeconds: new(secs * sign), Period: period}, nil
}

// swapDateOrder assigns a and b to month and day according to order, then
// swaps if the configured assignment produces an impossible month (> 12).
// Any value > 12 can only be a day: "30-01-2016" with MDY → month=30 impossible
// → swap → Jan 30, 2016.
func swapDateOrder(a, b int, order DateOrder) (month, day int) {
	switch order {
	case DMY:
		day, month = a, b
	default: // MDY (zero value — US English default)
		month, day = a, b
	}
	if month > 12 && day <= 12 {
		month, day = day, month
	}
	return month, day
}

// makeIntegerIntegerYearHandler returns a Handler for the INTEGER INTEGER YEAR
// signature that interprets the two leading integers according to order:
//   - MDY: first=month, second=day  ("12/04/2026" → Dec 4)
//   - DMY: first=day,   second=month ("04/12/2026" → Dec 4)
//   - YMD: first=year,  second=month — not meaningful here since YEAR token
//     already captures four-digit years; included for completeness.
func makeIntegerIntegerYearHandler(order DateOrder) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := FilterFillers(tokens)
		a := toks[0].Value.(int)
		b := toks[1].Value.(int)
		y := toks[2].Value.(int)
		m, d := swapDateOrder(a, b, order)
		if m < 1 || m > 12 || d < 1 || d > 31 {
			return nil, fmt.Errorf("nowandlater: invalid date %d/%d/%d", a, b, y)
		}
		return &ParsedDateSlots{Year: y, Month: m, Day: d, Period: PeriodDay}, nil
	}
}

// handleIntegerIntegerYear is the MDY fallback used by the global handlers map.
// Lang-specific dispatch calls makeIntegerIntegerYearHandler(lang.DateOrder) instead.
var handleIntegerIntegerYear = makeIntegerIntegerYearHandler(MDY)

// ---------------------------------------------------------------------------
// "second" as ordinal (day 2) — unit/ordinal conflict resolution
//
// "second" is mapped to TokenUnit(PeriodSecond) in language Words maps so that
// "in 5 seconds" works. This means "march second" produces MONTH UNIT instead
// of MONTH INTEGER. The helpers below replace that unit token with INTEGER(2)
// so the standard month/day handlers can be reused unchanged.
// ---------------------------------------------------------------------------

// replaceSecondUnit returns a copy of tokens with the first TokenUnit(PeriodSecond)
// replaced by TokenInteger(2). Used to treat "second" as the ordinal "2nd".
func replaceSecondUnit(tokens []Token) []Token {
	out := make([]Token, len(tokens))
	copy(out, tokens)
	for i, t := range out {
		if t.Type == TokenUnit && t.Value.(Period) == PeriodSecond {
			out[i] = Token{Type: TokenInteger, Value: 2}
			break
		}
	}
	return out
}

// secondOrdinal wraps a handler so that a TokenUnit(PeriodSecond) in the token
// list is treated as TokenInteger(2) before the base handler runs.
// If no PeriodSecond token is present, the UNIT cannot be treated as an ordinal
// and ErrUnknownSignature is returned.
func secondOrdinal(base Handler) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		replaced := replaceSecondUnit(tokens)
		for _, t := range replaced {
			if t.Type == TokenUnit {
				return nil, ErrUnknownSignature
			}
		}
		return base(replaced)
	}
}

var (
	handleMonthSecondDay     = secondOrdinal(handleMonthDay)
	handleMonthSecondDayYear = secondOrdinal(handleMonthDayYear)
	handleSecondDayMonth     = secondOrdinal(handleDayMonth)
	handleSecondDayMonthYear = secondOrdinal(handleIntegerMonthYear)
)

// makeIntegerAMPMDateFragmentHandler returns a Handler for INTEGER AMPM DATE_FRAGMENT.
// DATE_FRAGMENT arises when a compound date has a 2-digit year (e.g. "20.07.21").
// The raw string is re-parsed applying DateOrder and Expand2DigitYear.
// Example: "1 a.m 20.07.21"
func makeIntegerAMPMDateFragmentHandler(order DateOrder) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := FilterFillers(tokens)
		// [0]=INTEGER (hour), [1]=AMPM, [2]=DATE_FRAGMENT
		hr, err := validateAndApplyAMPM(toks[0].Value.(int), toks[1].Value.(AMPM))
		if err != nil {
			return nil, err
		}
		mo, d, y, err := parseDateFragment(toks[2].Value.(string), order)
		if err != nil {
			return nil, err
		}
		return &ParsedDateSlots{
			Year: y, Month: mo, Day: d,
			Hour: hr, Period: PeriodHour,
		}, nil
	}
}

// makeIntegerAMPMIntegerIntegerYearHandler returns a Handler for INTEGER AMPM INTEGER INTEGER YEAR.
// The leading INTEGER+AMPM is a 12-hour time; the trailing INTEGER INTEGER YEAR is a date
// interpreted according to order (MDY or DMY), identical to makeIntegerIntegerYearHandler.
// Example: "1 a.m 20.07.2021"
func makeIntegerAMPMIntegerIntegerYearHandler(order DateOrder) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := FilterFillers(tokens)
		// [0]=INTEGER (hour), [1]=AMPM, [2]=INTEGER (a), [3]=INTEGER (b), [4]=YEAR
		hr, err := validateAndApplyAMPM(toks[0].Value.(int), toks[1].Value.(AMPM))
		if err != nil {
			return nil, err
		}
		a := toks[2].Value.(int)
		b := toks[3].Value.(int)
		y := toks[4].Value.(int)
		mo, d := swapDateOrder(a, b, order)
		if mo < 1 || mo > 12 || d < 1 || d > 31 {
			return nil, fmt.Errorf("nowandlater: invalid date %d/%d/%d", a, b, y)
		}
		return &ParsedDateSlots{
			Year: y, Month: mo, Day: d,
			Hour: hr, Period: PeriodHour,
		}, nil
	}
}

// handleTimeAMPMMonthIntegerYear handles: TIME AMPM MONTH INTEGER YEAR
// Example: "8:25 a.m. Dec. 12, 2014" — time precedes date
func handleTimeAMPMMonthIntegerYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens)
	// [0]=TIME, [1]=AMPM, [2]=MONTH, [3]=INTEGER (day), [4]=YEAR
	timeVal := toks[0].Value.(string)
	h, min, sec := MustParseTime(timeVal)
	h = ApplyAMPM(h, toks[1].Value.(AMPM))
	slots := &ParsedDateSlots{
		Year:   toks[4].Value.(int),
		Month:  int(toks[2].Value.(Month)),
		Day:    toks[3].Value.(int),
		Hour:   h,
		Minute: min,
		Period: TimePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleDirectionUnitWeekday handles: DIRECTION UNIT WEEKDAY
// Example: 来週の月曜日 → next Monday
// The UNIT (e.g. PeriodWeek from 来週) provides context but is redundant for
// resolution; weekday + direction is sufficient.
func handleDirectionUnitWeekday(tokens []Token) (*ParsedDateSlots, error) {
	toks := FilterFillers(tokens) // [DIRECTION, UNIT, WEEKDAY]
	return &ParsedDateSlots{
		Weekday:   toks[2].Value.(Weekday),
		Direction: toks[0].Value.(Direction),
		Period:    PeriodDay,
	}, nil
}
