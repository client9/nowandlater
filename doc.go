// Package nowandlater parses natural-language date and time strings into
// concrete [time.Time] values.
//
// The package is self-contained with no external dependencies.
//
// # Quick start
//
//	p := nowandlater.Parser{}
//	t, err := p.Parse("next Monday at 9:30 AM")
//	start, end, err := p.ParseInterval("this week")
//
// [Parser] is the primary entry point. Its zero value is valid and uses
// English, [time.Local], and [time.Now] as defaults.
//
// # Supported input formats
//
// Absolute dates:
//
//	2026-12-04          2026/12/04          2026.12.04
//	December 4, 2026    Dec 4 2026          04-Dec-2026
//	12/04/2026          2026-dec-04
//
// Relative expressions:
//
//	tomorrow            yesterday           now
//	next Monday         last Friday         this week
//	in 3 days           2 hours ago         a week from now
//	3 days before tomorrow
//
// Times:
//
//	9:30 AM             14:30:00            noon            midnight
//	next Monday at 9:30 AM
//	2026-12-04T09:30:00Z
//
// Machine formats (RFC 3339, RFC 2822, ANSI C, Unix, Ruby):
//
//	2026-03-22T10:00:00-07:00
//	Mon, 02 Jan 2006 15:04:05 -0700
//	Mon Jan  2 15:04:05 2006
//
// # Multi-language support
//
// Built-in languages: [English], [Spanish].
//
// Custom languages can be added by constructing a [Lang] with a [Lang.Words]
// map, [Lang.OrdinalSuffixes], and optional [Lang.DateOrder] and
// [Lang.Handlers] overrides. See CLAUDE.md for the full guide.
//
// # Interval support
//
// [Parser.ParseInterval] resolves an expression to a half-open calendar
// interval [start, end). [ResolveInterval] and [EndOf] are available for
// lower-level use.
//
// # Error handling
//
// An unrecognised input returns [ErrUnknownSignature]. No other sentinel
// errors are defined; all other errors indicate invalid timezone tokens.
package nowandlater
