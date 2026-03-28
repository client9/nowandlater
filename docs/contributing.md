# Contributing

## Adding a new language

File naming convention: `languages/lang_<iso639-1>.go` / `tests/lang_<iso639-1>_test.go`
(e.g. `languages/lang_fr.go` / `tests/lang_fr_test.go`).

1. Create `languages/lang_<code>.go` with a `var MyLang = Lang{Words: myWords, OrdinalSuffixes: [...]}`
2. Map everything into `Words map[string]WordEntry`:
   - Regular words: `"lunes": {TokenWeekday, WeekdayMonday}`
   - Multi-word phrases (space-containing keys; matched longest-first):
     `"il y a": {TokenModifier, ModifierPast}`
   - Time-word substitutes: `"midi": {TokenTime, "12:00"}`
   - Number words: `"cinq": {TokenInteger, "5"}`, `"vingt et un": {TokenInteger, "21"}`
   - Omit number words that conflict with TokenUnit (e.g. Spanish "segundo")
   - Do NOT add timezone abbreviations — they are recognised globally via `defaultTimezones`
     in `internal/engine/timezone.go`. Add to `lang.Timezones` only to override an
     ambiguous abbreviation (e.g. IST as Irish vs India vs Israel).
3. Set `OrdinalSuffixes` for digit-based ordinal stripping
   (e.g. `["er", "ème"]` for French "1er", "2ème" → "1", "2")
4. Set `DateOrder` if the locale uses day-first numeric dates: `DateOrder: DMY`.
   Most of Europe and Latin America use DMY. Leave unset (MDY) for US English.
5. If the language uses reversed word order, add handlers to `lang.Handlers`
   (e.g. "WEEKDAY DIRECTION" for "lunes próximo").
6. Create `tests/lang_<code>_test.go` with a `[]struct{input string; want time.Time}` table
   and a `TestMyLang` function mirroring the pattern in `tests/lang_es_test.go`.

See `languages/lang_en.go` / `languages/lang_es.go` and their test files as reference
implementations.

## Adding a new handler

1. Write `func handleFoo(tokens []Token) (*ParsedDateSlots, error)` in
   `internal/engine/handlers.go`.
2. Use `filterFillers(tokens)` to strip FILLER tokens before indexing.
3. Use `mustAtoi(token.Value)` and `mustParseTime(token.Value)` — these panic on
   invalid input. Token values are always well-formed after tokenization, so panics
   indicate a programming error (wrong handler wired to wrong signature), not user
   error.
4. Register the signature string in `internal/engine/dispatch.go`'s `handlers` map.
   - If a signature is recognisably date-like but genuinely ambiguous (e.g. a
     weekday abbreviation that also abbreviates a month), register `handleAmbiguous`
     in the language's `lang.Handlers` map instead of leaving the signature
     unregistered. This gives callers `ErrAmbiguous` rather than the misleading
     `ErrUnknownSignature`. See `languages/lang_es.go` for an example.
