package tests

import (
	//"maps"
	"testing"
	"time"

	"github.com/client9/nowandlater"
	. "github.com/client9/nowandlater/internal/engine"
	. "github.com/client9/nowandlater/languages"
)

// miniLang is a tiny custom language used to verify that the Lang architecture
// works end-to-end without depending on LangEn words. It uses invented words
// that don't appear in LangEn so there's no ambiguity.
//
// Vocabulary:
//
//	"manana"  → ANCHOR  "tomorrow"   (canonical: tomorrow)
//	"ayer"    → ANCHOR  "yesterday"  (canonical: yesterday)
//	"lunes"   → WEEKDAY "monday"     (canonical: monday)
//	"proximo" → DIRECTION "next"     (canonical: next)
//	"en"      → PREP    "in"         (canonical: in)
//	"hace"    → MODIFIER "ago"       (canonical: ago)
//	"dias"    → UNIT    "days"       (canonical: days)
//	"horas"   → UNIT    "hours"      (canonical: hours)
//	"mediodia"→ time substitution → "12:00"
var miniLang = Lang{
	Words: map[string]WordEntry{
		"manana":   {TokenAnchor, AnchorTomorrow},
		"ayer":     {TokenAnchor, AnchorYesterday},
		"lunes":    {TokenWeekday, WeekdayMonday},
		"proximo":  {TokenDirection, DirectionFuture},
		"en":       {TokenPrep, nil},
		"hace":     {TokenModifier, ModifierPast},
		"dias":     {TokenUnit, PeriodDay},
		"horas":    {TokenUnit, PeriodHour},
		"mediodia": {TokenTime, "12:00"}, // time-word substitution as a Words entry
		// AM/PM tokens must be present for time-glued parsing to work
		"am": {TokenAMPM, AMPMAm},
		"pm": {TokenAMPM, AMPMPm},
	},
	OrdinalSuffixes: []string{}, // no ordinal stripping in this mini language
}

// TestLangTokenize verifies that a custom Lang correctly classifies its words
// and stores canonical values in Token.Value.
func TestLangTokenize(t *testing.T) {
	cases := []struct {
		input  string
		tokens []Token
		sig    string
	}{
		{
			"manana",
			tt(TokenAnchor, AnchorTomorrow),
			"ANCHOR",
		},
		{
			"proximo lunes",
			tt(TokenDirection, DirectionFuture, TokenWeekday, WeekdayMonday),
			"DIRECTION WEEKDAY",
		},
		{
			"3 dias hace",
			tt(TokenInteger, 3, TokenUnit, PeriodDay, TokenModifier, ModifierPast),
			"INTEGER UNIT MODIFIER",
		},
		{
			"en 2 dias",
			tt(TokenPrep, nil, TokenInteger, 2, TokenUnit, PeriodDay),
			"PREP INTEGER UNIT",
		},
		{
			// Time substitution: "mediodia" → "12:00" before tokenization
			"mediodia",
			tt(TokenTime, "12:00"),
			"TIME",
		},
		{
			// Unknown words → UNKNOWN token
			"monday",
			tt(TokenUnknown, "monday"),
			"UNKNOWN",
		},
		{
			// Numbers and compound dates are language-neutral
			"2026-12-04",
			tt(TokenYear, 2026, TokenInteger, 12, TokenInteger, 4),
			"YEAR INTEGER INTEGER",
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := miniLang.Tokenize(tc.input)
			if len(got) != len(tc.tokens) {
				t.Fatalf("miniLang.Tokenize(%q)\n  got  %v\n  want %v", tc.input, got, tc.tokens)
			}
			for i, tok := range got {
				if tok != tc.tokens[i] {
					t.Errorf("miniLang.Tokenize(%q) token[%d]\n  got  %+v\n  want %+v",
						tc.input, i, tok, tc.tokens[i])
				}
			}
			sig := Signature(got)
			if sig != tc.sig {
				t.Errorf("Signature(miniLang.Tokenize(%q)) = %q, want %q", tc.input, sig, tc.sig)
			}
		})
	}
}

