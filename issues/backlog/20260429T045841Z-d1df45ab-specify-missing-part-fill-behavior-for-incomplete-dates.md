{
  "title": "Specify missing-part fill behavior for incomplete dates",
  "id": "20260429T045841Z-d1df45ab",
  "state": "backlog",
  "created": "2026-04-29T04:58:41Z",
  "labels": [
    "design"
  ],
  "assignees": [],
  "milestone": "",
  "projects": [],
  "template": "",
  "events": [
    {
      "ts": "2026-04-29T04:58:41Z",
      "type": "filed",
      "to": "backlog"
    }
  ]
}

## Context

Issue `9239a6e0` started as a general discussion about ambiguous date resolution, but it mixes two different problems:

1. occurrence selection / direction for underspecified dates
2. filling missing parts for incomplete absolute dates

The occurrence-selection part should stay in `9239a6e0`.

This issue is for the second part only: when some calendar components are missing but the input already identifies a year/month frame.

## Inputs in scope

Examples:

- `December 2015` — month + year, no day
- `2015` — year only
- possibly other forms that are clearly absolute but leave components unspecified

Examples that are **not** in scope here:

- `5 hours`
- `monday`
- `October`
- `March 5` when the question is which year occurrence to choose

Those belong to occurrence-selection policy, not missing-part fill policy.

## Problem statement

These inputs do not primarily have a past/future ambiguity. Instead, they are missing calendar components.

Examples:

- `December 2015` is already pinned to a specific month and year. The question is what day to fill in.
- `2015` is already pinned to a specific year. The question is what month/day to fill in.

Possible fill policies include:

- first value of missing component
- current value from the reference date context
- last value of missing component
- reject as too ambiguous

## Design questions

1. Should missing-part filling be configurable at all, or should the library choose one canonical interpretation?
2. For `December 2015`, should the default be first day, current day-of-month, or last day?
3. For `2015`, should the default be Jan 1, current month/day, Dec 31, or rejection?
4. Should strict mode reject these inputs instead of filling missing parts?
5. Should fill policy be independent from occurrence/direction presets such as scheduling vs historical?

## Draft direction

Keep this as a separate concern from occurrence-selection.

A likely internal model is:

- missing day-of-month policy
- missing month-of-year policy
- possibly strict rejection for incomplete absolute dates

This should not be folded back into the scheduling/historical direction presets without careful thought, because these are different ambiguities.
