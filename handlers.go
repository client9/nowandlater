package nowandlater

import (
	"fmt"
	"math"
)

// ---------------------------------------------------------------------------
// Handlers
// Each handler receives the full token list (including FILLER) and returns
// a populated ParsedDateSlots. All handlers call filterFillers first so that
// positional indices align exactly with the signature string.
// ---------------------------------------------------------------------------

// handleAnchor handles: ANCHOR
// Examples: "today", "tomorrow", "yesterday", "now"
func handleAnchor(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	anchor := toks[0].Value.(Anchor)
	period := PeriodDay
	if anchor == AnchorNow {
		period = PeriodSecond
	}
	return &ParsedDateSlots{DeltaSeconds: new(anchorToSeconds[anchor]), Period: period}, nil
}

// handleWeekday handles: WEEKDAY
// Example: "Monday"
func handleWeekday(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	return &ParsedDateSlots{
		Weekday:   toks[0].Value.(Weekday),
		Direction: DirectionNearest,
		Period:    PeriodDay,
	}, nil
}

// handleDirectionWeekday handles: DIRECTION WEEKDAY
// Examples: "next Monday", "last Friday"
func handleDirectionWeekday(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	return &ParsedDateSlots{
		Weekday:   toks[1].Value.(Weekday),
		Direction: toks[0].Value.(Direction),
		Period:    PeriodDay,
	}, nil
}

// handleDirectionWeekdayTime handles: DIRECTION WEEKDAY PREP TIME
// Example: "next Monday at 9:30" or "next Monday at 9:30:00"
func handleDirectionWeekdayTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	timeVal := toks[3].Value.(string)
	h, m, sec := mustParseTime(timeVal)
	slots := &ParsedDateSlots{
		Weekday:   toks[1].Value.(Weekday),
		Direction: toks[0].Value.(Direction),
		Hour:      h,
		Minute:    m,
		Period:    timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleDirectionWeekdayTimeAMPM handles: DIRECTION WEEKDAY PREP TIME AMPM
// Example: "next Monday at 9:30 AM"
func handleDirectionWeekdayTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	timeVal := toks[3].Value.(string)
	h, m, sec := mustParseTime(timeVal)
	h = applyAMPM(h, toks[4].Value.(AMPM))
	slots := &ParsedDateSlots{
		Weekday:   toks[1].Value.(Weekday),
		Direction: toks[0].Value.(Direction),
		Hour:      h,
		Minute:    m,
		Period:    timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleIntegerAMPM handles: INTEGER AMPM
// Examples: "3pm", "11am"
func handleIntegerAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	h := toks[0].Value.(int)
	if h < 1 || h > 12 {
		return nil, fmt.Errorf("nowandlater: hour %d out of range for 12-hour clock", h)
	}
	h = applyAMPM(h, toks[1].Value.(AMPM))
	return &ParsedDateSlots{Hour: h, Period: PeriodHour}, nil
}

// handlePrepTime handles: PREP TIME
// Example: "at 09:30", "at 09:30:45"
func handlePrepTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	timeVal := toks[1].Value.(string)
	h, m, sec := mustParseTime(timeVal)
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: timePeriod(timeVal)}
	slots.Second = sec
	return slots, nil
}

// handlePrepTimeAMPM handles: PREP TIME AMPM
// Example: "at 9:30 AM"
func handlePrepTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	timeVal := toks[1].Value.(string)
	h, m, sec := mustParseTime(timeVal)
	h = applyAMPM(h, toks[2].Value.(AMPM))
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: timePeriod(timeVal)}
	slots.Second = sec
	return slots, nil
}

// handleMonthDay handles: MONTH INTEGER
// Example: "March 5"
func handleMonthDay(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	d := toks[1].Value.(int)
	return &ParsedDateSlots{
		Month:  int(toks[0].Value.(Month)),
		Day:    d,
		Period: PeriodDay,
	}, nil
}

