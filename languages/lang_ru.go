package languages

import (
	. "github.com/client9/nowandlater/internal/engine"
)

// LangRu is the built-in Russian Lang.
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
var LangRu = Lang{
	Words:     russianWords,
	DateOrder: DMY,
}

var russianWords = map[string]WordEntry{
	// --- Weekdays (nominative + common oblique forms + 3-letter abbreviations) ---
	"понедельник":   {Type: TokenWeekday, Value: WeekdayMonday},
	"понедельника":  {Type: TokenWeekday, Value: WeekdayMonday},
	"понедельнику":  {Type: TokenWeekday, Value: WeekdayMonday},
	"понедельником": {Type: TokenWeekday, Value: WeekdayMonday},
	"понедельнике":  {Type: TokenWeekday, Value: WeekdayMonday},
	"пн":            {Type: TokenWeekday, Value: WeekdayMonday},
	"пнд":           {Type: TokenWeekday, Value: WeekdayMonday},

	"вторник":   {Type: TokenWeekday, Value: WeekdayTuesday},
	"вторника":  {Type: TokenWeekday, Value: WeekdayTuesday},
	"вторнику":  {Type: TokenWeekday, Value: WeekdayTuesday},
	"вторником": {Type: TokenWeekday, Value: WeekdayTuesday},
	"вторнике":  {Type: TokenWeekday, Value: WeekdayTuesday},
	"вт":        {Type: TokenWeekday, Value: WeekdayTuesday},
	"втр":       {Type: TokenWeekday, Value: WeekdayTuesday},

	"среда":  {Type: TokenWeekday, Value: WeekdayWednesday},
	"среды":  {Type: TokenWeekday, Value: WeekdayWednesday},
	"среде":  {Type: TokenWeekday, Value: WeekdayWednesday},
	"среду":  {Type: TokenWeekday, Value: WeekdayWednesday},
	"средой": {Type: TokenWeekday, Value: WeekdayWednesday},
	"ср":     {Type: TokenWeekday, Value: WeekdayWednesday},
	"срд":    {Type: TokenWeekday, Value: WeekdayWednesday},

	"четверг":   {Type: TokenWeekday, Value: WeekdayThursday},
	"четверга":  {Type: TokenWeekday, Value: WeekdayThursday},
	"четвергу":  {Type: TokenWeekday, Value: WeekdayThursday},
	"четвергом": {Type: TokenWeekday, Value: WeekdayThursday},
	"четверге":  {Type: TokenWeekday, Value: WeekdayThursday},
	"чт":        {Type: TokenWeekday, Value: WeekdayThursday},
	"чтв":       {Type: TokenWeekday, Value: WeekdayThursday},

	"пятница":  {Type: TokenWeekday, Value: WeekdayFriday},
	"пятницы":  {Type: TokenWeekday, Value: WeekdayFriday},
	"пятнице":  {Type: TokenWeekday, Value: WeekdayFriday},
	"пятницу":  {Type: TokenWeekday, Value: WeekdayFriday},
	"пятницей": {Type: TokenWeekday, Value: WeekdayFriday},
	"пт":       {Type: TokenWeekday, Value: WeekdayFriday},
	"птн":      {Type: TokenWeekday, Value: WeekdayFriday},

	"суббота":  {Type: TokenWeekday, Value: WeekdaySaturday},
	"субботы":  {Type: TokenWeekday, Value: WeekdaySaturday},
	"субботе":  {Type: TokenWeekday, Value: WeekdaySaturday},
	"субботу":  {Type: TokenWeekday, Value: WeekdaySaturday},
	"субботой": {Type: TokenWeekday, Value: WeekdaySaturday},
	"сб":       {Type: TokenWeekday, Value: WeekdaySaturday},
	"сбт":      {Type: TokenWeekday, Value: WeekdaySaturday},

	"воскресенье":  {Type: TokenWeekday, Value: WeekdaySunday},
	"воскресенья":  {Type: TokenWeekday, Value: WeekdaySunday},
	"воскресенью":  {Type: TokenWeekday, Value: WeekdaySunday},
	"воскресеньем": {Type: TokenWeekday, Value: WeekdaySunday},
	"воскресении":  {Type: TokenWeekday, Value: WeekdaySunday},
	"воскресение":  {Type: TokenWeekday, Value: WeekdaySunday}, // alternate spelling (без мягкого знака)
	"вс":           {Type: TokenWeekday, Value: WeekdaySunday},
	"вск":          {Type: TokenWeekday, Value: WeekdaySunday},

	// --- Months (nominative + genitive — genitive used in "24 марта 2026") ---
	"январь": {Type: TokenMonth, Value: MonthJanuary},
	"января": {Type: TokenMonth, Value: MonthJanuary},
	"январе": {Type: TokenMonth, Value: MonthJanuary},
	"янв":    {Type: TokenMonth, Value: MonthJanuary},

	"февраль": {Type: TokenMonth, Value: MonthFebruary},
	"февраля": {Type: TokenMonth, Value: MonthFebruary},
	"феврале": {Type: TokenMonth, Value: MonthFebruary},
	"фев":     {Type: TokenMonth, Value: MonthFebruary},
	"февр":    {Type: TokenMonth, Value: MonthFebruary}, // CLDR abbreviation

	"март":  {Type: TokenMonth, Value: MonthMarch},
	"марта": {Type: TokenMonth, Value: MonthMarch},
	"марте": {Type: TokenMonth, Value: MonthMarch},
	"мар":   {Type: TokenMonth, Value: MonthMarch}, // CLDR abbreviation

	"апрель": {Type: TokenMonth, Value: MonthApril},
	"апреля": {Type: TokenMonth, Value: MonthApril},
	"апреле": {Type: TokenMonth, Value: MonthApril},
	"апр":    {Type: TokenMonth, Value: MonthApril},

	"май": {Type: TokenMonth, Value: MonthMay},
	"мая": {Type: TokenMonth, Value: MonthMay},
	"мае": {Type: TokenMonth, Value: MonthMay},

	"июнь": {Type: TokenMonth, Value: MonthJune},
	"июня": {Type: TokenMonth, Value: MonthJune},
	"июне": {Type: TokenMonth, Value: MonthJune},
	"июн":  {Type: TokenMonth, Value: MonthJune}, // CLDR abbreviation

	"июль": {Type: TokenMonth, Value: MonthJuly},
	"июля": {Type: TokenMonth, Value: MonthJuly},
	"июле": {Type: TokenMonth, Value: MonthJuly},
	"июл":  {Type: TokenMonth, Value: MonthJuly}, // CLDR abbreviation

	"август":  {Type: TokenMonth, Value: MonthAugust},
	"августа": {Type: TokenMonth, Value: MonthAugust},
	"августе": {Type: TokenMonth, Value: MonthAugust},
	"авг":     {Type: TokenMonth, Value: MonthAugust},

	"сентябрь": {Type: TokenMonth, Value: MonthSeptember},
	"сентября": {Type: TokenMonth, Value: MonthSeptember},
	"сентябре": {Type: TokenMonth, Value: MonthSeptember},
	"сен":      {Type: TokenMonth, Value: MonthSeptember},
	"сент":     {Type: TokenMonth, Value: MonthSeptember},

	"октябрь": {Type: TokenMonth, Value: MonthOctober},
	"октября": {Type: TokenMonth, Value: MonthOctober},
	"октябре": {Type: TokenMonth, Value: MonthOctober},
	"окт":     {Type: TokenMonth, Value: MonthOctober},

	"ноябрь": {Type: TokenMonth, Value: MonthNovember},
	"ноября": {Type: TokenMonth, Value: MonthNovember},
	"ноябре": {Type: TokenMonth, Value: MonthNovember},
	"ноя":    {Type: TokenMonth, Value: MonthNovember},
	"нояб":   {Type: TokenMonth, Value: MonthNovember}, // CLDR abbreviation

	"декабрь": {Type: TokenMonth, Value: MonthDecember},
	"декабря": {Type: TokenMonth, Value: MonthDecember},
	"декабре": {Type: TokenMonth, Value: MonthDecember},
	"дек":     {Type: TokenMonth, Value: MonthDecember},

	// --- Direction — inflected adjective forms ---
	// Next (следующ-):
	"следующий":  {Type: TokenDirection, Value: DirectionFuture},
	"следующая":  {Type: TokenDirection, Value: DirectionFuture},
	"следующее":  {Type: TokenDirection, Value: DirectionFuture},
	"следующего": {Type: TokenDirection, Value: DirectionFuture},
	"следующей":  {Type: TokenDirection, Value: DirectionFuture},
	"следующему": {Type: TokenDirection, Value: DirectionFuture},
	"следующим":  {Type: TokenDirection, Value: DirectionFuture},
	"следующем":  {Type: TokenDirection, Value: DirectionFuture},
	"следующую":  {Type: TokenDirection, Value: DirectionFuture},
	// Last (прошл-):
	"прошлый":  {Type: TokenDirection, Value: DirectionPast},
	"прошлая":  {Type: TokenDirection, Value: DirectionPast},
	"прошлое":  {Type: TokenDirection, Value: DirectionPast},
	"прошлого": {Type: TokenDirection, Value: DirectionPast},
	"прошлой":  {Type: TokenDirection, Value: DirectionPast},
	"прошлому": {Type: TokenDirection, Value: DirectionPast},
	"прошлым":  {Type: TokenDirection, Value: DirectionPast},
	"прошлом":  {Type: TokenDirection, Value: DirectionPast},
	"прошлую":  {Type: TokenDirection, Value: DirectionPast},
	// This (эт-):
	"этот":     {Type: TokenDirection, Value: DirectionNearest},
	"эта":      {Type: TokenDirection, Value: DirectionNearest},
	"это":      {Type: TokenDirection, Value: DirectionNearest},
	"этого":    {Type: TokenDirection, Value: DirectionNearest},
	"этой":     {Type: TokenDirection, Value: DirectionNearest},
	"этому":    {Type: TokenDirection, Value: DirectionNearest},
	"этим":     {Type: TokenDirection, Value: DirectionNearest},
	"этом":     {Type: TokenDirection, Value: DirectionNearest},
	"эту":      {Type: TokenDirection, Value: DirectionNearest},
	"текущий":  {Type: TokenDirection, Value: DirectionNearest},
	"текущая":  {Type: TokenDirection, Value: DirectionNearest},
	"текущей":  {Type: TokenDirection, Value: DirectionNearest},
	"нынешний": {Type: TokenDirection, Value: DirectionNearest},
	"нынешняя": {Type: TokenDirection, Value: DirectionNearest},
	"нынешнее": {Type: TokenDirection, Value: DirectionNearest},

	// --- Anchors ---
	"сейчас":           {Type: TokenAnchor, Value: AnchorNow},
	"сегодня":          {Type: TokenAnchor, Value: AnchorToday},
	"завтра":           {Type: TokenAnchor, Value: AnchorTomorrow},
	"вчера":            {Type: TokenAnchor, Value: AnchorYesterday},
	"послезавтра":      {Type: TokenAnchor, Value: Anchor2DaysFromNow},
	"позавчера":        {Type: TokenAnchor, Value: Anchor2DaysAgo},
	"послепослезавтра": {Type: TokenAnchor, Value: Anchor3DaysFromNow}, // supplementary data

	// --- Modifiers ---
	"назад": {Type: TokenModifier, Value: ModifierPast}, // "3 дня назад" = 3 days ago → INTEGER UNIT MODIFIER

	// --- Prepositions ---
	"в":      {Type: TokenPrep, Value: nil}, // в марте, в 15:00
	"через":  {Type: TokenPrep, Value: nil}, // через 3 дня = in 3 days → PREP INTEGER UNIT
	"на":     {Type: TokenPrep, Value: nil}, // на следующей неделе
	"спустя": {Type: TokenPrep, Value: nil}, // спустя 3 дня = after/in 3 days → PREP INTEGER UNIT

	// --- Multi-word preposition ---
	"в течение": {Type: TokenPrep, Value: nil}, // в течение 3 дней = within 3 days → PREP INTEGER UNIT

	// --- Fillers ---
	"и":        {Type: TokenFiller, Value: nil},
	"во":       {Type: TokenFiller, Value: nil}, // variant of "в" before consonant clusters ("во вторник")
	"около":    {Type: TokenFiller, Value: nil}, // "около 3 дней назад" = approximately 3 days ago
	"примерно": {Type: TokenFiller, Value: nil}, // "примерно 3 дня назад" = approximately 3 days ago

	// --- Units (nominative + numeral agreement forms) ---
	// After 1: секунда, минута, час, день, неделя, месяц, год
	// After 2–4: секунды, минуты, часа, дня, недели, месяца, года
	// After 5+: секунд, минут, часов, дней, недель, месяцев, лет
	"секунда": {Type: TokenUnit, Value: PeriodSecond},
	"секунды": {Type: TokenUnit, Value: PeriodSecond},
	"секунд":  {Type: TokenUnit, Value: PeriodSecond},
	"секунду": {Type: TokenUnit, Value: PeriodSecond},
	"сек":     {Type: TokenUnit, Value: PeriodSecond},

	"минута": {Type: TokenUnit, Value: PeriodMinute},
	"минуты": {Type: TokenUnit, Value: PeriodMinute},
	"минут":  {Type: TokenUnit, Value: PeriodMinute},
	"минуту": {Type: TokenUnit, Value: PeriodMinute},
	"мин":    {Type: TokenUnit, Value: PeriodMinute},

	"час":   {Type: TokenUnit, Value: PeriodHour},
	"часа":  {Type: TokenUnit, Value: PeriodHour},
	"часов": {Type: TokenUnit, Value: PeriodHour},
	"часу":  {Type: TokenUnit, Value: PeriodHour},
	"часом": {Type: TokenUnit, Value: PeriodHour},
	"часе":  {Type: TokenUnit, Value: PeriodHour},

	"день":  {Type: TokenUnit, Value: PeriodDay},
	"дня":   {Type: TokenUnit, Value: PeriodDay},
	"дней":  {Type: TokenUnit, Value: PeriodDay},
	"дню":   {Type: TokenUnit, Value: PeriodDay},
	"днём":  {Type: TokenUnit, Value: PeriodDay},
	"дне":   {Type: TokenUnit, Value: PeriodDay},
	"дн":    {Type: TokenUnit, Value: PeriodDay}, // CLDR abbreviation
	"сутки": {Type: TokenUnit, Value: PeriodDay}, // 24-hour period; supplementary data
	"суток": {Type: TokenUnit, Value: PeriodDay}, // genitive plural of "сутки"

	"неделя":  {Type: TokenUnit, Value: PeriodWeek},
	"недели":  {Type: TokenUnit, Value: PeriodWeek},
	"недель":  {Type: TokenUnit, Value: PeriodWeek},
	"неделю":  {Type: TokenUnit, Value: PeriodWeek},
	"неделей": {Type: TokenUnit, Value: PeriodWeek},
	"неделе":  {Type: TokenUnit, Value: PeriodWeek},
	"нед":     {Type: TokenUnit, Value: PeriodWeek}, // CLDR abbreviation

	"месяц":   {Type: TokenUnit, Value: PeriodMonth},
	"месяца":  {Type: TokenUnit, Value: PeriodMonth},
	"месяцев": {Type: TokenUnit, Value: PeriodMonth},
	"месяцу":  {Type: TokenUnit, Value: PeriodMonth},
	"месяце":  {Type: TokenUnit, Value: PeriodMonth},
	"мес":     {Type: TokenUnit, Value: PeriodMonth}, // CLDR abbreviation

	"год":   {Type: TokenUnit, Value: PeriodYear},
	"года":  {Type: TokenUnit, Value: PeriodYear}, // also genitive after 4-digit year; see Known Limitations
	"лет":   {Type: TokenUnit, Value: PeriodYear},
	"году":  {Type: TokenUnit, Value: PeriodYear},
	"годом": {Type: TokenUnit, Value: PeriodYear},

	// --- AM/PM (CLDR) ---
	"дп": {Type: TokenAMPM, Value: AMPMAm}, // до полудня = before noon (AM)
	"пп": {Type: TokenAMPM, Value: AMPMPm}, // после полудня = after noon (PM)

	// --- Time-word substitutes ---
	"полдень": {Type: TokenTime, Value: "12:00"}, // noon
	"полночь": {Type: TokenTime, Value: "0:00"},  // midnight

	// --- Number words — Cardinals ---
	"один":         {Type: TokenInteger, Value: 1},
	"одна":         {Type: TokenInteger, Value: 1},
	"одно":         {Type: TokenInteger, Value: 1},
	"одного":       {Type: TokenInteger, Value: 1},
	"одной":        {Type: TokenInteger, Value: 1},
	"одному":       {Type: TokenInteger, Value: 1},
	"одним":        {Type: TokenInteger, Value: 1},
	"одну":         {Type: TokenInteger, Value: 1},
	"два":          {Type: TokenInteger, Value: 2},
	"две":          {Type: TokenInteger, Value: 2},
	"двух":         {Type: TokenInteger, Value: 2},
	"двум":         {Type: TokenInteger, Value: 2},
	"двумя":        {Type: TokenInteger, Value: 2},
	"три":          {Type: TokenInteger, Value: 3},
	"трёх":         {Type: TokenInteger, Value: 3},
	"трём":         {Type: TokenInteger, Value: 3},
	"тремя":        {Type: TokenInteger, Value: 3},
	"четыре":       {Type: TokenInteger, Value: 4},
	"четырёх":      {Type: TokenInteger, Value: 4},
	"пять":         {Type: TokenInteger, Value: 5},
	"пяти":         {Type: TokenInteger, Value: 5},
	"шесть":        {Type: TokenInteger, Value: 6},
	"шести":        {Type: TokenInteger, Value: 6},
	"семь":         {Type: TokenInteger, Value: 7},
	"семи":         {Type: TokenInteger, Value: 7},
	"восемь":       {Type: TokenInteger, Value: 8},
	"восьми":       {Type: TokenInteger, Value: 8},
	"девять":       {Type: TokenInteger, Value: 9},
	"девяти":       {Type: TokenInteger, Value: 9},
	"десять":       {Type: TokenInteger, Value: 10},
	"десяти":       {Type: TokenInteger, Value: 10},
	"одиннадцать":  {Type: TokenInteger, Value: 11},
	"одиннадцати":  {Type: TokenInteger, Value: 11},
	"двенадцать":   {Type: TokenInteger, Value: 12},
	"двенадцати":   {Type: TokenInteger, Value: 12},
	"тринадцать":   {Type: TokenInteger, Value: 13},
	"тринадцати":   {Type: TokenInteger, Value: 13},
	"четырнадцать": {Type: TokenInteger, Value: 14},
	"четырнадцати": {Type: TokenInteger, Value: 14},
	"пятнадцать":   {Type: TokenInteger, Value: 15},
	"пятнадцати":   {Type: TokenInteger, Value: 15},
	"шестнадцать":  {Type: TokenInteger, Value: 16},
	"шестнадцати":  {Type: TokenInteger, Value: 16},
	"семнадцать":   {Type: TokenInteger, Value: 17},
	"семнадцати":   {Type: TokenInteger, Value: 17},
	"восемнадцать": {Type: TokenInteger, Value: 18},
	"восемнадцати": {Type: TokenInteger, Value: 18},
	"девятнадцать": {Type: TokenInteger, Value: 19},
	"девятнадцати": {Type: TokenInteger, Value: 19},
	"двадцать":     {Type: TokenInteger, Value: 20},
	"двадцати":     {Type: TokenInteger, Value: 20},
	"тридцать":     {Type: TokenInteger, Value: 30},
	"тридцати":     {Type: TokenInteger, Value: 30},
}
