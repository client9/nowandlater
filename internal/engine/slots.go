package engine

import "time"

// Period describes the coarsest time component present in a parsed date/time string.
type Period int

const (
	PeriodSecond Period = iota + 1
	PeriodMinute
	PeriodHour
	PeriodDay
	PeriodFortnight // 14 days
	PeriodWeek
	PeriodMonth
	PeriodYear
)

// String implements fmt.Stringer.
func (p Period) String() string {
	switch p {
	case PeriodSecond:
		return "second"
	case PeriodMinute:
		return "minute"
	case PeriodHour:
		return "hour"
	case PeriodDay:
		return "day"
	case PeriodFortnight:
		return "fortnight"
	case PeriodWeek:
		return "week"
	case PeriodMonth:
		return "month"
	case PeriodYear:
		return "year"
	default:
		return ""
	}
}

// Weekday identifies a day of the week. WeekdayMonday = 1 … WeekdaySunday = 7;
// the zero value means "not specified".
type Weekday int

const (
	WeekdayMonday Weekday = iota + 1
	WeekdayTuesday
	WeekdayWednesday
	WeekdayThursday
	WeekdayFriday
	WeekdaySaturday
	WeekdaySunday
)

// String implements fmt.Stringer.
func (w Weekday) String() string {
	switch w {
	case WeekdayMonday:
		return "monday"
	case WeekdayTuesday:
		return "tuesday"
	case WeekdayWednesday:
		return "wednesday"
	case WeekdayThursday:
		return "thursday"
	case WeekdayFriday:
		return "friday"
	case WeekdaySaturday:
		return "saturday"
	case WeekdaySunday:
		return "sunday"
	default:
		return ""
	}
}

// Month identifies a calendar month. MonthJanuary = 1 … MonthDecember = 12;
// the zero value means "not specified".
type Month int

const (
	MonthJanuary Month = iota + 1
	MonthFebruary
	MonthMarch
	MonthApril
	MonthMay
	MonthJune
	MonthJuly
	MonthAugust
	MonthSeptember
	MonthOctober
	MonthNovember
	MonthDecember
)

// String implements fmt.Stringer.
func (m Month) String() string {
	switch m {
	case MonthJanuary:
		return "january"
	case MonthFebruary:
		return "february"
	case MonthMarch:
		return "march"
	case MonthApril:
		return "april"
	case MonthMay:
		return "may"
	case MonthJune:
		return "june"
	case MonthJuly:
		return "july"
	case MonthAugust:
		return "august"
	case MonthSeptember:
		return "september"
	case MonthOctober:
		return "october"
	case MonthNovember:
		return "november"
	case MonthDecember:
		return "december"
	default:
		return ""
	}
}

// AMPM distinguishes AM from PM in 12-hour clock notation.
type AMPM int

const (
	AMPMAm AMPM = iota + 1
	AMPMPm
)

// String implements fmt.Stringer.
func (a AMPM) String() string {
	if a == AMPMAm {
		return "am"
	}
	return "pm"
}

// Anchor names a fixed calendar point used as a reference in relative expressions
// such as "3 days before tomorrow". The zero value means "not specified".
type Anchor int

const (
	AnchorNow Anchor = iota + 1
	AnchorToday
	AnchorTomorrow
	AnchorYesterday
	Anchor2DaysAgo
	Anchor2DaysFromNow
	Anchor3DaysAgo
	Anchor3DaysFromNow
)

// String implements fmt.Stringer.
func (a Anchor) String() string {
	switch a {
	case AnchorNow:
		return "now"
	case AnchorToday:
		return "today"
	case AnchorTomorrow:
		return "tomorrow"
	case AnchorYesterday:
		return "yesterday"
	case Anchor2DaysAgo:
		return "2daysago"
	case Anchor2DaysFromNow:
		return "2daysfromnow"
	case Anchor3DaysAgo:
		return "3daysago"
	case Anchor3DaysFromNow:
		return "3daysfromnow"
	default:
		return ""
	}
}

// Modifier describes the temporal direction of a relative expression.
// The integer value is the sign multiplier: ModifierFuture = +1, ModifierPast = -1.
type Modifier int

const (
	ModifierFuture Modifier = 1
	ModifierPast   Modifier = -1
)

// String implements fmt.Stringer.
func (m Modifier) String() string {
	if m == ModifierFuture {
		return "future"
	}
	return "past"
}

// Direction qualifies weekday and anchor resolution.
type Direction int

const (
	DirectionFuture  Direction = iota + 1 // "next Monday", "coming Friday"
	DirectionPast                         // "last Monday"
	DirectionNearest                      // "this Monday" — whichever occurrence is closer
)

// String implements fmt.Stringer.
func (d Direction) String() string {
	switch d {
	case DirectionFuture:
		return "future"
	case DirectionPast:
		return "past"
	case DirectionNearest:
		return "nearest"
	default:
		return ""
	}
}

// AmbiguousForm records which underspecified input shape produced the slots.
// It lets the resolver apply policy decisions without changing token meaning.
type AmbiguousForm int

const (
	AmbiguousNone AmbiguousForm = iota
	AmbiguousImplicitDuration
	AmbiguousBareWeekday
	AmbiguousBareMonth
	AmbiguousMonthDay
)

// ParsedDateSlots holds the intermediate result of parsing a date/time string.
// Zero values mean "not present in input". DeltaSeconds uses a pointer because
// 0 is a valid delta ("now"/"today"). All other fields use value types with
// zero meaning "not specified".
//
// Modeled on libc struct tm with three additions: delta, direction, and anchor.
type ParsedDateSlots struct {
	// Absolute fields (0 = not mentioned in input, except Weekday/Hour/Minute/Second)
	Year    int     // full year, e.g. 2026; 0 = not specified
	Month   int     // 1–12; 0 = not specified
	Day     int     // 1–31; 0 = not specified
	Weekday Weekday // 0=not specified; WeekdayMonday…WeekdaySunday (1–7)
	Hour    int     // 0–23 (24-hour); 0 = unset or midnight — use Period to distinguish
	Minute  int     // 0–59
	Second  int     // 0–59

	// Relative offset from now, pre-normalized to seconds.
	// e.g. "3 days ago" → -259200
	// nil means no relative delta was expressed.
	DeltaSeconds *int

	// UnixTime holds an absolute Unix timestamp (seconds since 1970-01-01 UTC).
	// When set, all other fields are ignored by Resolve.
	// Zero means no Unix timestamp was expressed (timestamps < 10000 are rejected).
	UnixTime int64

	// Direction qualifies weekday/month resolution.
	// Zero value means no direction was expressed.
	Direction Direction

	// Anchor names a calendar boundary used with start/end expressions.
	// Valid values: PeriodDay | PeriodWeek | PeriodMonth | PeriodYear.
	// Zero value means no anchor was expressed.
	Anchor Period

	// Location is the timezone parsed from the input (e.g. "EST", "+05:30", "Z").
	// nil means no timezone was specified; the resolver uses now.Location().
	Location *time.Location

	// Period is the coarsest time component present in the input.
	// Defaults to PeriodDay when nothing finer or coarser is specified.
	// Zero value (0) means unset.
	Period Period

	// AmbiguousForm identifies underspecified inputs whose final meaning depends
	// on parser policy rather than explicit user wording.
	AmbiguousForm AmbiguousForm
}
