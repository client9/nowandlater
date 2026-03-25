package nowandlater

// Italian is the built-in Italian Lang.
//
// Italian uses standard whitespace tokenization, DMY date order, and the same
// Romance-language patterns already handled globally (WEEKDAY DIRECTION,
// UNIT DIRECTION, INTEGER UNIT MODIFIER, PREP INTEGER UNIT). No Handlers
// overrides are required.
//
// Elided article forms glued to the next word by an apostrophe become single
// whitespace-split chunks. A small set of commonly elided forms is pre-mapped
// in Words to avoid TokenUnknown:
//   - "l'anno" → TokenUnit PeriodYear  (for "l'anno prossimo")
//   - "un'ora" → TokenUnit PeriodHour  (for "fra un'ora" → PREP UNIT)
//   - "l'altroieri" → TokenAnchor Anchor2DaysAgo
//
// Known limitations:
//   - "mar" resolves to martedì (Tuesday); write "marzo" in full for March.
//   - "secondo"/"seconda" conflict with TokenUnit PeriodSecond; use "2" for
//     the 2nd day of the month.
//   - "quest'anno" (this year) is a single elided chunk that cannot expand to
//     two tokens; write "questo anno" (with a space) instead.
//   - Single-char unit abbreviations "g" (giorno), "h" (ora), "s" (secondo)
//     are intentionally omitted to avoid false positives.
//   - Elided contractions beyond the pre-mapped set become TokenUnknown
//     and are silently dropped from the signature.
var Italian = Lang{
	Words:           italianWords,
	OrdinalSuffixes: []string{"°", "º"}, // degree sign and ordinal indicator
	DateOrder:       DMY,
}

