package languages

import (
	. "github.com/client9/nowandlater/internal/engine"
)

// LangIt is the built-in Italian Lang.
//
// Italian uses standard whitespace tokenization, DMY date order, and the same
// Romance-language patterns already handled globally (WEEKDAY DIRECTION,
// UNIT DIRECTION, INTEGER UNIT MODIFIER, PREP INTEGER UNIT).
//
// Elided article forms glued to the next word by an apostrophe become single
// whitespace-split chunks. A small set of commonly elided forms is pre-mapped
// in Words to avoid TokenUnknown:
//   - "l'anno" → TokenUnit PeriodYear  (for "l'anno prossimo")
//   - "un'ora" → TokenUnit PeriodHour  (for "fra un'ora" → PREP UNIT)
//   - "l'altroieri" → TokenAnchor Anchor2DaysAgo
//
// Known limitations:
//   - "mar" resolves to martedì (Tuesday) in weekday contexts ("mar prossimo" =
//     next Tuesday). In numeric date position ("mar 5", "5 mar 2026") the input
//     is genuinely ambiguous and Parse returns [ErrAmbiguous]. Write "marzo" to
//     avoid ambiguity.
//   - "quest'anno" (this year) is a single elided chunk that cannot expand to
//     two tokens; write "questo anno" (with a space) instead.
//   - Single-char unit abbreviations "g" (giorno), "h" (ora), "s" (secondo)
//     are intentionally omitted to avoid false positives.
//   - Elided contractions beyond the pre-mapped set become TokenUnknown
//     and are silently dropped from the signature.
var LangIt = Lang{
	Words:           italianWords,
	OrdinalSuffixes: []string{"°", "º"}, // degree sign and ordinal indicator
	DateOrder:       DMY,
	Handlers: map[string]Handler{
		// "mar" abbreviates both martedì (Tuesday) and marzo (March).
		// These signatures are genuinely ambiguous; return ErrAmbiguous
		// so callers can ask for clarification rather than getting a wrong date.
		"WEEKDAY INTEGER":      HandleAmbiguous,
		"INTEGER WEEKDAY":      HandleAmbiguous,
		"WEEKDAY INTEGER YEAR": HandleAmbiguous,
		"INTEGER WEEKDAY YEAR": HandleAmbiguous,
	},
}