// TestLangParse verifies that a custom Lang can parse date expressions end-to-end.
// Because Token.Value stores canonical LangEn keys, the existing handlers and
// lookup tables work unchanged.
func TestLangParse(t *testing.T) {
	cases := []struct {
		input   string
		delta   *int
		weekday Weekday
		hour    *int
		minute  *int
		dir     Direction
		period  Period
	}{
		// "manana" → ANCHOR "tomorrow" → delta=+86400, period="day"
		{"manana", new(86400), 0, nil, nil, 0, PeriodDay},
		// "ayer" → ANCHOR "yesterday" → delta=-86400, period="day"
		{"ayer", new(-86400), 0, nil, nil, 0, PeriodDay},
		// "proximo lunes" → DIRECTION "next" + WEEKDAY "monday"
		{"proximo lunes", nil, WeekdayMonday, nil, nil, DirectionFuture, PeriodDay},
		// "3 dias hace" → INTEGER UNIT MODIFIER → delta=-3*86400
		{"3 dias hace", new(-259200), 0, nil, nil, 0, PeriodDay},
		// "en 2 horas" → PREP INTEGER UNIT → delta=+2*3600
		{"en 2 horas", new(7200), 0, nil, nil, 0, PeriodHour},
		// "mediodia" → time substitution → TIME "12:00" → hour=12, minute=0
		{"mediodia", nil, 0, new(12), new(0), 0, PeriodMinute},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := miniLang.Parse(tc.input)
			if err != nil {
				t.Fatalf("miniLang.Parse(%q) error: %v", tc.input, err)
			}
			checkInt(t, tc.input, "DeltaSeconds", got.DeltaSeconds, tc.delta)
			if tc.weekday != 0 && got.Weekday != tc.weekday {
				t.Errorf("miniLang.Parse(%q).Weekday = %v, want %v", tc.input, got.Weekday, tc.weekday)
			}
			checkInt(t, tc.input, "Hour", &got.Hour, tc.hour)
			checkInt(t, tc.input, "Minute", &got.Minute, tc.minute)
			if tc.dir != 0 && got.Direction != tc.dir {
				t.Errorf("miniLang.Parse(%q).Direction = %v, want %v", tc.input, got.Direction, tc.dir)
			}
			if tc.period != 0 && got.Period != tc.period {
				t.Errorf("miniLang.Parse(%q).Period = %q, want %q", tc.input, got.Period, tc.period)
			}
		})
	}
}

// TestLangHandlerOverride verifies that lang.Handlers takes priority over the
// global handlers map. This enables languages with different word order to
// supply their own handlers for signatures that don't exist in LangEn.
//
// Example: French "lundi prochain" (Monday next) produces WEEKDAY DIRECTION —
// the reverse of LangEn DIRECTION WEEKDAY. The lang can supply a handler for
// "WEEKDAY DIRECTION" that reads tokens in the correct order.
func TestLangHandlerOverride(t *testing.T) {
	// frenchLike extends miniLang with a WEEKDAY DIRECTION handler.
	frenchLike := Lang{
		Words: map[string]WordEntry{
			"lundi":    {TokenWeekday, WeekdayMonday},
			"prochain": {TokenDirection, DirectionFuture},
		},
		// Handler for reversed word order: WEEKDAY DIRECTION ("lundi prochain")
		Handlers: map[string]Handler{
			"WEEKDAY DIRECTION": func(tokens []Token) (*ParsedDateSlots, error) {
				toks := FilterFillers(tokens)
				// toks[0]=WEEKDAY, toks[1]=DIRECTION (reversed from LangEn)
				dir := toks[1].Value.(Direction)
				wd := toks[0].Value.(Weekday)
				return &ParsedDateSlots{
					Weekday:   wd,
					Direction: dir,
					Period:    PeriodDay,
				}, nil
			},
		},
	}

	got, err := frenchLike.Parse("lundi prochain")
	if err != nil {
		t.Fatalf("frenchLike.Parse(\"lundi prochain\") error: %v", err)
	}
	if got.Weekday != WeekdayMonday {
		t.Errorf("Weekday = %v, want %v", got.Weekday, WeekdayMonday)
	}
	if got.Direction != DirectionFuture {
		t.Errorf("Direction = %v, want %v", got.Direction, DirectionFuture)
	}
	if got.Period != PeriodDay {
		t.Errorf("Period = %q, want \"day\"", got.Period)
	}
}

