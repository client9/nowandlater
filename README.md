![Now-and-Later-Wrapper-Small](https://github.com/user-attachments/assets/6095d42b-a443-4363-8b73-29cbdf297725)

# nowandlater - fast natural date and time parsing

A self-contained, zero-dependency Go library for parsing natural-language
date and time strings into `time.Time` values.

[![Go Reference](https://pkg.go.dev/badge/github.com/client9/nowandlater.svg)](https://pkg.go.dev/github.com/client9/nowandlater)

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

### How does it work?

Read the CLAUDE.md summary, but simply it:

* Turns input into a list of tokens, e.g. "Monday" --> `{ WEEKDAY, Monday }`. This is 99% of variations between different human languages.
* It's data driven -- the tokens are in a map (e.g. `lang_en.go`)
* The list of tokens is a "signature" - across all langauges there under 100 signatures (see dispatch.go)
* The signature is used to find a handler function using a `map`.
* The handler converts the signature into a golang struct holding numeric values for day, month, hour, etc. The handler `func` is normally a few lines long.
* The struct is converted into go a `time.Time` object.

### Can I add another human languages?

Yes. Look at lang_en.go (English) or lang_es.go (Spanish). Translate that, run tests, perhaps add a handler. Done.

Your favorite coding AI assistant can do most of it in 5 minutes.

### Is it fast?

It can parse a date and time snippet in about 500ns.

Python's dateparser, takes about 1,000,000ns. That's 2000x faster.

Go's native time parser on a fixed format is 20ns. That's 25x slower.


### Can it be ported to another (computer) lanaguge?

Please do so!  It should be easy.  Let me know if you need help.

### Can't AI LLMs do this?

In 2026, LLMs are good at extracting and converting an arbitrary text into date snippets:

```
Tomorrow, let's get ice cream at 3pm"
```

into

```
tomorrow at 3pm
```

but they are not very good at converting that to a standard time format or converting to a native datetime object.  There are a lot of off by one errors, time zone issues, and confusion.

Also it's very slow (relative to converting locally).

### What about other Go libraries?

The existing projects in Go are mostly abandoned or obsolete.  As of 2026:

* [go-anytime](https://github.com/ijt/go-anytime) - last commit in 2023, English only
* [naturaltime](https://github.com/Sho0pi/naturaltime) - last commit in 2025, runs a javacript interpreter (!!) to execute[chrononode](https://github.com/wanasit/chrono/tree/master) which is pile of regexp.
* [when](https://github.com/olebedev/when) - last commit in 2025, complicated rule set using regexp.
* [go-naturaldate](https://github.com/tj/go-naturaldate) - last commit in 2020, parser/generator PEG based.  English only
* [naturaldate](https://github.com/anatol/naturaldate.go) - a fork of previous
* [go-dateparser](https://github.com/markusmobius/go-dateparser) - direct port of python's dateparser, last commit 1 year, has not kept up to date
* [araddon/dateparse](https://github.com/araddon/dateparse) - last commit in 2021, more for fixed computer formats
* [shadiestgoat/dateparse](https://github.com/shadiestgoat/dateparse) - last commit in 2024, fixed formats

All (?) use regexp soup or some parser/generator.  Go needs something better.

### What about porting code from other prgramming langauges or libraries?

There are good libraries in other programming languages, some with unique features:

* Python [dateparser](https://github.com/scrapinghub/dateparser)  - actively maintained, regexp based.
* Typescript [chrononode](https://github.com/wanasit/chrono/tree/master) - complicated, but has a lot of interesting features.
* Ruby [chronic](https://github.com/mojombo/chronic) - English only, last commit in 2023.
* Java [natty](https://github.com/joestelmach/natty) - Last commit in 2017 - English only - parser/generator
* GNU date has some interesting date parsing abilities

But all are either English-only or regexp/parser/generator based. Let me know if I'm missing one.

None are really good to port to Go.  And it's hard to keep track up subsequent changes.

### What's with the name?

[Now and Later](https://en.wikipedia.org/wiki/Now_and_Later) is a classic American candy.

---

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
- **Unix timestamps**: 10-digit integers are not yet recognised.
- **morning / afternoon / evening**: semantics undefined; not yet supported.

## Development tool

`cmd/nldate` is a CLI for parsing natural-language dates and inspecting the parse pipeline:

```
go run ./cmd/nldate "next Monday at 9:30 AM"
go run ./cmd/nldate -now 2026-03-22 -interval "this week"
go run ./cmd/nldate -unix "in 2 days"
echo "in 2 days" | go run ./cmd/nldate
```

Output shows the token list, signature, parsed period, and resolved time.
Use `-unix` to print only the resolved Unix timestamp (useful for scripting).

## Running tests

See Makefile

```
go test ./...
```