var italianWords = map[string]WordEntry{
	// --- Weekdays ---
	// Accented and unaccented forms both common in typed text.
	"lunedì":    {Type: TokenWeekday, Value: WeekdayMonday},
	"lunedi":    {Type: TokenWeekday, Value: WeekdayMonday},
	"lun":       {Type: TokenWeekday, Value: WeekdayMonday},
	"martedì":   {Type: TokenWeekday, Value: WeekdayTuesday},
	"martedi":   {Type: TokenWeekday, Value: WeekdayTuesday},
	"mar":       {Type: TokenWeekday, Value: WeekdayTuesday}, // ambiguous: also marzo abbrev; weekday wins
	"mercoledì": {Type: TokenWeekday, Value: WeekdayWednesday},
	"mercoledi": {Type: TokenWeekday, Value: WeekdayWednesday},
	"mer":       {Type: TokenWeekday, Value: WeekdayWednesday},
	"giovedì":   {Type: TokenWeekday, Value: WeekdayThursday},
	"giovedi":   {Type: TokenWeekday, Value: WeekdayThursday},
	"gio":       {Type: TokenWeekday, Value: WeekdayThursday},
	"venerdì":   {Type: TokenWeekday, Value: WeekdayFriday},
	"venerdi":   {Type: TokenWeekday, Value: WeekdayFriday},
	"ven":       {Type: TokenWeekday, Value: WeekdayFriday},
	"sabato":    {Type: TokenWeekday, Value: WeekdaySaturday},
	"sab":       {Type: TokenWeekday, Value: WeekdaySaturday},
	"domenica":  {Type: TokenWeekday, Value: WeekdaySunday},
	"dom":       {Type: TokenWeekday, Value: WeekdaySunday},

	// --- Months ---
	// "mar" omitted — maps to Tuesday above; write "marzo" in full.
	"gennaio":   {Type: TokenMonth, Value: MonthJanuary},
	"gen":       {Type: TokenMonth, Value: MonthJanuary},
	"febbraio":  {Type: TokenMonth, Value: MonthFebruary},
	"feb":       {Type: TokenMonth, Value: MonthFebruary},
	"marzo":     {Type: TokenMonth, Value: MonthMarch},
	"aprile":    {Type: TokenMonth, Value: MonthApril},
	"apr":       {Type: TokenMonth, Value: MonthApril},
	"maggio":    {Type: TokenMonth, Value: MonthMay},
	"mag":       {Type: TokenMonth, Value: MonthMay},
	"giugno":    {Type: TokenMonth, Value: MonthJune},
	"giu":       {Type: TokenMonth, Value: MonthJune},
	"luglio":    {Type: TokenMonth, Value: MonthJuly},
	"lug":       {Type: TokenMonth, Value: MonthJuly},
	"agosto":    {Type: TokenMonth, Value: MonthAugust},
	"ago":       {Type: TokenMonth, Value: MonthAugust},
	"settembre": {Type: TokenMonth, Value: MonthSeptember},
	"set":       {Type: TokenMonth, Value: MonthSeptember},
	"ottobre":   {Type: TokenMonth, Value: MonthOctober},
	"ott":       {Type: TokenMonth, Value: MonthOctober},
	"novembre":  {Type: TokenMonth, Value: MonthNovember},
	"nov":       {Type: TokenMonth, Value: MonthNovember},
	"dicembre":  {Type: TokenMonth, Value: MonthDecember},
	"dic":       {Type: TokenMonth, Value: MonthDecember},

	// --- Direction ---
	"prossimo":   {Type: TokenDirection, Value: DirectionFuture},
	"prossima":   {Type: TokenDirection, Value: DirectionFuture},
	"prossimi":   {Type: TokenDirection, Value: DirectionFuture},
	"prossime":   {Type: TokenDirection, Value: DirectionFuture},
	"seguente":   {Type: TokenDirection, Value: DirectionFuture},
	"successivo": {Type: TokenDirection, Value: DirectionFuture},
	"successiva": {Type: TokenDirection, Value: DirectionFuture},
	"scorso":     {Type: TokenDirection, Value: DirectionPast},
	"scorsa":     {Type: TokenDirection, Value: DirectionPast},
	"scorsi":     {Type: TokenDirection, Value: DirectionPast},
	"scorse":     {Type: TokenDirection, Value: DirectionPast},
	"ultimo":     {Type: TokenDirection, Value: DirectionPast},
	"ultima":     {Type: TokenDirection, Value: DirectionPast},
	"passato":    {Type: TokenDirection, Value: DirectionPast},
	"passata":    {Type: TokenDirection, Value: DirectionPast},
	"questo":     {Type: TokenDirection, Value: DirectionNearest},
	"questa":     {Type: TokenDirection, Value: DirectionNearest},
	"quest'":     {Type: TokenDirection, Value: DirectionNearest}, // "quest'anno" elided — rare standalone

	// --- Anchors ---
	"adesso":      {Type: TokenAnchor, Value: AnchorNow},
	"oggi":        {Type: TokenAnchor, Value: AnchorToday},
	"domani":      {Type: TokenAnchor, Value: AnchorTomorrow},
	"ieri":        {Type: TokenAnchor, Value: AnchorYesterday},
	"dopodomani":  {Type: TokenAnchor, Value: Anchor2DaysFromNow},
	"l'altroieri": {Type: TokenAnchor, Value: Anchor2DaysAgo}, // single elided token
	"altroieri":   {Type: TokenAnchor, Value: Anchor2DaysAgo},

	// --- Modifiers ---
	"fa":    {Type: TokenModifier, Value: ModifierPast},   // "3 giorni fa" = 3 days ago
	"dopo":  {Type: TokenModifier, Value: ModifierFuture}, // "3 giorni dopo" = 3 days later
	"prima": {Type: TokenModifier, Value: ModifierPast},   // "3 giorni prima" = 3 days before

	// --- Prepositions ---
	"fra":  {Type: TokenPrep, Value: nil}, // "fra 3 giorni" = in 3 days
	"tra":  {Type: TokenPrep, Value: nil}, // synonym for "fra"
	"in":   {Type: TokenPrep, Value: nil},
	"a":    {Type: TokenPrep, Value: nil}, // "a mezzogiorno" = at noon
	"al":   {Type: TokenPrep, Value: nil}, // "al mattino"
	"alle": {Type: TokenPrep, Value: nil}, // "alle 9:30" = at 9:30

	// --- Fillers ---
	"il":    {Type: TokenFiller, Value: nil},
	"lo":    {Type: TokenFiller, Value: nil},
	"la":    {Type: TokenFiller, Value: nil},
	"i":     {Type: TokenFiller, Value: nil},
	"gli":   {Type: TokenFiller, Value: nil},
	"le":    {Type: TokenFiller, Value: nil},
	"di":    {Type: TokenFiller, Value: nil},
	"del":   {Type: TokenFiller, Value: nil},
	"della": {Type: TokenFiller, Value: nil},
	"dei":   {Type: TokenFiller, Value: nil},
	"e":     {Type: TokenFiller, Value: nil},
	"circa": {Type: TokenFiller, Value: nil}, // "circa 3 giorni fa" = approximately 3 days ago
	"l'":    {Type: TokenFiller, Value: nil}, // elided article not followed by a known word
	"d'":    {Type: TokenFiller, Value: nil},

	// --- Units ---
	// "secondo"/"seconda" are also ordinal "2nd"; replaceSecondUnit handles ordinal day-2.
	"secondo":    {Type: TokenUnit, Value: PeriodSecond}, // masculine; also ordinal "2nd"
	"seconda":    {Type: TokenUnit, Value: PeriodSecond}, // feminine; also ordinal "2nd" — handled by replaceSecondUnit
	"secondi":    {Type: TokenUnit, Value: PeriodSecond},
	"sec":        {Type: TokenUnit, Value: PeriodSecond}, // abbreviation: "3 sec fa"
	"minuto":     {Type: TokenUnit, Value: PeriodMinute},
	"minuti":     {Type: TokenUnit, Value: PeriodMinute},
	"min":        {Type: TokenUnit, Value: PeriodMinute},
	"ora":        {Type: TokenUnit, Value: PeriodHour},
	"ore":        {Type: TokenUnit, Value: PeriodHour},
	"giorno":     {Type: TokenUnit, Value: PeriodDay},
	"giorni":     {Type: TokenUnit, Value: PeriodDay},
	"gg":         {Type: TokenUnit, Value: PeriodDay}, // Italian plural abbrev: "2 gg fa"
	"settimana":  {Type: TokenUnit, Value: PeriodWeek},
	"settimane":  {Type: TokenUnit, Value: PeriodWeek},
	"sett":       {Type: TokenUnit, Value: PeriodWeek}, // abbreviation: "2 sett fa"
	"mese":       {Type: TokenUnit, Value: PeriodMonth},
	"mesi":       {Type: TokenUnit, Value: PeriodMonth},
	"anno":       {Type: TokenUnit, Value: PeriodYear},
	"anni":       {Type: TokenUnit, Value: PeriodYear},
	"quindicina": {Type: TokenUnit, Value: PeriodFortnight},
	"quindicine": {Type: TokenUnit, Value: PeriodFortnight},

	// --- Elided unit forms (single chunk after whitespace split) ---
	"l'anno": {Type: TokenUnit, Value: PeriodYear}, // for "l'anno prossimo" → UNIT DIRECTION
	"un'ora": {Type: TokenUnit, Value: PeriodHour}, // for "fra un'ora" → PREP UNIT

	// --- Multi-word anchors ---
	"l'altro ieri": {Type: TokenAnchor, Value: Anchor2DaysAgo}, // elided article form
	"altro ieri":   {Type: TokenAnchor, Value: Anchor2DaysAgo}, // bare form from supplementary data

	// --- AM/PM ---
	"am": {Type: TokenAMPM, Value: AMPMAm},
	"pm": {Type: TokenAMPM, Value: AMPMPm},

	// --- AM/PM time-of-day phrases ---
	"di mattina":     {Type: TokenAMPM, Value: AMPMAm},
	"della mattina":  {Type: TokenAMPM, Value: AMPMAm},
	"del mattino":    {Type: TokenAMPM, Value: AMPMAm},
	"di pomeriggio":  {Type: TokenAMPM, Value: AMPMPm},
	"del pomeriggio": {Type: TokenAMPM, Value: AMPMPm},
	"di sera":        {Type: TokenAMPM, Value: AMPMPm},
	"della sera":     {Type: TokenAMPM, Value: AMPMPm},

	// --- Time-word substitutes ---
	"mezzogiorno": {Type: TokenTime, Value: "12:00"},
	"mezzanotte":  {Type: TokenTime, Value: "0:00"},

	// --- Number words — Cardinals ---
	// "secondo"/"seconda" omitted — conflict with TokenUnit PeriodSecond.
	"uno": {Type: TokenInteger, Value: 1}, "una": {Type: TokenInteger, Value: 1}, "un": {Type: TokenInteger, Value: 1},
	"due":         {Type: TokenInteger, Value: 2},
	"tre":         {Type: TokenInteger, Value: 3},
	"quattro":     {Type: TokenInteger, Value: 4},
	"cinque":      {Type: TokenInteger, Value: 5},
	"sei":         {Type: TokenInteger, Value: 6},
	"sette":       {Type: TokenInteger, Value: 7},
	"otto":        {Type: TokenInteger, Value: 8},
	"nove":        {Type: TokenInteger, Value: 9},
	"dieci":       {Type: TokenInteger, Value: 10},
	"undici":      {Type: TokenInteger, Value: 11},
	"dodici":      {Type: TokenInteger, Value: 12},
	"tredici":     {Type: TokenInteger, Value: 13},
	"quattordici": {Type: TokenInteger, Value: 14},
	"quindici":    {Type: TokenInteger, Value: 15},
	"sedici":      {Type: TokenInteger, Value: 16},
	"diciassette": {Type: TokenInteger, Value: 17},
	"diciotto":    {Type: TokenInteger, Value: 18},
	"diciannove":  {Type: TokenInteger, Value: 19},
	"venti":       {Type: TokenInteger, Value: 20},
	"trenta":      {Type: TokenInteger, Value: 30},

	// 21–29 are single compound words in Italian
	"ventuno": {Type: TokenInteger, Value: 21}, "ventuna": {Type: TokenInteger, Value: 21},
	"ventidue": {Type: TokenInteger, Value: 22},
	"ventitré": {Type: TokenInteger, Value: 23}, "ventitre": {Type: TokenInteger, Value: 23},
	"ventiquattro": {Type: TokenInteger, Value: 24},
	"venticinque":  {Type: TokenInteger, Value: 25},
	"ventisei":     {Type: TokenInteger, Value: 26},
	"ventisette":   {Type: TokenInteger, Value: 27},
	"ventotto":     {Type: TokenInteger, Value: 28},
	"ventinove":    {Type: TokenInteger, Value: 29},
	"trentuno":     {Type: TokenInteger, Value: 31}, "trentuna": {Type: TokenInteger, Value: 31},

	// --- Number words — Ordinals ---
	// "secondo"/"seconda" omitted. "quarto"/"quarta" safe (no weekday conflict in Italian).
	"primo": {Type: TokenInteger, Value: 1}, // "prima" omitted — maps to ModifierPast above
	"terzo": {Type: TokenInteger, Value: 3}, "terza": {Type: TokenInteger, Value: 3},
	"quarto": {Type: TokenInteger, Value: 4}, "quarta": {Type: TokenInteger, Value: 4},
	"quinto": {Type: TokenInteger, Value: 5}, "quinta": {Type: TokenInteger, Value: 5},
	"sesto": {Type: TokenInteger, Value: 6}, "sesta": {Type: TokenInteger, Value: 6},
	"settimo": {Type: TokenInteger, Value: 7}, "settima": {Type: TokenInteger, Value: 7},
	"ottavo": {Type: TokenInteger, Value: 8}, "ottava": {Type: TokenInteger, Value: 8},
	"nono": {Type: TokenInteger, Value: 9}, "nona": {Type: TokenInteger, Value: 9},
	"decimo": {Type: TokenInteger, Value: 10}, "decima": {Type: TokenInteger, Value: 10},
}
