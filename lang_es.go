package nowandlater

// Spanish is the built-in Spanish Lang.
//
// Known limitations:
//   - "mar" resolves to martes (Tuesday); write "marzo" in full for March.
//   - "segundo"/"segunda" (ordinal: 2nd) conflicts with the second time-unit token;
//     use "2" for the 2nd day of the month.
//   - Single-char unit abbreviations "h" (hora), "d" (día), "s" (segundo),
//     "m" (mes), "a" (año) are intentionally omitted to avoid false positives.
//     "a" would also shadow the preposition entry.
//   - "mi" (Wednesday abbreviation from supplementary data) also means the
//     possessive pronoun "my" in Spanish; it maps to Wednesday in date context.
var Spanish = Lang{
	Words:           spanishWords,
	OrdinalSuffixes: []string{},
	DateOrder:       DMY,
}

// spanishWords is the word table for Spanish.
// It covers single words, multi-word phrases, time-word substitutions,
// and number words — all in one map.
var spanishWords = map[string]WordEntry{
	// --- Weekdays ---
	// 2-letter forms from supplementary data. Note: "mi" also means the possessive
	// pronoun "my" in Spanish — see Known Limitations.
	"lunes":     {TokenWeekday, WeekdayMonday},
	"lun":       {TokenWeekday, WeekdayMonday},
	"lu":        {TokenWeekday, WeekdayMonday},
	"martes":    {TokenWeekday, WeekdayTuesday},
	"mar":       {TokenWeekday, WeekdayTuesday}, // ambiguous: also marzo abbrev; weekday wins
	"miércoles": {TokenWeekday, WeekdayWednesday},
	"miercoles": {TokenWeekday, WeekdayWednesday},
	"mié":       {TokenWeekday, WeekdayWednesday},
	"mie":       {TokenWeekday, WeekdayWednesday},
	"mi":        {TokenWeekday, WeekdayWednesday},
	"jueves":    {TokenWeekday, WeekdayThursday},
	"jue":       {TokenWeekday, WeekdayThursday},
	"ju":        {TokenWeekday, WeekdayThursday},
	"viernes":   {TokenWeekday, WeekdayFriday},
	"vie":       {TokenWeekday, WeekdayFriday},
	"vi":        {TokenWeekday, WeekdayFriday},
	"sábado":    {TokenWeekday, WeekdaySaturday},
	"sabado":    {TokenWeekday, WeekdaySaturday},
	"sáb":       {TokenWeekday, WeekdaySaturday},
	"sab":       {TokenWeekday, WeekdaySaturday},
	"sa":        {TokenWeekday, WeekdaySaturday},
	"domingo":   {TokenWeekday, WeekdaySunday},
	"dom":       {TokenWeekday, WeekdaySunday},
	"do":        {TokenWeekday, WeekdaySunday},

	// --- Months ---
	"enero":   {TokenMonth, MonthJanuary},
	"ene":     {TokenMonth, MonthJanuary},
	"febrero": {TokenMonth, MonthFebruary},
	"feb":     {TokenMonth, MonthFebruary},
	"marzo":   {TokenMonth, MonthMarch},
	// "mar" → tuesday (see above)
	"abril":      {TokenMonth, MonthApril},
	"abr":        {TokenMonth, MonthApril},
	"mayo":       {TokenMonth, MonthMay},
	"may":        {TokenMonth, MonthMay}, // CLDR standard abbreviation
	"junio":      {TokenMonth, MonthJune},
	"jun":        {TokenMonth, MonthJune},
	"julio":      {TokenMonth, MonthJuly},
	"jul":        {TokenMonth, MonthJuly},
	"agosto":     {TokenMonth, MonthAugust},
	"ago":        {TokenMonth, MonthAugust},
	"septiembre": {TokenMonth, MonthSeptember},
	"setiembre":  {TokenMonth, MonthSeptember},
	"sept":       {TokenMonth, MonthSeptember}, // CLDR base es standard abbreviation
	"sep":        {TokenMonth, MonthSeptember}, // Latin America (es-419 and most locales)
	"set":        {TokenMonth, MonthSeptember}, // es-PE, es-UY
	"octubre":    {TokenMonth, MonthOctober},
	"oct":        {TokenMonth, MonthOctober},
	"noviembre":  {TokenMonth, MonthNovember},
	"nov":        {TokenMonth, MonthNovember},
	"diciembre":  {TokenMonth, MonthDecember},
	"dic":        {TokenMonth, MonthDecember},

	// --- Direction ---
	"próximo":   {TokenDirection, DirectionFuture},
	"proximo":   {TokenDirection, DirectionFuture},
	"próxima":   {TokenDirection, DirectionFuture},
	"proxima":   {TokenDirection, DirectionFuture},
	"siguiente": {TokenDirection, DirectionFuture},
	"pasado":    {TokenDirection, DirectionPast}, // "pasado mañana" handled by multi-word key below
	"pasada":    {TokenDirection, DirectionPast},
	"último":    {TokenDirection, DirectionPast},
	"ultimo":    {TokenDirection, DirectionPast},
	"última":    {TokenDirection, DirectionPast},
	"ultima":    {TokenDirection, DirectionPast},
	"anterior":  {TokenDirection, DirectionPast},
	"este":      {TokenDirection, DirectionNearest},
	"esta":      {TokenDirection, DirectionNearest},

	// --- Anchors ---
	"ahora":    {TokenAnchor, AnchorNow},
	"hoy":      {TokenAnchor, AnchorToday},
	"mañana":   {TokenAnchor, AnchorTomorrow},
	"manana":   {TokenAnchor, AnchorTomorrow},
	"ayer":     {TokenAnchor, AnchorYesterday},
	"anteayer": {TokenAnchor, Anchor2DaysAgo},
	"antier":   {TokenAnchor, Anchor2DaysAgo}, // colloquial variant

	// --- Modifiers ---
	"hace":    {TokenModifier, ModifierPast},
	"atrás":   {TokenModifier, ModifierPast},
	"atras":   {TokenModifier, ModifierPast},
	"después": {TokenModifier, ModifierFuture}, // "X días después" = X days from now
	"despues": {TokenModifier, ModifierFuture},
	"antes":   {TokenModifier, ModifierPast},

	// --- Prepositions (value not consumed semantically) ---
	"en": {TokenPrep, nil},
	"a":  {TokenPrep, nil}, // "a las" handled by multi-word key below

	// --- Fillers (value not consumed semantically) ---
	"el":    {TokenFiller, nil},
	"la":    {TokenFiller, nil},
	"lo":    {TokenFiller, nil},
	"los":   {TokenFiller, nil},
	"las":   {TokenFiller, nil},
	"de":    {TokenFiller, nil},
	"del":   {TokenFiller, nil},
	"al":    {TokenFiller, nil},
	"y":     {TokenFiller, nil}, // "treinta y uno" — "y" consumed as filler (but see multi-word number below)
	"cerca": {TokenFiller, nil}, // "hace cerca de 3 días" = approximately 3 days ago

	// --- Units (singular and plural — all variants carry the same Period constant) ---
	"segundo":   {TokenUnit, PeriodSecond},
	"segundos":  {TokenUnit, PeriodSecond},
	"seg":       {TokenUnit, PeriodSecond}, // es-AR, es-PY abbreviation
	"minuto":    {TokenUnit, PeriodMinute},
	"minutos":   {TokenUnit, PeriodMinute},
	"min":       {TokenUnit, PeriodMinute}, // CLDR abbreviation
	"hora":      {TokenUnit, PeriodHour},
	"horas":     {TokenUnit, PeriodHour},
	"día":       {TokenUnit, PeriodDay},
	"dia":       {TokenUnit, PeriodDay},
	"días":      {TokenUnit, PeriodDay},
	"dias":      {TokenUnit, PeriodDay},
	"semana":    {TokenUnit, PeriodWeek},
	"semanas":   {TokenUnit, PeriodWeek},
	"sem":       {TokenUnit, PeriodWeek},      // CLDR abbreviation
	"quincena":  {TokenUnit, PeriodFortnight}, // 14-day period; "quincena" is literally 15 days but used colloquially for a fortnight
	"quincenas": {TokenUnit, PeriodFortnight},
	"mes":       {TokenUnit, PeriodMonth},
	"meses":     {TokenUnit, PeriodMonth},
	"año":       {TokenUnit, PeriodYear},
	"ano":       {TokenUnit, PeriodYear},
	"años":      {TokenUnit, PeriodYear},
	"anos":      {TokenUnit, PeriodYear},

	// --- AM/PM ---
	"am": {TokenAMPM, AMPMAm},
	"pm": {TokenAMPM, AMPMPm},

	// --- Time-word substitutions — produce TokenTime directly ---
	"mediodía":   {TokenTime, "12:00"},
	"mediodia":   {TokenTime, "12:00"},
	"medianoche": {TokenTime, "0:00"},

	// --- Multi-word phrases (space-containing keys; matched longest-first) ---

	// Anchors
	"pasado mañana": {TokenAnchor, Anchor2DaysFromNow},
	"pasado manana": {TokenAnchor, Anchor2DaysFromNow},

	// Prepositions / compound preps
	"dentro de": {TokenPrep, nil}, // "dentro de 3 días" → "in 3 days"
	"a las":     {TokenPrep, nil}, // "a las 9:30" → "at 9:30"

	// AM/PM expressed as time-of-day phrases
	"de la mañana": {TokenAMPM, AMPMAm}, // "a las 9 de la mañana" → 9 AM
	"de la manana": {TokenAMPM, AMPMAm},
	"de la tarde":  {TokenAMPM, AMPMPm}, // "a las 3 de la tarde" → 3 PM
	"de la noche":  {TokenAMPM, AMPMPm}, // "a las 10 de la noche" → 10 PM

	// --- Number words — Cardinals ---
	// "segundo"/"segunda" (ordinal: 2nd) omitted — conflict with TokenUnit PeriodSecond.
	"uno": {TokenInteger, 1}, "un": {TokenInteger, 1}, "una": {TokenInteger, 1},
	"dos": {TokenInteger, 2}, "tres": {TokenInteger, 3}, "cuatro": {TokenInteger, 4},
	"cinco": {TokenInteger, 5}, "seis": {TokenInteger, 6}, "siete": {TokenInteger, 7},
	"ocho": {TokenInteger, 8}, "nueve": {TokenInteger, 9}, "diez": {TokenInteger, 10},
	"once": {TokenInteger, 11}, "doce": {TokenInteger, 12}, "trece": {TokenInteger, 13},
	"catorce": {TokenInteger, 14}, "quince": {TokenInteger, 15},
	"dieciséis": {TokenInteger, 16}, "dieciseis": {TokenInteger, 16},
	"diecisiete": {TokenInteger, 17}, "dieciocho": {TokenInteger, 18}, "diecinueve": {TokenInteger, 19},
	"veinte": {TokenInteger, 20},
	// 21–29 are single compound words in Spanish
	"veintiuno": {TokenInteger, 21}, "veintiuna": {TokenInteger, 21},
	"veintidós": {TokenInteger, 22}, "veintidos": {TokenInteger, 22},
	"veintitrés": {TokenInteger, 23}, "veintitres": {TokenInteger, 23},
	"veinticuatro": {TokenInteger, 24}, "veinticinco": {TokenInteger, 25},
	"veintiséis": {TokenInteger, 26}, "veintiseis": {TokenInteger, 26},
	"veintisiete": {TokenInteger, 27}, "veintiocho": {TokenInteger, 28}, "veintinueve": {TokenInteger, 29},
	"treinta": {TokenInteger, 30},

	// --- Number words — Ordinals ---
	"primero": {TokenInteger, 1}, "primera": {TokenInteger, 1},
	"tercero": {TokenInteger, 3}, "tercera": {TokenInteger, 3},
	"cuarto": {TokenInteger, 4}, "cuarta": {TokenInteger, 4},
	"quinto": {TokenInteger, 5}, "quinta": {TokenInteger, 5},
	"sexto": {TokenInteger, 6}, "sexta": {TokenInteger, 6},
	"séptimo": {TokenInteger, 7}, "septimo": {TokenInteger, 7}, "séptima": {TokenInteger, 7}, "septima": {TokenInteger, 7},
	"octavo": {TokenInteger, 8}, "octava": {TokenInteger, 8},
	"noveno": {TokenInteger, 9}, "novena": {TokenInteger, 9},
	"décimo": {TokenInteger, 10}, "decimo": {TokenInteger, 10}, "décima": {TokenInteger, 10}, "decima": {TokenInteger, 10},

	// Multi-word number: "treinta y uno" = 31 (matched before "treinta" alone)
	"treinta y uno": {TokenInteger, 31}, "treinta y un": {TokenInteger, 31}, "treinta y una": {TokenInteger, 31},
}
