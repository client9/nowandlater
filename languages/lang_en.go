package languages

import (
	. "github.com/client9/nowandlater/internal/engine"
)

// LangEn is the built-in English Lang.
var LangEn = Lang{
	Words:           englishWords,
	OrdinalSuffixes: []string{"st", "nd", "rd", "th"},
}

// englishWords is the word table for English.
// It covers single words, multi-word phrases (space-containing keys),
// time-word substitutions, and number words — all in one map.
var englishWords = map[string]WordEntry{
	// Weekdays
	"monday":    {Type: TokenWeekday, Value: WeekdayMonday},
	"mon":       {Type: TokenWeekday, Value: WeekdayMonday},
	"tuesday":   {Type: TokenWeekday, Value: WeekdayTuesday},
	"tu":        {Type: TokenWeekday, Value: WeekdayTuesday},
	"tue":       {Type: TokenWeekday, Value: WeekdayTuesday},
	"tues":      {Type: TokenWeekday, Value: WeekdayTuesday},
	"wednesday": {Type: TokenWeekday, Value: WeekdayWednesday},
	"wed":       {Type: TokenWeekday, Value: WeekdayWednesday},
	"wednes":    {Type: TokenWeekday, Value: WeekdayWednesday},
	"thursday":  {Type: TokenWeekday, Value: WeekdayThursday},
	"thu":       {Type: TokenWeekday, Value: WeekdayThursday},
	"thur":      {Type: TokenWeekday, Value: WeekdayThursday},
	"thurs":     {Type: TokenWeekday, Value: WeekdayThursday},
	"friday":    {Type: TokenWeekday, Value: WeekdayFriday},
	"fri":       {Type: TokenWeekday, Value: WeekdayFriday},
	"saturday":  {Type: TokenWeekday, Value: WeekdaySaturday},
	"sat":       {Type: TokenWeekday, Value: WeekdaySaturday},
	"sunday":    {Type: TokenWeekday, Value: WeekdaySunday},
	"su":        {Type: TokenWeekday, Value: WeekdaySunday},
	"sun":       {Type: TokenWeekday, Value: WeekdaySunday},

	// Months
	"january":   {Type: TokenMonth, Value: MonthJanuary},
	"jan":       {Type: TokenMonth, Value: MonthJanuary},
	"february":  {Type: TokenMonth, Value: MonthFebruary},
	"feb":       {Type: TokenMonth, Value: MonthFebruary},
	"march":     {Type: TokenMonth, Value: MonthMarch},
	"mar":       {Type: TokenMonth, Value: MonthMarch},
	"april":     {Type: TokenMonth, Value: MonthApril},
	"apr":       {Type: TokenMonth, Value: MonthApril},
	"may":       {Type: TokenMonth, Value: MonthMay},
	"june":      {Type: TokenMonth, Value: MonthJune},
	"jun":       {Type: TokenMonth, Value: MonthJune},
	"july":      {Type: TokenMonth, Value: MonthJuly},
	"jul":       {Type: TokenMonth, Value: MonthJuly},
	"august":    {Type: TokenMonth, Value: MonthAugust},
	"aug":       {Type: TokenMonth, Value: MonthAugust},
	"september": {Type: TokenMonth, Value: MonthSeptember},
	"sep":       {Type: TokenMonth, Value: MonthSeptember},
	"sept":      {Type: TokenMonth, Value: MonthSeptember},
	"october":   {Type: TokenMonth, Value: MonthOctober},
	"oct":       {Type: TokenMonth, Value: MonthOctober},
	"november":  {Type: TokenMonth, Value: MonthNovember},
	"nov":       {Type: TokenMonth, Value: MonthNovember},
	"december":  {Type: TokenMonth, Value: MonthDecember},
	"dec":       {Type: TokenMonth, Value: MonthDecember},

	// Direction
	"next":     {Type: TokenDirection, Value: DirectionFuture},
	"last":     {Type: TokenDirection, Value: DirectionPast},
	"previous": {Type: TokenDirection, Value: DirectionPast},
	"prev":     {Type: TokenDirection, Value: DirectionPast},
	"coming":   {Type: TokenDirection, Value: DirectionFuture},
	"this":     {Type: TokenDirection, Value: DirectionNearest},

	// Modifier
	"ago":    {Type: TokenModifier, Value: ModifierPast},
	"hence":  {Type: TokenModifier, Value: ModifierFuture},
	"from":   {Type: TokenModifier, Value: ModifierFuture},
	"later":  {Type: TokenModifier, Value: ModifierFuture},
	"before": {Type: TokenModifier, Value: ModifierPast},
	"after":  {Type: TokenModifier, Value: ModifierFuture},

	// Anchor
	"now":       {Type: TokenAnchor, Value: AnchorNow},
	"today":     {Type: TokenAnchor, Value: AnchorToday},
	"tomorrow":  {Type: TokenAnchor, Value: AnchorTomorrow},
	"yesterday": {Type: TokenAnchor, Value: AnchorYesterday},

	// Preposition (value not consumed semantically)
	"at": {Type: TokenPrep, Value: nil},
	"on": {Type: TokenPrep, Value: nil},
	"in": {Type: TokenPrep, Value: nil},
	"by": {Type: TokenPrep, Value: nil},

	// AM/PM (already normalized from a.m./p.m. by the normalize step)
	"am": {Type: TokenAMPM, Value: AMPMAm},
	"pm": {Type: TokenAMPM, Value: AMPMPm},

	// Units (singular and plural — all variants of the same period carry the same Period constant)
	"second":     {Type: TokenUnit, Value: PeriodSecond},
	"seconds":    {Type: TokenUnit, Value: PeriodSecond},
	"sec":        {Type: TokenUnit, Value: PeriodSecond},
	"secs":       {Type: TokenUnit, Value: PeriodSecond},
	"minute":     {Type: TokenUnit, Value: PeriodMinute},
	"minutes":    {Type: TokenUnit, Value: PeriodMinute},
	"min":        {Type: TokenUnit, Value: PeriodMinute},
	"mins":       {Type: TokenUnit, Value: PeriodMinute},
	"hour":       {Type: TokenUnit, Value: PeriodHour},
	"hours":      {Type: TokenUnit, Value: PeriodHour},
	"hr":         {Type: TokenUnit, Value: PeriodHour},
	"hrs":        {Type: TokenUnit, Value: PeriodHour},
	"day":        {Type: TokenUnit, Value: PeriodDay},
	"days":       {Type: TokenUnit, Value: PeriodDay},
	"fortnight":  {Type: TokenUnit, Value: PeriodFortnight},
	"fortnights": {Type: TokenUnit, Value: PeriodFortnight},
	"week":       {Type: TokenUnit, Value: PeriodWeek},
	"weeks":      {Type: TokenUnit, Value: PeriodWeek},
	"wk":         {Type: TokenUnit, Value: PeriodWeek},
	"month":      {Type: TokenUnit, Value: PeriodMonth},
	"months":     {Type: TokenUnit, Value: PeriodMonth},
	"mo":         {Type: TokenUnit, Value: PeriodMonth},
	"year":       {Type: TokenUnit, Value: PeriodYear},
	"years":      {Type: TokenUnit, Value: PeriodYear},
	"yr":         {Type: TokenUnit, Value: PeriodYear},
	"yrs":        {Type: TokenUnit, Value: PeriodYear},

	// Filler (grammatical noise — value not consumed semantically)
	"the":   {Type: TokenFiller, Value: nil},
	"of":    {Type: TokenFiller, Value: nil},
	"a":     {Type: TokenFiller, Value: nil},
	"an":    {Type: TokenFiller, Value: nil},
	"and":   {Type: TokenFiller, Value: nil},
	"about": {Type: TokenFiller, Value: nil},
	"just":  {Type: TokenFiller, Value: nil},

	// Multi-word anchors (matched longest-first by tokenizer)
	"day before yesterday": {Type: TokenAnchor, Value: Anchor2DaysAgo},
	"day after tomorrow":   {Type: TokenAnchor, Value: Anchor2DaysFromNow},

	// Time-word substitutions — produce TokenTime directly
	// ("noon" and "midnight" as words tokenize to TIME rather than requiring preprocessing)
	"noon":     {Type: TokenTime, Value: "12:00"},
	"midnight": {Type: TokenTime, Value: "0:00"},

	// Number words — Cardinals (1–30)
	// "second" is intentionally absent: it conflicts with TokenUnit PeriodSecond.
	// Use "2nd" (ordinal suffix stripped by OrdinalSuffixes) for the 2nd day of month.
	"one": {Type: TokenInteger, Value: 1}, "two": {Type: TokenInteger, Value: 2}, "three": {Type: TokenInteger, Value: 3},
	"four": {Type: TokenInteger, Value: 4}, "five": {Type: TokenInteger, Value: 5}, "six": {Type: TokenInteger, Value: 6},
	"seven": {Type: TokenInteger, Value: 7}, "eight": {Type: TokenInteger, Value: 8}, "nine": {Type: TokenInteger, Value: 9},
	"ten": {Type: TokenInteger, Value: 10}, "eleven": {Type: TokenInteger, Value: 11}, "twelve": {Type: TokenInteger, Value: 12},
	"thirteen": {Type: TokenInteger, Value: 13}, "fourteen": {Type: TokenInteger, Value: 14}, "fifteen": {Type: TokenInteger, Value: 15},
	"sixteen": {Type: TokenInteger, Value: 16}, "seventeen": {Type: TokenInteger, Value: 17}, "eighteen": {Type: TokenInteger, Value: 18},
	"nineteen": {Type: TokenInteger, Value: 19}, "twenty": {Type: TokenInteger, Value: 20}, "thirty": {Type: TokenInteger, Value: 30},

	// Ordinal words (day-of-month range: 1–31)
	"first": {Type: TokenInteger, Value: 1}, "third": {Type: TokenInteger, Value: 3}, "fourth": {Type: TokenInteger, Value: 4},
	"fifth": {Type: TokenInteger, Value: 5}, "sixth": {Type: TokenInteger, Value: 6}, "seventh": {Type: TokenInteger, Value: 7},
	"eighth": {Type: TokenInteger, Value: 8}, "ninth": {Type: TokenInteger, Value: 9}, "tenth": {Type: TokenInteger, Value: 10},
	"eleventh": {Type: TokenInteger, Value: 11}, "twelfth": {Type: TokenInteger, Value: 12}, "thirteenth": {Type: TokenInteger, Value: 13},
	"fourteenth": {Type: TokenInteger, Value: 14}, "fifteenth": {Type: TokenInteger, Value: 15}, "sixteenth": {Type: TokenInteger, Value: 16},
	"seventeenth": {Type: TokenInteger, Value: 17}, "eighteenth": {Type: TokenInteger, Value: 18}, "nineteenth": {Type: TokenInteger, Value: 19},
	"twentieth": {Type: TokenInteger, Value: 20}, "thirtieth": {Type: TokenInteger, Value: 30},

	// Multi-word ordinals — hyphenated variants (single tokens after splitting)
	"twenty-first": {Type: TokenInteger, Value: 21}, "twenty-second": {Type: TokenInteger, Value: 22},
	"twenty-third": {Type: TokenInteger, Value: 23}, "twenty-fourth": {Type: TokenInteger, Value: 24},
	"twenty-fifth": {Type: TokenInteger, Value: 25}, "twenty-sixth": {Type: TokenInteger, Value: 26},
	"twenty-seventh": {Type: TokenInteger, Value: 27}, "twenty-eighth": {Type: TokenInteger, Value: 28},
	"twenty-ninth": {Type: TokenInteger, Value: 29}, "thirty-first": {Type: TokenInteger, Value: 31},

	// Multi-word ordinals — space-separated phrases (matched longest-first by tokenizer)
	"twenty first": {Type: TokenInteger, Value: 21}, "twenty second": {Type: TokenInteger, Value: 22},
	"twenty third": {Type: TokenInteger, Value: 23}, "twenty fourth": {Type: TokenInteger, Value: 24},
	"twenty fifth": {Type: TokenInteger, Value: 25}, "twenty sixth": {Type: TokenInteger, Value: 26},
	"twenty seventh": {Type: TokenInteger, Value: 27}, "twenty eighth": {Type: TokenInteger, Value: 28},
	"twenty ninth": {Type: TokenInteger, Value: 29}, "thirty first": {Type: TokenInteger, Value: 31},
}
