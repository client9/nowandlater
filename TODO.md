# TODO

- [ ] **morning / afternoon / evening / night** — semantics unclear
- [ ] **start / end / beginning of week/month/year** — needs `Anchor` field in
      `ParsedDateSlots`; design the resolver behaviour first
- [ ] Support "later" in English.  "Two days later", etc.
- [ ] **Unix timestamp** — 10-digit integer (seconds since epoch); easy to detect
      in number sub-parser (`len == 10 && allDigits`), resolver just wraps it
- [ ] **2-digit year** — partially done in some fixed formats. `26` → infer century (e.g. ±50 years from now window)
- [ ] **before / after relative to a non-now anchor plus a time**
      e.g. "3 hours after noon" — complex; resolver needs to combine time + delta
- [ ] **per-language preprocessing step** — the global dotted-abbreviation normalizer
      strips trailing dots before tokenization, so "sept." and "sept" are identical by
      the time the Words map is consulted. This causes silent wrong answers in two
      confirmed cases:
        - French "10 sept. 2026" → INTEGER(7) → July 10 instead of September 10
        - Portuguese "10 dez. 2026" → INTEGER(10) → October 10 instead of December 10
      Note: Portuguese "set.", "out.", "nov." are unaffected — their bare forms ("set",
      "out", "nov") map unambiguously to months, so stripping the dot causes no harm.
      Only "dez." is broken because "dez" collides with the number 10. In formal PT
      documentation the dotted form "dez." is standard, making this a realistic input.
      The simplest fix is a per-language substitution pass that runs *before*
      normalization (e.g. "sept." → "septembre", "dez." → "dezembro"). The general
      case may need a `Lang.Preprocess func(string) string` hook so each language can
      expand ambiguous dotted abbreviations before the shared pipeline runs.
