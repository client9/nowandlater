{
  "title": "support compressed ISO 8601 format YYYYMMDDThhmmssZ 20260429T030444Z",
  "id": "20260429T033201Z-02b4024d",
  "state": "done",
  "created": "2026-04-29T03:32:01Z",
  "labels": [],
  "assignees": [],
  "milestone": "",
  "projects": [],
  "template": "",
  "events": [
    {
      "ts": "2026-04-29T03:32:01Z",
      "type": "filed",
      "to": "backlog"
    },
    {
      "ts": "2026-04-29T18:11:56Z",
      "type": "moved",
      "from": "backlog",
      "to": "active"
    },
    {
      "ts": "2026-04-29T18:13:23Z",
      "type": "moved",
      "from": "active",
      "to": "done"
    }
  ]
}

currently supported is standard ISO 8601 format '2026-04-29T03:04:44Z'

```sh
$ go run cmd/nldate/main.go '2026-04-29T03:04:44Z'
input:     "2026-04-29T03:04:44Z"
signature: "YEAR INTEGER INTEGER TIME TIMEZONE"
tokens:
  [0] YEAR           2026
  [1] INTEGER        4
  [2] INTEGER        29
  [3] TIME           "03:04:44"
  [4] TIMEZONE       "z"
period:    second
now:       2026-04-28T20:32:55-07:00
resolved:  2026-04-29T03:04:44Z
```

However the compressed format is not '20260429T030444Z'  (YYYYMMDDThhmmssZ).  It's a fixed format with 'T' in the middle and 'Z' at the end, total 18 characters.

```sh
$ go run cmd/nldate/main.go '20260429T030444Z'
input:     "20260429T030444Z"
signature: "INTEGER UNKNOWN"
tokens:
  [0] INTEGER        20260429
  [1] UNKNOWN        "030444z"
parse:     error: nowandlater: unknown date signature
```

This is special case fixed format date and time representation.  There is no internationalization or special langauge support needed.

## Scope

Z suffix is required. `YYYYMMDDThhmmss` without a trailing Z (local time) is out of scope for now.

## Root cause

`preprocess()` in `tokenizer.go` already fires the T-separator substitution on this input — the `t` at position 8 is between two digits — producing `20260429 030444z`. The problem is that `20260429` is an 8-digit INTEGER (not a YEAR token), and `030444z` has no tokenizer rule, becoming UNKNOWN. The existing dashed form works because `-` separators produce a proper YEAR + two INTEGERs before the T fires.

## Implementation plan

**One file changed, no new handlers or dispatch entries.**

`internal/engine/tokenizer.go` — add compact ISO detection in `preprocess()` *before* the existing T-separator block:

```go
// Compact ISO 8601: "20260429T030444Z" → "2026-04-29t03:04:44z"
// Rewrite before the T-separator pass so the result goes through the
// normal dashed-ISO path (YEAR INTEGER INTEGER TIME TIMEZONE).
if len(s) == 16 && s[8] == 't' && s[15] == 'z' &&
    allDigits(s, 0, 8) && allDigits(s, 9, 15) {
    s = s[0:4] + "-" + s[4:6] + "-" + s[6:8] +
        "t" + s[9:11] + ":" + s[11:13] + ":" + s[13:15] + "z"
}
```

`allDigits(s string, lo, hi int) bool` — a small private helper using the existing `IsDigitByte`.

After rewriting, the T-separator pass produces `2026-04-29 03:04:44z`, which tokenizes to `YEAR INTEGER INTEGER TIME TIMEZONE` — already dispatched. No resolver changes.

**Tests** — add to `tests/lang_test.go`:
- `"20260429T030444Z"` → `2026-04-29 03:04:44 UTC`
- `"20260429t030444z"` → same (lowercase handled by `preprocess` lowercasing first)

## Resolution

Implemented as designed, with two small deviations from the plan:

- `allDigits(s string, lo, hi int) bool` was not needed — the existing `allDigits(s string) bool` handles the full string, so the detection uses `allDigits(s[0:8])` and `allDigits(s[9:15])` directly.
- Tests landed in `tests/resolve_test.go` (`TestResolveTimezone`) rather than `tests/lang_test.go`, alongside the existing dashed ISO 8601 UTC case — a better structural fit.

What landed:
- `internal/engine/tokenizer.go`: 9-line compact ISO block added to `preprocess()`, before the T-separator pass; docstring updated.
- `tests/resolve_test.go`: two cases added to `TestResolveTimezone` — uppercase Z and lowercase z forms.
- No handler, dispatch, or resolver changes.