// handleDayMonth handles: INTEGER MONTH
// Example: "3rd of January" (fillers already dropped in signature)
func handleDayMonth(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	d := toks[0].Value.(int)
	return &ParsedDateSlots{
		Month:  int(toks[1].Value.(Month)),
		Day:    d,
		Period: PeriodDay,
	}, nil
}

// handleMonthDayYear handles: MONTH INTEGER YEAR
// Example: "Dec 3rd 2026"
func handleMonthDayYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	sign := int(toks[0].Value.(Modifier))
	secs, period := sumTwoUnits(toks, 1)
	return &ParsedDateSlots{DeltaSeconds: new(secs * sign), Period: period}, nil
}

// handleWeekdayDirection handles: WEEKDAY DIRECTION
// Example: "lunes próximo" (Spanish: next Monday), "lundi prochain" (French: next Monday)
// This is the weekday-first word order used by many non-English languages.
func handleWeekdayDirection(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	n := toks[0].Value.(int)
	period := toks[1].Value.(Period)
	secs := periodToSeconds[period]
	sign := int(toks[2].Value.(Modifier))
	aDelta := anchorToSeconds[toks[3].Value.(Anchor)]
	return &ParsedDateSlots{
		DeltaSeconds: new(n*secs*sign + aDelta),
		Period:       period,
	}, nil
}

// handleIntegerUnit handles: INTEGER UNIT
// Examples: "4 hours", "3 días" — bare count with no preposition or modifier.
// Implied direction is future (positive delta), matching GNU date behaviour.
func handleIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	n := toks[0].Value.(int)
	period := toks[1].Value.(Period)
	secs := periodToSeconds[period]
	return &ParsedDateSlots{
		DeltaSeconds: new(n * secs),
		Period:       period,
	}, nil
}

// decimalUnitSeconds converts a fractional count and period to whole seconds.
func decimalUnitSeconds(n float64, period Period) int {
	return int(math.Round(n * float64(periodToSeconds[period])))
}

// handleDecimalUnit handles: DECIMAL UNIT
// Examples: "3.5 days", "1.5 horas" — implied future.
func handleDecimalUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	n := toks[0].Value.(float64)
	period := toks[1].Value.(Period)
	return &ParsedDateSlots{DeltaSeconds: new(decimalUnitSeconds(n, period)), Period: period}, nil
}

// handlePrepDecimalUnit handles: PREP DECIMAL UNIT
// Examples: "in 1.5 hours", "en 2.5 días"
func handlePrepDecimalUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	n := toks[1].Value.(float64)
	period := toks[2].Value.(Period)
	return &ParsedDateSlots{DeltaSeconds: new(decimalUnitSeconds(n, period)), Period: period}, nil
}

// handleDecimalUnitModifier handles: DECIMAL UNIT MODIFIER
// Examples: "1.5 hours ago", "2.5 días después"
func handleDecimalUnitModifier(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	n := toks[0].Value.(float64)
	period := toks[1].Value.(Period)
	sign := int(toks[2].Value.(Modifier))
	return &ParsedDateSlots{DeltaSeconds: new(decimalUnitSeconds(n, period) * sign), Period: period}, nil
}

// handleModifierDecimalUnit handles: MODIFIER DECIMAL UNIT
// Examples: "hace 1.5 días" (Spanish modifier-first word order)
func handleModifierDecimalUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	sign := int(toks[0].Value.(Modifier))
	n := toks[1].Value.(float64)
	period := toks[2].Value.(Period)
	return &ParsedDateSlots{DeltaSeconds: new(decimalUnitSeconds(n, period) * sign), Period: period}, nil
}

// handlePrepIntegerUnit handles: PREP INTEGER UNIT
// Example: "in 2 days"
func handlePrepIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	secs, period := sumTwoUnits(toks, 1)
	return &ParsedDateSlots{DeltaSeconds: new(secs), Period: period}, nil
}

// handleIntegerUnitIntegerUnit handles: INTEGER UNIT INTEGER UNIT
// Example: "1 hour and 10 minutes" — implied future.
func handleIntegerUnitIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	secs, period := sumTwoUnits(toks, 0)
	return &ParsedDateSlots{DeltaSeconds: new(secs), Period: period}, nil
}

