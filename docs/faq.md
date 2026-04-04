# FAQ

### How does it work?

See `docs/architecture.md` for the full pipeline diagram. In brief:

* Turns input into a list of tokens, e.g. "Monday" --> `{ WEEKDAY, Monday }`. This is 99% of variations between different human languages.
* It's data driven -- the tokens are in a map (e.g. `languages/lang_en.go`)
* The list of tokens is a "signature" - across all languages there are under 100 signatures (see `internal/engine/dispatch.go`)
* The signature is used to find a handler function using a `map`.
* The handler converts the signature into a golang struct holding numeric values for day, month, hour, etc. The handler `func` is normally a few lines long.
* The struct is converted into a `time.Time` object.

### Project layout

```
nowandlater/
├── parser.go          # Public API — Parser, Parse, ParseInterval
├── internal/engine/   # Core pipeline: tokenizer, handlers, dispatch, resolve
├── languages/         # One file per human language (lang_en.go, lang_es.go, …)
├── tests/             # All tests
├── cmd/nldate/        # CLI dev tool
└── docs/              # Architecture and contributor docs
```

### Can I add another human language?

Yes. Look at `languages/lang_en.go` (English) or `languages/lang_es.go` (Spanish). Translate that, run tests, perhaps add a handler. Done. See `docs/contributing.md` for a step-by-step guide.

Your favorite coding AI assistant can do most of it in 5 minutes.

### Is it fast?

It can parse a date and time snippet in about 500ns.

Python's dateparser, takes about 1,000,000ns. That's 2000x faster.

Go's native time parser on a fixed format is 20ns. That's 25x slower.


### Can it be ported to another (computer) language?

Please do so!  It should be easy.  Let me know if you need help.

### Can't AI LLMs do this?

In 2026, LLMs are good at extracting and converting an arbitrary text into date snippets:

```
Tomorrow, let's get ice cream at 3pm"
```

into

```
tomorrow at 3pm
```

but they are not very good at converting that to a standard time format or converting to a native datetime object.  There are a lot of off by one errors, time zone issues, and confusion.

Also it's very slow (relative to converting locally).

### What about other Go libraries?

The existing projects in Go are mostly abandoned or obsolete.  As of 2026:

* [go-anytime](https://github.com/ijt/go-anytime) - last commit in 2023, English only
* [naturaltime](https://github.com/Sho0pi/naturaltime) - last commit in 2025, runs a javascript interpreter (!!) to execute[chrononode](https://github.com/wanasit/chrono/tree/master) which is pile of regexp.
* [when](https://github.com/olebedev/when) - last commit in 2025, complicated rule set using regexp.
* [go-naturaldate](https://github.com/tj/go-naturaldate) - last commit in 2020, parser/generator PEG based.  English only
* [naturaldate](https://github.com/anatol/naturaldate.go) - a fork of previous
* [go-dateparser](https://github.com/markusmobius/go-dateparser) - direct port of python's dateparser, last commit 1 year, has not kept up to date
* [araddon/dateparse](https://github.com/araddon/dateparse) - last commit in 2021, more for fixed computer formats
* [shadiestgoat/dateparse](https://github.com/shadiestgoat/dateparse) - last commit in 2024, fixed formats

All (?) use regexp soup or some parser/generator.  Go needs something better.

### What about porting code from other programming languages or libraries?

There are good libraries in other programming languages, some with unique features:

* Python [dateparser](https://github.com/scrapinghub/dateparser)  - actively maintained, regexp based.
* Typescript [chrononode](https://github.com/wanasit/chrono/tree/master) - complicated, but has a lot of interesting features.
* Ruby [chronic](https://github.com/mojombo/chronic) - English only, last commit in 2023.
* Java [natty](https://github.com/joestelmach/natty) - Last commit in 2017 - English only - parser/generator
* GNU date has some interesting date parsing abilities

But all are either English-only or regexp/parser/generator based. Let me know if I'm missing one.

None are really good to port to Go.  And it's hard to keep track up subsequent changes.

### What's with the name?

[Now and Later](https://en.wikipedia.org/wiki/Now_and_Later) is a classic American candy.


