package lorem

import (
	"github.com/platipy-io/d2s/app"
	"github.com/platipy-io/d2s/server"
)

func Index(ctx *server.Context) error {
	defer ctx.LogWrapper("lorem endpoint")()
	component := IndexTplt()
	if _, ok := ctx.Request.Header["Hx-Request"]; !ok {
		component = app.BaseTplt(app.IndexTplt(component))
	}
	return ctx.Render(component)
}
