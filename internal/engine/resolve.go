package engine

import (
	"fmt"
	"time"
)

// ResolvePolicy controls how underspecified forms are turned into concrete
// dates and times.
type ResolvePolicy struct {
	ImplicitDurationDirection Direction
	CalendarDirection         Direction
	MonthDayDirection         Direction
	RejectAmbiguous           bool
	WeekStartSunday           bool
}

func legacyResolvePolicy() ResolvePolicy {
	return ResolvePolicy{
		ImplicitDurationDirection: DirectionFuture,
		CalendarDirection:         DirectionNearest,
		MonthDayDirection:         DirectionFuture,
	}
}

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
	return ResolveWithPolicy(slots, now, legacyResolvePolicy())
}

// ResolveWithPolicy converts a ParsedDateSlots into a concrete time.Time using
// now as the reference point and policy to resolve underspecified forms.
func ResolveWithPolicy(slots *ParsedDateSlots, now time.Time, policy ResolvePolicy) (time.Time, error) {
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
		if slots.AmbiguousForm == AmbiguousImplicitDuration {
			if policy.RejectAmbiguous {
				return time.Time{}, fmt.Errorf("%w: relative duration %q needs an explicit modifier", ErrAmbiguous, describeSlots(slots))
			}
			base := now.Add(time.Duration(absInt(*slots.DeltaSeconds)*directionSign(policyDurationDirection(policy))) * time.Second)
			if slots.Period <= PeriodHour && (slots.Hour != 0 || slots.Minute != 0) {
				return applyTimePart(base, slots), nil
			}
			return base, nil
		}
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
		direction := slots.Direction
		if slots.AmbiguousForm == AmbiguousBareWeekday {
			if policy.RejectAmbiguous {
				return time.Time{}, fmt.Errorf("%w: weekday %q needs an explicit direction", ErrAmbiguous, describeSlots(slots))
			}
			direction = policy.CalendarDirection
		} else if direction == 0 {
			direction = DirectionNearest
		}
		base := ResolveWeekday(slots.Weekday, direction, now)
		if slots.Period <= PeriodHour {
			return applyTimePart(base, slots), nil
		}
		return base, nil

	// --- Direction + anchor: "next week", "last month", "this year" ---
	// Produced by: handleDirectionUnit.
	case slots.Direction != 0 && slots.Anchor != 0:
		return resolveDirectionAnchor(slots.Direction, slots.Anchor, now, policy.WeekStartSunday)

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
		if slots.AmbiguousForm == AmbiguousMonthDay {
			if policy.RejectAmbiguous {
				return time.Time{}, fmt.Errorf("%w: month/day %q needs an explicit year", ErrAmbiguous, describeSlots(slots))
			}
			return resolveMonthDay(slots.Month, slots.Day, slots, now, policy.MonthDayDirection)
		}
		return resolveMonthDay(slots.Month, slots.Day, slots, now, DirectionFuture)

	// --- Month only: "October" ---
	case slots.Month != 0:
		if slots.AmbiguousForm == AmbiguousBareMonth {
			if policy.RejectAmbiguous {
				return time.Time{}, fmt.Errorf("%w: month %q needs an explicit year", ErrAmbiguous, describeSlots(slots))
			}
			return resolveMonthOnly(slots.Month, now, policy.CalendarDirection), nil
		}
		return resolveMonthOnly(slots.Month, now, DirectionFuture), nil

	// --- Year only: "2026" ---
	// Produced by: handleYear.
	case slots.Year != 0:
		return time.Date(slots.Year, time.January, 1, 0, 0, 0, 0, loc), nil

	// --- Time only: "at 9:30", "3pm", "noon" ---
	// Produced by: handleTime, handleTimeAMPM, handleIntegerAMPM,
	// handlePrepTime, handlePrepTimeAMPM, handlePrepIntegerAMPM.
	case slots.Period <= PeriodHour:
		return applyTimePart(StartOfDay(now), slots), nil

	default:
		return time.Time{}, fmt.Errorf("nowandlater: resolve: cannot determine a time from the given slots")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// StartOfDay returns t with the clock reset to 00:00:00.000 in t's location.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// applyTimePart returns base with its clock replaced by slots.Hour / Minute / Second.
func applyTimePart(base time.Time, slots *ParsedDateSlots) time.Time {
	y, mo, d := base.Date()
	return time.Date(y, mo, d, slots.Hour, slots.Minute, slots.Second, 0, base.Location())
}

// ResolveWeekday finds the date of targetWd relative to now according to
// direction ("future", "past", or "nearest").
// Returns the start of that day (00:00:00).
func ResolveWeekday(targetWd Weekday, direction Direction, now time.Time) time.Time {
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
			return StartOfDay(now)
		}
		bwd := 7 - fwd
		if bwd < fwd {
			days = -bwd
		} else {
			days = fwd
		}
	}
	return StartOfDay(now.AddDate(0, 0, days))
}

