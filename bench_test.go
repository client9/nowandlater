package nowandlater

import (
	"testing"
	"time"
)

// fixedBenchNow is a pre-built time value used as the reference clock in all
// Parser benchmarks, avoiding time.Now() overhead in the measurement.
var fixedBenchNow = time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

var benchParser = Parser{
	Lang: &English,
	Now:  func() time.Time { return fixedBenchNow },
}

// BenchmarkStdlibRFC3339 is the baseline: Go's own time.Parse on a well-formed
// RFC3339 string. Use this as the lower-bound reference.
func BenchmarkStdlibRFC3339(b *testing.B) {
	b.ReportAllocs()
	const input = "2026-03-22T09:30:00Z"
	b.ResetTimer()
	for b.Loop() {
		_, _ = time.Parse(time.RFC3339, input)
	}
}

// BenchmarkParserRFC3339 measures our full pipeline (preprocess → tokenize →
// dispatch → resolve) on the same RFC3339 string that BenchmarkStdlibRFC3339
// uses. This shows the overhead of natural-language support on a structured input.
func BenchmarkParserRFC3339(b *testing.B) {
	b.ReportAllocs()
	const input = "2026-03-22T09:30:00Z"
	b.ResetTimer()
	for b.Loop() {
		_, _ = benchParser.Parse(input)
	}
}

// BenchmarkParserNaturalAbsolute measures parsing of a natural-language absolute
// date: a form that time.Parse cannot handle at all.
func BenchmarkParserNaturalAbsolute(b *testing.B) {
	b.ReportAllocs()
	const input = "March 22, 2026 at 9:30am"
	b.ResetTimer()
	for b.Loop() {
		_, _ = benchParser.Parse(input)
	}
}

// BenchmarkParserRelative measures parsing of a purely relative expression,
// which requires resolution against the reference clock.
func BenchmarkParserRelative(b *testing.B) {
	b.ReportAllocs()
	const input = "next Monday at 9:30"
	b.ResetTimer()
	for b.Loop() {
		_, _ = benchParser.Parse(input)
	}
}

// BenchmarkParserSuite runs a representative cross-section of input types so a
// single run gives a broad picture of average throughput.
func BenchmarkParserSuite(b *testing.B) {
	cases := []string{
		"2026-03-22T09:30:00Z",     // RFC3339 (structured)
		"March 22, 2026",           // natural absolute date
		"March 22, 2026 at 9:30am", // natural absolute date+time
		"next Monday",              // relative weekday
		"next Monday at 9:30",      // relative weekday + time
		"in 3 days",                // relative delta
		"tomorrow at noon",         // anchor + time-of-day
		"next week",                // direction + unit
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, input := range cases {
			_, _ = benchParser.Parse(input)
		}
	}
}