var italianWords = map[string]WordEntry{
	// --- Weekdays ---
	// Accented and unaccented forms both common in typed text.
	"lunedì":    {TokenWeekday, WeekdayMonday},
	"lunedi":    {TokenWeekday, WeekdayMonday},
	"lun":       {TokenWeekday, WeekdayMonday},
	"martedì":   {TokenWeekday, WeekdayTuesday},
	"martedi":   {TokenWeekday, WeekdayTuesday},
	"mar":       {TokenWeekday, WeekdayTuesday}, // ambiguous: also marzo abbrev; weekday wins
	"mercoledì": {TokenWeekday, WeekdayWednesday},
	"mercoledi": {TokenWeekday, WeekdayWednesday},
	"mer":       {TokenWeekday, WeekdayWednesday},
	"giovedì":   {TokenWeekday, WeekdayThursday},
	"giovedi":   {TokenWeekday, WeekdayThursday},
	"gio":       {TokenWeekday, WeekdayThursday},
	"venerdì":   {TokenWeekday, WeekdayFriday},
	"venerdi":   {TokenWeekday, WeekdayFriday},
	"ven":       {TokenWeekday, WeekdayFriday},
	"sabato":    {TokenWeekday, WeekdaySaturday},
	"sab":       {TokenWeekday, WeekdaySaturday},
	"domenica":  {TokenWeekday, WeekdaySunday},
	"dom":       {TokenWeekday, WeekdaySunday},

	// --- Months ---
	// "mar" omitted — maps to Tuesday above; write "marzo" in full.
	"gennaio":   {TokenMonth, MonthJanuary},
	"gen":       {TokenMonth, MonthJanuary},
	"febbraio":  {TokenMonth, MonthFebruary},
	"feb":       {TokenMonth, MonthFebruary},
	"marzo":     {TokenMonth, MonthMarch},
	"aprile":    {TokenMonth, MonthApril},
	"apr":       {TokenMonth, MonthApril},
	"maggio":    {TokenMonth, MonthMay},
	"mag":       {TokenMonth, MonthMay},
	"giugno":    {TokenMonth, MonthJune},
	"giu":       {TokenMonth, MonthJune},
	"luglio":    {TokenMonth, MonthJuly},
	"lug":       {TokenMonth, MonthJuly},
	"agosto":    {TokenMonth, MonthAugust},
	"ago":       {TokenMonth, MonthAugust},
	"settembre": {TokenMonth, MonthSeptember},
	"set":       {TokenMonth, MonthSeptember},
	"ottobre":   {TokenMonth, MonthOctober},
	"ott":       {TokenMonth, MonthOctober},
	"novembre":  {TokenMonth, MonthNovember},
	"nov":       {TokenMonth, MonthNovember},
	"dicembre":  {TokenMonth, MonthDecember},
	"dic":       {TokenMonth, MonthDecember},

	// --- Direction ---
	"prossimo":   {TokenDirection, DirectionFuture},
	"prossima":   {TokenDirection, DirectionFuture},
	"prossimi":   {TokenDirection, DirectionFuture},
	"prossime":   {TokenDirection, DirectionFuture},
	"seguente":   {TokenDirection, DirectionFuture},
	"successivo": {TokenDirection, DirectionFuture},
	"successiva": {TokenDirection, DirectionFuture},
	"scorso":     {TokenDirection, DirectionPast},
	"scorsa":     {TokenDirection, DirectionPast},
	"scorsi":     {TokenDirection, DirectionPast},
	"scorse":     {TokenDirection, DirectionPast},
	"ultimo":     {TokenDirection, DirectionPast},
	"ultima":     {TokenDirection, DirectionPast},
	"passato":    {TokenDirection, DirectionPast},
	"passata":    {TokenDirection, DirectionPast},
	"questo":     {TokenDirection, DirectionNearest},
	"questa":     {TokenDirection, DirectionNearest},
	"quest'":     {TokenDirection, DirectionNearest}, // "quest'anno" elided — rare standalone

	// --- Anchors ---
	"adesso":      {TokenAnchor, AnchorNow},
	"oggi":        {TokenAnchor, AnchorToday},
	"domani":      {TokenAnchor, AnchorTomorrow},
	"ieri":        {TokenAnchor, AnchorYesterday},
	"dopodomani":  {TokenAnchor, Anchor2DaysFromNow},
	"l'altroieri": {TokenAnchor, Anchor2DaysAgo}, // single elided token
	"altroieri":   {TokenAnchor, Anchor2DaysAgo},

	// --- Modifiers ---
	"fa":    {TokenModifier, ModifierPast},   // "3 giorni fa" = 3 days ago
	"dopo":  {TokenModifier, ModifierFuture}, // "3 giorni dopo" = 3 days later
	"prima": {TokenModifier, ModifierPast},   // "3 giorni prima" = 3 days before

	// --- Prepositions ---
	"fra":  {TokenPrep, nil}, // "fra 3 giorni" = in 3 days
	"tra":  {TokenPrep, nil}, // synonym for "fra"
	"in":   {TokenPrep, nil},
	"a":    {TokenPrep, nil}, // "a mezzogiorno" = at noon
	"al":   {TokenPrep, nil}, // "al mattino"
	"alle": {TokenPrep, nil}, // "alle 9:30" = at 9:30

	// --- Fillers ---
	"il":    {TokenFiller, nil},
	"lo":    {TokenFiller, nil},
	"la":    {TokenFiller, nil},
	"i":     {TokenFiller, nil},
	"gli":   {TokenFiller, nil},
	"le":    {TokenFiller, nil},
	"di":    {TokenFiller, nil},
	"del":   {TokenFiller, nil},
	"della": {TokenFiller, nil},
	"dei":   {TokenFiller, nil},
	"e":     {TokenFiller, nil},
	"circa": {TokenFiller, nil}, // "circa 3 giorni fa" = approximately 3 days ago
	"l'":    {TokenFiller, nil}, // elided article not followed by a known word
	"d'":    {TokenFiller, nil},

	// --- Units ---
	// "secondo"/"seconda" → TokenUnit PeriodSecond; ordinal "secondo" (2nd) conflicts.
	"secondo":    {TokenUnit, PeriodSecond},
	"secondi":    {TokenUnit, PeriodSecond},
	"sec":        {TokenUnit, PeriodSecond}, // abbreviation: "3 sec fa"
	"minuto":     {TokenUnit, PeriodMinute},
	"minuti":     {TokenUnit, PeriodMinute},
	"min":        {TokenUnit, PeriodMinute},
	"ora":        {TokenUnit, PeriodHour},
	"ore":        {TokenUnit, PeriodHour},
	"giorno":     {TokenUnit, PeriodDay},
	"giorni":     {TokenUnit, PeriodDay},
	"gg":         {TokenUnit, PeriodDay}, // Italian plural abbrev: "2 gg fa"
	"settimana":  {TokenUnit, PeriodWeek},
	"settimane":  {TokenUnit, PeriodWeek},
	"sett":       {TokenUnit, PeriodWeek}, // abbreviation: "2 sett fa"
	"mese":       {TokenUnit, PeriodMonth},
	"mesi":       {TokenUnit, PeriodMonth},
	"anno":       {TokenUnit, PeriodYear},
	"anni":       {TokenUnit, PeriodYear},
	"quindicina": {TokenUnit, PeriodFortnight},
	"quindicine": {TokenUnit, PeriodFortnight},

	// --- Elided unit forms (single chunk after whitespace split) ---
	"l'anno": {TokenUnit, PeriodYear}, // for "l'anno prossimo" → UNIT DIRECTION
	"un'ora": {TokenUnit, PeriodHour}, // for "fra un'ora" → PREP UNIT

	// --- Multi-word anchors ---
	"l'altro ieri": {TokenAnchor, Anchor2DaysAgo}, // elided article form
	"altro ieri":   {TokenAnchor, Anchor2DaysAgo}, // bare form from supplementary data

	// --- AM/PM ---
	"am": {TokenAMPM, AMPMAm},
	"pm": {TokenAMPM, AMPMPm},

	// --- AM/PM time-of-day phrases ---
	"di mattina":     {TokenAMPM, AMPMAm},
	"della mattina":  {TokenAMPM, AMPMAm},
	"del mattino":    {TokenAMPM, AMPMAm},
	"di pomeriggio":  {TokenAMPM, AMPMPm},
	"del pomeriggio": {TokenAMPM, AMPMPm},
	"di sera":        {TokenAMPM, AMPMPm},
	"della sera":     {TokenAMPM, AMPMPm},

	// --- Time-word substitutes ---
	"mezzogiorno": {TokenTime, "12:00"},
	"mezzanotte":  {TokenTime, "0:00"},

	// --- Number words — Cardinals ---
	// "secondo"/"seconda" omitted — conflict with TokenUnit PeriodSecond.
	"uno": {TokenInteger, 1}, "una": {TokenInteger, 1}, "un": {TokenInteger, 1},
	"due":         {TokenInteger, 2},
	"tre":         {TokenInteger, 3},
	"quattro":     {TokenInteger, 4},
	"cinque":      {TokenInteger, 5},
	"sei":         {TokenInteger, 6},
	"sette":       {TokenInteger, 7},
	"otto":        {TokenInteger, 8},
	"nove":        {TokenInteger, 9},
	"dieci":       {TokenInteger, 10},
	"undici":      {TokenInteger, 11},
	"dodici":      {TokenInteger, 12},
	"tredici":     {TokenInteger, 13},
	"quattordici": {TokenInteger, 14},
	"quindici":    {TokenInteger, 15},
	"sedici":      {TokenInteger, 16},
	"diciassette": {TokenInteger, 17},
	"diciotto":    {TokenInteger, 18},
	"diciannove":  {TokenInteger, 19},
	"venti":       {TokenInteger, 20},
	"trenta":      {TokenInteger, 30},

	// 21–29 are single compound words in Italian
	"ventuno": {TokenInteger, 21}, "ventuna": {TokenInteger, 21},
	"ventidue": {TokenInteger, 22},
	"ventitré": {TokenInteger, 23}, "ventitre": {TokenInteger, 23},
	"ventiquattro": {TokenInteger, 24},
	"venticinque":  {TokenInteger, 25},
	"ventisei":     {TokenInteger, 26},
	"ventisette":   {TokenInteger, 27},
	"ventotto":     {TokenInteger, 28},
	"ventinove":    {TokenInteger, 29},
	"trentuno":     {TokenInteger, 31}, "trentuna": {TokenInteger, 31},

	// --- Number words — Ordinals ---
	// "secondo"/"seconda" omitted. "quarto"/"quarta" safe (no weekday conflict in Italian).
	"primo": {TokenInteger, 1}, // "prima" omitted — maps to ModifierPast above
	"terzo": {TokenInteger, 3}, "terza": {TokenInteger, 3},
	"quarto": {TokenInteger, 4}, "quarta": {TokenInteger, 4},
	"quinto": {TokenInteger, 5}, "quinta": {TokenInteger, 5},
	"sesto": {TokenInteger, 6}, "sesta": {TokenInteger, 6},
	"settimo": {TokenInteger, 7}, "settima": {TokenInteger, 7},
	"ottavo": {TokenInteger, 8}, "ottava": {TokenInteger, 8},
	"nono": {TokenInteger, 9}, "nona": {TokenInteger, 9},
	"decimo": {TokenInteger, 10}, "decima": {TokenInteger, 10},
}
