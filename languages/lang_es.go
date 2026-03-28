package languages

import (
	. "github.com/client9/nowandlater/internal/engine"
)

// LangEs is the built-in Spanish Lang.
//
// Known limitations:
//   - "mar" resolves to Tuesday (martes) in weekday contexts ("el mar pasado" = last
//     Tuesday). In numeric date position ("mar 5", "5 de mar") the input is genuinely
//     ambiguous and Parse returns [ErrAmbiguous]. Write "marzo" to avoid ambiguity.
//   - Single-char unit abbreviations "h" (hora), "d" (día), "s" (segundo),
//     "m" (mes), "a" (año) are intentionally omitted to avoid false positives.
//     "a" would also shadow the preposition entry.
//   - "mi" (Wednesday abbreviation from supplementary data) also means the
//     possessive pronoun "my" in Spanish; it maps to Wednesday in date context.
var LangEs = Lang{
	Words:           spanishWords,
	OrdinalSuffixes: []string{},
	DateOrder:       DMY,
	Handlers: map[string]Handler{
		// "mar" abbreviates both martes (Tuesday) and marzo (March).
		// These signatures are genuinely ambiguous; return ErrAmbiguous
		// so callers can ask for clarification rather than getting a wrong date.
		"WEEKDAY INTEGER":      HandleAmbiguous,
		"INTEGER WEEKDAY":      HandleAmbiguous,
		"WEEKDAY INTEGER YEAR": HandleAmbiguous,
		"INTEGER WEEKDAY YEAR": HandleAmbiguous,
	},
}

