package nowandlater

import (
	"fmt"
	"strconv"
	"time"
)

// defaultTimezones maps lowercase timezone abbreviations to their fixed-offset
// *time.Location values. All entries use time.FixedZone so the offset is always
// applied as specified, regardless of DST rules.
//
// Ambiguous abbreviations:
//   - IST: chosen as India Standard Time (UTC+5:30); could also mean Irish (UTC+1)
//     or Israel (UTC+2). Override via Lang.Timezones if a different meaning is needed.
var defaultTimezones = map[string]*time.Location{
	// UTC / universal
	"utc": time.UTC,
	"gmt": time.UTC,
	"z":   time.UTC,
	// US Alaska
	"akst": time.FixedZone("AKST", -9*3600),
	"akdt": time.FixedZone("AKDT", -8*3600),
	// US Hawaii / Samoa
	"hst":  time.FixedZone("HST", -10*3600),
	"hast": time.FixedZone("HAST", -10*3600),
	"hadt": time.FixedZone("HADT", -9*3600),
	"sst":  time.FixedZone("SST", -11*3600), // Samoa Standard Time
	// US Pacific
	"pst": time.FixedZone("PST", -8*3600),
	"pdt": time.FixedZone("PDT", -7*3600),
	// US Mountain
	"mst": time.FixedZone("MST", -7*3600),
	"mdt": time.FixedZone("MDT", -6*3600),
	// US Central
	"cst": time.FixedZone("CST", -6*3600),
	"cdt": time.FixedZone("CDT", -5*3600),
	// US Eastern
	"est": time.FixedZone("EST", -5*3600),
	"edt": time.FixedZone("EDT", -4*3600),
	// Canada / Atlantic
	"ast": time.FixedZone("AST", -4*3600),           // Atlantic Standard Time
	"adt": time.FixedZone("ADT", -3*3600),           // Atlantic Daylight Time
	"nst": time.FixedZone("NST", -(3*3600 + 30*60)), // Newfoundland Standard
	"ndt": time.FixedZone("NDT", -(2*3600 + 30*60)), // Newfoundland Daylight
	// South America
	"art":  time.FixedZone("ART", -3*3600),  // Argentina Time
	"brt":  time.FixedZone("BRT", -3*3600),  // Brasilia Time
	"brst": time.FixedZone("BRST", -2*3600), // Brasilia Summer Time
	"clt":  time.FixedZone("CLT", -4*3600),  // Chile Standard Time
	"clst": time.FixedZone("CLST", -3*3600), // Chile Summer Time
	// Western Europe
	"wet":  time.FixedZone("WET", 0),       // Western European Time
	"west": time.FixedZone("WEST", 1*3600), // Western European Summer Time
	"bst":  time.FixedZone("BST", 1*3600),  // British Summer Time
	// Central Europe
	"cet":  time.FixedZone("CET", 1*3600),
	"cest": time.FixedZone("CEST", 2*3600),
	"met":  time.FixedZone("MET", 1*3600),  // Middle European Time
	"mez":  time.FixedZone("MEZ", 1*3600),  // Mitteleuropäische Zeit
	"mest": time.FixedZone("MEST", 2*3600), // Middle European Summer Time
	"mesz": time.FixedZone("MESZ", 2*3600), // Mitteleuropäische Sommerzeit
	// Eastern Europe
	"eet":  time.FixedZone("EET", 2*3600),  // Eastern European Time
	"eest": time.FixedZone("EEST", 3*3600), // Eastern European Summer Time
	// Africa
	"wat":  time.FixedZone("WAT", 1*3600),  // West Africa Time
	"cat":  time.FixedZone("CAT", 2*3600),  // Central Africa Time
	"sast": time.FixedZone("SAST", 2*3600), // South Africa Standard Time
	"eat":  time.FixedZone("EAT", 3*3600),  // East Africa Time
	// Russia / Middle East
	"msk": time.FixedZone("MSK", 3*3600), // Moscow Standard Time
	"msd": time.FixedZone("MSD", 4*3600), // Moscow Daylight Time (historical)
	"gst": time.FixedZone("GST", 4*3600), // Gulf Standard Time
	// Asia / Pacific
	"ist":  time.FixedZone("IST", 5*3600+30*60), // India Standard Time (ambiguous; see note)
	"sgt":  time.FixedZone("SGT", 8*3600),
	"hkt":  time.FixedZone("HKT", 8*3600),
	"cst8": time.FixedZone("CST", 8*3600), // China Standard Time — not exposed as "cst" (conflicts with US Central)
	"jst":  time.FixedZone("JST", 9*3600),
	"kst":  time.FixedZone("KST", 9*3600),
	"aest": time.FixedZone("AEST", 10*3600),
	"aedt": time.FixedZone("AEDT", 11*3600),
	// New Zealand
	"nzst": time.FixedZone("NZST", 12*3600),
	"nzdt": time.FixedZone("NZDT", 13*3600),
}

// parseTimezoneValue converts a timezone token value to a *time.Location.
// Resolution order:
//  1. lang.Timezones (caller-supplied overrides)
//  2. defaultTimezones (built-in abbreviation table)
//  3. numeric offset parsing (+HH, +HHMM, +HH:MM and negative equivalents)
func parseTimezoneValue(value string, lang *Lang) (*time.Location, error) {
	if lang.Timezones != nil {
		if loc, ok := lang.Timezones[value]; ok {
			return loc, nil
		}
	}
	if loc, ok := defaultTimezones[value]; ok {
		return loc, nil
	}
	return parseNumericOffset(value)
}

// parseNumericOffset parses a numeric timezone offset string into a *time.Location.
// Accepted forms: +HH, -HH, +HHMM, -HHMM, +HH:MM, -HH:MM.
func parseNumericOffset(s string) (*time.Location, error) {
	if len(s) < 3 {
		return nil, fmt.Errorf("nowandlater: unrecognised timezone %q", s)
	}
	sign := 1
	switch s[0] {
	case '+':
	case '-':
		sign = -1
	default:
		return nil, fmt.Errorf("nowandlater: unrecognised timezone %q", s)
	}
	rest := s[1:]

	var hours, minutes int
	var err error
	switch len(rest) {
	case 2: // HH
		hours, err = strconv.Atoi(rest)
	case 4: // HHMM
		hours, err = strconv.Atoi(rest[:2])
		if err == nil {
			minutes, err = strconv.Atoi(rest[2:])
		}
	case 5: // HH:MM
		if rest[2] != ':' {
			return nil, fmt.Errorf("nowandlater: unrecognised timezone %q", s)
		}
		hours, err = strconv.Atoi(rest[:2])
		if err == nil {
			minutes, err = strconv.Atoi(rest[3:])
		}
	default:
		return nil, fmt.Errorf("nowandlater: unrecognised timezone %q", s)
	}
	if err != nil || hours > 14 || minutes > 59 {
		return nil, fmt.Errorf("nowandlater: unrecognised timezone %q", s)
	}
	offset := sign * (hours*3600 + minutes*60)
	return time.FixedZone(s, offset), nil
}
