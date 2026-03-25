package nowandlater

// Portuguese is the built-in Portuguese Lang (covers both European and
// Brazilian Portuguese).
//
// Key differences from Spanish:
//   - Weekday names use the "feira" system: segunda(-feira), terça(-feira), …
//   - Past modifier: "há" (European) / "faz" (Brazilian) for "N days ago".
//   - "depois de amanhã" (day after tomorrow) is a 3-word phrase.
//   - Ordinal suffixes: º (masculine) and ª (feminine) stripped from digits.
//
// Known limitations:
//   - "segundo"/"segundos" is mapped to TokenUnit PeriodSecond; the ordinal
//     "segundo" (2nd) conflicts and must be written as "2".
//   - "dez" maps to the number 10; write "dezembro" in full for December.
//   - "quarta"/"quinta"/"sexta" map to weekdays (Wednesday/Thursday/Friday);
//     ordinal uses (4th/5th/6th) must be written as digits or "quarto"/"quinto"/"sexto".
//   - "seg" maps to WeekdayMonday (segunda-feira); write "segundo"/"segundos" in
//     full for the second time unit. CLDR lists "seg" as a second abbreviation too.
//   - Single-char unit abbreviations "h" (hora), "m" (minuto), "s" (segundo)
//     are intentionally omitted to avoid false positives.
var Portuguese = Lang{
	Words:           portugueseWords,
	OrdinalSuffixes: []string{"º", "ª"},
	DateOrder:       DMY,
}