// TestLangHandlerFallthrough verifies that language-neutral signatures
// (e.g. ISO date formats) fall through to the global handlers even when
// a lang defines its own Handlers map.
func TestLangHandlerFallthrough(t *testing.T) {

	// Copy English
	langWithHandlers := LangEn

	// over-write with limited handler
	langWithHandlers.Handlers = map[string]Handler{
		"WEEKDAY DIRECTION": func(tokens []Token) (*ParsedDateSlots, error) {
			return &ParsedDateSlots{Period: PeriodDay}, nil
		},
	}

	// ISO date — not in lang.Handlers → should fall through to global handler
	got, err := langWithHandlers.Parse("2026-12-04")
	if err != nil {
		t.Fatalf("Parse(\"2026-12-04\") error: %v", err)
	}
	if got.Year != 2026 {
		t.Errorf("Year = %v, want 2026", got.Year)
	}
	if got.Month != 12 {
		t.Errorf("Month = %v, want 12", got.Month)
	}
	if got.Day != 4 {
		t.Errorf("Day = %v, want 4", got.Day)
	}
}

/* TODO

// TestLangTimezones verifies that Lang.Timezones overrides defaultTimezones and
// that custom abbreviations can be added alongside the built-in table.
func TestLangTimezones(t *testing.T) {
	// IST is ambiguous: built-in default is India (UTC+5:30).
	// Override it to Irish Standard Time (UTC+1) and add a custom "NZST".
	irish := time.FixedZone("IST", 1*3600)
	nzst := time.FixedZone("NZST", 12*3600)

	// englishWords doesn't include NZST, so add it alongside the override.
	customWords := make(map[string]WordEntry, len(englishWords)+1)
	maps.Copy(customWords, englishWords)
	customWords["nzst"] = WordEntry{TokenTimezone, "nzst"}

	customLang := Lang{
		Words: customWords,
		Timezones: map[string]*time.Location{
			"ist":  irish,
			"nzst": nzst,
		},
	}

	// IST should resolve to Irish (UTC+1), not India (UTC+5:30).
	slots, err := customLang.Parse("at 9:30 IST")
	if err != nil {
		t.Fatalf("Parse(\"at 9:30 IST\") error: %v", err)
	}
	if slots.Location == nil {
		t.Fatal("slots.Location is nil, want IST")
	}
	_, offset := time.Unix(0, 0).In(slots.Location).Zone()
	if offset != 1*3600 {
		t.Errorf("IST offset = %d, want %d (Irish, UTC+1)", offset, 1*3600)
	}

	// NZST is a custom abbreviation not in defaultTimezones.
	slots2, err := customLang.Parse("at 9:30 NZST")
	if err != nil {
		t.Fatalf("Parse(\"at 9:30 NZST\") error: %v", err)
	}
	if slots2.Location == nil {
		t.Fatal("slots2.Location is nil, want NZST")
	}
	_, offset2 := time.Unix(0, 0).In(slots2.Location).Zone()
	if offset2 != 12*3600 {
		t.Errorf("NZST offset = %d, want %d (UTC+12)", offset2, 12*3600)
	}
}
*/