// handleIntegerUnitIntegerUnitModifier handles: INTEGER UNIT INTEGER UNIT MODIFIER
// Example: "1 hour and 10 minutes ago"
func handleIntegerUnitIntegerUnitModifier(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	secs, period := sumTwoUnits(toks, 0)
	sign := int(toks[4].Value.(Modifier))
	return &ParsedDateSlots{DeltaSeconds: new(secs * sign), Period: period}, nil
}

// handlePrepDirectionIntegerUnit handles: PREP DIRECTION INTEGER UNIT
// Example: "in next 2 days"
func handlePrepDirectionIntegerUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	y := toks[0].Value.(int)
	m := toks[1].Value.(int)
	d := toks[2].Value.(int)
	return &ParsedDateSlots{Year: y, Month: m, Day: d, Period: PeriodDay}, nil
}

// handleYearMonthInteger handles: YEAR MONTH INTEGER
// Example: "2026-dec-04"
func handleYearMonthInteger(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	d := toks[0].Value.(int)
	y := toks[2].Value.(int)
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[1].Value.(Month)),
		Day:    d,
		Period: PeriodDay,
	}, nil
}

// handleWeekdayIntegerMonthYear handles: WEEKDAY INTEGER MONTH YEAR
// The weekday is informational (derivable from the date) and is ignored.
// Example: "Mon, 02 Jan 2006" (RFC 2822 date portion)
func handleWeekdayIntegerMonthYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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

// handleWeekdayIntegerMonthYearTime handles: WEEKDAY INTEGER MONTH YEAR TIME
// Example: "Mon, 02 Jan 2006 15:04:05" (RFC 2822 without timezone)
func handleWeekdayIntegerMonthYearTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	// [0]=WEEKDAY (ignored), [1]=INTEGER, [2]=MONTH, [3]=YEAR, [4]=TIME
	d := toks[1].Value.(int)
	y := toks[3].Value.(int)
	h, m, s := mustParseTime(toks[4].Value.(string))
	return &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[2].Value.(Month)),
		Day:    d,
		Hour:   h,
		Minute: m,
		Second: s,
		Period: PeriodSecond,
	}, nil
}

// handleIntegerMonthIntegerTime handles: INTEGER MONTH INTEGER TIME
// Examples: "02 Jan 06 15:04 MST" (RFC822), "22 Mar 26 10:04 -0700" (RFC822Z)
// The second INTEGER is a 2-digit year expanded via the RFC 2822 rule.
func handleIntegerMonthIntegerTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	// [0]=INTEGER (day), [1]=MONTH, [2]=INTEGER (2-digit year), [3]=TIME
	h, m, s := mustParseTime(toks[3].Value.(string))
	return &ParsedDateSlots{
		Year:   expand2DigitYear(toks[2].Value.(int)),
		Month:  int(toks[1].Value.(Month)),
		Day:    toks[0].Value.(int),
		Hour:   h,
		Minute: m,
		Second: s,
		Period: timePeriod(toks[3].Value.(string)),
	}, nil
}

// handleWeekdayIntegerMonthIntegerTime handles: WEEKDAY INTEGER MONTH INTEGER TIME
// Example: "Monday, 02-Jan-06 15:04:05 MST" (RFC850)
// The weekday is informational and is ignored. Second INTEGER is a 2-digit year.
func handleWeekdayIntegerMonthIntegerTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	// [0]=WEEKDAY (ignored), [1]=INTEGER (day), [2]=MONTH, [3]=INTEGER (2-digit year), [4]=TIME
	h, m, s := mustParseTime(toks[4].Value.(string))
	return &ParsedDateSlots{
		Year:   expand2DigitYear(toks[3].Value.(int)),
		Month:  int(toks[2].Value.(Month)),
		Day:    toks[1].Value.(int),
		Hour:   h,
		Minute: m,
		Second: s,
		Period: timePeriod(toks[4].Value.(string)),
	}, nil
}

