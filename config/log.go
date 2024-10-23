package config

import (
	"strings"

	"github.com/rs/zerolog"
)

func (c Configuration) MarshalZerologObject(e *zerolog.Event) {
	e.Bool("dev", c.Dev)
	e.Str("host", c.Host)
	e.Int("port", c.Port)

	e.Object("logger", c.Logger)
	e.Object("tracer", c.Tracer)
}

func (l Logger) MarshalZerologObject(e *zerolog.Event) {
	e.Str("level", l.Level)
}

var sensibleHeader = map[string]struct{}{"set-cookie": {}, "authorization": {}}

func (t Tracer) MarshalZerologObject(e *zerolog.Event) {
	e.Bool("enabled", t.Enabled)
	if t.Endpoint != "" {
		e.Str("endpoint", t.Endpoint)
	}
	if len(t.Headers) != 0 {
		dict := zerolog.Dict()
		for k, v := range t.Headers {
			k = strings.ToLower(k)
			if _, ok := sensibleHeader[k]; ok {
				v = "*****"
			}
			dict.Str(k, v)
		}
		e.Dict("headers", dict)
	}
}
