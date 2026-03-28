package languages

import (
	. "github.com/client9/nowandlater/internal/engine"
)

// LangDe is the built-in German Lang.
//
// Ordinal dots ("1. März 2026") are handled globally in classifyNumber, so no
// special OrdinalSuffixes entry is required. Compound numeric dates
// ("01.03.2026") are handled by splitCompoundDate in the tokenizer.
//
// Known limitations:
//   - Case inflection for weekdays and months in prepositional phrases
//     (e.g. "am Montag" → "Montag" correct, "eines Montags" → genitive not mapped).
//   - "die" (abbreviation for Tuesday/Dienstag) is mapped to TokenFiller because it
//     is the most common German article. TokenFiller tokens are stripped from the
//     signature before dispatch, making conflict resolution architecturally impossible:
//     no handler can ever see "die" in a signature. Use "di" or "dienstag" instead.
//   - Single-char unit abbreviations "h" (Stunde), "m" (Minute), "s" (Sekunde)
//     are intentionally omitted to avoid false positives.
var LangDe = Lang{
	Words:     germanWords,
	DateOrder: DMY,
}

// germanWords is the word table for German.
// Direction words are inflected adjectives that vary by gender and case;
// all common forms are mapped to the same Direction constant so that patterns
// like "nächsten Montag", "nächste Woche", "nächstes Jahr" all parse correctly
// without requiring separate handlers.
var germanWords = map[string]WordEntry{
	// --- Weekdays ---
	// "die" (Tuesday abbrev) is a TokenFiller (German article) — see Known Limitations.
	"montag":     {Type: TokenWeekday, Value: WeekdayMonday},
	"mo":         {Type: TokenWeekday, Value: WeekdayMonday},
	"mon":        {Type: TokenWeekday, Value: WeekdayMonday},
	"dienstag":   {Type: TokenWeekday, Value: WeekdayTuesday},
	"di":         {Type: TokenWeekday, Value: WeekdayTuesday},
	"mittwoch":   {Type: TokenWeekday, Value: WeekdayWednesday},
	"mi":         {Type: TokenWeekday, Value: WeekdayWednesday},
	"donnerstag": {Type: TokenWeekday, Value: WeekdayThursday},
	"do":         {Type: TokenWeekday, Value: WeekdayThursday},
	"don":        {Type: TokenWeekday, Value: WeekdayThursday},
	"freitag":    {Type: TokenWeekday, Value: WeekdayFriday},
	"fr":         {Type: TokenWeekday, Value: WeekdayFriday},
	"fre":        {Type: TokenWeekday, Value: WeekdayFriday},
	"samstag":    {Type: TokenWeekday, Value: WeekdaySaturday},
	"sonnabend":  {Type: TokenWeekday, Value: WeekdaySaturday}, // North German variant
	"sa":         {Type: TokenWeekday, Value: WeekdaySaturday},
	"sam":        {Type: TokenWeekday, Value: WeekdaySaturday},
	"sonntag":    {Type: TokenWeekday, Value: WeekdaySunday},
	"so":         {Type: TokenWeekday, Value: WeekdaySunday},
	"son":        {Type: TokenWeekday, Value: WeekdaySunday},

	// --- Months ---
	"januar":    {Type: TokenMonth, Value: MonthJanuary},
	"jan":       {Type: TokenMonth, Value: MonthJanuary},
	"jänner":    {Type: TokenMonth, Value: MonthJanuary}, // Austrian German variant
	"janner":    {Type: TokenMonth, Value: MonthJanuary},
	"februar":   {Type: TokenMonth, Value: MonthFebruary},
	"feb":       {Type: TokenMonth, Value: MonthFebruary},
	"feber":     {Type: TokenMonth, Value: MonthFebruary}, // Austrian German variant
	"märz":      {Type: TokenMonth, Value: MonthMarch},
	"marz":      {Type: TokenMonth, Value: MonthMarch},
	"mär":       {Type: TokenMonth, Value: MonthMarch},
	"mar":       {Type: TokenMonth, Value: MonthMarch},
	"mrz":       {Type: TokenMonth, Value: MonthMarch}, // supplementary abbreviation
	"april":     {Type: TokenMonth, Value: MonthApril},
	"apr":       {Type: TokenMonth, Value: MonthApril},
	"mai":       {Type: TokenMonth, Value: MonthMay},
	"juni":      {Type: TokenMonth, Value: MonthJune},
	"jun":       {Type: TokenMonth, Value: MonthJune},
	"juli":      {Type: TokenMonth, Value: MonthJuly},
	"jul":       {Type: TokenMonth, Value: MonthJuly},
	"august":    {Type: TokenMonth, Value: MonthAugust},
	"aug":       {Type: TokenMonth, Value: MonthAugust},
	"september": {Type: TokenMonth, Value: MonthSeptember},
	"sep":       {Type: TokenMonth, Value: MonthSeptember},
	"sept":      {Type: TokenMonth, Value: MonthSeptember},
	"oktober":   {Type: TokenMonth, Value: MonthOctober},
	"okt":       {Type: TokenMonth, Value: MonthOctober},
	"november":  {Type: TokenMonth, Value: MonthNovember},
	"nov":       {Type: TokenMonth, Value: MonthNovember},
	"dezember":  {Type: TokenMonth, Value: MonthDecember},
	"dez":       {Type: TokenMonth, Value: MonthDecember},

	// --- Direction — inflected adjective forms ---
	// Next (nächst-): varies by gender and case
	"nächste":   {Type: TokenDirection, Value: DirectionFuture},
	"nachste":   {Type: TokenDirection, Value: DirectionFuture},
	"nächsten":  {Type: TokenDirection, Value: DirectionFuture},
	"nachsten":  {Type: TokenDirection, Value: DirectionFuture},
	"nächstes":  {Type: TokenDirection, Value: DirectionFuture},
	"nachstes":  {Type: TokenDirection, Value: DirectionFuture},
	"nächster":  {Type: TokenDirection, Value: DirectionFuture},
	"nachster":  {Type: TokenDirection, Value: DirectionFuture},
	"kommende":  {Type: TokenDirection, Value: DirectionFuture},
	"kommenden": {Type: TokenDirection, Value: DirectionFuture},
	"folgende":  {Type: TokenDirection, Value: DirectionFuture},
	"folgenden": {Type: TokenDirection, Value: DirectionFuture},
	// Last (letzt-): varies by gender and case
	"letzte":      {Type: TokenDirection, Value: DirectionPast},
	"letzten":     {Type: TokenDirection, Value: DirectionPast},
	"letztes":     {Type: TokenDirection, Value: DirectionPast},
	"letzter":     {Type: TokenDirection, Value: DirectionPast},
	"vorige":      {Type: TokenDirection, Value: DirectionPast},
	"vorigen":     {Type: TokenDirection, Value: DirectionPast},
	"voriger":     {Type: TokenDirection, Value: DirectionPast},
	"vergangene":  {Type: TokenDirection, Value: DirectionPast},
	"vergangenen": {Type: TokenDirection, Value: DirectionPast},
	// This (dies-):
	"diese":    {Type: TokenDirection, Value: DirectionNearest},
	"diesen":   {Type: TokenDirection, Value: DirectionNearest},
	"dieses":   {Type: TokenDirection, Value: DirectionNearest},
	"dieser":   {Type: TokenDirection, Value: DirectionNearest},
	"diesem":   {Type: TokenDirection, Value: DirectionNearest},
	"aktuelle": {Type: TokenDirection, Value: DirectionNearest},

	// --- Anchors ---
	"jetzt":      {Type: TokenAnchor, Value: AnchorNow},
	"heute":      {Type: TokenAnchor, Value: AnchorToday},
	"morgen":     {Type: TokenAnchor, Value: AnchorTomorrow},
	"gestern":    {Type: TokenAnchor, Value: AnchorYesterday},
	"vorgestern": {Type: TokenAnchor, Value: Anchor2DaysAgo},
	"übermorgen": {Type: TokenAnchor, Value: Anchor2DaysFromNow},
	"ubermorgen": {Type: TokenAnchor, Value: Anchor2DaysFromNow},

	// --- Modifiers ---
	// "vor" means "before/ago": "vor 3 Tagen" = 3 days ago → MODIFIER INTEGER UNIT
	// "später" means "later": "2 Stunden später" = 2 hours later → INTEGER UNIT MODIFIER
	"vor":    {Type: TokenModifier, Value: ModifierPast},
	"später": {Type: TokenModifier, Value: ModifierFuture},
	"spater": {Type: TokenModifier, Value: ModifierFuture},

	// --- Prepositions ---
	"in": {Type: TokenPrep, Value: nil},
	"am": {Type: TokenPrep, Value: nil},
	"um": {Type: TokenPrep, Value: nil},
	"an": {Type: TokenPrep, Value: nil},

	// --- Fillers ---
	"der":  {Type: TokenFiller, Value: nil},
	"die":  {Type: TokenFiller, Value: nil}, // article; "Die"  abbrev for Tuesday loses to article
	"das":  {Type: TokenFiller, Value: nil},
	"den":  {Type: TokenFiller, Value: nil},
	"dem":  {Type: TokenFiller, Value: nil},
	"des":  {Type: TokenFiller, Value: nil},
	"und":  {Type: TokenFiller, Value: nil},
	"im":   {Type: TokenFiller, Value: nil}, // contraction of "in dem"; "im nächsten Monat" → DIRECTION UNIT
	"etwa": {Type: TokenFiller, Value: nil}, // "vor etwa 3 Tagen" = approximately 3 days ago

	// --- Units (nominative, accusative, genitive, dative) ---
	"sekunde":  {Type: TokenUnit, Value: PeriodSecond},
	"sekunden": {Type: TokenUnit, Value: PeriodSecond},
	"minute":   {Type: TokenUnit, Value: PeriodMinute},
	"minuten":  {Type: TokenUnit, Value: PeriodMinute},
	"min":      {Type: TokenUnit, Value: PeriodMinute},
	"stunde":   {Type: TokenUnit, Value: PeriodHour},
	"stunden":  {Type: TokenUnit, Value: PeriodHour},
	"std":      {Type: TokenUnit, Value: PeriodHour},
	"tag":      {Type: TokenUnit, Value: PeriodDay},
	"tage":     {Type: TokenUnit, Value: PeriodDay},
	"tagen":    {Type: TokenUnit, Value: PeriodDay},
	"tags":     {Type: TokenUnit, Value: PeriodDay},
	"woche":    {Type: TokenUnit, Value: PeriodWeek},
	"wochen":   {Type: TokenUnit, Value: PeriodWeek},
	"monat":    {Type: TokenUnit, Value: PeriodMonth},
	"monate":   {Type: TokenUnit, Value: PeriodMonth},
	"monaten":  {Type: TokenUnit, Value: PeriodMonth},
	"monats":   {Type: TokenUnit, Value: PeriodMonth},
	"jahr":     {Type: TokenUnit, Value: PeriodYear},
	"jahre":    {Type: TokenUnit, Value: PeriodYear},
	"jahren":   {Type: TokenUnit, Value: PeriodYear},
	"jahres":   {Type: TokenUnit, Value: PeriodYear},

	// --- AM/PM ---
	// German time uses 24-hour notation; "pm" is accepted from mixed-language input.
	"pm": {Type: TokenAMPM, Value: AMPMPm},
	// "uhr" (o'clock) acts as a no-op time qualifier; treated as filler.
	"uhr": {Type: TokenFiller, Value: nil},

	// --- Time-word substitutions ---
	"mittags":     {Type: TokenTime, Value: "12:00"}, // midday
	"mitternacht": {Type: TokenTime, Value: "0:00"},  // midnight

	// --- Number words — Cardinals ---
	"ein": {Type: TokenInteger, Value: 1}, "eine": {Type: TokenInteger, Value: 1}, "einen": {Type: TokenInteger, Value: 1},
	"einer": {Type: TokenInteger, Value: 1}, "einem": {Type: TokenInteger, Value: 1},
	"zwei":     {Type: TokenInteger, Value: 2},
	"drei":     {Type: TokenInteger, Value: 3},
	"vier":     {Type: TokenInteger, Value: 4},
	"fünf":     {Type: TokenInteger, Value: 5},
	"funf":     {Type: TokenInteger, Value: 5},
	"sechs":    {Type: TokenInteger, Value: 6},
	"sieben":   {Type: TokenInteger, Value: 7},
	"acht":     {Type: TokenInteger, Value: 8},
	"neun":     {Type: TokenInteger, Value: 9},
	"zehn":     {Type: TokenInteger, Value: 10},
	"elf":      {Type: TokenInteger, Value: 11},
	"zwölf":    {Type: TokenInteger, Value: 12},
	"zwolf":    {Type: TokenInteger, Value: 12},
	"dreizehn": {Type: TokenInteger, Value: 13},
	"vierzehn": {Type: TokenInteger, Value: 14},
	"fünfzehn": {Type: TokenInteger, Value: 15},
	"funfzehn": {Type: TokenInteger, Value: 15},
	"sechzehn": {Type: TokenInteger, Value: 16},
	"siebzehn": {Type: TokenInteger, Value: 17},
	"achtzehn": {Type: TokenInteger, Value: 18},
	"neunzehn": {Type: TokenInteger, Value: 19},
	"zwanzig":  {Type: TokenInteger, Value: 20},
	"dreißig":  {Type: TokenInteger, Value: 30},
	"dreissig": {Type: TokenInteger, Value: 30},

	// --- Number words — Ordinals (for day-of-month in text) ---
	"ersten": {Type: TokenInteger, Value: 1}, "erste": {Type: TokenInteger, Value: 1}, "erstem": {Type: TokenInteger, Value: 1},
	"zweiten": {Type: TokenInteger, Value: 2}, "zweite": {Type: TokenInteger, Value: 2},
	"dritten": {Type: TokenInteger, Value: 3}, "dritte": {Type: TokenInteger, Value: 3},
	"vierten": {Type: TokenInteger, Value: 4}, "vierte": {Type: TokenInteger, Value: 4},
	"fünften": {Type: TokenInteger, Value: 5}, "funften": {Type: TokenInteger, Value: 5},
	"sechsten": {Type: TokenInteger, Value: 6},
	"siebten":  {Type: TokenInteger, Value: 7},
	"achten":   {Type: TokenInteger, Value: 8},
	"neunten":  {Type: TokenInteger, Value: 9},
	"zehnten":  {Type: TokenInteger, Value: 10},
}