// resolveDirectionAnchor resolves "next week", "last month", "this year", etc.
// Returns the start of the target period:
//   - week  → configurable week boundary at 00:00:00
//   - month → 1st of month 00:00:00
//   - year  → Jan 1 00:00:00
//   - day   → start of day 00:00:00
func resolveDirectionAnchor(direction Direction, anchor Period, now time.Time, weekStartSunday bool) (time.Time, error) {
	loc := now.Location()
	y, m, _ := now.Date()

	currentWeekStart := func() time.Time {
		return StartOfDay(now.AddDate(0, 0, -weekStartOffset(now, weekStartSunday)))
	}

	switch anchor {
	case PeriodWeek:
		weekStart := currentWeekStart()
		switch direction {
		case DirectionFuture:
			return weekStart.AddDate(0, 0, 7), nil
		case DirectionPast:
			return weekStart.AddDate(0, 0, -7), nil
		default: // nearest
			return weekStart, nil
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
			return StartOfDay(now.AddDate(0, 0, 1)), nil
		case DirectionPast:
			return StartOfDay(now.AddDate(0, 0, -1)), nil
		default:
			return StartOfDay(now), nil
		}
	default:
		return time.Time{}, fmt.Errorf("nowandlater: resolve: unsupported anchor %v", anchor)
	}
}

// resolveMonthDay resolves a month+day expression with no year, inferring the
// year from now: uses the current year if the resulting time is after now,
// otherwise advances to next year.
func resolveMonthDay(month, day int, slots *ParsedDateSlots, now time.Time, direction Direction) (time.Time, error) {
	loc := now.Location()
	h, m, s := 0, 0, 0
	if slots.Period <= PeriodHour {
		h, m, s = slots.Hour, slots.Minute, slots.Second
	}
	y := now.Year()
	t := time.Date(y, time.Month(month), day, h, m, s, 0, loc)
	if direction == DirectionPast {
		if !t.Before(now) {
			t = time.Date(y-1, time.Month(month), day, h, m, s, 0, loc)
		}
		return t, nil
	}
	if !t.After(now) {
		t = time.Date(y+1, time.Month(month), day, h, m, s, 0, loc)
	}
	return t, nil
}

func resolveMonthOnly(month int, now time.Time, direction Direction) time.Time {
	loc := now.Location()
	y := now.Year()
	t := time.Date(y, time.Month(month), 1, 0, 0, 0, 0, loc)
	if direction == DirectionPast {
		currentMonthStart := time.Date(y, now.Month(), 1, 0, 0, 0, 0, loc)
		if !t.Before(currentMonthStart) {
			return time.Date(y-1, time.Month(month), 1, 0, 0, 0, 0, loc)
		}
		return t
	}
	if !t.After(now) {
		return time.Date(y+1, time.Month(month), 1, 0, 0, 0, 0, loc)
	}
	return t
}

// ---------------------------------------------------------------------------
// Interval helpers
// ---------------------------------------------------------------------------

// startOfPeriod truncates t to the start of the given period in t's location.
// For PeriodWeek, the start of week is configurable.
func startOfPeriod(t time.Time, period Period, weekStartSunday bool) time.Time {
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
		return StartOfDay(t)
	case PeriodWeek:
		return StartOfDay(t.AddDate(0, 0, -weekStartOffset(t, weekStartSunday)))
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
	return ResolveIntervalWithPolicy(slots, now, legacyResolvePolicy())
}

// ResolveIntervalWithPolicy resolves slots to a half-open calendar interval
// [start, end) using the supplied ambiguity policy.
func ResolveIntervalWithPolicy(slots *ParsedDateSlots, now time.Time, policy ResolvePolicy) (start, end time.Time, err error) {
	pt, err := ResolveWithPolicy(slots, now, policy)
	if err != nil {
		return
	}
	start = startOfPeriod(pt, slots.Period, policy.WeekStartSunday)
	end = EndOf(start, slots.Period)
	return
}

// weekStartOffset returns the number of days to subtract to reach the start of
// the current week in t's location.
func weekStartOffset(t time.Time, sundayStart bool) int {
	if sundayStart {
		return int(t.Weekday())
	}
	// Monday-start: Go Sunday=0 becomes 6, Monday=1 becomes 0, etc.
	return (int(t.Weekday()) + 6) % 7
}

func policyDurationDirection(policy ResolvePolicy) Direction {
	if policy.ImplicitDurationDirection == 0 {
		return DirectionFuture
	}
	return policy.ImplicitDurationDirection
}

func directionSign(direction Direction) int {
	if direction == DirectionPast {
		return -1
	}
	return 1
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func describeSlots(slots *ParsedDateSlots) string {
	switch slots.AmbiguousForm {
	case AmbiguousImplicitDuration:
		return "unsigned duration"
	case AmbiguousBareWeekday:
		return slots.Weekday.String()
	case AmbiguousBareMonth:
		return time.Month(slots.Month).String()
	case AmbiguousMonthDay:
		return fmt.Sprintf("%s %d", time.Month(slots.Month), slots.Day)
	default:
		return "date expression"
	}
}
