package languages

import (
	. "github.com/client9/nowandlater/internal/engine"
)

// LangFr is the built-in French Lang.
//
// Known limitations:
//   - "mar" resolves to Tuesday (mardi) in weekday contexts ("mar prochain" = next Tuesday).
//     In numeric date position ("mar 5", "5 mar 2026") the input is genuinely ambiguous
//     and Parse returns [ErrAmbiguous]. Write "mars" to avoid ambiguity.
//   - "sept" resolves to the number 7 and produces SILENT WRONG ANSWERS in date
//     expressions: "10 sept 2026" parses as July 10 (month=7), not September 10.
//     The trailing-dot form "sept." is identical after normalization, so it has
//     the same problem. The INTEGER INTEGER YEAR signature is handled by the DMY
//     date-order handler with no way to distinguish "sept-as-7" from a real
//     integer month; conflict resolution is architecturally impossible.
//     Use "sep" or "septembre" instead.
//   - Single-char unit abbreviations "j" (jour), "h" (heure), "s" (seconde),
//     "m" (mois), "a" (an) are intentionally omitted to avoid false positives.
//     "a" would also shadow the preposition entry.
//   - Elided articles glued to the following noun (e.g. "l'an", "d'ici") are
//     treated as single tokens. Only "l'année"/"l'annee"/"l'an" is pre-mapped to
//     TokenUnit; other contractions become TokenUnknown and are ignored at dispatch.
//   - 2-letter weekday abbreviations (lu/ma/me/je/ve/sa/di) from supplementary data
//     conflict with common French words: "ma" (my), "me" (me), "je" (I), "sa"
//     (his/her). These are mapped to weekdays; inputs containing these words in a
//     non-weekday role will produce unexpected signatures.
var LangFr = Lang{
	Words:           frenchWords,
	OrdinalSuffixes: []string{"ière", "iere", "ère", "ere", "er", "ième", "ieme", "ème", "eme"},
	DateOrder:       DMY,
	Handlers: map[string]Handler{
		// "mar" abbreviates both mardi (Tuesday) and mars (March).
		// These signatures are genuinely ambiguous; return ErrAmbiguous
		// so callers can ask for clarification rather than getting a wrong date.
		"WEEKDAY INTEGER":      HandleAmbiguous,
		"INTEGER WEEKDAY":      HandleAmbiguous,
		"WEEKDAY INTEGER YEAR": HandleAmbiguous,
		"INTEGER WEEKDAY YEAR": HandleAmbiguous,
	},
}

