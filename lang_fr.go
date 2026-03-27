package nowandlater

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
		"WEEKDAY INTEGER":      handleAmbiguous,
		"INTEGER WEEKDAY":      handleAmbiguous,
		"WEEKDAY INTEGER YEAR": handleAmbiguous,
		"INTEGER WEEKDAY YEAR": handleAmbiguous,
	},
}

// frenchWords is the word table for French.
// It covers single words, multi-word phrases, time-word substitutions,
// and number words — all in one map.
var frenchWords = map[string]WordEntry{
	// --- Weekdays ---
	// 2-letter forms from supplementary data; note "ma"/"me"/"je"/"sa" are common
	// French words — they win over their pronoun/possessive meanings in date context.
	"lundi":    {TokenWeekday, WeekdayMonday},
	"lun":      {TokenWeekday, WeekdayMonday},
	"lu":       {TokenWeekday, WeekdayMonday},
	"mardi":    {TokenWeekday, WeekdayTuesday},
	"mar":      {TokenWeekday, WeekdayTuesday}, // ambiguous: "mars" is March; weekday wins for abbreviation
	"ma":       {TokenWeekday, WeekdayTuesday},
	"mercredi": {TokenWeekday, WeekdayWednesday},
	"mer":      {TokenWeekday, WeekdayWednesday},
	"me":       {TokenWeekday, WeekdayWednesday},
	"jeudi":    {TokenWeekday, WeekdayThursday},
	"jeu":      {TokenWeekday, WeekdayThursday},
	"je":       {TokenWeekday, WeekdayThursday},
	"vendredi": {TokenWeekday, WeekdayFriday},
	"ven":      {TokenWeekday, WeekdayFriday},
	"ve":       {TokenWeekday, WeekdayFriday},
	"samedi":   {TokenWeekday, WeekdaySaturday},
	"sam":      {TokenWeekday, WeekdaySaturday},
	"sa":       {TokenWeekday, WeekdaySaturday},
	"dimanche": {TokenWeekday, WeekdaySunday},
	"dim":      {TokenWeekday, WeekdaySunday},
	"di":       {TokenWeekday, WeekdaySunday},

	// --- Months ---
	"janvier": {TokenMonth, MonthJanuary},
	"janv":    {TokenMonth, MonthJanuary}, // CLDR standard abbreviation
	"jan":     {TokenMonth, MonthJanuary},
	"février": {TokenMonth, MonthFebruary},
	"fevrier": {TokenMonth, MonthFebruary},
	"févr":    {TokenMonth, MonthFebruary}, // CLDR standard abbreviation
	"fevr":    {TokenMonth, MonthFebruary},
	"fév":     {TokenMonth, MonthFebruary},
	"fev":     {TokenMonth, MonthFebruary},
	"mars":    {TokenMonth, MonthMarch},
	// "mar" → mardi (Tuesday) — weekday wins; spell out "mars" for March
	"avril":     {TokenMonth, MonthApril},
	"avr":       {TokenMonth, MonthApril},
	"mai":       {TokenMonth, MonthMay},
	"juin":      {TokenMonth, MonthJune},
	"jun":       {TokenMonth, MonthJune}, // supplementary abbreviation
	"juillet":   {TokenMonth, MonthJuly},
	"juil":      {TokenMonth, MonthJuly},
	"jul":       {TokenMonth, MonthJuly}, // supplementary abbreviation
	"août":      {TokenMonth, MonthAugust},
	"aout":      {TokenMonth, MonthAugust},
	"aoû":       {TokenMonth, MonthAugust}, // supplementary abbreviation
	"septembre": {TokenMonth, MonthSeptember},
	"sep":       {TokenMonth, MonthSeptember},
	"octobre":   {TokenMonth, MonthOctober},
	"oct":       {TokenMonth, MonthOctober},
	"novembre":  {TokenMonth, MonthNovember},
	"nov":       {TokenMonth, MonthNovember},
	"décembre":  {TokenMonth, MonthDecember},
	"decembre":  {TokenMonth, MonthDecember},
	"déc":       {TokenMonth, MonthDecember},
	"dec":       {TokenMonth, MonthDecember},

	// --- Direction ---
	"prochain":   {TokenDirection, DirectionFuture},
	"prochaine":  {TokenDirection, DirectionFuture},
	"prochains":  {TokenDirection, DirectionFuture},
	"prochaines": {TokenDirection, DirectionFuture},
	"suivant":    {TokenDirection, DirectionFuture},
	"suivante":   {TokenDirection, DirectionFuture},
	"dernier":    {TokenDirection, DirectionPast},
	"dernière":   {TokenDirection, DirectionPast},
	"derniere":   {TokenDirection, DirectionPast},
	"précédent":  {TokenDirection, DirectionPast},
	"precedent":  {TokenDirection, DirectionPast},
	"passé":      {TokenDirection, DirectionPast},
	"passe":      {TokenDirection, DirectionPast},
	"ce":         {TokenDirection, DirectionNearest},
	"cette":      {TokenDirection, DirectionNearest},
	"cet":        {TokenDirection, DirectionNearest},

	// --- Anchors ---
	"maintenant":   {TokenAnchor, AnchorNow},
	"aujourd'hui":  {TokenAnchor, AnchorToday},
	"demain":       {TokenAnchor, AnchorTomorrow},
	"hier":         {TokenAnchor, AnchorYesterday},
	"avant-hier":   {TokenAnchor, Anchor2DaysAgo},
	"après-demain": {TokenAnchor, Anchor2DaysFromNow},
	"apres-demain": {TokenAnchor, Anchor2DaysFromNow},

	// --- Modifiers ---
	// "il y a" is the canonical 3-word past modifier (e.g. "il y a 3 jours" = 3 days ago).
	// This is the primary test of the phrase-match infrastructure for 3-word phrases.
	"il y a": {TokenModifier, ModifierPast},
	"il ya":  {TokenModifier, ModifierPast}, // no-space typo variant

	// --- Prepositions ---
	"dans":  {TokenPrep, nil},
	"en":    {TokenPrep, nil},
	"à":     {TokenPrep, nil},
	"a":     {TokenPrep, nil},
	"après": {TokenPrep, nil}, // "après 3 jours" = in/after 3 days
	"apres": {TokenPrep, nil},

	// --- Fillers ---
	"le":      {TokenFiller, nil},
	"environ": {TokenFiller, nil}, // "il y a environ 3 jours" = approximately 3 days ago
	"la":      {TokenFiller, nil},
	"les":     {TokenFiller, nil},
	"l'":      {TokenFiller, nil},
	"du":      {TokenFiller, nil},
	"de":      {TokenFiller, nil},
	"d'":      {TokenFiller, nil},
	"et":      {TokenFiller, nil},
	"au":      {TokenFiller, nil},

	// --- Units (singular and plural) ---
	"second":     {TokenUnit, PeriodSecond}, // masculine; also ordinal "2nd" — handled by replaceSecondUnit
	"seconde":    {TokenUnit, PeriodSecond}, // feminine; also ordinal "2nd" — handled by replaceSecondUnit
	"secondes":   {TokenUnit, PeriodSecond},
	"minute":     {TokenUnit, PeriodMinute},
	"minutes":    {TokenUnit, PeriodMinute},
	"min":        {TokenUnit, PeriodMinute},
	"heure":      {TokenUnit, PeriodHour},
	"heures":     {TokenUnit, PeriodHour},
	"jour":       {TokenUnit, PeriodDay},
	"jours":      {TokenUnit, PeriodDay},
	"semaine":    {TokenUnit, PeriodWeek},
	"semaines":   {TokenUnit, PeriodWeek},
	"sem":        {TokenUnit, PeriodWeek}, // abbreviation: "il y a 2 sem"
	"quinzaine":  {TokenUnit, PeriodFortnight},
	"quinzaines": {TokenUnit, PeriodFortnight},
	"mois":       {TokenUnit, PeriodMonth},
	"an":         {TokenUnit, PeriodYear},
	"ans":        {TokenUnit, PeriodYear},
	"année":      {TokenUnit, PeriodYear},
	"annee":      {TokenUnit, PeriodYear},
	"années":     {TokenUnit, PeriodYear},
	"annees":     {TokenUnit, PeriodYear},
	// Elided forms are single whitespace-delimited tokens; map them directly.
	"l'année": {TokenUnit, PeriodYear},
	"l'annee": {TokenUnit, PeriodYear},
	"l'an":    {TokenUnit, PeriodYear},

	// --- AM/PM ---
	"am": {TokenAMPM, AMPMAm},
	"pm": {TokenAMPM, AMPMPm},

	// --- Multi-word AM/PM phrases ---
	"du matin":        {TokenAMPM, AMPMAm}, // "9:00 du matin" = 9 AM
	"du soir":         {TokenAMPM, AMPMPm}, // "9:00 du soir" = 9 PM
	"de la nuit":      {TokenAMPM, AMPMPm}, // "10:00 de la nuit" = 10 PM
	"de l'après-midi": {TokenAMPM, AMPMPm}, // "3:00 de l'après-midi" = 3 PM
	"de l'apres-midi": {TokenAMPM, AMPMPm}, // unaccented variant

	// --- Time-word substitutions — produce TokenTime directly ---
	"midi":   {TokenTime, "12:00"},
	"minuit": {TokenTime, "0:00"},

	// --- Number words — Cardinals ---
	// "second"/"seconde" are mapped to TokenUnit above; replaceSecondUnit handles them as ordinal day-2.
	// "sept" = 7; use "sep" or "septembre" for September.
	"un": {TokenInteger, 1}, "une": {TokenInteger, 1},
	"deux":      {TokenInteger, 2},
	"trois":     {TokenInteger, 3},
	"quatre":    {TokenInteger, 4},
	"cinq":      {TokenInteger, 5},
	"six":       {TokenInteger, 6},
	"sept":      {TokenInteger, 7},
	"huit":      {TokenInteger, 8},
	"neuf":      {TokenInteger, 9},
	"dix":       {TokenInteger, 10},
	"onze":      {TokenInteger, 11},
	"douze":     {TokenInteger, 12},
	"treize":    {TokenInteger, 13},
	"quatorze":  {TokenInteger, 14},
	"quinze":    {TokenInteger, 15},
	"seize":     {TokenInteger, 16},
	"dix-sept":  {TokenInteger, 17},
	"dix-huit":  {TokenInteger, 18},
	"dix-neuf":  {TokenInteger, 19},
	"vingt":     {TokenInteger, 20},
	"trente":    {TokenInteger, 30},
	"quarante":  {TokenInteger, 40},
	"cinquante": {TokenInteger, 50},

	// Compound numbers (3-word: matched before 1-word "vingt"/"trente")
	"vingt et un": {TokenInteger, 21}, "vingt et une": {TokenInteger, 21},
	"trente et un": {TokenInteger, 31}, "trente et une": {TokenInteger, 31},

	// Compound numbers (2-word hyphenated — single tokens after whitespace split)
	"vingt-deux":   {TokenInteger, 22},
	"vingt-trois":  {TokenInteger, 23},
	"vingt-quatre": {TokenInteger, 24},
	"vingt-cinq":   {TokenInteger, 25},
	"vingt-six":    {TokenInteger, 26},
	"vingt-sept":   {TokenInteger, 27},
	"vingt-huit":   {TokenInteger, 28},
	"vingt-neuf":   {TokenInteger, 29},

	// --- Number words — Ordinals ---
	"premier": {TokenInteger, 1}, "première": {TokenInteger, 1}, "premiere": {TokenInteger, 1},
	// "second"/"seconde" → mapped to TokenUnit above; replaceSecondUnit handles ordinal day-2
	"troisième": {TokenInteger, 3}, "troisieme": {TokenInteger, 3},
	"quatrième": {TokenInteger, 4}, "quatrieme": {TokenInteger, 4},
	"cinquième": {TokenInteger, 5}, "cinquieme": {TokenInteger, 5},
	"sixième": {TokenInteger, 6}, "sixieme": {TokenInteger, 6},
	"septième": {TokenInteger, 7}, "septieme": {TokenInteger, 7},
	"huitième": {TokenInteger, 8}, "huitieme": {TokenInteger, 8},
	"neuvième": {TokenInteger, 9}, "neuvieme": {TokenInteger, 9},
	"dixième": {TokenInteger, 10}, "dixieme": {TokenInteger, 10},
}
