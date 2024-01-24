package app

import (
	"github.com/platipy-io/d2s/server"
)

func Index(ctx *server.Context) error {
	span := ctx.NewSpan("index")
	defer span.End()
	defer ctx.LogWrapper("index endpoint")()
	return ctx.Render(BaseTplt(IndexTplt(nil)))
}

func ErrorHandler(ctx *server.Context, err error) {
	ctx.WriteHeader(500)
	component := ErrorTplt(500, err.Error())
	if _, ok := ctx.Request.Header["Hx-Request"]; !ok {
		component = BaseTplt(component)
	}
	ctx.Render(component)
}
