package app

import (
	"net/http"

	"github.com/platipy-io/d2s/server"
)

func Index(ctx *server.Context) error {
	span := ctx.NewSpan("index")
	defer span.End()
	defer ctx.LogWrapper("index endpoint")()
	return ctx.Render(BaseTplt(ctx, IndexTplt(ctx, nil)))
}

type HTTPError struct {
	Code int
	Msg  string
	Err  error
}

func (he HTTPError) Error() string { return he.Err.Error() }

func (he HTTPError) Render(ctx *server.Context) {
	ctx.WriteHeader(he.Code)
	component := ErrorTplt(he)
	if _, ok := ctx.Request.Header["Hx-Request"]; !ok {
		component = BaseTplt(ctx, component)
	}
	if err := ctx.Render(component); err != nil {
		ctx.Logger.Error().Ctx(ctx.Context()).Stack().Err(err).Msg("failed rendering template")
	}
}

func ErrorHandler(ctx *server.Context, err error) {
	errHTTP := HTTPError{Code: 500, Msg: "We encountered an issue", Err: err}
	if e, ok := err.(HTTPError); ok {
		errHTTP = e
	}
	ctx.Logger.Error().Ctx(ctx.Context()).Stack().Err(errHTTP.Err).Msg("handling error")
	errHTTP.Render(ctx)

}

func NotFoundHandler(ctx *server.Context) {
	ctx.Logger.Warn().Msg("path not found")
	HTTPError{Code: http.StatusNotFound, Msg: "The page you are looking for does not exist"}.Render(ctx)
}
