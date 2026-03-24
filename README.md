# nowandlater

A self-contained, zero-dependency Go library for parsing natural-language
date and time strings into `time.Time` values.

```go
import "github.com/client9/nowandlater"
```

## Installation

```
go get github.com/client9/nowandlater
```

No external dependencies.

## Usage

```go
p := nowandlater.Parser{}

// Single point in time
t, err := p.Parse("next Monday at 9:30 AM")
t, err := p.Parse("in 3 days")
t, err := p.Parse("2026-12-04T09:30:00Z")

// Calendar interval [start, end)
start, end, err := p.ParseInterval("this week")
start, end, err := p.ParseInterval("last month")
```

`Parser` is safe for concurrent use. Its zero value is valid and defaults to
English, `time.Local`, and `time.Now`.

```go
// Custom reference time and timezone (useful in tests)
p := nowandlater.Parser{
    Now:      func() time.Time { return fixedNow },
    Location: time.UTC,
}
```

## Supported formats

### Absolute dates

| Input | Notes |
|---|---|
| `2026-12-04` | ISO 8601 |
| `2026/12/04`, `2026.12.04` | Alternative separators |
| `2026-dec-04`, `04-Dec-2026` | With month name |
| `December 4, 2026` | Full month name |
| `Dec 4 2026`, `4 Dec 2026` | Abbreviated |
| `12/04/2026` | MM/DD/YYYY (US default; see DateOrder) |

### Times

| Input | Notes |
|---|---|
| `9:30`, `14:30:00` | 24-hour |
| `9:30 AM`, `3 PM` | 12-hour with AM/PM |
| `noon`, `midnight` | Special words |
| `7.15pm` | Dot separator with AM/PM |
| `09:30:00-07:00`, `9:30+05:30` | With UTC offset |

### Date + time combined

```
next Monday at 9:30 AM
2026-12-04 09:30
2026-12-04T09:30:00Z              (ISO 8601 / RFC 3339)
Mon, 02 Jan 2006 15:04:05 -0700   (RFC 2822)
Mon Jan  2 15:04:05 MST 2006      (ANSI C / Go reference time)
```

### Relative expressions

```
now             today           tomorrow        yesterday
next Monday     last Friday     this week
in 3 days       2 hours ago     a week from now
3 days before tomorrow
in 1 hour and 30 minutes
1.5 hours ago
```

### Timezone tokens

Recognized globally across all languages: `UTC`, `EST`, `PST`, `CET`, and
many others (see `timezone.go`). Numeric offsets (`+05:30`, `-07:00`, `Z`)
are also supported.

## Languages

### Built-in

| Variable | Language |
|---|---|
| `nowandlater.English` | English (default) |
| `nowandlater.Spanish` | Spanish / Español |

```go
p := nowandlater.Parser{Lang: &nowandlater.Spanish}
t, err := p.Parse("próximo lunes")   // next Monday
t, err := p.Parse("hace 3 días")     // 3 days ago
```

### Custom languages

Implement a `Lang` with a `Words` map, `OrdinalSuffixes`, and optionally
`DateOrder` and `Handlers`:

```go
myLang := nowandlater.Lang{
    Words: map[string]nowandlater.WordEntry{
        "lundi":    {nowandlater.TokenWeekday, nowandlater.WeekdayMonday},
        "prochain": {nowandlater.TokenDirection, nowandlater.DirectionFuture},
        "il y a":   {nowandlater.TokenModifier, nowandlater.ModifierPast}, // multi-word phrase
        "midi":     {nowandlater.TokenTime, "12:00"},                      // time-word substitution
    },
    OrdinalSuffixes: []string{"er", "re", "me", "ème"},
    DateOrder:       nowandlater.DMY,
}
p := nowandlater.Parser{Lang: &myLang}
```

See `lang_en.go` and `lang_es.go` for complete reference implementations.

## Interval support

`ParseInterval` resolves an expression to a half-open calendar interval
`[start, end)` where `start` is always aligned to the period boundary:

| Input | Period | start | end |
|---|---|---|---|
| `today` | day | 2026-03-22 00:00 | 2026-03-23 00:00 |
| `this week` | week | Monday 00:00 | next Monday 00:00 |
| `last month` | month | 2026-03-01 00:00 | 2026-04-01 00:00 |
| `2026` | year | 2026-01-01 00:00 | 2027-01-01 00:00 |
| `in 2 hours` | hour | 12:00:00 | 13:00:00 |

Weeks start on Monday. `EndOf(start, period)` and `ResolveInterval(slots, now)`
are available for lower-level use.

## DateOrder

Ambiguous all-numeric dates like `02/03/2026` are interpreted according to
`Lang.DateOrder`:

| Value | Interpretation | Locale |
|---|---|---|
| `MDY` (default) | month/day/year | US English |
| `DMY` | day/month/year | Europe, Latin America |
| `YMD` | year/month/day | ISO 8601 (2-digit year) |

`DateOrder` only affects the `INTEGER INTEGER YEAR` signature. Dates with a
month name or an ISO year-first form are always unambiguous.

## Error handling

An unrecognised input returns `ErrUnknownSignature`:

```go
t, err := p.Parse("something unrecognisable")
if errors.Is(err, nowandlater.ErrUnknownSignature) {
    // input did not match any known pattern
}
```

## Known limitations

- **"second" ambiguity**: the ordinal "second" (2nd) conflicts with the time unit.
  Use `"2nd"` instead of `"second"` when referring to the second day.
- **Spanish "mar"**: resolves to Tuesday (martes), not March. Write "marzo" in full.
- **2-digit years**: not currently supported.
- **Unix timestamps**: 10-digit integers are not yet recognised.
- **morning / afternoon / evening**: semantics undefined; not yet supported.

## Development tool

`cmd/tokenize` is a CLI for inspecting the parse pipeline interactively:

```
go run ./cmd/tokenize "next Monday at 9:30 AM"
go run ./cmd/tokenize -now 2026-03-22 -interval "this week"
echo "in 2 days" | go run ./cmd/tokenize
```

Output shows the token list, signature, parsed period, and resolved time.

## Running tests

```
go test ./...
```
