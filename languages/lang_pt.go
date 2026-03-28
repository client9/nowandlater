package languages

import (
	. "github.com/client9/nowandlater/internal/engine"
)

// LangPt is the built-in Portuguese Lang (covers both European and
// Brazilian Portuguese).
//
// Key differences from Spanish:
//   - Weekday names use the "feira" system: segunda(-feira), terça(-feira), …
//   - Past modifier: "há" (European) / "faz" (Brazilian) for "N days ago".
//   - "depois de amanhã" (day after tomorrow) is a 3-word phrase.
//   - Ordinal suffixes: º (masculine) and ª (feminine) stripped from digits.
//
// Known limitations:
//   - "dez" maps to the number 10 and produces SILENT WRONG ANSWERS in date
//     expressions: "10 dez 2026" parses as October 10 (month=10), not December 10.
//     The INTEGER INTEGER YEAR signature is handled by the DMY date-order handler
//     with no way to distinguish "dez-as-10" from a real integer month; conflict
//     resolution is architecturally impossible. Write "dezembro" in full.
//   - "segunda", "quarta", "quinta", "sexta" are both weekday names and feminine
//     ordinals (2nd/4th/5th/6th). In WEEKDAY MONTH position ("quarta de março")
//     the input is genuinely ambiguous and Parse returns [ErrAmbiguous]. Use the
//     masculine ordinals "quarto"/"quinto"/"sexto" or digits to avoid ambiguity.
//     The masculine "segundo" is unambiguous in this context (UNIT MONTH is handled
//     by replaceSecondUnit).
//   - "seg" maps to WeekdayMonday (segunda-feira); write "segundo"/"segundos" in
//     full for the second time unit. CLDR lists "seg" as a second abbreviation too.
//   - Single-char unit abbreviations "h" (hora), "m" (minuto), "s" (segundo)
//     are intentionally omitted to avoid false positives.
var LangPt = Lang{
	Words:           portugueseWords,
	OrdinalSuffixes: []string{"º", "ª"},
	DateOrder:       DMY,
	Handlers: map[string]Handler{
		// "segunda", "quarta", "quinta", "sexta" are both weekday names and
		// feminine ordinals (2nd/4th/5th/6th). WEEKDAY MONTH is genuinely
		// ambiguous ("quarta de março" = Wednesday in March OR 4th of March).
		"WEEKDAY MONTH":      HandleAmbiguous,
		"WEEKDAY MONTH YEAR": HandleAmbiguous,
	},
}