// TestLangOrdinalSuffixes verifies that OrdinalSuffixes are applied correctly
// and that a lang with no suffixes leaves numbers unmodified.
func TestLangOrdinalSuffixes(t *testing.T) {
	// miniLang has no ordinal suffixes, so "3rd" stays as "3rd" → UNKNOWN
	toks := miniLang.Tokenize("3rd dias")
	// "3rd" → not a pure digit sequence after 'd', so ClassifyBareInteger gets "3rd"
	// which fails allDigits → UNKNOWN. "dias" → UNIT "days".
	if len(toks) != 2 {
		t.Fatalf("miniLang.Tokenize(\"3rd dias\") got %d tokens, want 2: %v", len(toks), toks)
	}
	if toks[0].Type != TokenUnknown {
		t.Errorf("token[0].Type = %v, want UNKNOWN (no ordinal stripping in miniLang)", toks[0].Type)
	}

	// LangEn strips "rd" → "3" → INTEGER
	engToks := LangEn.Tokenize("3rd of January")
	// Expected: INTEGER "3", FILLER "of", MONTH "january"
	found := false
	for _, tok := range engToks {
		if tok.Type == TokenInteger && tok.Value == 3 {
			found = true
		}
	}
	if !found {
		t.Errorf("LangEn.Tokenize(\"3rd of January\") did not strip ordinal; tokens: %v", engToks)
	}
}

// TestLangNumberWords verifies that number words in Words are recognized as
// TokenInteger, enabling relative expressions like "en una hora" = "in one hour".
func TestLangNumberWords(t *testing.T) {
	numLang := Lang{
		Words: map[string]WordEntry{
			"en":    {TokenPrep, nil},
			"hace":  {TokenModifier, ModifierPast},
			"horas": {TokenUnit, PeriodHour},
			"dias":  {TokenUnit, PeriodDay},
			"am":    {TokenAMPM, AMPMAm},
			"pm":    {TokenAMPM, AMPMPm},
			// number words as TokenInteger entries
			"una":  {TokenInteger, 1},
			"un":   {TokenInteger, 1},
			"dos":  {TokenInteger, 2},
			"tres": {TokenInteger, 3},
		},
	}

	cases := []struct {
		input      string
		wantDelta  int
		wantPeriod Period
	}{
		{"en una horas", 3600, PeriodHour},
		{"en dos horas", 2 * 3600, PeriodHour},
		{"tres dias hace", -3 * 86400, PeriodDay},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := numLang.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			if slots.DeltaSeconds == nil || *slots.DeltaSeconds != tc.wantDelta {
				t.Errorf("DeltaSeconds = %v, want %d", slots.DeltaSeconds, tc.wantDelta)
			}
			if slots.Period != tc.wantPeriod {
				t.Errorf("Period = %q, want %q", slots.Period, tc.wantPeriod)
			}
		})
	}
}

// TestLangPhrases verifies that multi-word phrases (space-containing keys in Words)
// are matched before single-word lookup, with longer phrases taking priority.
func TestLangPhrases(t *testing.T) {
	phraseLang := Lang{
		Words: map[string]WordEntry{
			"dans":  {TokenPrep, nil},
			"jours": {TokenUnit, PeriodDay},
			"hier":  {TokenAnchor, AnchorYesterday},
			// Hyphenated forms are single chunks → Words, not phrases
			"avant-hier": {TokenAnchor, Anchor2DaysAgo},
			"am":         {TokenAMPM, AMPMAm},
			"pm":         {TokenAMPM, AMPMPm},
			// Number words as TokenInteger entries
			"trois": {TokenInteger, 3},
			// Multi-word phrases (space-containing keys)
			"il y a":     {TokenModifier, ModifierPast}, // 3-word phrase
			"il y":       {TokenUnknown, "unknown"},     // 2-word — must not shadow 3-word above
			"avant hier": {TokenAnchor, Anchor2DaysAgo}, // space-separated 2-word phrase
		},
	}

	cases := []struct {
		input   string
		wantSig string
		wantTok []Token
	}{
		{
			// 3-word phrase takes priority over 2-word prefix
			"trois jours il y a",
			"INTEGER UNIT MODIFIER",
			tt(TokenInteger, 3, TokenUnit, PeriodDay, TokenModifier, ModifierPast),
		},
		{
			// hyphenated single-chunk → Words lookup
			"avant-hier",
			"ANCHOR",
			tt(TokenAnchor, Anchor2DaysAgo),
		},
		{
			// space-separated 2-word phrase → Phrases lookup
			"avant hier",
			"ANCHOR",
			tt(TokenAnchor, Anchor2DaysAgo),
		},
		{
			// phrase at start, single word after
			"il y a trois jours",
			"MODIFIER INTEGER UNIT",
			tt(TokenModifier, ModifierPast, TokenInteger, 3, TokenUnit, PeriodDay),
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := phraseLang.Tokenize(tc.input)
			if Signature(got) != tc.wantSig {
				t.Errorf("Signature = %q, want %q (tokens: %v)", Signature(got), tc.wantSig, got)
			}
			if len(got) != len(tc.wantTok) {
				t.Fatalf("token count = %d, want %d: %v", len(got), len(tc.wantTok), got)
			}
			for i, tok := range got {
				if tok != tc.wantTok[i] {
					t.Errorf("token[%d] = %+v, want %+v", i, tok, tc.wantTok[i])
				}
			}
		})
	}
}

