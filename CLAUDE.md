# nowandlater — developer notes for Claude

## What this is

A self-contained, zero-dependency Go library for parsing natural-language date/time
strings into a concrete `time.Time`. No external packages; no regex engine beyond the
standard library. Module path: `github.com/client9/nowandlater`.

## Pipeline

```
input string
    │
    ▼ preprocess(input, lang)
    │   • Lowercases
    │   • ISO T-separator: digit-T-digit → digit-space-digit
    │   • Dot-time normalization: "7.15pm" → "7:15pm"
    │   (time-word and number-word substitutions are NOT done here —
    │    they are handled by the tokenizer's Words map lookup below)
    │
    ▼ normalize(s, lang)
    │   • Lowercases, strips OrdinalSuffixes ("3rd"→"3")
    │   • Expands dotted abbreviations ("a.m."→"am")
    │   • Collapses whitespace/commas to single spaces
    │
    ▼ lang.Tokenize(s) → []Token
    │   • Phrase lookahead first (longest match, span≥2): lang.Words space-keyed entries
    │   • Then single-word: lang.Words, then classifyNumber
    │   • TokenTime substitutions ("noon"→TokenTime "12:00") and number words
    │     ("five"→TokenInteger 5) are expressed as Words entries matched here
    │   • Unknown words → TokenUnknown (not an error — ignored by Signature)
    │
    ▼ Signature(tokens) → string   (FILLER tokens excluded)
    │   e.g. "MODIFIER INTEGER UNIT", "ANCHOR PREP TIME AMPM"
    │
    ▼ Lang.Parse dispatches to handler   (dispatch.go)
    │   • lang.Handlers checked first (language-specific overrides)
    │   • global handlers map as fallback (language-neutral patterns)
    │   • No match → ErrUnknownSignature
    │
    ▼ handler(tokens) → *ParsedDateSlots
    │
    ▼ Resolve(slots, now) → time.Time          (point in time)
    ▼ ResolveInterval(slots, now) → [start, end)  (calendar interval)
```

## Key types

**`Parser`** (`parser.go`) — the primary high-level entry point. Configure once,
call `Parse` or `ParseInterval` repeatedly. Zero value is valid.
- `Lang *Lang` — language to use; nil → `&English`
- `Location *time.Location` — default timezone when input has none; nil → `time.Local`
- `Now func() time.Time` — clock source for relative expressions; nil → `time.Now`
- `Parse(input string) (time.Time, error)` — single-call parse + resolve
- `ParseInterval(input string) (start, end time.Time, err error)` — single-call parse + interval resolve

**`ParsedDateSlots`** (`slots.go`) — intermediate parse result. All fields are
pointers; nil means "not present". Key fields:
- `Year/Month/Day/Weekday/Hour/Minute/Second *int` — absolute components
- `DeltaSeconds *int` — pre-normalized relative offset in seconds
- `Direction *string` — "future" | "past" | "nearest"
- `Anchor *string` — "week" | "month" | "year" (for "next week" etc.)
- `Period string` — coarsest unit present ("day", "hour", "second", …)
- `Location *time.Location` — parsed timezone; nil → use now.Location()

**`Token`** (`tokenizer.go`) — `{Type TokenType, Value any}`. Value holds a typed
constant for semantic tokens, keeping handlers language-neutral. Raw/numeric tokens
carry strings. See the table below for per-type values.

**`Lang`** (`lang.go`) — language configuration:
- `Words map[string]WordEntry` — all word/phrase lookups in one map: single words, multi-word phrases (space-containing keys, matched longest-first), time-word substitutes (`"noon": {TokenTime, "12:00"}`), and number words (`"five": {TokenInteger, "5"}`)
- `OrdinalSuffixes []string` — suffixes to strip from trailing digits ("st","nd"…)
- `DateOrder DateOrder` — controls interpretation of ambiguous all-numeric dates (`INTEGER INTEGER YEAR`): `MDY` (default, US English), `DMY` (Europe/Latin America), `YMD`. Only affects this one signature; ISO year-first forms (`YEAR INTEGER INTEGER`) and letter-month forms (`INTEGER MONTH YEAR`) are always unambiguous. Spanish sets `DMY`.
- `Timezones map[string]*time.Location` — override/extend built-in tz table
- `Handlers map[string]Handler` — language-specific signature overrides (checked before `DateOrder` dispatch and before the global handlers map)

