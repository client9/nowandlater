package nowandlater

import (
	"github.com/client9/nowandlater/internal/engine"
	"github.com/client9/nowandlater/languages"
	"strings"
)

type Lang = engine.Lang

// langRegistry maps ISO 639-1 codes to the built-in Lang pointers.
var langRegistry = map[string]*engine.Lang{
	"en": &languages.LangEn,
	"es": &languages.LangEs,
	"fr": &languages.LangFr,
	"de": &languages.LangDe,
	"it": &languages.LangIt,
	"pt": &languages.LangPt,
	"ru": &languages.LangRu,
	"ja": &languages.LangJa,
	"zh": &languages.LangZh,
}

// LookupLang returns the built-in Lang for the given ISO 639-1 code, or nil if
// the code is not recognised. The lookup is case-insensitive and region suffixes
// are ignored ("en_US" and "en-GB" both return &LangEn).
func LookupLang(code string) *engine.Lang {
	code = strings.ToLower(strings.TrimSpace(code))
	if i := strings.IndexAny(code, "-_"); i >= 0 {
		code = code[:i]
	}
	return langRegistry[code]
}
