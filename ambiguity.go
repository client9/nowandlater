package nowandlater

import "github.com/client9/nowandlater/internal/engine"

type ambiguityPreset uint8

const (
	ambiguityScheduling ambiguityPreset = iota + 1
	ambiguityHistorical
	ambiguityStrict
)

// AmbiguityConfig controls how the parser resolves underspecified inputs.
//
// Use the exported preset values rather than constructing this type directly.
// Its internal representation may expand in future versions.
type AmbiguityConfig struct {
	preset ambiguityPreset
}

// Named ambiguity presets for the Parser.
var (
	AmbiguityScheduling = AmbiguityConfig{preset: ambiguityScheduling}
	AmbiguityHistorical = AmbiguityConfig{preset: ambiguityHistorical}
	AmbiguityStrict     = AmbiguityConfig{preset: ambiguityStrict}
)

func (c AmbiguityConfig) resolvePolicy() engine.ResolvePolicy {
	switch c.preset {
	case ambiguityHistorical:
		return engine.ResolvePolicy{
			ImplicitDurationDirection: engine.DirectionPast,
			CalendarDirection:         engine.DirectionPast,
			MonthDayDirection:         engine.DirectionPast,
			RejectAmbiguous:           false,
		}
	case ambiguityStrict:
		return engine.ResolvePolicy{
			ImplicitDurationDirection: engine.DirectionFuture,
			CalendarDirection:         engine.DirectionFuture,
			MonthDayDirection:         engine.DirectionFuture,
			RejectAmbiguous:           true,
		}
	default:
		return engine.ResolvePolicy{
			ImplicitDurationDirection: engine.DirectionFuture,
			CalendarDirection:         engine.DirectionFuture,
			MonthDayDirection:         engine.DirectionFuture,
			RejectAmbiguous:           false,
		}
	}
}
