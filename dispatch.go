package nowandlater

import "errors"

// Handler is a function that extracts slot values from a token list.
// It receives the full token list (including FILLER tokens) and returns
// a populated ParsedDateSlots on success.
type Handler func(tokens []Token) (*ParsedDateSlots, error)

// ErrUnknownSignature is returned by Parse when no handler matches the input.
var ErrUnknownSignature = errors.New("nowandlater: unknown date signature")

// handlers maps each signature string to its handler function.
// The signature is produced by Signature(Tokenize(input)) — token type names
// joined by spaces, with FILLER tokens excluded.
var handlers = map[string]Handler{
	// Stand-alone anchors
	"ANCHOR": handleAnchor,

	// Stand-alone time (e.g. noon/midnight after preprocessing)
	"TIME": handleTime,

	// Weekday
	"WEEKDAY":                          handleWeekday,
	"DIRECTION WEEKDAY":                handleDirectionWeekday,
	"DIRECTION WEEKDAY PREP TIME":      handleDirectionWeekdayTime,
	"DIRECTION WEEKDAY PREP TIME AMPM": handleDirectionWeekdayTimeAMPM,
	"DIRECTION WEEKDAY PREP INTEGER":   handleDirectionWeekdayPrepInteger,

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
	"ANCHOR PREP TIME":         handleAnchorPrepTime,
	"ANCHOR PREP TIME AMPM":    handleAnchorPrepTimeAMPM,
	"ANCHOR PREP INTEGER AMPM": handleAnchorPrepIntegerAMPM,
	"ANCHOR PREP INTEGER":      handleAnchorPrepInteger,

	// Calendar date: month-name forms
	"MONTH INTEGER":                   handleMonthDay,
	"INTEGER MONTH":                   handleDayMonth,
	"MONTH INTEGER YEAR":              handleMonthDayYear,
	"MONTH YEAR":                      handleMonthYear,
	"YEAR MONTH":                      handleYearMonth,
	"WEEKDAY INTEGER MONTH YEAR":      handleWeekdayIntegerMonthYear,     // RFC 2822 date-only
	"WEEKDAY INTEGER MONTH YEAR TIME": handleWeekdayIntegerMonthYearTime, // RFC 2822 full

	// Calendar date: numeric compound forms (all separators → same signature)
	"YEAR INTEGER INTEGER": handleYearIntegerInteger,
	"YEAR MONTH INTEGER":   handleYearMonthInteger,
	"INTEGER MONTH YEAR":   handleIntegerMonthYear,
	"INTEGER INTEGER YEAR": handleIntegerIntegerYear, // MM/DD/YYYY — American default

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

	// Combined date + time
	"YEAR INTEGER INTEGER TIME":      handleYearIntegerIntegerTime,
	"YEAR INTEGER INTEGER TIME AMPM": handleYearIntegerIntegerTimeAMPM,
	"YEAR MONTH INTEGER TIME":        handleYearMonthIntegerTime,
	"YEAR MONTH INTEGER TIME AMPM":   handleYearMonthIntegerTimeAMPM,
	"MONTH INTEGER YEAR TIME":        handleMonthIntegerYearTime,
	"MONTH INTEGER YEAR TIME AMPM":   handleMonthIntegerYearTimeAMPM,

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
	case "INTEGER INTEGER YEAR PREP TIME":
		return withPrepTime(base)
	case "INTEGER INTEGER YEAR PREP TIME AMPM":
		return withPrepTime(base)
	case "INTEGER INTEGER YEAR PREP INTEGER AMPM":
		return withPrepTime(base)
	}
	return nil
}

// Parse tokenizes input, computes its signature, and dispatches to the
// matching handler. It returns ErrUnknownSignature if no handler is registered
// for the input's signature.
func (lang *Lang) Parse(input string) (*ParsedDateSlots, error) {
	tokens := lang.Tokenize(input)

	// Strip a trailing TIMEZONE token before signature dispatch.
	// Timezone is always a suffix so we can peel it off without affecting handlers.
	var tzValue string
	filtered := filterFillers(tokens)
	if len(filtered) > 0 && filtered[len(filtered)-1].Type == TokenTimezone {
		tzValue = filtered[len(filtered)-1].Value.(string)
		for i := len(tokens) - 1; i >= 0; i-- {
			if tokens[i].Type == TokenTimezone {
				tokens = append(tokens[:i], tokens[i+1:]...)
				break
			}
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
