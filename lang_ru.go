package nowandlater

// Russian is the built-in Russian Lang.
//
// Russian uses standard whitespace tokenization but has significant case
// inflection: weekdays, months, and units each appear in multiple grammatical
// forms. All common inflected forms are mapped to the same token constant so
// that patterns like "следующей неделе" (next week, prepositional) and
// "следующая неделя" (next week, nominative) resolve identically.
//
// Numeral agreement: Russian distinguishes секунда (1), секунды (2–4),
// секунд (5+). All three forms map to the same TokenUnit; the library does
// not enforce agreement, so "5 секунда" parses as well as "5 секунд".
//
// Known limitations:
//   - Ordinal day numbers (первое, второго…) are not supported; use digits.
//   - "2026 года" (year + genitive suffix) produces YEAR UNIT and fails;
//     write the year alone: "24 марта 2026".
//   - Compound relative expressions ("позавчера вечером") are not supported.
//   - Single-char unit abbreviations "г" (год), "ч" (час), "д" (день), "с" (секунда)
//     are intentionally omitted to avoid false positives.
var Russian = Lang{
	Words:     russianWords,
	DateOrder: DMY,
}

var russianWords = map[string]WordEntry{
	// --- Weekdays (nominative + common oblique forms + 3-letter abbreviations) ---
	"понедельник":   {TokenWeekday, WeekdayMonday},
	"понедельника":  {TokenWeekday, WeekdayMonday},
	"понедельнику":  {TokenWeekday, WeekdayMonday},
	"понедельником": {TokenWeekday, WeekdayMonday},
	"понедельнике":  {TokenWeekday, WeekdayMonday},
	"пн":            {TokenWeekday, WeekdayMonday},
	"пнд":           {TokenWeekday, WeekdayMonday},

	"вторник":   {TokenWeekday, WeekdayTuesday},
	"вторника":  {TokenWeekday, WeekdayTuesday},
	"вторнику":  {TokenWeekday, WeekdayTuesday},
	"вторником": {TokenWeekday, WeekdayTuesday},
	"вторнике":  {TokenWeekday, WeekdayTuesday},
	"вт":        {TokenWeekday, WeekdayTuesday},
	"втр":       {TokenWeekday, WeekdayTuesday},

	"среда":  {TokenWeekday, WeekdayWednesday},
	"среды":  {TokenWeekday, WeekdayWednesday},
	"среде":  {TokenWeekday, WeekdayWednesday},
	"среду":  {TokenWeekday, WeekdayWednesday},
	"средой": {TokenWeekday, WeekdayWednesday},
	"ср":     {TokenWeekday, WeekdayWednesday},
	"срд":    {TokenWeekday, WeekdayWednesday},

	"четверг":   {TokenWeekday, WeekdayThursday},
	"четверга":  {TokenWeekday, WeekdayThursday},
	"четвергу":  {TokenWeekday, WeekdayThursday},
	"четвергом": {TokenWeekday, WeekdayThursday},
	"четверге":  {TokenWeekday, WeekdayThursday},
	"чт":        {TokenWeekday, WeekdayThursday},
	"чтв":       {TokenWeekday, WeekdayThursday},

	"пятница":  {TokenWeekday, WeekdayFriday},
	"пятницы":  {TokenWeekday, WeekdayFriday},
	"пятнице":  {TokenWeekday, WeekdayFriday},
	"пятницу":  {TokenWeekday, WeekdayFriday},
	"пятницей": {TokenWeekday, WeekdayFriday},
	"пт":       {TokenWeekday, WeekdayFriday},
	"птн":      {TokenWeekday, WeekdayFriday},

	"суббота":  {TokenWeekday, WeekdaySaturday},
	"субботы":  {TokenWeekday, WeekdaySaturday},
	"субботе":  {TokenWeekday, WeekdaySaturday},
	"субботу":  {TokenWeekday, WeekdaySaturday},
	"субботой": {TokenWeekday, WeekdaySaturday},
	"сб":       {TokenWeekday, WeekdaySaturday},
	"сбт":      {TokenWeekday, WeekdaySaturday},

	"воскресенье":  {TokenWeekday, WeekdaySunday},
	"воскресенья":  {TokenWeekday, WeekdaySunday},
	"воскресенью":  {TokenWeekday, WeekdaySunday},
	"воскресеньем": {TokenWeekday, WeekdaySunday},
	"воскресении":  {TokenWeekday, WeekdaySunday},
	"воскресение":  {TokenWeekday, WeekdaySunday}, // alternate spelling (без мягкого знака)
	"вс":           {TokenWeekday, WeekdaySunday},
	"вск":          {TokenWeekday, WeekdaySunday},

	// --- Months (nominative + genitive — genitive used in "24 марта 2026") ---
	"январь": {TokenMonth, MonthJanuary},
	"января": {TokenMonth, MonthJanuary},
	"январе": {TokenMonth, MonthJanuary},
	"янв":    {TokenMonth, MonthJanuary},

	"февраль": {TokenMonth, MonthFebruary},
	"февраля": {TokenMonth, MonthFebruary},
	"феврале": {TokenMonth, MonthFebruary},
	"фев":     {TokenMonth, MonthFebruary},
	"февр":    {TokenMonth, MonthFebruary}, // CLDR abbreviation

	"март":  {TokenMonth, MonthMarch},
	"марта": {TokenMonth, MonthMarch},
	"марте": {TokenMonth, MonthMarch},
	"мар":   {TokenMonth, MonthMarch}, // CLDR abbreviation

	"апрель": {TokenMonth, MonthApril},
	"апреля": {TokenMonth, MonthApril},
	"апреле": {TokenMonth, MonthApril},
	"апр":    {TokenMonth, MonthApril},

	"май": {TokenMonth, MonthMay},
	"мая": {TokenMonth, MonthMay},
	"мае": {TokenMonth, MonthMay},

	"июнь": {TokenMonth, MonthJune},
	"июня": {TokenMonth, MonthJune},
	"июне": {TokenMonth, MonthJune},
	"июн":  {TokenMonth, MonthJune}, // CLDR abbreviation

	"июль": {TokenMonth, MonthJuly},
	"июля": {TokenMonth, MonthJuly},
	"июле": {TokenMonth, MonthJuly},
	"июл":  {TokenMonth, MonthJuly}, // CLDR abbreviation

	"август":  {TokenMonth, MonthAugust},
	"августа": {TokenMonth, MonthAugust},
	"августе": {TokenMonth, MonthAugust},
	"авг":     {TokenMonth, MonthAugust},

	"сентябрь": {TokenMonth, MonthSeptember},
	"сентября": {TokenMonth, MonthSeptember},
	"сентябре": {TokenMonth, MonthSeptember},
	"сен":      {TokenMonth, MonthSeptember},
	"сент":     {TokenMonth, MonthSeptember},

	"октябрь": {TokenMonth, MonthOctober},
	"октября": {TokenMonth, MonthOctober},
	"октябре": {TokenMonth, MonthOctober},
	"окт":     {TokenMonth, MonthOctober},

	"ноябрь": {TokenMonth, MonthNovember},
	"ноября": {TokenMonth, MonthNovember},
	"ноябре": {TokenMonth, MonthNovember},
	"ноя":    {TokenMonth, MonthNovember},
	"нояб":   {TokenMonth, MonthNovember}, // CLDR abbreviation

	"декабрь": {TokenMonth, MonthDecember},
	"декабря": {TokenMonth, MonthDecember},
	"декабре": {TokenMonth, MonthDecember},
	"дек":     {TokenMonth, MonthDecember},

	// --- Direction — inflected adjective forms ---
	// Next (следующ-):
	"следующий":  {TokenDirection, DirectionFuture},
	"следующая":  {TokenDirection, DirectionFuture},
	"следующее":  {TokenDirection, DirectionFuture},
	"следующего": {TokenDirection, DirectionFuture},
	"следующей":  {TokenDirection, DirectionFuture},
	"следующему": {TokenDirection, DirectionFuture},
	"следующим":  {TokenDirection, DirectionFuture},
	"следующем":  {TokenDirection, DirectionFuture},
	"следующую":  {TokenDirection, DirectionFuture},
	// Last (прошл-):
	"прошлый":  {TokenDirection, DirectionPast},
	"прошлая":  {TokenDirection, DirectionPast},
	"прошлое":  {TokenDirection, DirectionPast},
	"прошлого": {TokenDirection, DirectionPast},
	"прошлой":  {TokenDirection, DirectionPast},
	"прошлому": {TokenDirection, DirectionPast},
	"прошлым":  {TokenDirection, DirectionPast},
	"прошлом":  {TokenDirection, DirectionPast},
	"прошлую":  {TokenDirection, DirectionPast},
	// This (эт-):
	"этот":     {TokenDirection, DirectionNearest},
	"эта":      {TokenDirection, DirectionNearest},
	"это":      {TokenDirection, DirectionNearest},
	"этого":    {TokenDirection, DirectionNearest},
	"этой":     {TokenDirection, DirectionNearest},
	"этому":    {TokenDirection, DirectionNearest},
	"этим":     {TokenDirection, DirectionNearest},
	"этом":     {TokenDirection, DirectionNearest},
	"эту":      {TokenDirection, DirectionNearest},
	"текущий":  {TokenDirection, DirectionNearest},
	"текущая":  {TokenDirection, DirectionNearest},
	"текущей":  {TokenDirection, DirectionNearest},
	"нынешний": {TokenDirection, DirectionNearest},
	"нынешняя": {TokenDirection, DirectionNearest},
	"нынешнее": {TokenDirection, DirectionNearest},

	// --- Anchors ---
	"сейчас":           {TokenAnchor, AnchorNow},
	"сегодня":          {TokenAnchor, AnchorToday},
	"завтра":           {TokenAnchor, AnchorTomorrow},
	"вчера":            {TokenAnchor, AnchorYesterday},
	"послезавтра":      {TokenAnchor, Anchor2DaysFromNow},
	"позавчера":        {TokenAnchor, Anchor2DaysAgo},
	"послепослезавтра": {TokenAnchor, Anchor3DaysFromNow}, // supplementary data

	// --- Modifiers ---
	"назад": {TokenModifier, ModifierPast}, // "3 дня назад" = 3 days ago → INTEGER UNIT MODIFIER

	// --- Prepositions ---
	"в":      {TokenPrep, nil}, // в марте, в 15:00
	"через":  {TokenPrep, nil}, // через 3 дня = in 3 days → PREP INTEGER UNIT
	"на":     {TokenPrep, nil}, // на следующей неделе
	"спустя": {TokenPrep, nil}, // спустя 3 дня = after/in 3 days → PREP INTEGER UNIT

	// --- Multi-word preposition ---
	"в течение": {TokenPrep, nil}, // в течение 3 дней = within 3 days → PREP INTEGER UNIT

	// --- Fillers ---
	"и":        {TokenFiller, nil},
	"во":       {TokenFiller, nil}, // variant of "в" before consonant clusters ("во вторник")
	"около":    {TokenFiller, nil}, // "около 3 дней назад" = approximately 3 days ago
	"примерно": {TokenFiller, nil}, // "примерно 3 дня назад" = approximately 3 days ago

	// --- Units (nominative + numeral agreement forms) ---
	// After 1: секунда, минута, час, день, неделя, месяц, год
	// After 2–4: секунды, минуты, часа, дня, недели, месяца, года
	// After 5+: секунд, минут, часов, дней, недель, месяцев, лет
	"секунда": {TokenUnit, PeriodSecond},
	"секунды": {TokenUnit, PeriodSecond},
	"секунд":  {TokenUnit, PeriodSecond},
	"секунду": {TokenUnit, PeriodSecond},
	"сек":     {TokenUnit, PeriodSecond},

	"минута": {TokenUnit, PeriodMinute},
	"минуты": {TokenUnit, PeriodMinute},
	"минут":  {TokenUnit, PeriodMinute},
	"минуту": {TokenUnit, PeriodMinute},
	"мин":    {TokenUnit, PeriodMinute},

	"час":   {TokenUnit, PeriodHour},
	"часа":  {TokenUnit, PeriodHour},
	"часов": {TokenUnit, PeriodHour},
	"часу":  {TokenUnit, PeriodHour},
	"часом": {TokenUnit, PeriodHour},
	"часе":  {TokenUnit, PeriodHour},

	"день":  {TokenUnit, PeriodDay},
	"дня":   {TokenUnit, PeriodDay},
	"дней":  {TokenUnit, PeriodDay},
	"дню":   {TokenUnit, PeriodDay},
	"днём":  {TokenUnit, PeriodDay},
	"дне":   {TokenUnit, PeriodDay},
	"дн":    {TokenUnit, PeriodDay}, // CLDR abbreviation
	"сутки": {TokenUnit, PeriodDay}, // 24-hour period; supplementary data
	"суток": {TokenUnit, PeriodDay}, // genitive plural of "сутки"

	"неделя":  {TokenUnit, PeriodWeek},
	"недели":  {TokenUnit, PeriodWeek},
	"недель":  {TokenUnit, PeriodWeek},
	"неделю":  {TokenUnit, PeriodWeek},
	"неделей": {TokenUnit, PeriodWeek},
	"неделе":  {TokenUnit, PeriodWeek},
	"нед":     {TokenUnit, PeriodWeek}, // CLDR abbreviation

	"месяц":   {TokenUnit, PeriodMonth},
	"месяца":  {TokenUnit, PeriodMonth},
	"месяцев": {TokenUnit, PeriodMonth},
	"месяцу":  {TokenUnit, PeriodMonth},
	"месяце":  {TokenUnit, PeriodMonth},
	"мес":     {TokenUnit, PeriodMonth}, // CLDR abbreviation

	"год":   {TokenUnit, PeriodYear},
	"года":  {TokenUnit, PeriodYear}, // also genitive after 4-digit year; see Known Limitations
	"лет":   {TokenUnit, PeriodYear},
	"году":  {TokenUnit, PeriodYear},
	"годом": {TokenUnit, PeriodYear},

	// --- AM/PM (CLDR) ---
	"дп": {TokenAMPM, AMPMAm}, // до полудня = before noon (AM)
	"пп": {TokenAMPM, AMPMPm}, // после полудня = after noon (PM)

	// --- Time-word substitutes ---
	"полдень": {TokenTime, "12:00"}, // noon
	"полночь": {TokenTime, "0:00"},  // midnight

	// --- Number words — Cardinals ---
	"один":         {TokenInteger, 1},
	"одна":         {TokenInteger, 1},
	"одно":         {TokenInteger, 1},
	"одного":       {TokenInteger, 1},
	"одной":        {TokenInteger, 1},
	"одному":       {TokenInteger, 1},
	"одним":        {TokenInteger, 1},
	"одну":         {TokenInteger, 1},
	"два":          {TokenInteger, 2},
	"две":          {TokenInteger, 2},
	"двух":         {TokenInteger, 2},
	"двум":         {TokenInteger, 2},
	"двумя":        {TokenInteger, 2},
	"три":          {TokenInteger, 3},
	"трёх":         {TokenInteger, 3},
	"трём":         {TokenInteger, 3},
	"тремя":        {TokenInteger, 3},
	"четыре":       {TokenInteger, 4},
	"четырёх":      {TokenInteger, 4},
	"пять":         {TokenInteger, 5},
	"пяти":         {TokenInteger, 5},
	"шесть":        {TokenInteger, 6},
	"шести":        {TokenInteger, 6},
	"семь":         {TokenInteger, 7},
	"семи":         {TokenInteger, 7},
	"восемь":       {TokenInteger, 8},
	"восьми":       {TokenInteger, 8},
	"девять":       {TokenInteger, 9},
	"девяти":       {TokenInteger, 9},
	"десять":       {TokenInteger, 10},
	"десяти":       {TokenInteger, 10},
	"одиннадцать":  {TokenInteger, 11},
	"одиннадцати":  {TokenInteger, 11},
	"двенадцать":   {TokenInteger, 12},
	"двенадцати":   {TokenInteger, 12},
	"тринадцать":   {TokenInteger, 13},
	"тринадцати":   {TokenInteger, 13},
	"четырнадцать": {TokenInteger, 14},
	"четырнадцати": {TokenInteger, 14},
	"пятнадцать":   {TokenInteger, 15},
	"пятнадцати":   {TokenInteger, 15},
	"шестнадцать":  {TokenInteger, 16},
	"шестнадцати":  {TokenInteger, 16},
	"семнадцать":   {TokenInteger, 17},
	"семнадцати":   {TokenInteger, 17},
	"восемнадцать": {TokenInteger, 18},
	"восемнадцати": {TokenInteger, 18},
	"девятнадцать": {TokenInteger, 19},
	"девятнадцати": {TokenInteger, 19},
	"двадцать":     {TokenInteger, 20},
	"двадцати":     {TokenInteger, 20},
	"тридцать":     {TokenInteger, 30},
	"тридцати":     {TokenInteger, 30},
}
