package nowandlater

import (
	"fmt"
	"time"
)

// Resolve converts a ParsedDateSlots into a concrete time.Time using now as the
// reference point for all relative expressions.
//
// The location of the returned time matches now.Location(). Callers that need
// a specific timezone should pass a now value in that location.
//
// The slots.Period field describes the granularity of the input ("day", "minute",
// etc.) but does not affect the returned time — resolution is always to the
// second. Callers can use Period to decide how to display or truncate the result.
func Resolve(slots *ParsedDateSlots, now time.Time) (time.Time, error) {
	loc := now.Location()
	if slots.Location != nil {
		loc = slots.Location
	}
	// Rebind now in the target location so every helper that calls now.Location()
	// or now.AddDate(...) automatically uses the right timezone.
	now = now.In(loc)

	switch {

	// --- Unix timestamp: absolute seconds since 1970-01-01 UTC ---
	// Produced by: handleUnixTimestamp.
	case slots.UnixTime != 0:
		return time.Unix(slots.UnixTime, 0).In(loc), nil

	// --- Delta (relative offset), optionally with a time-of-day ---
	// Produced by: handleAnchor, handleRelativeDelta, handlePrepIntegerUnit,
	// handlePrepUnit, handleUnitModifier, handleAnchorPrep* variants.
	case slots.DeltaSeconds != nil:
		base := now.Add(time.Duration(*slots.DeltaSeconds) * time.Second)
		// Apply a time-of-day only when one was explicitly expressed alongside
		// the delta (e.g. "today at 9:30"). Period <= PeriodHour is necessary but
		// not sufficient: delta-unit expressions like "in 2 hours" also set
		// Period=PeriodHour but leave Hour=Minute=0, so check that at least one
		// time component is non-zero.
		if slots.Period <= PeriodHour && (slots.Hour != 0 || slots.Minute != 0) {
			return applyTimePart(base, slots), nil
		}
		return base, nil

	// --- Weekday (with direction), optionally with a time-of-day ---
	// Produced by: handleWeekday, handleDirectionWeekday, and withPrepTime variants.
	case slots.Weekday != 0:
		direction := DirectionNearest
		if slots.Direction != 0 {
			direction = slots.Direction
		}
		base := resolveWeekday(slots.Weekday, direction, now)
		if slots.Period <= PeriodHour {
			return applyTimePart(base, slots), nil
		}
		return base, nil

	// --- Direction + anchor: "next week", "last month", "this year" ---
	// Produced by: handleDirectionUnit.
	case slots.Direction != 0 && slots.Anchor != 0:
		return resolveDirectionAnchor(slots.Direction, slots.Anchor, now)

	// --- Full absolute date (year + month + day), optionally with time ---
	// Produced by: handleMonthDayYear, handleYearIntegerInteger,
	// handleYearMonthInteger, handleIntegerMonthYear, handleIntegerIntegerYear,
	// and all combined date+time handlers.
	case slots.Year != 0 && slots.Month != 0 && slots.Day != 0:
		base := time.Date(slots.Year, time.Month(slots.Month), slots.Day, 0, 0, 0, 0, loc)
		if slots.Period <= PeriodHour {
			return applyTimePart(base, slots), nil
		}
		return base, nil

	// --- Year + month (no day): "December 2026" ---
	// Produced by: handleMonthYear, handleYearMonth.
	case slots.Year != 0 && slots.Month != 0:
		return time.Date(slots.Year, time.Month(slots.Month), 1, 0, 0, 0, 0, loc), nil

	// --- Month + day (no year): "March 5", "April 10 at 3 PM" ---
	// Produced by: handleMonthDay, handleDayMonth, and withPrepTime variants.
	case slots.Month != 0 && slots.Day != 0:
		return resolveMonthDay(slots.Month, slots.Day, slots, now)

	// --- Year only: "2026" ---
	// Produced by: handleYear.
	case slots.Year != 0:
		return time.Date(slots.Year, time.January, 1, 0, 0, 0, 0, loc), nil

	// --- Time only: "at 9:30", "3pm", "noon" ---
	// Produced by: handleTime, handleTimeAMPM, handleIntegerAMPM,
	// handlePrepTime, handlePrepTimeAMPM, handlePrepIntegerAMPM.
	case slots.Period <= PeriodHour:
		return applyTimePart(startOfDay(now), slots), nil

	default:
		return time.Time{}, fmt.Errorf("nowandlater: resolve: cannot determine a time from the given slots")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// startOfDay returns t with the clock reset to 00:00:00.000 in t's location.
func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// applyTimePart returns base with its clock replaced by slots.Hour / Minute / Second.
func applyTimePart(base time.Time, slots *ParsedDateSlots) time.Time {
	y, mo, d := base.Date()
	return time.Date(y, mo, d, slots.Hour, slots.Minute, slots.Second, 0, base.Location())
}

// resolveWeekday finds the date of targetWd relative to now according to
// direction ("future", "past", or "nearest").
// Returns the start of that day (00:00:00).
func resolveWeekday(targetWd Weekday, direction Direction, now time.Time) time.Time {
	// Convert Go weekday (Sunday=0, Monday=1, …) to 0-based Monday system.
	todayWd := (int(now.Weekday()) + 6) % 7
	// Convert typed Weekday (Monday=1…Sunday=7) to 0-based Monday system.
	wd := int(targetWd) - 1

	var days int
	switch direction {
	case DirectionFuture:
		// At least 1 day ahead; if today matches, skip to next week.
		days = ((wd-todayWd-1+7)%7 + 1)
	case DirectionPast:
		// At least 1 day back; if today matches, go to last week.
		days = -((todayWd-wd-1+7)%7 + 1)
	default: // DirectionNearest
		fwd := (wd - todayWd + 7) % 7
		if fwd == 0 {
			// Today is the target weekday.
			return startOfDay(now)
		}
		bwd := 7 - fwd
		if bwd < fwd {
			days = -bwd
		} else {
			days = fwd
		}
	}
	return startOfDay(now.AddDate(0, 0, days))
}

// resolveDirectionAnchor resolves "next week", "last month", "this year", etc.
// Returns the start of the target period:
//   - week  → Monday 00:00:00
//   - month → 1st of month 00:00:00
//   - year  → Jan 1 00:00:00
//   - day   → start of day 00:00:00
func resolveDirectionAnchor(direction Direction, anchor Period, now time.Time) (time.Time, error) {
	loc := now.Location()
	y, m, _ := now.Date()

	// Monday of the current week (days back to Monday).
	currentMonday := func() time.Time {
		todayWd := (int(now.Weekday()) + 6) % 7
		return startOfDay(now.AddDate(0, 0, -todayWd))
	}

	switch anchor {
	case PeriodWeek:
		monday := currentMonday()
		switch direction {
		case DirectionFuture:
			return monday.AddDate(0, 0, 7), nil
		case DirectionPast:
			return monday.AddDate(0, 0, -7), nil
		default: // nearest
			return monday, nil
		}
	case PeriodMonth:
		switch direction {
		case DirectionFuture:
			return time.Date(y, m+1, 1, 0, 0, 0, 0, loc), nil
		case DirectionPast:
			return time.Date(y, m-1, 1, 0, 0, 0, 0, loc), nil
		default:
			return time.Date(y, m, 1, 0, 0, 0, 0, loc), nil
		}
	case PeriodYear:
		switch direction {
		case DirectionFuture:
			return time.Date(y+1, time.January, 1, 0, 0, 0, 0, loc), nil
		case DirectionPast:
			return time.Date(y-1, time.January, 1, 0, 0, 0, 0, loc), nil
		default:
			return time.Date(y, time.January, 1, 0, 0, 0, 0, loc), nil
		}
	case PeriodDay:
		switch direction {
		case DirectionFuture:
			return startOfDay(now.AddDate(0, 0, 1)), nil
		case DirectionPast:
			return startOfDay(now.AddDate(0, 0, -1)), nil
		default:
			return startOfDay(now), nil
		}
	default:
		return time.Time{}, fmt.Errorf("nowandlater: resolve: unsupported anchor %v", anchor)
	}
}

// resolveMonthDay resolves a month+day expression with no year, inferring the
// year from now: uses the current year if the resulting time is after now,
// otherwise advances to next year.
func resolveMonthDay(month, day int, slots *ParsedDateSlots, now time.Time) (time.Time, error) {
	loc := now.Location()
	h, m, s := 0, 0, 0
	if slots.Period <= PeriodHour {
		h, m, s = slots.Hour, slots.Minute, slots.Second
	}
	y := now.Year()
	t := time.Date(y, time.Month(month), day, h, m, s, 0, loc)
	if !t.After(now) {
		t = time.Date(y+1, time.Month(month), day, h, m, s, 0, loc)
	}
	return t, nil
}

// ---------------------------------------------------------------------------
// Interval helpers
// ---------------------------------------------------------------------------

// startOfPeriod truncates t to the start of the given period in t's location.
// For PeriodWeek, the week starts on Monday.
func startOfPeriod(t time.Time, period Period) time.Time {
	loc := t.Location()
	y, m, d := t.Date()
	h, min, s := t.Clock()
	switch period {
	case PeriodSecond:
		return time.Date(y, m, d, h, min, s, 0, loc)
	case PeriodMinute:
		return time.Date(y, m, d, h, min, 0, 0, loc)
	case PeriodHour:
		return time.Date(y, m, d, h, 0, 0, 0, loc)
	case PeriodDay, PeriodFortnight:
		return startOfDay(t)
	case PeriodWeek:
		// Monday of the week containing t (Go: Sun=0, Mon=1 → 0-based Mon system).
		daysToMonday := (int(t.Weekday()) + 6) % 7
		return startOfDay(t.AddDate(0, 0, -daysToMonday))
	case PeriodMonth:
		return time.Date(y, m, 1, 0, 0, 0, 0, loc)
	case PeriodYear:
		return time.Date(y, time.January, 1, 0, 0, 0, 0, loc)
	default:
		return t
	}
}

// EndOf returns the exclusive end of the period starting at start, forming a
// half-open interval [start, EndOf(start, period)).
//
// start should be calendar-aligned to the period boundary (e.g. midnight for
// PeriodDay, first of month 00:00 for PeriodMonth). For unrecognised periods,
// start is returned unchanged.
func EndOf(start time.Time, period Period) time.Time {
	loc := start.Location()
	y, m, _ := start.Date()
	switch period {
	case PeriodSecond:
		return start.Add(time.Second)
	case PeriodMinute:
		return start.Add(time.Minute)
	case PeriodHour:
		return start.Add(time.Hour)
	case PeriodDay:
		return start.AddDate(0, 0, 1)
	case PeriodFortnight:
		return start.AddDate(0, 0, 14)
	case PeriodWeek:
		return start.AddDate(0, 0, 7)
	case PeriodMonth:
		return time.Date(y, m+1, 1, 0, 0, 0, 0, loc)
	case PeriodYear:
		return time.Date(y+1, time.January, 1, 0, 0, 0, 0, loc)
	default:
		return start
	}
}

// ResolveInterval resolves slots to a half-open calendar interval [start, end).
//
// start is always calendar-aligned to the period boundary (midnight for
// day/week expressions, 1st of month for month, Jan 1 for year, truncated to
// the hour/minute boundary for sub-day expressions). For delta-based
// expressions ("tomorrow", "in 3 days"), start may therefore differ from what
// Resolve returns — Resolve returns now+delta, while ResolveInterval aligns
// start to the calendar boundary of the target period.
//
// end is the first moment after the period: end == EndOf(start, slots.Period).
func ResolveInterval(slots *ParsedDateSlots, now time.Time) (start, end time.Time, err error) {
	pt, err := Resolve(slots, now)
	if err != nil {
		return
	}
	start = startOfPeriod(pt, slots.Period)
	end = EndOf(start, slots.Period)
	return
}