## Token.Value types

Handlers extract values via type assertion. Language files use the same typed
constants — no magic strings required.

| TokenType      | Value type   | Constants / examples                             |
|---|---|---|
| TokenWeekday   | `Weekday`    | `WeekdayMonday` … `WeekdaySunday`                |
| TokenMonth     | `Month`      | `MonthJanuary` … `MonthDecember`                 |
| TokenDirection | `Direction`  | `DirectionFuture`, `DirectionPast`, `DirectionNearest` |
| TokenModifier  | `Modifier`   | `ModifierFuture` (+1), `ModifierPast` (-1)       |
| TokenAnchor    | `Anchor`     | `AnchorNow`, `AnchorToday`, `AnchorTomorrow`, `AnchorYesterday`, `Anchor2DaysAgo`, `Anchor2DaysFromNow`, `Anchor3DaysAgo`, `Anchor3DaysFromNow` |
| TokenUnit      | `Period`     | `PeriodSecond` … `PeriodYear`                    |
| TokenAMPM      | `AMPM`       | `AMPMAm`, `AMPMPm`                               |
| TokenPrep      | `nil`        | (value never consumed)                           |
| TokenFiller    | `nil`        | (value never consumed; filtered before dispatch) |
| TokenInteger   | `int`        | e.g. `3`, `15`                                   |
| TokenInteger2  | `int`        | leading-zero 2-digit, e.g. `3` (from "03"); type encodes the leading zero |
| TokenYear      | `int`        | 4-digit year, e.g. `2026`                        |
| TokenTime      | `string`     | `"H:MM"` or `"H:MM:SS"`, e.g. `"9:30"`          |
| TokenDecimal   | `float64`    | e.g. `3.5`, `1.5`, `0.25` — dot-separated all-digit pairs without AM/PM |
| TokenTimezone  | `string`     | lowercase abbreviation or numeric offset; recognised globally (not per-language) via `defaultTimezones` in `timezone.go` |
| TokenUnknown   | `string`     | the raw word                                     |

## Adding a new language

File naming convention: `lang_<iso639-1>.go` / `lang_<iso639-1>_test.go`
(e.g. `lang_fr.go` / `lang_fr_test.go`). The `lang_` prefix groups all language
files alphabetically between `lang.go` (framework) and the test files.

1. Create `lang_<code>.go` with a `var MyLang = Lang{Words: myWords, OrdinalSuffixes: [...]}`
2. Map everything into `Words map[string]WordEntry`:
   - Regular words: `"lunes": {TokenWeekday, WeekdayMonday}`
   - Multi-word phrases (space-containing keys; matched longest-first):
     `"il y a": {TokenModifier, ModifierPast}`
   - Time-word substitutes: `"midi": {TokenTime, "12:00"}`
   - Number words: `"cinq": {TokenInteger, "5"}`, `"vingt et un": {TokenInteger, "21"}`
   - Omit number words that conflict with TokenUnit (e.g. Spanish "segundo")
   - Do NOT add timezone abbreviations — they are recognised globally via `defaultTimezones` in `timezone.go`. Add to `lang.Timezones` only to override an ambiguous abbreviation (e.g. IST as Irish vs India vs Israel).
3. Set `OrdinalSuffixes` for digit-based ordinal stripping
   (e.g. `["er", "ème"]` for French "1er", "2ème" → "1", "2")
4. Set `DateOrder` if the locale uses day-first numeric dates: `DateOrder: DMY`.
   Most of Europe and Latin America use DMY. Leave unset (MDY) for US English.
