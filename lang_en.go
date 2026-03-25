package nowandlater

// English is the built-in English Lang.
var English = Lang{
	Words:           englishWords,
	OrdinalSuffixes: []string{"st", "nd", "rd", "th"},
}

// englishWords is the word table for English.
// It covers single words, multi-word phrases (space-containing keys),
// time-word substitutions, and number words — all in one map.
var englishWords = map[string]WordEntry{
	// Weekdays
	"monday":    {TokenWeekday, WeekdayMonday},
	"mon":       {TokenWeekday, WeekdayMonday},
	"tuesday":   {TokenWeekday, WeekdayTuesday},
	"tu":        {TokenWeekday, WeekdayTuesday},
	"tue":       {TokenWeekday, WeekdayTuesday},
	"tues":      {TokenWeekday, WeekdayTuesday},
	"wednesday": {TokenWeekday, WeekdayWednesday},
	"wed":       {TokenWeekday, WeekdayWednesday},
	"wednes":    {TokenWeekday, WeekdayWednesday},
	"thursday":  {TokenWeekday, WeekdayThursday},
	"thu":       {TokenWeekday, WeekdayThursday},
	"thur":      {TokenWeekday, WeekdayThursday},
	"thurs":     {TokenWeekday, WeekdayThursday},
	"friday":    {TokenWeekday, WeekdayFriday},
	"fri":       {TokenWeekday, WeekdayFriday},
	"saturday":  {TokenWeekday, WeekdaySaturday},
	"sat":       {TokenWeekday, WeekdaySaturday},
	"sunday":    {TokenWeekday, WeekdaySunday},
	"su":        {TokenWeekday, WeekdaySunday},
	"sun":       {TokenWeekday, WeekdaySunday},

	// Months
	"january":   {TokenMonth, MonthJanuary},
	"jan":       {TokenMonth, MonthJanuary},
	"february":  {TokenMonth, MonthFebruary},
	"feb":       {TokenMonth, MonthFebruary},
	"march":     {TokenMonth, MonthMarch},
	"mar":       {TokenMonth, MonthMarch},
	"april":     {TokenMonth, MonthApril},
	"apr":       {TokenMonth, MonthApril},
	"may":       {TokenMonth, MonthMay},
	"june":      {TokenMonth, MonthJune},
	"jun":       {TokenMonth, MonthJune},
	"july":      {TokenMonth, MonthJuly},
	"jul":       {TokenMonth, MonthJuly},
	"august":    {TokenMonth, MonthAugust},
	"aug":       {TokenMonth, MonthAugust},
	"september": {TokenMonth, MonthSeptember},
	"sep":       {TokenMonth, MonthSeptember},
	"sept":      {TokenMonth, MonthSeptember},
	"october":   {TokenMonth, MonthOctober},
	"oct":       {TokenMonth, MonthOctober},
	"november":  {TokenMonth, MonthNovember},
	"nov":       {TokenMonth, MonthNovember},
	"december":  {TokenMonth, MonthDecember},
	"dec":       {TokenMonth, MonthDecember},

	// Direction
	"next":     {TokenDirection, DirectionFuture},
	"last":     {TokenDirection, DirectionPast},
	"previous": {TokenDirection, DirectionPast},
	"prev":     {TokenDirection, DirectionPast},
	"coming":   {TokenDirection, DirectionFuture},
	"this":     {TokenDirection, DirectionNearest},

	// Modifier
	"ago":    {TokenModifier, ModifierPast},
	"hence":  {TokenModifier, ModifierFuture},
	"from":   {TokenModifier, ModifierFuture},
	"before": {TokenModifier, ModifierPast},
	"after":  {TokenModifier, ModifierFuture},

	// Anchor
	"now":       {TokenAnchor, AnchorNow},
	"today":     {TokenAnchor, AnchorToday},
	"tomorrow":  {TokenAnchor, AnchorTomorrow},
	"yesterday": {TokenAnchor, AnchorYesterday},

	// Preposition (value not consumed semantically)
	"at": {TokenPrep, nil},
	"on": {TokenPrep, nil},
	"in": {TokenPrep, nil},
	"by": {TokenPrep, nil},

	// AM/PM (already normalized from a.m./p.m. by the normalize step)
	"am": {TokenAMPM, AMPMAm},
	"pm": {TokenAMPM, AMPMPm},

	// Units (singular and plural — all variants of the same period carry the same Period constant)
	"second":     {TokenUnit, PeriodSecond},
	"seconds":    {TokenUnit, PeriodSecond},
	"sec":        {TokenUnit, PeriodSecond},
	"secs":       {TokenUnit, PeriodSecond},
	"minute":     {TokenUnit, PeriodMinute},
	"minutes":    {TokenUnit, PeriodMinute},
	"min":        {TokenUnit, PeriodMinute},
	"mins":       {TokenUnit, PeriodMinute},
	"hour":       {TokenUnit, PeriodHour},
	"hours":      {TokenUnit, PeriodHour},
	"hr":         {TokenUnit, PeriodHour},
	"hrs":        {TokenUnit, PeriodHour},
	"day":        {TokenUnit, PeriodDay},
	"days":       {TokenUnit, PeriodDay},
	"fortnight":  {TokenUnit, PeriodFortnight},
	"fortnights": {TokenUnit, PeriodFortnight},
	"week":       {TokenUnit, PeriodWeek},
	"weeks":      {TokenUnit, PeriodWeek},
	"wk":         {TokenUnit, PeriodWeek},
	"month":      {TokenUnit, PeriodMonth},
	"months":     {TokenUnit, PeriodMonth},
	"mo":         {TokenUnit, PeriodMonth},
	"year":       {TokenUnit, PeriodYear},
	"years":      {TokenUnit, PeriodYear},
	"yr":         {TokenUnit, PeriodYear},
	"yrs":        {TokenUnit, PeriodYear},

	// Filler (grammatical noise — value not consumed semantically)
	"the":   {TokenFiller, nil},
	"of":    {TokenFiller, nil},
	"a":     {TokenFiller, nil},
	"an":    {TokenFiller, nil},
	"and":   {TokenFiller, nil},
	"about": {TokenFiller, nil},
	"just":  {TokenFiller, nil},

	// Multi-word anchors (matched longest-first by tokenizer)
	"day before yesterday": {TokenAnchor, Anchor2DaysAgo},
	"day after tomorrow":   {TokenAnchor, Anchor2DaysFromNow},

	// Time-word substitutions — produce TokenTime directly
	// ("noon" and "midnight" as words tokenize to TIME rather than requiring preprocessing)
	"noon":     {TokenTime, "12:00"},
	"midnight": {TokenTime, "0:00"},

	// Number words — Cardinals (1–30)
	// "second" is intentionally absent: it conflicts with TokenUnit PeriodSecond.
	// Use "2nd" (ordinal suffix stripped by OrdinalSuffixes) for the 2nd day of month.
	"one": {TokenInteger, 1}, "two": {TokenInteger, 2}, "three": {TokenInteger, 3},
	"four": {TokenInteger, 4}, "five": {TokenInteger, 5}, "six": {TokenInteger, 6},
	"seven": {TokenInteger, 7}, "eight": {TokenInteger, 8}, "nine": {TokenInteger, 9},
	"ten": {TokenInteger, 10}, "eleven": {TokenInteger, 11}, "twelve": {TokenInteger, 12},
	"thirteen": {TokenInteger, 13}, "fourteen": {TokenInteger, 14}, "fifteen": {TokenInteger, 15},
	"sixteen": {TokenInteger, 16}, "seventeen": {TokenInteger, 17}, "eighteen": {TokenInteger, 18},
	"nineteen": {TokenInteger, 19}, "twenty": {TokenInteger, 20}, "thirty": {TokenInteger, 30},

	// Ordinal words (day-of-month range: 1–31)
	"first": {TokenInteger, 1}, "third": {TokenInteger, 3}, "fourth": {TokenInteger, 4},
	"fifth": {TokenInteger, 5}, "sixth": {TokenInteger, 6}, "seventh": {TokenInteger, 7},
	"eighth": {TokenInteger, 8}, "ninth": {TokenInteger, 9}, "tenth": {TokenInteger, 10},
	"eleventh": {TokenInteger, 11}, "twelfth": {TokenInteger, 12}, "thirteenth": {TokenInteger, 13},
	"fourteenth": {TokenInteger, 14}, "fifteenth": {TokenInteger, 15}, "sixteenth": {TokenInteger, 16},
	"seventeenth": {TokenInteger, 17}, "eighteenth": {TokenInteger, 18}, "nineteenth": {TokenInteger, 19},
	"twentieth": {TokenInteger, 20}, "thirtieth": {TokenInteger, 30},

	// Multi-word ordinals — hyphenated variants (single tokens after splitting)
	"twenty-first": {TokenInteger, 21}, "twenty-second": {TokenInteger, 22},
	"twenty-third": {TokenInteger, 23}, "twenty-fourth": {TokenInteger, 24},
	"twenty-fifth": {TokenInteger, 25}, "twenty-sixth": {TokenInteger, 26},
	"twenty-seventh": {TokenInteger, 27}, "twenty-eighth": {TokenInteger, 28},
	"twenty-ninth": {TokenInteger, 29}, "thirty-first": {TokenInteger, 31},

	// Multi-word ordinals — space-separated phrases (matched longest-first by tokenizer)
	"twenty first": {TokenInteger, 21}, "twenty second": {TokenInteger, 22},
	"twenty third": {TokenInteger, 23}, "twenty fourth": {TokenInteger, 24},
	"twenty fifth": {TokenInteger, 25}, "twenty sixth": {TokenInteger, 26},
	"twenty seventh": {TokenInteger, 27}, "twenty eighth": {TokenInteger, 28},
	"twenty ninth": {TokenInteger, 29}, "thirty first": {TokenInteger, 31},
}
