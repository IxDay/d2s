package app

import (
	"net/http"

	"github.com/platipy-io/d2s/internal/github"
	"github.com/platipy-io/d2s/server"
)

func Index(ctx *server.Context) error {
	span := ctx.NewSpan("index")
	defer span.End()
	defer ctx.LogWrapper("index endpoint")()
	if ctx.User == nil {
		return ctx.Render(BaseTplt(ctx, IndexTplt(nil, nil)))
	}
	repos, err := github.Starred(ctx.Context(), ctx.User)
	if err != nil {
		return err
	}
	return ctx.Render(BaseTplt(ctx, IndexTplt(repos, nil)))
}

type HTTPError struct {
	Code int
	Msg  string
	Err  error
}

func New500HTTPError(err error) HTTPError {
	return HTTPError{Code: http.StatusInternalServerError, Msg: "We encountered an issue", Err: err}
}

func New400HTTPError(err error) HTTPError {
	return HTTPError{Code: http.StatusBadRequest, Msg: "The request provided is invalid", Err: err}
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
	errHTTP := New500HTTPError(err)
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
