{
  "title": "make configuable if time offset is forwards or backwards",
  "id": "20260429T033753Z-9239a6e0",
  "state": "backlog",
  "created": "2026-04-29T03:37:53Z",
  "labels": [],
  "assignees": [],
  "milestone": "",
  "projects": [],
  "template": "",
  "events": [
    {
      "ts": "2026-04-29T03:37:53Z",
      "type": "filed",
      "to": "backlog"
    }
  ]
}

Allow configuration if a relative date is in the future (forwards) or in the past (backwards).

The current system was designed for scheduling which is in the future.  So when given a duration-like input, it resolves to a future time:

```sh
% go run cmd/nldate/main.go '5 hours'
input:     "5 hours"
signature: "INTEGER UNIT"
tokens:
  [0] INTEGER        5
  [1] UNIT           hour
period:    hour
now:       2026-04-28T20:38:15-07:00
resolved:  2026-04-29T01:38:15-07:00  <-- in future
```

This also should apply raw weekdays, but somehow it's defaulting to the past.  I would have expected this to be "next monday" in the future:

```
% go run cmd/nldate/main.go 'monday'
input:     "monday"
signature: "WEEKDAY"
tokens:
  [0] WEEKDAY        monday
period:    day
now:       2026-04-28T20:40:05-07:00
resolved:  2026-04-27T00:00:00-07:00  <-- in past ???
```

The resolution direction (backwards or forwards) should be configurable in the `type Parser struct` in parser.go.

pythons's [dateparser](https://dateparser.readthedocs.io/en/latest/settings.html#handling-incomplete-dates) calls these "incomplete dates" and has a number of settings.

This should be investigated.

## Notes: categorization of ambiguous dates

The original issue statement mixes several different kinds of ambiguity. For design purposes they should be separated:

### 1. Durations

Examples: `5 hours`, `2 days`, `5 years`

These are direction-only ambiguities. They can sensibly mean:

- `past`
- `future`

They do not have a meaningful `nearest` or `current` interpretation.

### 2. Calendar occurrences

Examples: `monday`, `October`

These refer to recurring positions in the calendar. They can sensibly mean:

- `past`
- `future`
- `nearest`
- `current_period` / `current`

Note: `current` is not the same as `nearest`.

### 3. Incomplete dates

Examples:

- `March 5` / `5 March` (day + month, no year)
- `December 2015` (month + year, no day)

This category actually contains two different problems:

- occurrence selection: e.g. which year should `March 5` resolve to?
- component filling: e.g. which day should `December 2015` resolve to?

So "incomplete dates" is a bucket containing at least two separate policy decisions.

### 4. Bare year

Example: `2025`

This is a missing-component fill problem, not an occurrence-selection problem.

Possible interpretations include:

- canonical start: `2025-01-01`
- canonical end: `2025-12-31`
- current fill: `2025-04-28` if today is April 28
- reject as too ambiguous

### Recommended framing

For this issue, the scope should be limited to occurrence/direction ambiguity:

- durations
- bare weekdays
- bare months
- possibly month/day without year

Missing-component fill behavior such as `December 2015` and bare `2025` should be handled in a separate issue.

## Design summary

After discussion, the scope of this issue should be narrowed and the API should stay simple.

### Scope for this issue

This issue should address occurrence/direction ambiguity for underspecified inputs such as:

- durations: `5 hours`, `2 days`, `5 years`
- bare weekdays: `monday`
- bare months: `October`
- possibly month/day without year: `March 5`

This issue should **not** address missing-component fill behavior such as:

- `December 2015` (which day should be filled?)
- `2025` (which month/day should be filled?)

Those are separate problems and should be tracked separately.

### Simplified public API

Do not expose a large matrix of fine-grained ambiguity settings yet.

Instead, expose:

```go
type Parser struct {
    Lang      *Lang
    Location  *time.Location
    Now       func() time.Time
    Ambiguity AmbiguityConfig
}
```

`AmbiguityConfig` should be exported so callers can pass preset values around, but its internal fields can remain unexported for now.

The intended public usage is:

```go
Parser{Ambiguity: AmbiguityScheduling}
Parser{Ambiguity: AmbiguityHistorical}
Parser{Ambiguity: AmbiguityStrict}
```

No public fine-grained knobs are needed in this version. Those can be added later if a real use case appears.

### Draft preset values

Export named preset values:

```go
var (
    AmbiguityScheduling AmbiguityConfig
    AmbiguityHistorical AmbiguityConfig
    AmbiguityStrict     AmbiguityConfig
)
```

The zero value of `Parser` should default to one of these presets, most likely `AmbiguityScheduling`.

### Internal model (not necessarily public yet)

Internally, the ambiguity system can still distinguish:

- duration direction (`past` / `future`)
- calendar occurrence selection (`past` / `future` / maybe strict rejection)
- partial date occurrence selection (`past` / `future` / maybe strict rejection)
- missing-component fill policies (to be handled in a separate issue)

However, `nearest` should be dropped from the public presets for now. It is not a target design mode for this version unless a real need emerges later.

## Preset behavior table

This table is a draft reference for intended behavior.

| Input form | Scheduling | Historical | Strict |
| --- | --- | --- | --- |
| `5 hours` | future | past | reject or require explicit modifier |
| `monday` | future | past | reject or require explicit modifier |
| `October` | future | past | reject or require explicit modifier |
| `March 5` | future occurrence | past occurrence | reject or require explicit year |
| `December 2015` | separate issue | separate issue | separate issue |
| `2025` | separate issue | separate issue | separate issue |

### Notes on defaults

- For scheduling, bare weekdays should likely resolve to `future`, not `nearest`.
- For historical queries such as `--since "5 hours"`, unsigned durations should resolve to `past`.
- Strict mode should prefer rejection over guessing when the input is materially underspecified.

## Internal implementation plan

The public API is intentionally small, but the implementation should still separate the underlying ambiguity decisions clearly.

### 1. Add parser-level ambiguity preset handling

Extend `Parser` with an `Ambiguity` field of type `AmbiguityConfig`.

`AmbiguityConfig` should be an exported type, but its fields can remain unexported for now. Export only named preset values:

- `AmbiguityScheduling`
- `AmbiguityHistorical`
- `AmbiguityStrict`

The zero-value `Parser` should default to the scheduling preset unless there is a strong reason to choose otherwise.

### 2. Thread ambiguity policy into resolution

Today `Parser.Parse` only passes `slots` and `now` into `engine.Resolve`.

The implementation should add a resolver-level policy input rather than trying to encode all behavior changes indirectly in token handlers. That keeps:

- parsing responsible for recognizing structure
- resolution responsible for choosing a concrete occurrence

This policy input should control at least:

- unsigned duration direction
- bare calendar occurrence direction
- partial date occurrence direction
- strict rejection of materially underspecified inputs

### 3. Keep structure-recognition and occurrence-selection separate

Avoid baking scheduling-vs-historical behavior directly into token handlers where possible.

For example:

- `monday` should still parse as a bare weekday
- `October` should still parse as a bare month
- `March 5` should still parse as a partial date

The ambiguity preset should determine how those parsed forms resolve, rather than changing their lexical meaning.

The main exception is strict mode, where some forms may need to be rejected rather than resolved.

### 4. Replace current implicit defaults where needed

Current behavior is inconsistent across categories:

- unsigned durations are implicitly future
- bare weekdays are implicitly nearest
- month/day without year already has a future bias

The new ambiguity system should make these category-level defaults explicit and consistent under each preset.

### 5. Test by input category and preset

Tests should be organized around categories and preset behavior, not just individual phrases in isolation.

At minimum, add coverage for:

- durations under scheduling / historical / strict
- bare weekdays under scheduling / historical / strict
- bare months under scheduling / historical / strict
- month/day without year under scheduling / historical / strict

The important assertions are semantic:

- scheduling chooses future occurrences
- historical chooses past occurrences
- strict rejects ambiguous inputs

### 6. Leave missing-part fill for a separate issue

Do not expand this issue to decide how to fill:

- month + year with missing day
- bare year with missing month/day

That work should be handled in a dedicated follow-up issue so occurrence-selection and missing-component filling do not get conflated again.