// handleWeekdayMonthIntegerYear handles: WEEKDAY MONTH INTEGER YEAR
// The weekday is informational and is ignored.
// Example: "Mon Jan 2 2026"
func handleWeekdayMonthIntegerYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	// [0]=WEEKDAY (ignored), [1]=MONTH, [2]=INTEGER (day), [3]=TIME, [4]=YEAR
	h, m, s := mustParseTime(toks[3].Value.(string))
	return &ParsedDateSlots{
		Year:   toks[4].Value.(int),
		Month:  int(toks[1].Value.(Month)),
		Day:    toks[2].Value.(int),
		Hour:   h,
		Minute: m,
		Second: s,
		Period: timePeriod(toks[3].Value.(string)),
	}, nil
}

// handleYear handles: YEAR
// Example: "2026"
func handleYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	y := toks[0].Value.(int)
	return &ParsedDateSlots{Year: y, Period: PeriodYear}, nil
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
		h, m, sec := mustParseTime(timeVal)
		if len(timeToks) > 1 && timeToks[1].Type == TokenAMPM {
			h = applyAMPM(h, timeToks[1].Value.(AMPM))
		}
		slots.Hour = h
		slots.Minute = m
		slots.Period = timePeriod(timeVal)
		slots.Second = sec
	case TokenInteger:
		h := timeToks[0].Value.(int)
		if len(timeToks) > 1 && timeToks[1].Type == TokenAMPM {
			h = applyAMPM(h, timeToks[1].Value.(AMPM))
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
		toks := filterFillers(tokens)
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

// handlePrepIntegerAMPM handles: PREP INTEGER AMPM
// Example: "at 3 PM"
func handlePrepIntegerAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	h := toks[1].Value.(int)
	if h < 1 || h > 12 {
		return nil, fmt.Errorf("nowandlater: hour %d out of range for 12-hour clock", h)
	}
	h = applyAMPM(h, toks[2].Value.(AMPM))
	return &ParsedDateSlots{Hour: h, Period: PeriodHour}, nil
}

// ---------------------------------------------------------------------------
// New handlers added in TODO pass
// ---------------------------------------------------------------------------

// handleTime handles: TIME
// Example: "noon" (preprocessed to "12:00"), "midnight" (preprocessed to "0:00")
func handleTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	return parseTimeSlots(toks[0].Value.(string))
}

// handleTimeAMPM handles: TIME AMPM
// Example: "7.15pm" (dot notation preprocessed to "7:15pm" → TIME + AMPM)
func handleTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	timeVal := toks[0].Value.(string)
	h, m, sec := mustParseTime(timeVal)
	h = applyAMPM(h, toks[1].Value.(AMPM))
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: timePeriod(timeVal)}
	slots.Second = sec
	return slots, nil
}

// parseTimeSlots is a shared helper that builds slots from a TIME token value.
func parseTimeSlots(timeValue string) (*ParsedDateSlots, error) {
	h, m, sec := mustParseTime(timeValue)
	slots := &ParsedDateSlots{Hour: h, Minute: m, Period: timePeriod(timeValue)}
	slots.Second = sec
	return slots, nil
}

// handleDirectionUnit handles: DIRECTION UNIT
// Examples: "next week", "last month", "this year"
// Sets Direction and Anchor; the resolver computes the calendar boundary.
func handleDirectionUnit(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	period := toks[0].Value.(Period)
	return &ParsedDateSlots{
		Direction: toks[1].Value.(Direction),
		Anchor:    period,
		Period:    period,
	}, nil
}

