package app

import (
	"net/http"

	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
)

func Index(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.NewSpan(r.Context(), "index")
	defer span.End()
	defer log.FnWrapper(ctx, "index endpoint")()
	component := BaseTplt(IndexTplt(nil))
	component.Render(ctx, w)
}

func Error(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	component := ErrorTplt(500, "Something bad happened")
	if _, ok := r.Header["Hx-Request"]; !ok {
		component = BaseTplt(component)
	}
	component.Render(r.Context(), w)
}
