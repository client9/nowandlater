package engine

import "errors"

// Handler is a function that extracts slot values from a token list.
// It receives the full token list (including FILLER tokens) and returns
// a populated ParsedDateSlots on success.
type Handler func(tokens []Token) (*ParsedDateSlots, error)

// ErrUnknownSignature is returned by Parse when no handler matches the input.
var ErrUnknownSignature = errors.New("nowandlater: unknown date signature")

// ErrAmbiguous is returned when the input matches a recognisable date pattern
// but the meaning cannot be determined without additional context.
// Example: "mar 5" in Spanish, where "mar" abbreviates both martes (Tuesday)
// and marzo (March).
var ErrAmbiguous = errors.New("nowandlater: ambiguous date expression")

// handlers maps each signature string to its handler function.
// The signature is produced by Signature(Tokenize(input)) — token type names
// joined by spaces, with FILLER tokens excluded.
var handlers = map[string]Handler{
	// Unix timestamp: bare integer interpreted as seconds since 1970-01-01 UTC.
	"INTEGER": handleUnixTimestamp,

	// Stand-alone anchors
	"ANCHOR": handleAnchor,

	// Stand-alone time (e.g. noon/midnight after preprocessing)
	"TIME": handleTime,

	// Weekday
	"WEEKDAY":                          handleWeekday,
	"DIRECTION WEEKDAY":                handleDirectionWeekday,
	"DIRECTION WEEKDAY PREP TIME":      withPrepTime(handleDirectionWeekday),
	"DIRECTION WEEKDAY PREP TIME AMPM": withPrepTime(handleDirectionWeekday),
	"DIRECTION WEEKDAY PREP INTEGER":   withPrepTime(handleDirectionWeekday),

	// Direction + unit: "next week", "last month"
	"DIRECTION UNIT": handleDirectionUnit,

	// Time of day
	"INTEGER AMPM":       handleIntegerAMPM,
	"TIME AMPM":          handleTimeAMPM, // "7:15pm" after dot preprocessing
	"PREP TIME":          handlePrepTime,
	"PREP TIME AMPM":     handlePrepTimeAMPM,
	"MODIFIER TIME":      handlePrepTime,
	"MODIFIER TIME AMPM": handlePrepTimeAMPM,

	// Anchor + time: "today at 9:30", "tomorrow at 3pm"
	"ANCHOR PREP TIME":         withPrepTime(handleAnchor),
	"ANCHOR PREP TIME AMPM":    withPrepTime(handleAnchor),
	"ANCHOR PREP INTEGER AMPM": withPrepTime(handleAnchor),
	"ANCHOR PREP INTEGER":      withPrepTime(handleAnchor),

	// "second" as ordinal day-2 in month/day expressions ("march second", "second of march")
	"MONTH UNIT":      handleMonthSecondDay,
	"MONTH UNIT YEAR": handleMonthSecondDayYear,
	"UNIT MONTH":      handleSecondDayMonth,
	"UNIT MONTH YEAR": handleSecondDayMonthYear,

	"MONTH UNIT PREP TIME":         withPrepTime(handleMonthSecondDay),
	"MONTH UNIT PREP TIME AMPM":    withPrepTime(handleMonthSecondDay),
	"MONTH UNIT PREP INTEGER AMPM": withPrepTime(handleMonthSecondDay),

	"MONTH UNIT YEAR PREP TIME":         withPrepTime(handleMonthSecondDayYear),
	"MONTH UNIT YEAR PREP TIME AMPM":    withPrepTime(handleMonthSecondDayYear),
	"MONTH UNIT YEAR PREP INTEGER AMPM": withPrepTime(handleMonthSecondDayYear),

	// Calendar date: month-name forms
	"MONTH INTEGER":                        handleMonthDay,
	"MONTH INTEGER TIME":                   withTrailingTime(handleMonthDay), // Go Stamp format
	"INTEGER MONTH":                        handleDayMonth,
	"MONTH INTEGER YEAR":                   handleMonthDayYear,
	"MONTH YEAR":                           handleMonthYear,
	"YEAR MONTH":                           handleYearMonth,
	"WEEKDAY INTEGER MONTH YEAR":           handleWeekdayIntegerMonthYear,                   // RFC 2822 date-only
	"WEEKDAY INTEGER MONTH YEAR TIME":      withTrailingTime(handleWeekdayIntegerMonthYear), // RFC 2822 full
	"WEEKDAY INTEGER MONTH YEAR TIME AMPM": withTrailingTime(handleWeekdayIntegerMonthYear), // RFC 2822 with AM/PM
	"WEEKDAY MONTH INTEGER YEAR":           handleWeekdayMonthIntegerYear,                   // ANSIC date-only
	"WEEKDAY MONTH INTEGER TIME YEAR":      handleWeekdayMonthIntegerTimeYear,               // ANSIC, UnixDate, RubyDate
	"INTEGER MONTH INTEGER TIME":           handleIntegerMonthIntegerTime,                   // RFC822, RFC822Z
	"WEEKDAY INTEGER MONTH INTEGER TIME":   handleWeekdayIntegerMonthIntegerTime,            // RFC850

	// Calendar date: numeric compound forms (all separators → same signature)
	"YEAR INTEGER INTEGER":         handleYearIntegerInteger,
	"YEAR MONTH INTEGER":           handleYearMonthInteger,
	"INTEGER MONTH YEAR":           handleIntegerMonthYear,
	"INTEGER MONTH YEAR TIME":      withTrailingTime(handleIntegerMonthYear),
	"INTEGER MONTH YEAR TIME AMPM": withTrailingTime(handleIntegerMonthYear),
	"INTEGER INTEGER YEAR":         handleIntegerIntegerYear, // MM/DD/YYYY — American default

	// Calendar date: compound forms with leading preposition
	"PREP YEAR INTEGER INTEGER": prepDelegate(handleYearIntegerInteger),
	"PREP YEAR MONTH INTEGER":   prepDelegate(handleYearMonthInteger),

	// Year only
	"YEAR": handleYear,

	// Relative deltas
	"INTEGER UNIT":                       handleIntegerUnit,                    // "4 hours", "3 días" — implied future
	"INTEGER UNIT INTEGER UNIT":          handleIntegerUnitIntegerUnit,         // "1 hour and 10 minutes"
	"INTEGER UNIT INTEGER UNIT MODIFIER": handleIntegerUnitIntegerUnitModifier, // "1 hour and 10 minutes ago"
	"DECIMAL UNIT":                       handleDecimalUnit,                    // "3.5 days", "1.5 horas" — implied future
	"PREP DECIMAL UNIT":                  handlePrepDecimalUnit,                // "in 1.5 hours"
	"DECIMAL UNIT MODIFIER":              handleDecimalUnitModifier,            // "1.5 hours ago"
	"MODIFIER DECIMAL UNIT":              handleModifierDecimalUnit,            // "hace 1.5 días"
	"INTEGER UNIT MODIFIER":              handleRelativeDelta,
	"INTEGER UNIT MODIFIER ANCHOR":       handleRelativeDeltaAnchor,
	"MODIFIER INTEGER UNIT":              handleModifierIntegerUnit,            // "hace 3 días" word order
	"MODIFIER INTEGER UNIT INTEGER UNIT": handleModifierIntegerUnitIntegerUnit, // "hace 1 hora y 10 minutos"
	"PREP INTEGER UNIT":                  handlePrepIntegerUnit,
	"PREP INTEGER UNIT INTEGER UNIT":     handlePrepIntegerUnitIntegerUnit, // "in 1 hour and 10 minutes"
	"PREP DIRECTION INTEGER UNIT":        handlePrepDirectionIntegerUnit,
	"PREP UNIT":                          handlePrepUnit,     // "in a week", "in an hour"
	"UNIT MODIFIER":                      handleUnitModifier, // "a week ago"

	// Weekday-first word order: "lunes próximo", "lundi prochain"
	"WEEKDAY DIRECTION": handleWeekdayDirection,

	// Unit-first word order: "mes pasado", "semaine prochaine"
	"UNIT DIRECTION": handleUnitDirection,

	// Direction + unit + weekday: "来週の月曜日" (next Monday in Japanese)
	"DIRECTION UNIT WEEKDAY": handleDirectionUnitWeekday,

	// Combined date + time
	"YEAR INTEGER INTEGER TIME":      withTrailingTime(handleYearIntegerInteger),
	"YEAR INTEGER INTEGER TIME AMPM": withTrailingTime(handleYearIntegerInteger),
	"YEAR MONTH INTEGER TIME":        withTrailingTime(handleYearMonthInteger),
	"YEAR MONTH INTEGER TIME AMPM":   withTrailingTime(handleYearMonthInteger),
	"TIME AMPM MONTH INTEGER YEAR":   handleTimeAMPMMonthIntegerYear,
	"MONTH INTEGER YEAR TIME":        withTrailingTime(handleMonthDayYear),
	"MONTH INTEGER YEAR TIME AMPM":   withTrailingTime(handleMonthDayYear),

	// Time-only: PREP INTEGER AMPM ("at 3 PM")
	"PREP INTEGER AMPM": handlePrepIntegerAMPM,

	// Weekday + time
	"WEEKDAY PREP TIME":         withPrepTime(handleWeekday),
	"WEEKDAY PREP TIME AMPM":    withPrepTime(handleWeekday),
	"WEEKDAY PREP INTEGER AMPM": withPrepTime(handleWeekday),

	// Direction + weekday + integer time (no AMPM already handled; add AMPM variant)
	"DIRECTION WEEKDAY PREP INTEGER AMPM": withPrepTime(handleDirectionWeekday),

	// Month + day + time
	"MONTH INTEGER PREP TIME":         withPrepTime(handleMonthDay),
	"MONTH INTEGER PREP TIME AMPM":    withPrepTime(handleMonthDay),
	"MONTH INTEGER PREP INTEGER AMPM": withPrepTime(handleMonthDay),

	// Month + day + year + time
	"MONTH INTEGER YEAR PREP TIME":         withPrepTime(handleMonthDayYear),
	"MONTH INTEGER YEAR PREP TIME AMPM":    withPrepTime(handleMonthDayYear),
	"MONTH INTEGER YEAR PREP INTEGER AMPM": withPrepTime(handleMonthDayYear),

	// YEAR INTEGER INTEGER + time
	"YEAR INTEGER INTEGER PREP TIME":         withPrepTime(handleYearIntegerInteger),
	"YEAR INTEGER INTEGER PREP TIME AMPM":    withPrepTime(handleYearIntegerInteger),
	"YEAR INTEGER INTEGER PREP INTEGER AMPM": withPrepTime(handleYearIntegerInteger),

	// YEAR MONTH INTEGER + time
	"YEAR MONTH INTEGER PREP TIME":         withPrepTime(handleYearMonthInteger),
	"YEAR MONTH INTEGER PREP TIME AMPM":    withPrepTime(handleYearMonthInteger),
	"YEAR MONTH INTEGER PREP INTEGER AMPM": withPrepTime(handleYearMonthInteger),
}

