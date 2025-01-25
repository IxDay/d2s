package server

import (
	"net/http"
	"time"

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

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{Logger: *log.Ctx(r.Context()), Request: r, ResponseWriter: w}
}

func (c *Context) NewSpan(name string) trace.Span {
	ctx, span := telemetry.NewSpan(c.Context(), name)
	c.Request = c.Request.WithContext(ctx)
	return span
}

func (c *Context) LogWrapper(name string) func() {
	return log.FnWrapper(c.Context(), c.Logger, name)
}

func (c *Context) Redirect(url string, code int) {
	http.Redirect(c.ResponseWriter, c.Request, url, code)
}

func (c *Context) SetCookie(name, value string, duration time.Duration) {
	cookie := http.Cookie{Name: name, Value: value, Expires: time.Now().Add(duration)}
	http.SetCookie(c.ResponseWriter, &cookie)
}

func (c *Context) Render(component templ.Component) error {
	return component.Render(c.Context(), c.ResponseWriter)
}

type Handler interface {
	Handle(*Context) error
}

type HandlerFunc func(*Context) error

func (hf HandlerFunc) Handle(c *Context) error { return hf(c) }
