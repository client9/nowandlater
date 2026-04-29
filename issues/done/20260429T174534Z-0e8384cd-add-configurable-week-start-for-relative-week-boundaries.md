{
  "title": "Add configurable week start for relative week boundaries",
  "id": "20260429T174534Z-0e8384cd",
  "state": "done",
  "created": "2026-04-29T17:45:34Z",
  "labels": [
    "feature",
    "design"
  ],
  "assignees": [],
  "milestone": "",
  "projects": [],
  "template": "",
  "events": [
    {
      "ts": "2026-04-29T17:45:34Z",
      "type": "filed",
      "to": "backlog"
    },
    {
      "ts": "2026-04-29T17:56:56Z",
      "type": "moved",
      "from": "backlog",
      "to": "active"
    },
    {
      "ts": "2026-04-29T17:58:33Z",
      "type": "moved",
      "from": "active",
      "to": "done"
    }
  ]
}

## Context
`next week`, `last week`, `this week`, and `ParseInterval("this week")` currently assume Monday as the start of the week. That is hard-coded in the resolver and in the interval boundary helpers.

## Problem
Some consumers need a different calendar convention, especially Sunday-start week numbering. Python `dateparser` exposes a similar configuration, and this library should probably offer the same capability instead of baking Monday into every week-based resolution.

## Proposed behavior
Add a parser-level or resolver-level configuration for the start of week, with an initial supported set of values like:

- Monday
- Sunday

That setting should affect:

- `next week`
- `last week`
- `this week`
- week intervals returned by `ParseInterval`
- week alignment used by `EndOf(..., PeriodWeek)`

It should not change bare weekday resolution such as `Monday` or `next Monday`.

## Design note
This is a calendar-boundary concern, not an ambiguity policy concern. It should stay separate from the existing scheduling/historical/strict ambiguity presets.

## Design decisions

**Public API: `WeekStartSunday bool` on `Parser`.**

```go
type Parser struct {
    Lang          *Lang
    Location      *time.Location
    Now           func() time.Time
    Ambiguity     AmbiguityConfig
    WeekStartSunday bool  // false (default) = Monday-start; true = Sunday-start
}
```

`time.Weekday` was rejected: its zero value is `time.Sunday`, which would silently break the Monday default. `WeekStartSunday bool` is zero-value-safe — `false` means Monday, which is correct for Europe and most of Asia.

The field is orthogonal to `AmbiguityConfig` and must not become a preset on it; this is a calendar convention setting, not an ambiguity policy.

**Internal plumbing: `ResolvePolicy` gets a matching `WeekStartSunday bool`.** `Parser.Parse` and `Parser.ParseInterval` build the policy from `Ambiguity` and then inject the flag:

```go
policy := p.Ambiguity.resolvePolicy()
policy.WeekStartSunday = p.WeekStartSunday
```

## Implementation plan

1. **`internal/engine/resolve.go` — add `WeekStartSunday bool` to `ResolvePolicy`.**

2. **`resolveDirectionAnchor`** — replace the hard-coded `currentMonday` closure with a `currentWeekStart` closure parameterised by policy:
   - Monday formula: `(int(now.Weekday()) + 6) % 7`
   - Sunday formula: `int(now.Weekday())`

3. **`startOfPeriod`** — same formula change for the `PeriodWeek` case. Function signature gains a `policy ResolvePolicy` parameter (or a `bool`); update all call sites (there are few).

4. **`EndOf`** — no change needed; it adds 7 days from the aligned start, so it tracks whatever start-of-week was used.

5. **`parser.go`** — add `WeekStartSunday bool` to `Parser`; inject into policy in both `Parse` and `ParseInterval`.

6. **Tests** — add cases to `tests/resolve_test.go` covering `next/last/this week` and `ParseInterval("this week")` under Sunday-start, using the existing `now = 2026-03-22` (a Sunday) reference time.