// resolveHandler returns the handler for sig, applying the three-level priority:
//  1. lang.Handlers (language-specific overrides — highest priority)
//  2. date-order-aware handlers for ambiguous numeric signatures (uses lang.DateOrder)
//  3. global handlers map (language-neutral fallback)
//
// Returns nil if no handler is found.
func (lang *Lang) resolveHandler(sig string) Handler {
	if h, ok := lang.Handlers[sig]; ok {
		return h
	}
	if h := lang.dateOrderHandler(sig); h != nil {
		return h
	}
	if h, ok := handlers[sig]; ok {
		return h
	}
	return nil
}

// dateOrderHandler returns a handler for signatures that are sensitive to
// lang.DateOrder (ambiguous all-numeric dates). Returns nil for all other
// signatures — those fall through to the global handlers map.
func (lang *Lang) dateOrderHandler(sig string) Handler {
	base := makeIntegerIntegerYearHandler(lang.DateOrder)
	switch sig {
	case "INTEGER INTEGER YEAR":
		return base
	case "INTEGER INTEGER YEAR TIME", "INTEGER INTEGER YEAR TIME AMPM":
		return withTrailingTime(makeIntegerIntegerYearHandler(lang.DateOrder))
	case "INTEGER INTEGER YEAR PREP TIME":
		return withPrepTime(base)
	case "INTEGER INTEGER YEAR PREP TIME AMPM":
		return withPrepTime(base)
	case "INTEGER INTEGER YEAR PREP INTEGER AMPM":
		return withPrepTime(base)
	case "DATE_FRAGMENT":
		return makeDateFragmentHandler(lang.DateOrder)
	case "DATE_FRAGMENT TIME", "DATE_FRAGMENT TIME AMPM":
		return withTrailingTime(makeDateFragmentHandler(lang.DateOrder))
	case "DATE_FRAGMENT PREP TIME":
		return withPrepTime(makeDateFragmentHandler(lang.DateOrder))
	case "DATE_FRAGMENT PREP TIME AMPM":
		return withPrepTime(makeDateFragmentHandler(lang.DateOrder))
	case "DATE_FRAGMENT PREP INTEGER AMPM":
		return withPrepTime(makeDateFragmentHandler(lang.DateOrder))
	case "INTEGER AMPM INTEGER INTEGER YEAR":
		return makeIntegerAMPMIntegerIntegerYearHandler(lang.DateOrder)
	case "INTEGER AMPM DATE_FRAGMENT":
		return makeIntegerAMPMDateFragmentHandler(lang.DateOrder)
	case "TIME AMPM DATE_FRAGMENT":
		return makeTimeAMPMDateFragmentHandler(lang.DateOrder)
	case "TIME AMPM PREP DATE_FRAGMENT":
		return makeTimeAMPMDateFragmentHandler(lang.DateOrder)
	}
	return nil
}

// Parse tokenizes input, computes its signature, and dispatches to the
// matching handler. It returns ErrUnknownSignature if no handler is registered
// for the input's signature.
func (lang *Lang) Parse(input string) (*ParsedDateSlots, error) {
	tokens := lang.Tokenize(input)

	// Strip any TIMEZONE token before signature dispatch.
	// Timezone is a free-standing modifier that can appear in any position
	// (trailing in most formats, but mid-sequence in UnixDate/RubyDate where
	// the year follows the timezone). Handlers never index it.
	var tzValue string
	for i := len(tokens) - 1; i >= 0; i-- {
		if tokens[i].Type == TokenTimezone {
			tzValue = tokens[i].Value.(string)
			tokens = append(tokens[:i], tokens[i+1:]...)
			break
		}
	}

	sig := Signature(tokens)
	h := lang.resolveHandler(sig)
	if h == nil {
		return nil, ErrUnknownSignature
	}
	slots, err := h(tokens)
	if err != nil {
		return nil, err
	}
	if tzValue != "" {
		loc, err := parseTimezoneValue(tzValue, lang)
		if err != nil {
			return nil, err
		}
		slots.Location = loc
	}
	return slots, nil
}
