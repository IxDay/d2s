package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
	"go.opentelemetry.io/otel/trace"
)

type Context struct {
	log.Logger
	*http.Request
	http.ResponseWriter
}

func (c *Context) NewSpan(name string) trace.Span {
	ctx, span := telemetry.NewSpan(c.Context(), name)
	c.Request = c.Request.WithContext(ctx)
	return span
}

func (c *Context) LogWrapper(name string) func() {
	return log.FnWrapper(c.Context(), c.Logger, name)
}

func (c *Context) Render(component templ.Component) error {
	return component.Render(c.Context(), c.ResponseWriter)
}

type Handler interface {
	Handle(*Context) error
}

type HandlerFunc func(*Context) error

func (hf HandlerFunc) Handle(c *Context) error { return hf(c) }
