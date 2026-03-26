package nowandlater

// German is the built-in German Lang.
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
var German = Lang{
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
	"montag":     {TokenWeekday, WeekdayMonday},
	"mo":         {TokenWeekday, WeekdayMonday},
	"mon":        {TokenWeekday, WeekdayMonday},
	"dienstag":   {TokenWeekday, WeekdayTuesday},
	"di":         {TokenWeekday, WeekdayTuesday},
	"mittwoch":   {TokenWeekday, WeekdayWednesday},
	"mi":         {TokenWeekday, WeekdayWednesday},
	"donnerstag": {TokenWeekday, WeekdayThursday},
	"do":         {TokenWeekday, WeekdayThursday},
	"don":        {TokenWeekday, WeekdayThursday},
	"freitag":    {TokenWeekday, WeekdayFriday},
	"fr":         {TokenWeekday, WeekdayFriday},
	"fre":        {TokenWeekday, WeekdayFriday},
	"samstag":    {TokenWeekday, WeekdaySaturday},
	"sonnabend":  {TokenWeekday, WeekdaySaturday}, // North German variant
	"sa":         {TokenWeekday, WeekdaySaturday},
	"sam":        {TokenWeekday, WeekdaySaturday},
	"sonntag":    {TokenWeekday, WeekdaySunday},
	"so":         {TokenWeekday, WeekdaySunday},
	"son":        {TokenWeekday, WeekdaySunday},

	// --- Months ---
	"januar":    {TokenMonth, MonthJanuary},
	"jan":       {TokenMonth, MonthJanuary},
	"jänner":    {TokenMonth, MonthJanuary}, // Austrian German variant
	"janner":    {TokenMonth, MonthJanuary},
	"februar":   {TokenMonth, MonthFebruary},
	"feb":       {TokenMonth, MonthFebruary},
	"feber":     {TokenMonth, MonthFebruary}, // Austrian German variant
	"märz":      {TokenMonth, MonthMarch},
	"marz":      {TokenMonth, MonthMarch},
	"mär":       {TokenMonth, MonthMarch},
	"mar":       {TokenMonth, MonthMarch},
	"mrz":       {TokenMonth, MonthMarch}, // supplementary abbreviation
	"april":     {TokenMonth, MonthApril},
	"apr":       {TokenMonth, MonthApril},
	"mai":       {TokenMonth, MonthMay},
	"juni":      {TokenMonth, MonthJune},
	"jun":       {TokenMonth, MonthJune},
	"juli":      {TokenMonth, MonthJuly},
	"jul":       {TokenMonth, MonthJuly},
	"august":    {TokenMonth, MonthAugust},
	"aug":       {TokenMonth, MonthAugust},
	"september": {TokenMonth, MonthSeptember},
	"sep":       {TokenMonth, MonthSeptember},
	"sept":      {TokenMonth, MonthSeptember},
	"oktober":   {TokenMonth, MonthOctober},
	"okt":       {TokenMonth, MonthOctober},
	"november":  {TokenMonth, MonthNovember},
	"nov":       {TokenMonth, MonthNovember},
	"dezember":  {TokenMonth, MonthDecember},
	"dez":       {TokenMonth, MonthDecember},

	// --- Direction — inflected adjective forms ---
	// Next (nächst-): varies by gender and case
	"nächste":   {TokenDirection, DirectionFuture},
	"nachste":   {TokenDirection, DirectionFuture},
	"nächsten":  {TokenDirection, DirectionFuture},
	"nachsten":  {TokenDirection, DirectionFuture},
	"nächstes":  {TokenDirection, DirectionFuture},
	"nachstes":  {TokenDirection, DirectionFuture},
	"nächster":  {TokenDirection, DirectionFuture},
	"nachster":  {TokenDirection, DirectionFuture},
	"kommende":  {TokenDirection, DirectionFuture},
	"kommenden": {TokenDirection, DirectionFuture},
	"folgende":  {TokenDirection, DirectionFuture},
	"folgenden": {TokenDirection, DirectionFuture},
	// Last (letzt-): varies by gender and case
	"letzte":      {TokenDirection, DirectionPast},
	"letzten":     {TokenDirection, DirectionPast},
	"letztes":     {TokenDirection, DirectionPast},
	"letzter":     {TokenDirection, DirectionPast},
	"vorige":      {TokenDirection, DirectionPast},
	"vorigen":     {TokenDirection, DirectionPast},
	"voriger":     {TokenDirection, DirectionPast},
	"vergangene":  {TokenDirection, DirectionPast},
	"vergangenen": {TokenDirection, DirectionPast},
	// This (dies-):
	"diese":    {TokenDirection, DirectionNearest},
	"diesen":   {TokenDirection, DirectionNearest},
	"dieses":   {TokenDirection, DirectionNearest},
	"dieser":   {TokenDirection, DirectionNearest},
	"diesem":   {TokenDirection, DirectionNearest},
	"aktuelle": {TokenDirection, DirectionNearest},

	// --- Anchors ---
	"jetzt":      {TokenAnchor, AnchorNow},
	"heute":      {TokenAnchor, AnchorToday},
	"morgen":     {TokenAnchor, AnchorTomorrow},
	"gestern":    {TokenAnchor, AnchorYesterday},
	"vorgestern": {TokenAnchor, Anchor2DaysAgo},
	"übermorgen": {TokenAnchor, Anchor2DaysFromNow},
	"ubermorgen": {TokenAnchor, Anchor2DaysFromNow},

	// --- Modifiers ---
	// "vor" means "before/ago": "vor 3 Tagen" = 3 days ago → MODIFIER INTEGER UNIT
	"vor": {TokenModifier, ModifierPast},

	// --- Prepositions ---
	"in": {TokenPrep, nil},
	"am": {TokenPrep, nil},
	"um": {TokenPrep, nil},
	"an": {TokenPrep, nil},

	// --- Fillers ---
	"der":  {TokenFiller, nil},
	"die":  {TokenFiller, nil}, // article; "Die"  abbrev for Tuesday loses to article
	"das":  {TokenFiller, nil},
	"den":  {TokenFiller, nil},
	"dem":  {TokenFiller, nil},
	"des":  {TokenFiller, nil},
	"und":  {TokenFiller, nil},
	"im":   {TokenFiller, nil}, // contraction of "in dem"; "im nächsten Monat" → DIRECTION UNIT
	"etwa": {TokenFiller, nil}, // "vor etwa 3 Tagen" = approximately 3 days ago

	// --- Units (nominative, accusative, genitive, dative) ---
	"sekunde":  {TokenUnit, PeriodSecond},
	"sekunden": {TokenUnit, PeriodSecond},
	"minute":   {TokenUnit, PeriodMinute},
	"minuten":  {TokenUnit, PeriodMinute},
	"min":      {TokenUnit, PeriodMinute},
	"stunde":   {TokenUnit, PeriodHour},
	"stunden":  {TokenUnit, PeriodHour},
	"std":      {TokenUnit, PeriodHour},
	"tag":      {TokenUnit, PeriodDay},
	"tage":     {TokenUnit, PeriodDay},
	"tagen":    {TokenUnit, PeriodDay},
	"tags":     {TokenUnit, PeriodDay},
	"woche":    {TokenUnit, PeriodWeek},
	"wochen":   {TokenUnit, PeriodWeek},
	"monat":    {TokenUnit, PeriodMonth},
	"monate":   {TokenUnit, PeriodMonth},
	"monaten":  {TokenUnit, PeriodMonth},
	"monats":   {TokenUnit, PeriodMonth},
	"jahr":     {TokenUnit, PeriodYear},
	"jahre":    {TokenUnit, PeriodYear},
	"jahren":   {TokenUnit, PeriodYear},
	"jahres":   {TokenUnit, PeriodYear},

	// --- AM/PM ---
	// German time uses 24-hour notation; "pm" is accepted from mixed-language input.
	"pm": {TokenAMPM, AMPMPm},
	// "uhr" (o'clock) acts as a no-op time qualifier; treated as filler.
	"uhr": {TokenFiller, nil},

	// --- Time-word substitutions ---
	"mittags":     {TokenTime, "12:00"}, // midday
	"mitternacht": {TokenTime, "0:00"},  // midnight

	// --- Number words — Cardinals ---
	"ein": {TokenInteger, 1}, "eine": {TokenInteger, 1}, "einen": {TokenInteger, 1},
	"einer": {TokenInteger, 1}, "einem": {TokenInteger, 1},
	"zwei":     {TokenInteger, 2},
	"drei":     {TokenInteger, 3},
	"vier":     {TokenInteger, 4},
	"fünf":     {TokenInteger, 5},
	"funf":     {TokenInteger, 5},
	"sechs":    {TokenInteger, 6},
	"sieben":   {TokenInteger, 7},
	"acht":     {TokenInteger, 8},
	"neun":     {TokenInteger, 9},
	"zehn":     {TokenInteger, 10},
	"elf":      {TokenInteger, 11},
	"zwölf":    {TokenInteger, 12},
	"zwolf":    {TokenInteger, 12},
	"dreizehn": {TokenInteger, 13},
	"vierzehn": {TokenInteger, 14},
	"fünfzehn": {TokenInteger, 15},
	"funfzehn": {TokenInteger, 15},
	"sechzehn": {TokenInteger, 16},
	"siebzehn": {TokenInteger, 17},
	"achtzehn": {TokenInteger, 18},
	"neunzehn": {TokenInteger, 19},
	"zwanzig":  {TokenInteger, 20},
	"dreißig":  {TokenInteger, 30},
	"dreissig": {TokenInteger, 30},

	// --- Number words — Ordinals (for day-of-month in text) ---
	"ersten": {TokenInteger, 1}, "erste": {TokenInteger, 1}, "erstem": {TokenInteger, 1},
	"zweiten": {TokenInteger, 2}, "zweite": {TokenInteger, 2},
	"dritten": {TokenInteger, 3}, "dritte": {TokenInteger, 3},
	"vierten": {TokenInteger, 4}, "vierte": {TokenInteger, 4},
	"fünften": {TokenInteger, 5}, "funften": {TokenInteger, 5},
	"sechsten": {TokenInteger, 6},
	"siebten":  {TokenInteger, 7},
	"achten":   {TokenInteger, 8},
	"neunten":  {TokenInteger, 9},
	"zehnten":  {TokenInteger, 10},
}