5. If the language uses reversed word order, add handlers to `lang.Handlers`
   (e.g. "WEEKDAY DIRECTION" for "lunes próximo").
6. Create `lang_<code>_test.go` with a `[]struct{input string; want time.Time}` table
   and a `TestMyLang` function mirroring the pattern in `lang_es_test.go`.

See `lang_en.go` / `lang_es.go` and their test files as reference implementations.

## Adding a new handler

1. Write `func handleFoo(tokens []Token) (*ParsedDateSlots, error)` in `handlers.go`.
2. Use `filterFillers(tokens)` to strip FILLER tokens before indexing.
3. Use `mustAtoi(token.Value)` and `mustParseTime(token.Value)` — these panic on
   invalid input. Token values are always well-formed after tokenization, so panics
   indicate a programming error (wrong handler wired to wrong signature), not user
   error.
4. Register the signature string in `dispatch.go`'s `handlers` map.

## mustAtoi / mustParseTime

Handlers use panic-based helpers rather than `(value, bool)` returns because token
values are guaranteed well-formed by the tokenizer. The panic documents the invariant
and eliminates unreachable error branches that suppress coverage.

## withPrepTime combinator

`withPrepTime(base Handler) Handler` wraps any date-only handler to also accept a
trailing `PREP TIME`, `PREP TIME AMPM`, or `PREP INTEGER AMPM` suffix. Register all
three variants in `dispatch.go`. This avoids duplicating date+time handler logic.

## Testing conventions

- `resolveNow` in `lang_test.go` / `parse_test.go`: `2026-03-22 10:00:00 UTC` (a Sunday).
- `spNow` in `lang_es_test.go`: `2026-03-22 10:00:00 UTC` (same date, for Spanish tests).
- `parserNow` in `parser_test.go`: `2026-03-22 10:00:00 UTC` (same reference).
- Helper `u(year, month, day, hour, min, sec)` constructs UTC times (defined in `resolve_test.go`).
- Run all tests: `go test ./...`

## Known limitations / design decisions

- **"second" conflict**: English ordinal "second" (2nd) is omitted from
  `englishNumberWords` because it collides with `TokenUnit "second"`. "Second of
  March" does not parse; use "2nd of March".
- **Spanish "mar"**: resolves to Tuesday (martes), not March. Write "marzo" in full.
- **Spanish "segundo/segunda"**: omitted from `spanishNumberWords` for the same unit
  conflict reason.
- **Phrase keys must be normalized** (lowercase, no punctuation) to match the output
  of `normalize`. Hyphenated words like "avant-hier" are single chunks after
  `strings.Fields` and belong in `Words`, not `Phrases`.

## Interval support

`resolve.go` exports two interval functions:

- `EndOf(start time.Time, period Period) time.Time` — returns the exclusive end of the
  period starting at `start`. `start` must be calendar-aligned (e.g. midnight for
  `PeriodDay`). Returns `[start, end)` where end is the first moment of the next period.
- `ResolveInterval(slots *ParsedDateSlots, now time.Time) (start, end time.Time, err error)` —
  resolves `slots` to a half-open calendar interval. `start` is always calendar-aligned
  (may differ from `Resolve` for delta-path expressions like "tomorrow"). Uses Monday
  as the week boundary.

`Parser.ParseInterval` is the single-call convenience wrapper.

**Period → end computation:**

| Period | End |
|---|---|
| `second` | `start + 1s` |
| `minute` | `start + 1m` |
| `hour` | `start + 1h` |
| `day` | `start.AddDate(0,0,1)` |
| `fortnight` | `start.AddDate(0,0,14)` |
| `week` | `start.AddDate(0,0,7)` |
| `month` | first of next month |
| `year` | Jan 1 of next year |

## Deferred / planned work

See `TODO.md` for the full list. Highlights:
- DST-aware timezone resolution
- "morning/afternoon/evening" time-of-day ranges
- "start/end of week/month/year"
- Unix timestamp (10-digit integer)