// spanishWords is the word table for Spanish.
// It covers single words, multi-word phrases, time-word substitutions,
// and number words — all in one map.
var spanishWords = map[string]WordEntry{
	// --- Weekdays ---
	// 2-letter forms from supplementary data. Note: "mi" also means the possessive
	// pronoun "my" in Spanish — see Known Limitations.
	"lunes":     {Type: TokenWeekday, Value: WeekdayMonday},
	"lun":       {Type: TokenWeekday, Value: WeekdayMonday},
	"lu":        {Type: TokenWeekday, Value: WeekdayMonday},
	"martes":    {Type: TokenWeekday, Value: WeekdayTuesday},
	"mar":       {Type: TokenWeekday, Value: WeekdayTuesday}, // ambiguous: also marzo abbrev; weekday wins
	"miércoles": {Type: TokenWeekday, Value: WeekdayWednesday},
	"miercoles": {Type: TokenWeekday, Value: WeekdayWednesday},
	"mié":       {Type: TokenWeekday, Value: WeekdayWednesday},
	"mie":       {Type: TokenWeekday, Value: WeekdayWednesday},
	"mi":        {Type: TokenWeekday, Value: WeekdayWednesday},
	"jueves":    {Type: TokenWeekday, Value: WeekdayThursday},
	"jue":       {Type: TokenWeekday, Value: WeekdayThursday},
	"ju":        {Type: TokenWeekday, Value: WeekdayThursday},
	"viernes":   {Type: TokenWeekday, Value: WeekdayFriday},
	"vie":       {Type: TokenWeekday, Value: WeekdayFriday},
	"vi":        {Type: TokenWeekday, Value: WeekdayFriday},
	"sábado":    {Type: TokenWeekday, Value: WeekdaySaturday},
	"sabado":    {Type: TokenWeekday, Value: WeekdaySaturday},
	"sáb":       {Type: TokenWeekday, Value: WeekdaySaturday},
	"sab":       {Type: TokenWeekday, Value: WeekdaySaturday},
	"sa":        {Type: TokenWeekday, Value: WeekdaySaturday},
	"domingo":   {Type: TokenWeekday, Value: WeekdaySunday},
	"dom":       {Type: TokenWeekday, Value: WeekdaySunday},
	"do":        {Type: TokenWeekday, Value: WeekdaySunday},

	// --- Months ---
	"enero":   {Type: TokenMonth, Value: MonthJanuary},
	"ene":     {Type: TokenMonth, Value: MonthJanuary},
	"febrero": {Type: TokenMonth, Value: MonthFebruary},
	"feb":     {Type: TokenMonth, Value: MonthFebruary},
	"marzo":   {Type: TokenMonth, Value: MonthMarch},
	// "mar" → tuesday (see above)
	"abril":      {Type: TokenMonth, Value: MonthApril},
	"abr":        {Type: TokenMonth, Value: MonthApril},
	"mayo":       {Type: TokenMonth, Value: MonthMay},
	"may":        {Type: TokenMonth, Value: MonthMay}, // CLDR standard abbreviation
	"junio":      {Type: TokenMonth, Value: MonthJune},
	"jun":        {Type: TokenMonth, Value: MonthJune},
	"julio":      {Type: TokenMonth, Value: MonthJuly},
	"jul":        {Type: TokenMonth, Value: MonthJuly},
	"agosto":     {Type: TokenMonth, Value: MonthAugust},
	"ago":        {Type: TokenMonth, Value: MonthAugust},
	"septiembre": {Type: TokenMonth, Value: MonthSeptember},
	"setiembre":  {Type: TokenMonth, Value: MonthSeptember},
	"sept":       {Type: TokenMonth, Value: MonthSeptember}, // CLDR base es standard abbreviation
	"sep":        {Type: TokenMonth, Value: MonthSeptember}, // Latin America (es-419 and most locales)
	"set":        {Type: TokenMonth, Value: MonthSeptember}, // es-PE, es-UY
	"octubre":    {Type: TokenMonth, Value: MonthOctober},
	"oct":        {Type: TokenMonth, Value: MonthOctober},
	"noviembre":  {Type: TokenMonth, Value: MonthNovember},
	"nov":        {Type: TokenMonth, Value: MonthNovember},
	"diciembre":  {Type: TokenMonth, Value: MonthDecember},
	"dic":        {Type: TokenMonth, Value: MonthDecember},

	// --- Direction ---
	"próximo":   {Type: TokenDirection, Value: DirectionFuture},
	"proximo":   {Type: TokenDirection, Value: DirectionFuture},
	"próxima":   {Type: TokenDirection, Value: DirectionFuture},
	"proxima":   {Type: TokenDirection, Value: DirectionFuture},
	"siguiente": {Type: TokenDirection, Value: DirectionFuture},
	"pasado":    {Type: TokenDirection, Value: DirectionPast}, // "pasado mañana" handled by multi-word key below
	"pasada":    {Type: TokenDirection, Value: DirectionPast},
	"último":    {Type: TokenDirection, Value: DirectionPast},
	"ultimo":    {Type: TokenDirection, Value: DirectionPast},
	"última":    {Type: TokenDirection, Value: DirectionPast},
	"ultima":    {Type: TokenDirection, Value: DirectionPast},
	"anterior":  {Type: TokenDirection, Value: DirectionPast},
	"este":      {Type: TokenDirection, Value: DirectionNearest},
	"esta":      {Type: TokenDirection, Value: DirectionNearest},

	// --- Anchors ---
	"ahora":    {Type: TokenAnchor, Value: AnchorNow},
	"hoy":      {Type: TokenAnchor, Value: AnchorToday},
	"mañana":   {Type: TokenAnchor, Value: AnchorTomorrow},
	"manana":   {Type: TokenAnchor, Value: AnchorTomorrow},
	"ayer":     {Type: TokenAnchor, Value: AnchorYesterday},
	"anteayer": {Type: TokenAnchor, Value: Anchor2DaysAgo},
	"antier":   {Type: TokenAnchor, Value: Anchor2DaysAgo}, // colloquial variant

	// --- Modifiers ---
	"hace":    {Type: TokenModifier, Value: ModifierPast},
	"atrás":   {Type: TokenModifier, Value: ModifierPast},
	"atras":   {Type: TokenModifier, Value: ModifierPast},
	"después": {Type: TokenModifier, Value: ModifierFuture}, // "X días después" = X days from now
	"despues": {Type: TokenModifier, Value: ModifierFuture},
	"antes":   {Type: TokenModifier, Value: ModifierPast},

	// --- Prepositions (value not consumed semantically) ---
	"en": {Type: TokenPrep, Value: nil},
	"a":  {Type: TokenPrep, Value: nil}, // "a las" handled by multi-word key below

	// --- Fillers (value not consumed semantically) ---
	"el":    {Type: TokenFiller, Value: nil},
	"la":    {Type: TokenFiller, Value: nil},
	"lo":    {Type: TokenFiller, Value: nil},
	"los":   {Type: TokenFiller, Value: nil},
	"las":   {Type: TokenFiller, Value: nil},
	"de":    {Type: TokenFiller, Value: nil},
	"del":   {Type: TokenFiller, Value: nil},
	"al":    {Type: TokenFiller, Value: nil},
	"y":     {Type: TokenFiller, Value: nil}, // "treinta y uno" — "y" consumed as filler (but see multi-word number below)
	"cerca": {Type: TokenFiller, Value: nil}, // "hace cerca de 3 días" = approximately 3 days ago

	// --- Units (singular and plural — all variants carry the same Period constant) ---
	"segundo":   {Type: TokenUnit, Value: PeriodSecond},
	"segunda":   {Type: TokenUnit, Value: PeriodSecond}, // feminine; also ordinal "2nd" — handled by replaceSecondUnit
	"segundos":  {Type: TokenUnit, Value: PeriodSecond},
	"seg":       {Type: TokenUnit, Value: PeriodSecond}, // es-AR, es-PY abbreviation
	"minuto":    {Type: TokenUnit, Value: PeriodMinute},
	"minutos":   {Type: TokenUnit, Value: PeriodMinute},
	"min":       {Type: TokenUnit, Value: PeriodMinute}, // CLDR abbreviation
	"hora":      {Type: TokenUnit, Value: PeriodHour},
	"horas":     {Type: TokenUnit, Value: PeriodHour},
	"día":       {Type: TokenUnit, Value: PeriodDay},
	"dia":       {Type: TokenUnit, Value: PeriodDay},
	"días":      {Type: TokenUnit, Value: PeriodDay},
	"dias":      {Type: TokenUnit, Value: PeriodDay},
	"semana":    {Type: TokenUnit, Value: PeriodWeek},
	"semanas":   {Type: TokenUnit, Value: PeriodWeek},
	"sem":       {Type: TokenUnit, Value: PeriodWeek},      // CLDR abbreviation
	"quincena":  {Type: TokenUnit, Value: PeriodFortnight}, // 14-day period; "quincena" is literally 15 days but used colloquially for a fortnight
	"quincenas": {Type: TokenUnit, Value: PeriodFortnight},
	"mes":       {Type: TokenUnit, Value: PeriodMonth},
	"meses":     {Type: TokenUnit, Value: PeriodMonth},
	"año":       {Type: TokenUnit, Value: PeriodYear},
	"ano":       {Type: TokenUnit, Value: PeriodYear},
	"años":      {Type: TokenUnit, Value: PeriodYear},
	"anos":      {Type: TokenUnit, Value: PeriodYear},

	// --- AM/PM ---
	"am": {Type: TokenAMPM, Value: AMPMAm},
	"pm": {Type: TokenAMPM, Value: AMPMPm},

	// --- Time-word substitutions — produce TokenTime directly ---
	"mediodía":   {Type: TokenTime, Value: "12:00"},
	"mediodia":   {Type: TokenTime, Value: "12:00"},
	"medianoche": {Type: TokenTime, Value: "0:00"},

	// --- Multi-word phrases (space-containing keys; matched longest-first) ---

	// Anchors
	"pasado mañana": {Type: TokenAnchor, Value: Anchor2DaysFromNow},
	"pasado manana": {Type: TokenAnchor, Value: Anchor2DaysFromNow},

	// Prepositions / compound preps
	"dentro de": {Type: TokenPrep, Value: nil}, // "dentro de 3 días" → "in 3 days"
	"a las":     {Type: TokenPrep, Value: nil}, // "a las 9:30" → "at 9:30"

	// AM/PM expressed as time-of-day phrases
	"de la mañana": {Type: TokenAMPM, Value: AMPMAm}, // "a las 9 de la mañana" → 9 AM
	"de la manana": {Type: TokenAMPM, Value: AMPMAm},
	"de la tarde":  {Type: TokenAMPM, Value: AMPMPm}, // "a las 3 de la tarde" → 3 PM
	"de la noche":  {Type: TokenAMPM, Value: AMPMPm}, // "a las 10 de la noche" → 10 PM

	// --- Number words — Cardinals ---
	// "segundo"/"segunda" are mapped to TokenUnit above; replaceSecondUnit handles them as ordinal day-2.
	"uno": {Type: TokenInteger, Value: 1}, "un": {Type: TokenInteger, Value: 1}, "una": {Type: TokenInteger, Value: 1},
	"dos": {Type: TokenInteger, Value: 2}, "tres": {Type: TokenInteger, Value: 3}, "cuatro": {Type: TokenInteger, Value: 4},
	"cinco": {Type: TokenInteger, Value: 5}, "seis": {Type: TokenInteger, Value: 6}, "siete": {Type: TokenInteger, Value: 7},
	"ocho": {Type: TokenInteger, Value: 8}, "nueve": {Type: TokenInteger, Value: 9}, "diez": {Type: TokenInteger, Value: 10},
	"once": {Type: TokenInteger, Value: 11}, "doce": {Type: TokenInteger, Value: 12}, "trece": {Type: TokenInteger, Value: 13},
	"catorce": {Type: TokenInteger, Value: 14}, "quince": {Type: TokenInteger, Value: 15},
	"dieciséis": {Type: TokenInteger, Value: 16}, "dieciseis": {Type: TokenInteger, Value: 16},
	"diecisiete": {Type: TokenInteger, Value: 17}, "dieciocho": {Type: TokenInteger, Value: 18}, "diecinueve": {Type: TokenInteger, Value: 19},
	"veinte": {Type: TokenInteger, Value: 20},
	// 21–29 are single compound words in Spanish
	"veintiuno": {Type: TokenInteger, Value: 21}, "veintiuna": {Type: TokenInteger, Value: 21},
	"veintidós": {Type: TokenInteger, Value: 22}, "veintidos": {Type: TokenInteger, Value: 22},
	"veintitrés": {Type: TokenInteger, Value: 23}, "veintitres": {Type: TokenInteger, Value: 23},
	"veinticuatro": {Type: TokenInteger, Value: 24}, "veinticinco": {Type: TokenInteger, Value: 25},
	"veintiséis": {Type: TokenInteger, Value: 26}, "veintiseis": {Type: TokenInteger, Value: 26},
	"veintisiete": {Type: TokenInteger, Value: 27}, "veintiocho": {Type: TokenInteger, Value: 28}, "veintinueve": {Type: TokenInteger, Value: 29},
	"treinta": {Type: TokenInteger, Value: 30},

	// --- Number words — Ordinals ---
	"primero": {Type: TokenInteger, Value: 1}, "primera": {Type: TokenInteger, Value: 1},
	"tercero": {Type: TokenInteger, Value: 3}, "tercera": {Type: TokenInteger, Value: 3},
	"cuarto": {Type: TokenInteger, Value: 4}, "cuarta": {Type: TokenInteger, Value: 4},
	"quinto": {Type: TokenInteger, Value: 5}, "quinta": {Type: TokenInteger, Value: 5},
	"sexto": {Type: TokenInteger, Value: 6}, "sexta": {Type: TokenInteger, Value: 6},
	"séptimo": {Type: TokenInteger, Value: 7}, "septimo": {Type: TokenInteger, Value: 7}, "séptima": {Type: TokenInteger, Value: 7}, "septima": {Type: TokenInteger, Value: 7},
	"octavo": {Type: TokenInteger, Value: 8}, "octava": {Type: TokenInteger, Value: 8},
	"noveno": {Type: TokenInteger, Value: 9}, "novena": {Type: TokenInteger, Value: 9},
	"décimo": {Type: TokenInteger, Value: 10}, "decimo": {Type: TokenInteger, Value: 10}, "décima": {Type: TokenInteger, Value: 10}, "decima": {Type: TokenInteger, Value: 10},

	// Multi-word number: "treinta y uno" = 31 (matched before "treinta" alone)
	"treinta y uno": {Type: TokenInteger, Value: 31}, "treinta y un": {Type: TokenInteger, Value: 31}, "treinta y una": {Type: TokenInteger, Value: 31},
}