// TestLangNDayAnchors verifies that canonical anchor values for ±2/±3 day offsets
// work correctly, enabling words like "anteayer" (Spanish: 2 days ago).
func TestLangNDayAnchors(t *testing.T) {
	anchorLang := Lang{
		Words: map[string]WordEntry{
			"anteayer":   {TokenAnchor, Anchor2DaysAgo},
			"vorgestern": {TokenAnchor, Anchor2DaysAgo},
			"übermorgen": {TokenAnchor, Anchor2DaysFromNow},
		},
	}

	cases := []struct {
		input     string
		wantDelta int
	}{
		{"anteayer", -2 * 86400},
		{"vorgestern", -2 * 86400},
		{"übermorgen", 2 * 86400},
	}

	now := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			slots, err := anchorLang.Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.input, err)
			}
			if slots.DeltaSeconds == nil || *slots.DeltaSeconds != tc.wantDelta {
				t.Errorf("DeltaSeconds = %v, want %d", slots.DeltaSeconds, tc.wantDelta)
			}
			if slots.Period != PeriodDay {
				t.Errorf("Period = %q, want \"day\"", slots.Period)
			}
			got, err := Resolve(slots, now)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tc.input, err)
			}
			want := now.Add(time.Duration(tc.wantDelta) * time.Second)
			if !got.Equal(want) {
				t.Errorf("Resolve(%q) = %v, want %v", tc.input, got, want)
			}
		})
	}
}

func TestLookupLang(t *testing.T) {
	cases := []struct {
		code string
		want *Lang
	}{
		{"en", &LangEn},
		{"es", &LangEs},
		{"fr", &LangFr},
		{"de", &LangDe},
		{"it", &LangIt},
		{"pt", &LangPt},
		{"ru", &LangRu},
		{"ja", &LangJa},
		{"zh", &LangZh},
		// Region suffixes stripped
		{"en_US", &LangEn},
		{"zh-CN", &LangZh},
		{"fr-FR", &LangFr},
		// Case normalised
		{"EN", &LangEn},
		{"FR", &LangFr},
		// Unknown / empty
		{"xx", nil},
		{"", nil},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			got := nowandlater.LookupLang(tc.code)
			if got != tc.want {
				t.Errorf("LookupLang(%q) = %p, want %p", tc.code, got, tc.want)
			}
		})
	}
}