// handleAnchorPrepTime handles: ANCHOR PREP TIME
// Examples: "today at 9:30", "tomorrow at 09:30:00"
func handleAnchorPrepTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	aDelta := anchorToSeconds[toks[0].Value.(Anchor)]
	timeVal := toks[2].Value.(string)
	h, m, sec := mustParseTime(timeVal)
	slots := &ParsedDateSlots{
		DeltaSeconds: new(aDelta),
		Hour:         h,
		Minute:       m,
		Period:       timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleAnchorPrepTimeAMPM handles: ANCHOR PREP TIME AMPM
// Examples: "today at 9:30 AM", "tomorrow at 11:00 PM"
func handleAnchorPrepTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	aDelta := anchorToSeconds[toks[0].Value.(Anchor)]
	timeVal := toks[2].Value.(string)
	h, m, sec := mustParseTime(timeVal)
	h = applyAMPM(h, toks[3].Value.(AMPM))
	slots := &ParsedDateSlots{
		DeltaSeconds: new(aDelta),
		Hour:         h,
		Minute:       m,
		Period:       timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleAnchorPrepIntegerAMPM handles: ANCHOR PREP INTEGER AMPM
// Examples: "today at 3pm", "tomorrow at 11am"
func handleAnchorPrepIntegerAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	aDelta := anchorToSeconds[toks[0].Value.(Anchor)]
	h := toks[2].Value.(int)
	h = applyAMPM(h, toks[3].Value.(AMPM))
	return &ParsedDateSlots{
		DeltaSeconds: new(aDelta),
		Hour:         h,
		Period:       PeriodHour,
	}, nil
}

// handleAnchorPrepInteger handles: ANCHOR PREP INTEGER
// Example: "today at 3" — hour is ambiguous (no AM/PM); stored as-is.
func handleAnchorPrepInteger(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	aDelta := anchorToSeconds[toks[0].Value.(Anchor)]
	h := toks[2].Value.(int)
	return &ParsedDateSlots{
		DeltaSeconds: new(aDelta),
		Hour:         h,
		Period:       PeriodHour,
	}, nil
}

// handleDirectionWeekdayPrepInteger handles: DIRECTION WEEKDAY PREP INTEGER
// Example: "next Tuesday at 3" — hour is ambiguous (no AM/PM).
func handleDirectionWeekdayPrepInteger(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	h := toks[3].Value.(int)
	return &ParsedDateSlots{
		Weekday:   toks[1].Value.(Weekday),
		Direction: toks[0].Value.(Direction),
		Hour:      h,
		Period:    PeriodHour,
	}, nil
}

// handleMonthYear handles: MONTH YEAR
// Examples: "December 2026", "Dec 2026"
func handleMonthYear(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
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
	toks := filterFillers(tokens)
	period := toks[1].Value.(Period)
	secs := periodToSeconds[period]
	return &ParsedDateSlots{DeltaSeconds: new(secs), Period: period}, nil
}

// handleUnitModifier handles: UNIT MODIFIER
// Examples: "a week ago", "an hour hence" ("a"/"an" are FILLER, already dropped)
// Implied quantity is 1.
func handleUnitModifier(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	period := toks[0].Value.(Period)
	secs := periodToSeconds[period]
	sign := int(toks[1].Value.(Modifier))
	return &ParsedDateSlots{DeltaSeconds: new(secs * sign), Period: period}, nil
}

// makeIntegerIntegerYearHandler returns a Handler for the INTEGER INTEGER YEAR
// signature that interprets the two leading integers according to order:
//   - MDY: first=month, second=day  ("12/04/2026" → Dec 4)
//   - DMY: first=day,   second=month ("04/12/2026" → Dec 4)
//   - YMD: first=year,  second=month — not meaningful here since YEAR token
//     already captures four-digit years; included for completeness.
func makeIntegerIntegerYearHandler(order DateOrder) Handler {
	return func(tokens []Token) (*ParsedDateSlots, error) {
		toks := filterFillers(tokens)
		a := toks[0].Value.(int)
		b := toks[1].Value.(int)
		y := toks[2].Value.(int)
		var m, d int
		switch order {
		case DMY:
			d, m = a, b
		default: // MDY (zero value — US English default)
			m, d = a, b
		}
		// If the configured order produces an impossible month but the other
		// assignment is valid, swap. Any value > 12 can only be a day, so the
		// result is unambiguous: "30-01-2016" with MDY → month=30 impossible
		// → swap → Jan 30, 2016.
		if m > 12 && d <= 12 {
			m, d = d, m
		}
		if m < 1 || m > 12 || d < 1 || d > 31 {
			return nil, fmt.Errorf("nowandlater: invalid date %d/%d/%d", a, b, y)
		}
		return &ParsedDateSlots{Year: y, Month: m, Day: d, Period: PeriodDay}, nil
	}
}

// handleIntegerIntegerYear is the MDY fallback used by the global handlers map.
// Lang-specific dispatch calls makeIntegerIntegerYearHandler(lang.DateOrder) instead.
var handleIntegerIntegerYear = makeIntegerIntegerYearHandler(MDY)

// handleYearIntegerIntegerTime handles: YEAR INTEGER INTEGER TIME
// Examples: "2026-12-04 09:30", "2026-12-04T09:30:00" (T preprocessed to space)
func handleYearIntegerIntegerTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	y := toks[0].Value.(int)
	m := toks[1].Value.(int)
	d := toks[2].Value.(int)
	timeVal := toks[3].Value.(string)
	h, min, sec := mustParseTime(timeVal)
	slots := &ParsedDateSlots{
		Year:   y,
		Month:  m,
		Day:    d,
		Hour:   h,
		Minute: min,
		Period: timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleYearIntegerIntegerTimeAMPM handles: YEAR INTEGER INTEGER TIME AMPM
// Example: "2026-12-04 9:30 AM"
func handleYearIntegerIntegerTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	y := toks[0].Value.(int)
	m := toks[1].Value.(int)
	d := toks[2].Value.(int)
	timeVal := toks[3].Value.(string)
	h, min, sec := mustParseTime(timeVal)
	h = applyAMPM(h, toks[4].Value.(AMPM))
	slots := &ParsedDateSlots{
		Year:   y,
		Month:  m,
		Day:    d,
		Hour:   h,
		Minute: min,
		Period: timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleYearMonthIntegerTime handles: YEAR MONTH INTEGER TIME
// Example: "2026-dec-04 09:30"
func handleYearMonthIntegerTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	y := toks[0].Value.(int)
	d := toks[2].Value.(int)
	timeVal := toks[3].Value.(string)
	h, min, sec := mustParseTime(timeVal)
	slots := &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[1].Value.(Month)),
		Day:    d,
		Hour:   h,
		Minute: min,
		Period: timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleYearMonthIntegerTimeAMPM handles: YEAR MONTH INTEGER TIME AMPM
// Example: "2026-dec-04 9:30 AM"
func handleYearMonthIntegerTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	y := toks[0].Value.(int)
	d := toks[2].Value.(int)
	timeVal := toks[3].Value.(string)
	h, min, sec := mustParseTime(timeVal)
	h = applyAMPM(h, toks[4].Value.(AMPM))
	slots := &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[1].Value.(Month)),
		Day:    d,
		Hour:   h,
		Minute: min,
		Period: timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleMonthIntegerYearTime handles: MONTH INTEGER YEAR TIME
// Example: "Dec 3 2026 09:30"
func handleMonthIntegerYearTime(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	d := toks[1].Value.(int)
	y := toks[2].Value.(int)
	timeVal := toks[3].Value.(string)
	h, min, sec := mustParseTime(timeVal)
	slots := &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[0].Value.(Month)),
		Day:    d,
		Hour:   h,
		Minute: min,
		Period: timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}

// handleMonthIntegerYearTimeAMPM handles: MONTH INTEGER YEAR TIME AMPM
// Example: "Dec 3 2026 9:30 AM"
func handleMonthIntegerYearTimeAMPM(tokens []Token) (*ParsedDateSlots, error) {
	toks := filterFillers(tokens)
	d := toks[1].Value.(int)
	y := toks[2].Value.(int)
	timeVal := toks[3].Value.(string)
	h, min, sec := mustParseTime(timeVal)
	h = applyAMPM(h, toks[4].Value.(AMPM))
	slots := &ParsedDateSlots{
		Year:   y,
		Month:  int(toks[0].Value.(Month)),
		Day:    d,
		Hour:   h,
		Minute: min,
		Period: timePeriod(timeVal),
	}
	slots.Second = sec
	return slots, nil
}
