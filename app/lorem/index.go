package lorem

import (
	"net/http"

	"github.com/platipy-io/d2s/app"
	"github.com/platipy-io/d2s/internal/log"
)

func index(w http.ResponseWriter, r *http.Request) {
	defer log.FnWrapper(r.Context(), "lorem endpoint")()
	component := IndexTplt()
	if _, ok := r.Header["Hx-Request"]; !ok {
		component = app.BaseTplt(app.IndexTplt(component))
	}
	component.Render(r.Context(), w)
}

var Index = http.HandlerFunc(index)
