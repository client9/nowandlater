![Now-and-Later-Wrapper-Small](https://github.com/user-attachments/assets/6095d42b-a443-4363-8b73-29cbdf297725)

# nowandlater - fast natural date and time parsing

A self-contained, zero-dependency Go library for parsing natural-language
date and time strings into `time.Time` values.

[![Go Reference](https://pkg.go.dev/badge/github.com/client9/nowandlater.svg)](https://pkg.go.dev/github.com/client9/nowandlater)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
🇬🇧🇪🇸🇩🇪🇫🇷🇮🇹🇵🇹🇷🇺🇯🇵🇨🇳

```go
import "github.com/client9/nowandlater"
```
---

### What makes it different?

* Fast - 500ns per call
* Only uses stdlib, no external dependencies
* No regular expressions
* No parser/generator langauge (e.g. ANTLR, PEG, Bison)
* Extensible to add new date, time, and duration expressions - PRs welcome!
* Extensible to add additional human languages - PRs welcome!
* Scalable - adding languages or rules has no additional performance cost.
* MIT License - do whatever you want with it!

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

## Language Support

| Variable | Language |
|---|---|
| `nowandlater.LangEn` | English (default) |
| `nowandlater.LangEs` | Spanish / Español |
| `nowandlater.LangFr` | French / Français |
| `nowandlater.LangDe` | German / Deutsch |
| `nowandlater.LangIt` | Italian / Italiano |
| `nowandlater.LangPt` | Portuguese / Português |
| `nowandlater.LangRu` | Russian / Русский |
| `nowandlater.LangJa` | Japanese / 日本語 |
| `nowandlater.LangZh` | Chinese / 中文 |

```go
p := nowandlater.Parser{Lang: &nowandlater.LangEs}
t, err := p.Parse("próximo lunes")   // next Monday
t, err := p.Parse("hace 3 días")     // 3 days ago
```

## Supported formats

An incomplete list.

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
1774711545                        (Unix timestamp, seconds since epoch; ≥ 5 digits)
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

An input that is recognisably date-like but genuinely ambiguous returns `ErrAmbiguous`:

```go
t, err := p.Parse("mar 5") // Spanish: "mar" = martes (Tuesday) or marzo (March)?
if errors.Is(err, nowandlater.ErrAmbiguous) {
    // input looks like a date but has multiple valid interpretations;
    // ask the user to be more specific
}
```

## Known limitations

- **Spanish "mar"**: works correctly as a weekday abbreviation in weekday contexts
  (`"el mar pasado"` → last Tuesday). In numeric date position (`"mar 5"`,
  `"5 de mar"`) it is genuinely ambiguous and returns `ErrAmbiguous`. Write
  `"marzo"` to avoid ambiguity.
- **morning / afternoon / evening**: semantics undefined; not yet supported.

## Development tool

The CLI API (args) is under development.  Perhaps it should be a drop-in for Linux's `date`?  Thoughts welcome.

`cmd/nldate` is a CLI for parsing natural-language dates and inspecting the parse pipeline:

```
go run ./cmd/nldate "next Monday at 9:30 AM"
go run ./cmd/nldate -now 2026-03-22 -interval "this week"
go run ./cmd/nldate -unix "in 2 days"
echo "in 2 days" | go run ./cmd/nldate
```

Output shows the token list, signature, parsed period, and resolved time.
Use `-unix` to print only the resolved Unix timestamp (useful for scripting).

## Building and Running Tests

| Make         | Description               |
|--------------|---------------------------|
| `make build` | Builds library            |
| `make test`  | Test and Lint             |
| `make fmt`   | Reformat source code      |
| `make cover` | Coverage tests            |
| `make fuzz`  | Run fuzz tests            |
| `make bench` | Run benchmark tests       |
| `make clean` | Clean up                  |