// frenchWords is the word table for French.
// It covers single words, multi-word phrases, time-word substitutions,
// and number words — all in one map.
var frenchWords = map[string]WordEntry{
	// --- Weekdays ---
	// 2-letter forms from supplementary data; note "ma"/"me"/"je"/"sa" are common
	// French words — they win over their pronoun/possessive meanings in date context.
	"lundi":    {Type: TokenWeekday, Value: WeekdayMonday},
	"lun":      {Type: TokenWeekday, Value: WeekdayMonday},
	"lu":       {Type: TokenWeekday, Value: WeekdayMonday},
	"mardi":    {Type: TokenWeekday, Value: WeekdayTuesday},
	"mar":      {Type: TokenWeekday, Value: WeekdayTuesday}, // ambiguous: "mars" is March; weekday wins for abbreviation
	"ma":       {Type: TokenWeekday, Value: WeekdayTuesday},
	"mercredi": {Type: TokenWeekday, Value: WeekdayWednesday},
	"mer":      {Type: TokenWeekday, Value: WeekdayWednesday},
	"me":       {Type: TokenWeekday, Value: WeekdayWednesday},
	"jeudi":    {Type: TokenWeekday, Value: WeekdayThursday},
	"jeu":      {Type: TokenWeekday, Value: WeekdayThursday},
	"je":       {Type: TokenWeekday, Value: WeekdayThursday},
	"vendredi": {Type: TokenWeekday, Value: WeekdayFriday},
	"ven":      {Type: TokenWeekday, Value: WeekdayFriday},
	"ve":       {Type: TokenWeekday, Value: WeekdayFriday},
	"samedi":   {Type: TokenWeekday, Value: WeekdaySaturday},
	"sam":      {Type: TokenWeekday, Value: WeekdaySaturday},
	"sa":       {Type: TokenWeekday, Value: WeekdaySaturday},
	"dimanche": {Type: TokenWeekday, Value: WeekdaySunday},
	"dim":      {Type: TokenWeekday, Value: WeekdaySunday},
	"di":       {Type: TokenWeekday, Value: WeekdaySunday},

	// --- Months ---
	"janvier": {Type: TokenMonth, Value: MonthJanuary},
	"janv":    {Type: TokenMonth, Value: MonthJanuary}, // CLDR standard abbreviation
	"jan":     {Type: TokenMonth, Value: MonthJanuary},
	"février": {Type: TokenMonth, Value: MonthFebruary},
	"fevrier": {Type: TokenMonth, Value: MonthFebruary},
	"févr":    {Type: TokenMonth, Value: MonthFebruary}, // CLDR standard abbreviation
	"fevr":    {Type: TokenMonth, Value: MonthFebruary},
	"fév":     {Type: TokenMonth, Value: MonthFebruary},
	"fev":     {Type: TokenMonth, Value: MonthFebruary},
	"mars":    {Type: TokenMonth, Value: MonthMarch},
	// "mar" → mardi (Tuesday) — weekday wins; spell out "mars" for March
	"avril":     {Type: TokenMonth, Value: MonthApril},
	"avr":       {Type: TokenMonth, Value: MonthApril},
	"mai":       {Type: TokenMonth, Value: MonthMay},
	"juin":      {Type: TokenMonth, Value: MonthJune},
	"jun":       {Type: TokenMonth, Value: MonthJune}, // supplementary abbreviation
	"juillet":   {Type: TokenMonth, Value: MonthJuly},
	"juil":      {Type: TokenMonth, Value: MonthJuly},
	"jul":       {Type: TokenMonth, Value: MonthJuly}, // supplementary abbreviation
	"août":      {Type: TokenMonth, Value: MonthAugust},
	"aout":      {Type: TokenMonth, Value: MonthAugust},
	"aoû":       {Type: TokenMonth, Value: MonthAugust}, // supplementary abbreviation
	"septembre": {Type: TokenMonth, Value: MonthSeptember},
	"sep":       {Type: TokenMonth, Value: MonthSeptember},
	"octobre":   {Type: TokenMonth, Value: MonthOctober},
	"oct":       {Type: TokenMonth, Value: MonthOctober},
	"novembre":  {Type: TokenMonth, Value: MonthNovember},
	"nov":       {Type: TokenMonth, Value: MonthNovember},
	"décembre":  {Type: TokenMonth, Value: MonthDecember},
	"decembre":  {Type: TokenMonth, Value: MonthDecember},
	"déc":       {Type: TokenMonth, Value: MonthDecember},
	"dec":       {Type: TokenMonth, Value: MonthDecember},

	// --- Direction ---
	"prochain":   {Type: TokenDirection, Value: DirectionFuture},
	"prochaine":  {Type: TokenDirection, Value: DirectionFuture},
	"prochains":  {Type: TokenDirection, Value: DirectionFuture},
	"prochaines": {Type: TokenDirection, Value: DirectionFuture},
	"suivant":    {Type: TokenDirection, Value: DirectionFuture},
	"suivante":   {Type: TokenDirection, Value: DirectionFuture},
	"dernier":    {Type: TokenDirection, Value: DirectionPast},
	"dernière":   {Type: TokenDirection, Value: DirectionPast},
	"derniere":   {Type: TokenDirection, Value: DirectionPast},
	"précédent":  {Type: TokenDirection, Value: DirectionPast},
	"precedent":  {Type: TokenDirection, Value: DirectionPast},
	"passé":      {Type: TokenDirection, Value: DirectionPast},
	"passe":      {Type: TokenDirection, Value: DirectionPast},
	"ce":         {Type: TokenDirection, Value: DirectionNearest},
	"cette":      {Type: TokenDirection, Value: DirectionNearest},
	"cet":        {Type: TokenDirection, Value: DirectionNearest},

	// --- Anchors ---
	"maintenant":   {Type: TokenAnchor, Value: AnchorNow},
	"aujourd'hui":  {Type: TokenAnchor, Value: AnchorToday},
	"demain":       {Type: TokenAnchor, Value: AnchorTomorrow},
	"hier":         {Type: TokenAnchor, Value: AnchorYesterday},
	"avant-hier":   {Type: TokenAnchor, Value: Anchor2DaysAgo},
	"après-demain": {Type: TokenAnchor, Value: Anchor2DaysFromNow},
	"apres-demain": {Type: TokenAnchor, Value: Anchor2DaysFromNow},

	// --- Modifiers ---
	// "il y a" is the canonical 3-word past modifier (e.g. "il y a 3 jours" = 3 days ago).
	// This is the primary test of the phrase-match infrastructure for 3-word phrases.
	"il y a": {Type: TokenModifier, Value: ModifierPast},
	"il ya":  {Type: TokenModifier, Value: ModifierPast}, // no-space typo variant

	// --- Prepositions ---
	"dans":  {Type: TokenPrep, Value: nil},
	"en":    {Type: TokenPrep, Value: nil},
	"à":     {Type: TokenPrep, Value: nil},
	"a":     {Type: TokenPrep, Value: nil},
	"après": {Type: TokenPrep, Value: nil}, // "après 3 jours" = in/after 3 days
	"apres": {Type: TokenPrep, Value: nil},

	// --- Fillers ---
	"le":      {Type: TokenFiller, Value: nil},
	"environ": {Type: TokenFiller, Value: nil}, // "il y a environ 3 jours" = approximately 3 days ago
	"la":      {Type: TokenFiller, Value: nil},
	"les":     {Type: TokenFiller, Value: nil},
	"l'":      {Type: TokenFiller, Value: nil},
	"du":      {Type: TokenFiller, Value: nil},
	"de":      {Type: TokenFiller, Value: nil},
	"d'":      {Type: TokenFiller, Value: nil},
	"et":      {Type: TokenFiller, Value: nil},
	"au":      {Type: TokenFiller, Value: nil},

	// --- Units (singular and plural) ---
	"second":     {Type: TokenUnit, Value: PeriodSecond}, // masculine; also ordinal "2nd" — handled by replaceSecondUnit
	"seconde":    {Type: TokenUnit, Value: PeriodSecond}, // feminine; also ordinal "2nd" — handled by replaceSecondUnit
	"secondes":   {Type: TokenUnit, Value: PeriodSecond},
	"minute":     {Type: TokenUnit, Value: PeriodMinute},
	"minutes":    {Type: TokenUnit, Value: PeriodMinute},
	"min":        {Type: TokenUnit, Value: PeriodMinute},
	"heure":      {Type: TokenUnit, Value: PeriodHour},
	"heures":     {Type: TokenUnit, Value: PeriodHour},
	"jour":       {Type: TokenUnit, Value: PeriodDay},
	"jours":      {Type: TokenUnit, Value: PeriodDay},
	"semaine":    {Type: TokenUnit, Value: PeriodWeek},
	"semaines":   {Type: TokenUnit, Value: PeriodWeek},
	"sem":        {Type: TokenUnit, Value: PeriodWeek}, // abbreviation: "il y a 2 sem"
	"quinzaine":  {Type: TokenUnit, Value: PeriodFortnight},
	"quinzaines": {Type: TokenUnit, Value: PeriodFortnight},
	"mois":       {Type: TokenUnit, Value: PeriodMonth},
	"an":         {Type: TokenUnit, Value: PeriodYear},
	"ans":        {Type: TokenUnit, Value: PeriodYear},
	"année":      {Type: TokenUnit, Value: PeriodYear},
	"annee":      {Type: TokenUnit, Value: PeriodYear},
	"années":     {Type: TokenUnit, Value: PeriodYear},
	"annees":     {Type: TokenUnit, Value: PeriodYear},
	// Elided forms are single whitespace-delimited tokens; map them directly.
	"l'année": {Type: TokenUnit, Value: PeriodYear},
	"l'annee": {Type: TokenUnit, Value: PeriodYear},
	"l'an":    {Type: TokenUnit, Value: PeriodYear},

	// --- AM/PM ---
	"am": {Type: TokenAMPM, Value: AMPMAm},
	"pm": {Type: TokenAMPM, Value: AMPMPm},

	// --- Multi-word AM/PM phrases ---
	"du matin":        {Type: TokenAMPM, Value: AMPMAm}, // "9:00 du matin" = 9 AM
	"du soir":         {Type: TokenAMPM, Value: AMPMPm}, // "9:00 du soir" = 9 PM
	"de la nuit":      {Type: TokenAMPM, Value: AMPMPm}, // "10:00 de la nuit" = 10 PM
	"de l'après-midi": {Type: TokenAMPM, Value: AMPMPm}, // "3:00 de l'après-midi" = 3 PM
	"de l'apres-midi": {Type: TokenAMPM, Value: AMPMPm}, // unaccented variant

	// --- Time-word substitutions — produce TokenTime directly ---
	"midi":   {Type: TokenTime, Value: "12:00"},
	"minuit": {Type: TokenTime, Value: "0:00"},

	// --- Number words — Cardinals ---
	// "second"/"seconde" are mapped to TokenUnit above; replaceSecondUnit handles them as ordinal day-2.
	// "sept" = 7; use "sep" or "septembre" for September.
	"un": {Type: TokenInteger, Value: 1}, "une": {Type: TokenInteger, Value: 1},
	"deux":      {Type: TokenInteger, Value: 2},
	"trois":     {Type: TokenInteger, Value: 3},
	"quatre":    {Type: TokenInteger, Value: 4},
	"cinq":      {Type: TokenInteger, Value: 5},
	"six":       {Type: TokenInteger, Value: 6},
	"sept":      {Type: TokenInteger, Value: 7},
	"huit":      {Type: TokenInteger, Value: 8},
	"neuf":      {Type: TokenInteger, Value: 9},
	"dix":       {Type: TokenInteger, Value: 10},
	"onze":      {Type: TokenInteger, Value: 11},
	"douze":     {Type: TokenInteger, Value: 12},
	"treize":    {Type: TokenInteger, Value: 13},
	"quatorze":  {Type: TokenInteger, Value: 14},
	"quinze":    {Type: TokenInteger, Value: 15},
	"seize":     {Type: TokenInteger, Value: 16},
	"dix-sept":  {Type: TokenInteger, Value: 17},
	"dix-huit":  {Type: TokenInteger, Value: 18},
	"dix-neuf":  {Type: TokenInteger, Value: 19},
	"vingt":     {Type: TokenInteger, Value: 20},
	"trente":    {Type: TokenInteger, Value: 30},
	"quarante":  {Type: TokenInteger, Value: 40},
	"cinquante": {Type: TokenInteger, Value: 50},

	// Compound numbers (3-word: matched before 1-word "vingt"/"trente")
	"vingt et un": {Type: TokenInteger, Value: 21}, "vingt et une": {Type: TokenInteger, Value: 21},
	"trente et un": {Type: TokenInteger, Value: 31}, "trente et une": {Type: TokenInteger, Value: 31},

	// Compound numbers (2-word hyphenated — single tokens after whitespace split)
	"vingt-deux":   {Type: TokenInteger, Value: 22},
	"vingt-trois":  {Type: TokenInteger, Value: 23},
	"vingt-quatre": {Type: TokenInteger, Value: 24},
	"vingt-cinq":   {Type: TokenInteger, Value: 25},
	"vingt-six":    {Type: TokenInteger, Value: 26},
	"vingt-sept":   {Type: TokenInteger, Value: 27},
	"vingt-huit":   {Type: TokenInteger, Value: 28},
	"vingt-neuf":   {Type: TokenInteger, Value: 29},

	// --- Number words — Ordinals ---
	"premier": {Type: TokenInteger, Value: 1}, "première": {Type: TokenInteger, Value: 1}, "premiere": {Type: TokenInteger, Value: 1},
	// "second"/"seconde" → mapped to TokenUnit above; replaceSecondUnit handles ordinal day-2
	"troisième": {Type: TokenInteger, Value: 3}, "troisieme": {Type: TokenInteger, Value: 3},
	"quatrième": {Type: TokenInteger, Value: 4}, "quatrieme": {Type: TokenInteger, Value: 4},
	"cinquième": {Type: TokenInteger, Value: 5}, "cinquieme": {Type: TokenInteger, Value: 5},
	"sixième": {Type: TokenInteger, Value: 6}, "sixieme": {Type: TokenInteger, Value: 6},
	"septième": {Type: TokenInteger, Value: 7}, "septieme": {Type: TokenInteger, Value: 7},
	"huitième": {Type: TokenInteger, Value: 8}, "huitieme": {Type: TokenInteger, Value: 8},
	"neuvième": {Type: TokenInteger, Value: 9}, "neuvieme": {Type: TokenInteger, Value: 9},
	"dixième": {Type: TokenInteger, Value: 10}, "dixieme": {Type: TokenInteger, Value: 10},
}
