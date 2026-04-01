# nowandlater — developer notes for Claude

## What this is

A self-contained, zero-dependency Go library for parsing natural-language date/time
strings into a concrete `time.Time`. No external packages; no regex engine beyond the
standard library. Module path: `github.com/client9/nowandlater`.

## Source layout

```
nowandlater/
├── parser.go          # Parser — public entry point (Parse, ParseInterval)
├── lookup.go          # Language lookup helpers exposed at package level
├── doc.go             # Package-level godoc
├── bench_test.go      # Benchmarks
├── internal/engine/   # Core pipeline: tokenizer, handlers, dispatch, resolve, slots, timezone
├── languages/         # One file per human language (lang_en.go, lang_es.go, …)
├── tests/             # All tests — language, handler, tokenizer, resolve
├── cmd/nldate/        # CLI dev tool (go run ./cmd/nldate "next Monday")
└── docs/              # Design and contributor docs
```

## Pipeline summary

```
input → preprocess → normalize → Tokenize → Signature → dispatch → handler → Resolve → time.Time
```

Full pipeline details, key types, and the Token.Value type table are in
`docs/architecture.md`.

## Testing conventions

- Reference "now": `2026-03-22 10:00:00 UTC` (a Sunday) — used across all test packages.
- Helper `u(year, month, day, hour, min, sec)` constructs UTC times (defined in `tests/resolve_test.go`).
- Test packages: `tests/lang_test.go`, `tests/lang_*_test.go`, `tests/resolve_test.go`, `tests/dispatch_test.go`, `tests/tokenizer_test.go`.
- Run all tests: `go test ./...`

## Adding languages and handlers

See `docs/contributing.md` for step-by-step guides.

## Deferred / planned work

See `TODO.md` for the full list. Highlights:
- "morning/afternoon/evening" time-of-day ranges (semantics unclear)
- "start/end of week/month/year" (needs new `Anchor` field design)
- Per-language preprocessing hook (`Lang.Preprocess`) for dotted abbreviations (French "sept.", Portuguese "dez.")
