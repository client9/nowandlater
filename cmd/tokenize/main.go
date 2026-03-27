// tokenize is a CLI for ad hoc testing of the nowandlater tokenizer and resolver.
//
// Usage:
//
//	tokenize [-now TIME] [-interval] "next Monday at 9:30 AM"   # single input from args
//	echo "in 2 days" | tokenize [-now TIME] [-interval]         # read from stdin
//	tokenize [-now TIME] [-interval]                            # interactive: one line per prompt
//
// Flags:
//
//	-now TIME      reference time in RFC3339 (2026-03-22T10:00:00Z) or date-only
//	               (2026-03-22, midnight local) format; defaults to time.Now()
//	-interval      also show the resolved calendar interval [start, end)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/client9/nowandlater"
)

func lookupLang(code string) (*nowandlater.Lang, error) {
	if l := nowandlater.LookupLang(code); l != nil {
		return l, nil
	}
	return nil, fmt.Errorf("unknown language %q; supported: en, es, fr, de, it, pt, ru, ja, zh", code)
}

func main() {
	nowFlag := flag.String("now", "", "reference time (RFC3339 or YYYY-MM-DD); default: time.Now()")
	intervalFlag := flag.Bool("interval", false, "show resolved interval [start, end)")
	langFlag := flag.String("lang", "en", "language code or locale (e.g. en, es, fr, de, it, pt, ru, ja, zh, en_US, zh-CN)")
	flag.Parse()

	lang, err := lookupLang(*langFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tokenize: -lang: %v\n", err)
		os.Exit(1)
	}

	now, err := parseNowFlag(*nowFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tokenize: -now: %v\n", err)
		os.Exit(1)
	}

	// Args mode: join non-flag arguments as one input string.
	if flag.NArg() > 0 {
		input := strings.Join(flag.Args(), " ")
		printTokens(input, now, *intervalFlag, lang)
		return
	}

	// Stdin mode: process one line at a time.
	scanner := bufio.NewScanner(os.Stdin)
	interactive := isTerminal()
	for {
		if interactive {
			fmt.Print("> ")
		}
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		printTokens(line, now, *intervalFlag, lang)
		if interactive {
			fmt.Println()
		}
	}
}

func printTokens(input string, now time.Time, showInterval bool, lang *nowandlater.Lang) {
	tokens := lang.Tokenize(input)
	sig := nowandlater.Signature(tokens)

	fmt.Printf("input:     %q\n", input)
	fmt.Printf("signature: %q\n", sig)
	fmt.Printf("tokens:\n")
	for i, tok := range tokens {
		fmt.Printf("  [%d] %-14s %s\n", i, tok.Type, formatValue(tok.Value))
	}

	slots, err := lang.Parse(input)
	if err != nil {
		fmt.Printf("parse:     error: %v\n", err)
		return
	}
	fmt.Printf("period:    %s\n", slots.Period)

	result, err := nowandlater.Resolve(slots, now)
	if err != nil {
		fmt.Printf("resolve:   error: %v\n", err)
		return
	}
	fmt.Printf("now:       %s\n", now.Format(time.RFC3339))
	fmt.Printf("resolved:  %s\n", result.Format(time.RFC3339))

	if showInterval {
		start, end, err := nowandlater.ResolveInterval(slots, now)
		if err != nil {
			fmt.Printf("interval:  error: %v\n", err)
		} else {
			fmt.Printf("start:     %s\n", start.Format(time.RFC3339))
			fmt.Printf("end:       %s\n", end.Format(time.RFC3339))
		}
	}
}

// parseNowFlag parses the -now flag value. Accepts RFC3339 or date-only (YYYY-MM-DD).
// Returns time.Now() if s is empty.
func parseNowFlag(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.In(time.Local), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q; use RFC3339 (2026-03-22T10:00:00Z) or date (2026-03-22)", s)
}

// formatValue formats a Token.Value for display:
//   - nil (Prep/Filler)  → "-"
//   - string (Time/TZ/…) → quoted "value"
//   - int (Integer/Year) → bare number
//   - typed constant     → String() representation
func formatValue(v any) string {
	if v == nil {
		return "-"
	}
	switch v := v.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// isTerminal reports whether stdin is an interactive terminal.
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
