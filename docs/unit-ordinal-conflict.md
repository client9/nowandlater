# Handling unit/ordinal conflicts

## The problem

Some words are simultaneously a **time unit** and an **ordinal number word**:

| Language | Word       | As unit          | As ordinal |
|----------|------------|------------------|------------|
| English  | second     | PeriodSecond     | 2nd        |
| French   | seconde    | PeriodSecond     | 2nd (f.)   |
| Spanish  | segundo/a  | PeriodSecond     | 2nd        |
| Italian  | secondo/a  | PeriodSecond     | 2nd        |
| Portuguese | segundo/a | PeriodSecond    | 2nd        |

If the word is added to a language's `Words` map as `TokenInteger`, expressions like
"in 5 seconds" break because "second" would tokenize as an integer instead of a unit.
If it is mapped to `TokenUnit`, expressions like "march second" fail with
`ErrUnknownSignature`.

## Why the tokenizer cannot resolve this

Tokenization is context-free: a word always maps to the same `Token` regardless of
surrounding tokens. Context-dependent disambiguation would require lookahead in the
tokenizer, which would add significant complexity and break the clean separation
between tokenization and parsing.

## The solution: keep the unit mapping, add ordinal handlers

Leave the word mapped to `TokenUnit` in the `Words` map. This keeps all unit-based
expressions working. Then register additional handlers for the signature patterns
where the word appears in an **ordinal (day-of-month) position**.

The helpers `replaceSecondUnit` and `secondOrdinal` in `handlers.go` perform this
transformation: they replace the first `TokenUnit(PeriodSecond)` in the token list
with `TokenInteger(2)`, then delegate to the normal month/day handler.

The affected signatures (registered in `dispatch.go`) are:

```
MONTH UNIT          →  "march second"
MONTH UNIT YEAR     →  "march second, 2010"
UNIT MONTH          →  "second of march"
UNIT MONTH YEAR     →  "second of march, 2010"
MONTH UNIT PREP TIME / PREP TIME AMPM / PREP INTEGER AMPM   (+ time variants)
MONTH UNIT YEAR PREP TIME / PREP TIME AMPM / PREP INTEGER AMPM
```

These are registered in the **global** `handlers` map, so they apply to every
language automatically.

## Applying this to a new language

If a new language maps its "second" equivalent to `TokenUnit`, the global handlers
already cover the common signatures above — no extra work needed, provided the word
is mapped to `TokenUnit` (not omitted entirely).

If the language uses **reversed word order** for ordinals and the affected signatures
differ, add language-specific overrides to `lang.Handlers` using the same
`secondOrdinal` wrapper:

```go
var MyLang = Lang{
    Words: myWords,
    Handlers: map[string]Handler{
        "UNIT DIRECTION MONTH": secondOrdinal(handleMyReversedHandler),
    },
}
```

## When NOT to use this pattern

If the conflict word exists **only as an ordinal** in that language (no unit homograph),
add it to `Words` as `TokenInteger` normally — no special handling needed.

Similarly, "minute" (English) and its equivalents are rarely used as ordinals ("the
minute of March" is not natural), so no conflict handler is needed for `PeriodMinute`.

## A conflict this pattern cannot resolve: TokenFiller words

Some words are mapped to `TokenFiller` because they serve as articles, conjunctions,
or other grammatical particles in the target language. `TokenFiller` tokens are stripped
from the signature before dispatch — they are **architecturally invisible** to the
handler system. No handler can ever be registered for a signature that includes a filler
word, because filler words never appear in signatures.

Example: German "die" is the most common German article ("the", feminine/plural). It is
also an uncommon abbreviation for *Dienstag* (Tuesday). Mapping it to `TokenWeekday`
would break expressions like "die nächste Woche" (next week). Keeping it as
`TokenFiller` means it is invisible to dispatch — `handleAmbiguous` cannot help here
because there is no signature to register it under.

**Resolution:** document the limitation and require the unambiguous form ("di" or
"dienstag" for Tuesday). Do NOT attempt to register `handleAmbiguous` for filler
conflicts — it is impossible by design.

## A conflict this pattern cannot resolve: weekday/month abbreviations

Spanish "mar" abbreviates both *martes* (Tuesday) and *marzo* (March). This looks
similar but is **not resolvable** with this pattern, for a fundamental reason:

Both "mar" and "martes" produce the same token — `TokenWeekday(WeekdayTuesday)`. By
the time a handler runs, the original word is gone. A handler for `WEEKDAY INTEGER`
that replaced `WeekdayTuesday` with `MonthMarch` would also incorrectly swallow
"martes 5" (Tuesday the 5th), a legitimate Spanish expression.

The unit/ordinal pattern works precisely because the conflicting context is
**unambiguous**: `MONTH UNIT(second)` can only ever mean an ordinal day — no natural
date expression puts a time-unit after a month name. The "mar" case lacks this
guarantee: `WEEKDAY INTEGER` is valid for *both* "mar 5" (March 5th) and "martes 5"
(Tuesday the 5th).

**Resolution:** keep the weekday-wins policy for "mar" and register `handleAmbiguous`
in `Spanish.Handlers` for the affected signatures (`WEEKDAY INTEGER`, `INTEGER WEEKDAY`,
and year variants). This ensures callers receive `ErrAmbiguous` — a signal that the
input was recognised as date-like but could not be resolved — rather than the
misleading `ErrUnknownSignature`. Users should write "marzo" to avoid ambiguity.