func TestStringMethods(t *testing.T) {
	// Period
	for _, tc := range []struct {
		p    Period
		want string
	}{
		{PeriodSecond, "second"}, {PeriodMinute, "minute"}, {PeriodHour, "hour"},
		{PeriodDay, "day"}, {PeriodFortnight, "fortnight"}, {PeriodWeek, "week"},
		{PeriodMonth, "month"}, {PeriodYear, "year"},
	} {
		if got := tc.p.String(); got != tc.want {
			t.Errorf("Period(%d).String() = %q, want %q", int(tc.p), got, tc.want)
		}
	}

	// Weekday
	for _, tc := range []struct {
		w    Weekday
		want string
	}{
		{WeekdayMonday, "monday"}, {WeekdayTuesday, "tuesday"}, {WeekdayWednesday, "wednesday"},
		{WeekdayThursday, "thursday"}, {WeekdayFriday, "friday"}, {WeekdaySaturday, "saturday"},
		{WeekdaySunday, "sunday"},
	} {
		if got := tc.w.String(); got != tc.want {
			t.Errorf("Weekday(%d).String() = %q, want %q", int(tc.w), got, tc.want)
		}
	}

	// Month
	for _, tc := range []struct {
		m    Month
		want string
	}{
		{MonthJanuary, "january"}, {MonthFebruary, "february"}, {MonthMarch, "march"},
		{MonthApril, "april"}, {MonthMay, "may"}, {MonthJune, "june"},
		{MonthJuly, "july"}, {MonthAugust, "august"}, {MonthSeptember, "september"},
		{MonthOctober, "october"}, {MonthNovember, "november"}, {MonthDecember, "december"},
	} {
		if got := tc.m.String(); got != tc.want {
			t.Errorf("Month(%d).String() = %q, want %q", int(tc.m), got, tc.want)
		}
	}

	// AMPM
	for _, tc := range []struct {
		a    AMPM
		want string
	}{
		{AMPMAm, "am"}, {AMPMPm, "pm"},
	} {
		if got := tc.a.String(); got != tc.want {
			t.Errorf("AMPM(%d).String() = %q, want %q", int(tc.a), got, tc.want)
		}
	}

	// Anchor
	for _, tc := range []struct {
		a    Anchor
		want string
	}{
		{AnchorNow, "now"}, {AnchorToday, "today"}, {AnchorTomorrow, "tomorrow"},
		{AnchorYesterday, "yesterday"}, {Anchor2DaysAgo, "2daysago"},
		{Anchor2DaysFromNow, "2daysfromnow"}, {Anchor3DaysAgo, "3daysago"},
		{Anchor3DaysFromNow, "3daysfromnow"},
	} {
		if got := tc.a.String(); got != tc.want {
			t.Errorf("Anchor(%d).String() = %q, want %q", int(tc.a), got, tc.want)
		}
	}

	// Modifier
	for _, tc := range []struct {
		m    Modifier
		want string
	}{
		{ModifierFuture, "future"}, {ModifierPast, "past"},
	} {
		if got := tc.m.String(); got != tc.want {
			t.Errorf("Modifier(%d).String() = %q, want %q", int(tc.m), got, tc.want)
		}
	}

	// Direction
	for _, tc := range []struct {
		d    Direction
		want string
	}{
		{DirectionFuture, "future"}, {DirectionPast, "past"}, {DirectionNearest, "nearest"},
	} {
		if got := tc.d.String(); got != tc.want {
			t.Errorf("Direction(%d).String() = %q, want %q", int(tc.d), got, tc.want)
		}
	}
}

// TestExpand2DigitYear covers all three branches of the RFC 2822 year expansion rule.
func TestExpand2DigitYear(t *testing.T) {
	for _, tc := range []struct{ y, want int }{
		{0, 2000}, {20, 2020}, {49, 2049}, // < 50 → 2000+y
		{50, 1950}, {75, 1975}, {99, 1999}, // 50–99 → 1900+y
		{100, 100}, {2026, 2026}, {999, 999}, // ≥ 100 → passthrough
	} {
		if got := Expand2DigitYear(tc.y); got != tc.want {
			t.Errorf("Expand2DigitYear(%d) = %d, want %d", tc.y, got, tc.want)
		}
	}
}