var portugueseWords = map[string]WordEntry{
	// --- Weekdays ---
	// Full hyphenated forms and bare stems are both in common use.
	"segunda-feira": {Type: TokenWeekday, Value: WeekdayMonday},
	"segunda":       {Type: TokenWeekday, Value: WeekdayMonday},
	"seg":           {Type: TokenWeekday, Value: WeekdayMonday},
	"terça-feira":   {Type: TokenWeekday, Value: WeekdayTuesday},
	"terca-feira":   {Type: TokenWeekday, Value: WeekdayTuesday},
	"terça":         {Type: TokenWeekday, Value: WeekdayTuesday},
	"terca":         {Type: TokenWeekday, Value: WeekdayTuesday},
	"ter":           {Type: TokenWeekday, Value: WeekdayTuesday},
	"quarta-feira":  {Type: TokenWeekday, Value: WeekdayWednesday},
	"quarta":        {Type: TokenWeekday, Value: WeekdayWednesday},
	"qua":           {Type: TokenWeekday, Value: WeekdayWednesday},
	"quinta-feira":  {Type: TokenWeekday, Value: WeekdayThursday},
	"quinta":        {Type: TokenWeekday, Value: WeekdayThursday},
	"qui":           {Type: TokenWeekday, Value: WeekdayThursday},
	"sexta-feira":   {Type: TokenWeekday, Value: WeekdayFriday},
	"sexta":         {Type: TokenWeekday, Value: WeekdayFriday},
	"sex":           {Type: TokenWeekday, Value: WeekdayFriday},
	"sábado":        {Type: TokenWeekday, Value: WeekdaySaturday},
	"sabado":        {Type: TokenWeekday, Value: WeekdaySaturday},
	"sáb":           {Type: TokenWeekday, Value: WeekdaySaturday},
	"sab":           {Type: TokenWeekday, Value: WeekdaySaturday},
	"domingo":       {Type: TokenWeekday, Value: WeekdaySunday},
	"dom":           {Type: TokenWeekday, Value: WeekdaySunday},

	// --- Months ---
	// "dez" is omitted — it maps to the number 10; write "dezembro" in full.
	"janeiro":   {Type: TokenMonth, Value: MonthJanuary},
	"jan":       {Type: TokenMonth, Value: MonthJanuary},
	"fevereiro": {Type: TokenMonth, Value: MonthFebruary},
	"fev":       {Type: TokenMonth, Value: MonthFebruary},
	"março":     {Type: TokenMonth, Value: MonthMarch},
	"marco":     {Type: TokenMonth, Value: MonthMarch},
	"mar":       {Type: TokenMonth, Value: MonthMarch},
	"abril":     {Type: TokenMonth, Value: MonthApril},
	"abr":       {Type: TokenMonth, Value: MonthApril},
	"maio":      {Type: TokenMonth, Value: MonthMay},
	"mai":       {Type: TokenMonth, Value: MonthMay}, // CLDR abbreviation
	"junho":     {Type: TokenMonth, Value: MonthJune},
	"jun":       {Type: TokenMonth, Value: MonthJune},
	"julho":     {Type: TokenMonth, Value: MonthJuly},
	"jul":       {Type: TokenMonth, Value: MonthJuly},
	"agosto":    {Type: TokenMonth, Value: MonthAugust},
	"ago":       {Type: TokenMonth, Value: MonthAugust},
	"setembro":  {Type: TokenMonth, Value: MonthSeptember},
	"septembro": {Type: TokenMonth, Value: MonthSeptember}, // supplementary alternate spelling
	"set":       {Type: TokenMonth, Value: MonthSeptember},
	"outubro":   {Type: TokenMonth, Value: MonthOctober},
	"out":       {Type: TokenMonth, Value: MonthOctober},
	"novembro":  {Type: TokenMonth, Value: MonthNovember},
	"nov":       {Type: TokenMonth, Value: MonthNovember},
	"dezembro":  {Type: TokenMonth, Value: MonthDecember},

	// --- Direction ---
	"próximo":  {Type: TokenDirection, Value: DirectionFuture},
	"proximo":  {Type: TokenDirection, Value: DirectionFuture},
	"próxima":  {Type: TokenDirection, Value: DirectionFuture},
	"proxima":  {Type: TokenDirection, Value: DirectionFuture},
	"seguinte": {Type: TokenDirection, Value: DirectionFuture},
	"passado":  {Type: TokenDirection, Value: DirectionPast},
	"passada":  {Type: TokenDirection, Value: DirectionPast},
	"último":   {Type: TokenDirection, Value: DirectionPast},
	"ultimo":   {Type: TokenDirection, Value: DirectionPast},
	"última":   {Type: TokenDirection, Value: DirectionPast},
	"ultima":   {Type: TokenDirection, Value: DirectionPast},
	"anterior": {Type: TokenDirection, Value: DirectionPast},
	"este":     {Type: TokenDirection, Value: DirectionNearest},
	"esta":     {Type: TokenDirection, Value: DirectionNearest},
	"atual":    {Type: TokenDirection, Value: DirectionNearest},
	"actual":   {Type: TokenDirection, Value: DirectionNearest}, // European variant

	// --- Anchors ---
	"agora":     {Type: TokenAnchor, Value: AnchorNow},
	"hoje":      {Type: TokenAnchor, Value: AnchorToday},
	"amanhã":    {Type: TokenAnchor, Value: AnchorTomorrow},
	"amanha":    {Type: TokenAnchor, Value: AnchorTomorrow},
	"ontem":     {Type: TokenAnchor, Value: AnchorYesterday},
	"anteontem": {Type: TokenAnchor, Value: Anchor2DaysAgo},

	// --- Modifiers ---
	"há":     {Type: TokenModifier, Value: ModifierPast}, // European PT: "há 3 dias" = 3 days ago
	"ha":     {Type: TokenModifier, Value: ModifierPast}, // unaccented variant
	"faz":    {Type: TokenModifier, Value: ModifierPast}, // Brazilian PT: "faz 3 dias"
	"atrás":  {Type: TokenModifier, Value: ModifierPast},
	"atras":  {Type: TokenModifier, Value: ModifierPast},
	"depois": {Type: TokenModifier, Value: ModifierFuture}, // "3 dias depois" = 3 days later
	"antes":  {Type: TokenModifier, Value: ModifierPast},

	// --- Prepositions ---
	"em": {Type: TokenPrep, Value: nil},
	"às": {Type: TokenPrep, Value: nil}, // "às 9:30" = at 9:30
	"ao": {Type: TokenPrep, Value: nil}, // "ao meio-dia" = at noon

	// --- Fillers ---
	"o":     {Type: TokenFiller, Value: nil},
	"a":     {Type: TokenFiller, Value: nil},
	"os":    {Type: TokenFiller, Value: nil},
	"as":    {Type: TokenFiller, Value: nil},
	"de":    {Type: TokenFiller, Value: nil},
	"do":    {Type: TokenFiller, Value: nil},
	"da":    {Type: TokenFiller, Value: nil},
	"dos":   {Type: TokenFiller, Value: nil},
	"das":   {Type: TokenFiller, Value: nil},
	"e":     {Type: TokenFiller, Value: nil},
	"cerca": {Type: TokenFiller, Value: nil}, // "há cerca de 3 dias" = approximately 3 days ago
	"na":    {Type: TokenFiller, Value: nil},
	"no":    {Type: TokenFiller, Value: nil},

	// --- Units ---
	"segundo":   {Type: TokenUnit, Value: PeriodSecond},
	"segundos":  {Type: TokenUnit, Value: PeriodSecond},
	"minuto":    {Type: TokenUnit, Value: PeriodMinute},
	"minutos":   {Type: TokenUnit, Value: PeriodMinute},
	"min":       {Type: TokenUnit, Value: PeriodMinute}, // CLDR abbreviation
	"mins":      {Type: TokenUnit, Value: PeriodMinute}, // CLDR abbreviation (plural form)
	"hora":      {Type: TokenUnit, Value: PeriodHour},
	"horas":     {Type: TokenUnit, Value: PeriodHour},
	"dia":       {Type: TokenUnit, Value: PeriodDay},
	"dias":      {Type: TokenUnit, Value: PeriodDay},
	"semana":    {Type: TokenUnit, Value: PeriodWeek},
	"semanas":   {Type: TokenUnit, Value: PeriodWeek},
	"sem":       {Type: TokenUnit, Value: PeriodWeek}, // CLDR abbreviation
	"quinzena":  {Type: TokenUnit, Value: PeriodFortnight},
	"quinzenas": {Type: TokenUnit, Value: PeriodFortnight},
	"mês":       {Type: TokenUnit, Value: PeriodMonth},
	"mes":       {Type: TokenUnit, Value: PeriodMonth},
	"meses":     {Type: TokenUnit, Value: PeriodMonth},
	"ano":       {Type: TokenUnit, Value: PeriodYear},
	"anos":      {Type: TokenUnit, Value: PeriodYear},

	// --- AM/PM ---
	"am": {Type: TokenAMPM, Value: AMPMAm},
	"pm": {Type: TokenAMPM, Value: AMPMPm},

	// --- Time-word substitutes ---
	"meio-dia":   {Type: TokenTime, Value: "12:00"},
	"meia-noite": {Type: TokenTime, Value: "0:00"},

	// --- Multi-word phrases ---

	// Anchors
	"depois de amanhã": {Type: TokenAnchor, Value: Anchor2DaysFromNow},
	"depois de amanha": {Type: TokenAnchor, Value: Anchor2DaysFromNow},

	// Prepositions
	"daqui a":   {Type: TokenPrep, Value: nil}, // "daqui a 3 dias" = in 3 days
	"dentro de": {Type: TokenPrep, Value: nil}, // "dentro de 3 dias" = in 3 days

	// AM/PM time-of-day phrases
	"da manhã": {Type: TokenAMPM, Value: AMPMAm}, // "9 da manhã" = 9 AM
	"da manha": {Type: TokenAMPM, Value: AMPMAm},
	"da tarde": {Type: TokenAMPM, Value: AMPMPm}, // "3 da tarde" = 3 PM
	"da noite": {Type: TokenAMPM, Value: AMPMPm}, // "10 da noite" = 10 PM

	// --- Number words — Cardinals ---
	// "segundo" is mapped to TokenUnit above; replaceSecondUnit handles it as ordinal day-2.
	// "segunda" is mapped to WeekdayMonday above; use "segundo" or digits for the 2nd day.
	"um": {Type: TokenInteger, Value: 1}, "uma": {Type: TokenInteger, Value: 1},
	"dois": {Type: TokenInteger, Value: 2}, "duas": {Type: TokenInteger, Value: 2},
	"três": {Type: TokenInteger, Value: 3}, "tres": {Type: TokenInteger, Value: 3},
	"quatro":  {Type: TokenInteger, Value: 4},
	"cinco":   {Type: TokenInteger, Value: 5},
	"seis":    {Type: TokenInteger, Value: 6},
	"sete":    {Type: TokenInteger, Value: 7},
	"oito":    {Type: TokenInteger, Value: 8},
	"nove":    {Type: TokenInteger, Value: 9},
	"dez":     {Type: TokenInteger, Value: 10},
	"onze":    {Type: TokenInteger, Value: 11},
	"doze":    {Type: TokenInteger, Value: 12},
	"treze":   {Type: TokenInteger, Value: 13},
	"catorze": {Type: TokenInteger, Value: 14}, "quatorze": {Type: TokenInteger, Value: 14},
	"quinze":    {Type: TokenInteger, Value: 15},
	"dezasseis": {Type: TokenInteger, Value: 16}, "dezesseis": {Type: TokenInteger, Value: 16}, // EP/BP
	"dezassete": {Type: TokenInteger, Value: 17}, "dezessete": {Type: TokenInteger, Value: 17},
	"dezoito":  {Type: TokenInteger, Value: 18},
	"dezanove": {Type: TokenInteger, Value: 19}, "dezenove": {Type: TokenInteger, Value: 19},
	"vinte":  {Type: TokenInteger, Value: 20},
	"trinta": {Type: TokenInteger, Value: 30},

	// --- Number words — Ordinals ---
	// "quarta"/"quinta"/"sexta" omitted — map to weekdays above.
	"primeiro": {Type: TokenInteger, Value: 1}, "primeira": {Type: TokenInteger, Value: 1},
	"terceiro": {Type: TokenInteger, Value: 3}, "terceira": {Type: TokenInteger, Value: 3},
	"quarto": {Type: TokenInteger, Value: 4},
	"quinto": {Type: TokenInteger, Value: 5},
	"sexto":  {Type: TokenInteger, Value: 6},
	"sétimo": {Type: TokenInteger, Value: 7}, "setimo": {Type: TokenInteger, Value: 7},
	"oitavo": {Type: TokenInteger, Value: 8},
	"nono":   {Type: TokenInteger, Value: 9},
	"décimo": {Type: TokenInteger, Value: 10}, "decimo": {Type: TokenInteger, Value: 10},

	// Multi-word numbers
	"vinte e um": {Type: TokenInteger, Value: 21}, "vinte e uma": {Type: TokenInteger, Value: 21},
	"vinte e dois": {Type: TokenInteger, Value: 22}, "vinte e duas": {Type: TokenInteger, Value: 22},
	"vinte e três": {Type: TokenInteger, Value: 23}, "vinte e tres": {Type: TokenInteger, Value: 23},
	"vinte e quatro": {Type: TokenInteger, Value: 24},
	"trinta e um":    {Type: TokenInteger, Value: 31}, "trinta e uma": {Type: TokenInteger, Value: 31},
}
