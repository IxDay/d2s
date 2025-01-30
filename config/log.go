package config

import (
	"strings"

	"github.com/rs/zerolog"
)

func (c Configuration) MarshalZerologObject(e *zerolog.Event) {
	e.Bool("dev", bool(c.Dev))
	e.Str("host", c.Host)
	e.Int("port", c.Port)

	e.Object("logger", c.Logger)
	e.Object("tracer", c.Tracer)
	e.Object("authentication", c.Authentication)
}

func (l Logger) MarshalZerologObject(e *zerolog.Event) {
	e.Str("level", l.LogLevel.String())
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

func (a Authentication) MarshalZerologObject(e *zerolog.Event) {
	e.Str("redirect", a.Redirect)
	e.Str("client-id", a.ClientID)
	if a.ClientSecret != "" {
		e.Str("client-secret", "*****")
	} else {
		e.Str("client-secret", "<unset>")
	}
	if a.BypassToken != "" {
		e.Str("bypass-token", "*****")
	} else {
		e.Str("bypass-token", "<unset>")
	}
}
