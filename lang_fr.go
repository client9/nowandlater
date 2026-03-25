package nowandlater

// French is the built-in French Lang.
//
// Known limitations:
//   - "mar" resolves to mardi (Tuesday); write "mars" in full for March.
//   - "sept" resolves to the number 7; use "sep" or "septembre" for September.
//   - "second"/"seconde" (ordinal: 2nd) conflicts with the second time-unit token;
//     use "2" for the 2nd day of the month.
//   - Single-char unit abbreviations "j" (jour), "h" (heure), "s" (seconde),
//     "m" (mois), "a" (an) are intentionally omitted to avoid false positives.
//     "a" would also shadow the preposition entry.
//   - Elided articles glued to the following noun (e.g. "l'an", "d'ici") are
//     treated as single tokens. Only "l'année"/"l'annee"/"l'an" is pre-mapped to
//     TokenUnit; other contractions become TokenUnknown and are ignored at dispatch.
var French = Lang{
	Words:           frenchWords,
	OrdinalSuffixes: []string{"ière", "iere", "ère", "ere", "er", "ième", "ieme", "ème", "eme"},
	DateOrder:       DMY,
}

// frenchWords is the word table for French.
// It covers single words, multi-word phrases, time-word substitutions,
// and number words — all in one map.
var frenchWords = map[string]WordEntry{
	// --- Weekdays ---
	"lundi":    {TokenWeekday, WeekdayMonday},
	"lun":      {TokenWeekday, WeekdayMonday},
	"mardi":    {TokenWeekday, WeekdayTuesday},
	"mar":      {TokenWeekday, WeekdayTuesday}, // ambiguous: "mars" is March; weekday wins for abbreviation
	"mercredi": {TokenWeekday, WeekdayWednesday},
	"mer":      {TokenWeekday, WeekdayWednesday},
	"jeudi":    {TokenWeekday, WeekdayThursday},
	"jeu":      {TokenWeekday, WeekdayThursday},
	"vendredi": {TokenWeekday, WeekdayFriday},
	"ven":      {TokenWeekday, WeekdayFriday},
	"samedi":   {TokenWeekday, WeekdaySaturday},
	"sam":      {TokenWeekday, WeekdaySaturday},
	"dimanche": {TokenWeekday, WeekdaySunday},
	"dim":      {TokenWeekday, WeekdaySunday},

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
	"juillet":   {TokenMonth, MonthJuly},
	"juil":      {TokenMonth, MonthJuly},
	"août":      {TokenMonth, MonthAugust},
	"aout":      {TokenMonth, MonthAugust},
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

	// --- Prepositions ---
	"dans": {TokenPrep, nil},
	"en":   {TokenPrep, nil},
	"à":    {TokenPrep, nil},
	"a":    {TokenPrep, nil},

	// --- Fillers ---
	"le":  {TokenFiller, nil},
	"la":  {TokenFiller, nil},
	"les": {TokenFiller, nil},
	"l'":  {TokenFiller, nil},
	"du":  {TokenFiller, nil},
	"de":  {TokenFiller, nil},
	"d'":  {TokenFiller, nil},
	"et":  {TokenFiller, nil},
	"au":  {TokenFiller, nil},

	// --- Units (singular and plural) ---
	"seconde":    {TokenUnit, PeriodSecond},
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
	// "second"/"seconde" omitted — conflict with TokenUnit PeriodSecond.
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
	// "second"/"seconde" → unit conflict; omitted
	"troisième": {TokenInteger, 3}, "troisieme": {TokenInteger, 3},
	"quatrième": {TokenInteger, 4}, "quatrieme": {TokenInteger, 4},
	"cinquième": {TokenInteger, 5}, "cinquieme": {TokenInteger, 5},
	"sixième": {TokenInteger, 6}, "sixieme": {TokenInteger, 6},
	"septième": {TokenInteger, 7}, "septieme": {TokenInteger, 7},
	"huitième": {TokenInteger, 8}, "huitieme": {TokenInteger, 8},
	"neuvième": {TokenInteger, 9}, "neuvieme": {TokenInteger, 9},
	"dixième": {TokenInteger, 10}, "dixieme": {TokenInteger, 10},
}