var portugueseWords = map[string]WordEntry{
	// --- Weekdays ---
	// Full hyphenated forms and bare stems are both in common use.
	"segunda-feira": {TokenWeekday, WeekdayMonday},
	"segunda":       {TokenWeekday, WeekdayMonday},
	"seg":           {TokenWeekday, WeekdayMonday},
	"terça-feira":   {TokenWeekday, WeekdayTuesday},
	"terca-feira":   {TokenWeekday, WeekdayTuesday},
	"terça":         {TokenWeekday, WeekdayTuesday},
	"terca":         {TokenWeekday, WeekdayTuesday},
	"ter":           {TokenWeekday, WeekdayTuesday},
	"quarta-feira":  {TokenWeekday, WeekdayWednesday},
	"quarta":        {TokenWeekday, WeekdayWednesday},
	"qua":           {TokenWeekday, WeekdayWednesday},
	"quinta-feira":  {TokenWeekday, WeekdayThursday},
	"quinta":        {TokenWeekday, WeekdayThursday},
	"qui":           {TokenWeekday, WeekdayThursday},
	"sexta-feira":   {TokenWeekday, WeekdayFriday},
	"sexta":         {TokenWeekday, WeekdayFriday},
	"sex":           {TokenWeekday, WeekdayFriday},
	"sábado":        {TokenWeekday, WeekdaySaturday},
	"sabado":        {TokenWeekday, WeekdaySaturday},
	"sáb":           {TokenWeekday, WeekdaySaturday},
	"sab":           {TokenWeekday, WeekdaySaturday},
	"domingo":       {TokenWeekday, WeekdaySunday},
	"dom":           {TokenWeekday, WeekdaySunday},

	// --- Months ---
	// "dez" is omitted — it maps to the number 10; write "dezembro" in full.
	"janeiro":   {TokenMonth, MonthJanuary},
	"jan":       {TokenMonth, MonthJanuary},
	"fevereiro": {TokenMonth, MonthFebruary},
	"fev":       {TokenMonth, MonthFebruary},
	"março":     {TokenMonth, MonthMarch},
	"marco":     {TokenMonth, MonthMarch},
	"mar":       {TokenMonth, MonthMarch},
	"abril":     {TokenMonth, MonthApril},
	"abr":       {TokenMonth, MonthApril},
	"maio":      {TokenMonth, MonthMay},
	"mai":       {TokenMonth, MonthMay}, // CLDR abbreviation
	"junho":     {TokenMonth, MonthJune},
	"jun":       {TokenMonth, MonthJune},
	"julho":     {TokenMonth, MonthJuly},
	"jul":       {TokenMonth, MonthJuly},
	"agosto":    {TokenMonth, MonthAugust},
	"ago":       {TokenMonth, MonthAugust},
	"setembro":  {TokenMonth, MonthSeptember},
	"set":       {TokenMonth, MonthSeptember},
	"outubro":   {TokenMonth, MonthOctober},
	"out":       {TokenMonth, MonthOctober},
	"novembro":  {TokenMonth, MonthNovember},
	"nov":       {TokenMonth, MonthNovember},
	"dezembro":  {TokenMonth, MonthDecember},

	// --- Direction ---
	"próximo":  {TokenDirection, DirectionFuture},
	"proximo":  {TokenDirection, DirectionFuture},
	"próxima":  {TokenDirection, DirectionFuture},
	"proxima":  {TokenDirection, DirectionFuture},
	"seguinte": {TokenDirection, DirectionFuture},
	"passado":  {TokenDirection, DirectionPast},
	"passada":  {TokenDirection, DirectionPast},
	"último":   {TokenDirection, DirectionPast},
	"ultimo":   {TokenDirection, DirectionPast},
	"última":   {TokenDirection, DirectionPast},
	"ultima":   {TokenDirection, DirectionPast},
	"anterior": {TokenDirection, DirectionPast},
	"este":     {TokenDirection, DirectionNearest},
	"esta":     {TokenDirection, DirectionNearest},
	"atual":    {TokenDirection, DirectionNearest},
	"actual":   {TokenDirection, DirectionNearest}, // European variant

	// --- Anchors ---
	"agora":     {TokenAnchor, AnchorNow},
	"hoje":      {TokenAnchor, AnchorToday},
	"amanhã":    {TokenAnchor, AnchorTomorrow},
	"amanha":    {TokenAnchor, AnchorTomorrow},
	"ontem":     {TokenAnchor, AnchorYesterday},
	"anteontem": {TokenAnchor, Anchor2DaysAgo},

	// --- Modifiers ---
	"há":     {TokenModifier, ModifierPast}, // European PT: "há 3 dias" = 3 days ago
	"ha":     {TokenModifier, ModifierPast}, // unaccented variant
	"faz":    {TokenModifier, ModifierPast}, // Brazilian PT: "faz 3 dias"
	"atrás":  {TokenModifier, ModifierPast},
	"atras":  {TokenModifier, ModifierPast},
	"depois": {TokenModifier, ModifierFuture}, // "3 dias depois" = 3 days later
	"antes":  {TokenModifier, ModifierPast},

	// --- Prepositions ---
	"em": {TokenPrep, nil},
	"às": {TokenPrep, nil}, // "às 9:30" = at 9:30
	"ao": {TokenPrep, nil}, // "ao meio-dia" = at noon

	// --- Fillers ---
	"o":   {TokenFiller, nil},
	"a":   {TokenFiller, nil},
	"os":  {TokenFiller, nil},
	"as":  {TokenFiller, nil},
	"de":  {TokenFiller, nil},
	"do":  {TokenFiller, nil},
	"da":  {TokenFiller, nil},
	"dos": {TokenFiller, nil},
	"das": {TokenFiller, nil},
	"e":   {TokenFiller, nil},
	"na":  {TokenFiller, nil},
	"no":  {TokenFiller, nil},

	// --- Units ---
	"segundo":   {TokenUnit, PeriodSecond},
	"segundos":  {TokenUnit, PeriodSecond},
	"minuto":    {TokenUnit, PeriodMinute},
	"minutos":   {TokenUnit, PeriodMinute},
	"min":       {TokenUnit, PeriodMinute}, // CLDR abbreviation
	"mins":      {TokenUnit, PeriodMinute}, // CLDR abbreviation (plural form)
	"hora":      {TokenUnit, PeriodHour},
	"horas":     {TokenUnit, PeriodHour},
	"dia":       {TokenUnit, PeriodDay},
	"dias":      {TokenUnit, PeriodDay},
	"semana":    {TokenUnit, PeriodWeek},
	"semanas":   {TokenUnit, PeriodWeek},
	"sem":       {TokenUnit, PeriodWeek}, // CLDR abbreviation
	"quinzena":  {TokenUnit, PeriodFortnight},
	"quinzenas": {TokenUnit, PeriodFortnight},
	"mês":       {TokenUnit, PeriodMonth},
	"mes":       {TokenUnit, PeriodMonth},
	"meses":     {TokenUnit, PeriodMonth},
	"ano":       {TokenUnit, PeriodYear},
	"anos":      {TokenUnit, PeriodYear},

	// --- AM/PM ---
	"am": {TokenAMPM, AMPMAm},
	"pm": {TokenAMPM, AMPMPm},

	// --- Time-word substitutes ---
	"meio-dia":   {TokenTime, "12:00"},
	"meia-noite": {TokenTime, "0:00"},

	// --- Multi-word phrases ---

	// Anchors
	"depois de amanhã": {TokenAnchor, Anchor2DaysFromNow},
	"depois de amanha": {TokenAnchor, Anchor2DaysFromNow},

	// Prepositions
	"daqui a":   {TokenPrep, nil}, // "daqui a 3 dias" = in 3 days
	"dentro de": {TokenPrep, nil}, // "dentro de 3 dias" = in 3 days

	// AM/PM time-of-day phrases
	"da manhã": {TokenAMPM, AMPMAm}, // "9 da manhã" = 9 AM
	"da manha": {TokenAMPM, AMPMAm},
	"da tarde": {TokenAMPM, AMPMPm}, // "3 da tarde" = 3 PM
	"da noite": {TokenAMPM, AMPMPm}, // "10 da noite" = 10 PM

	// --- Number words — Cardinals ---
	// "segundo"/"segunda" omitted — conflict with TokenUnit PeriodSecond.
	"um": {TokenInteger, 1}, "uma": {TokenInteger, 1},
	"dois": {TokenInteger, 2}, "duas": {TokenInteger, 2},
	"três": {TokenInteger, 3}, "tres": {TokenInteger, 3},
	"quatro":  {TokenInteger, 4},
	"cinco":   {TokenInteger, 5},
	"seis":    {TokenInteger, 6},
	"sete":    {TokenInteger, 7},
	"oito":    {TokenInteger, 8},
	"nove":    {TokenInteger, 9},
	"dez":     {TokenInteger, 10},
	"onze":    {TokenInteger, 11},
	"doze":    {TokenInteger, 12},
	"treze":   {TokenInteger, 13},
	"catorze": {TokenInteger, 14}, "quatorze": {TokenInteger, 14},
	"quinze":    {TokenInteger, 15},
	"dezasseis": {TokenInteger, 16}, "dezesseis": {TokenInteger, 16}, // EP/BP
	"dezassete": {TokenInteger, 17}, "dezessete": {TokenInteger, 17},
	"dezoito":  {TokenInteger, 18},
	"dezanove": {TokenInteger, 19}, "dezenove": {TokenInteger, 19},
	"vinte":  {TokenInteger, 20},
	"trinta": {TokenInteger, 30},

	// --- Number words — Ordinals ---
	// "quarta"/"quinta"/"sexta" omitted — map to weekdays above.
	"primeiro": {TokenInteger, 1}, "primeira": {TokenInteger, 1},
	"terceiro": {TokenInteger, 3}, "terceira": {TokenInteger, 3},
	"quarto": {TokenInteger, 4},
	"quinto": {TokenInteger, 5},
	"sexto":  {TokenInteger, 6},
	"sétimo": {TokenInteger, 7}, "setimo": {TokenInteger, 7},
	"oitavo": {TokenInteger, 8},
	"nono":   {TokenInteger, 9},
	"décimo": {TokenInteger, 10}, "decimo": {TokenInteger, 10},

	// Multi-word numbers
	"vinte e um": {TokenInteger, 21}, "vinte e uma": {TokenInteger, 21},
	"vinte e dois": {TokenInteger, 22}, "vinte e duas": {TokenInteger, 22},
	"vinte e três": {TokenInteger, 23}, "vinte e tres": {TokenInteger, 23},
	"vinte e quatro": {TokenInteger, 24},
	"trinta e um":    {TokenInteger, 31}, "trinta e uma": {TokenInteger, 31},
}
