package propagate

import (
	"net/http"

	"github.com/platipy-io/d2s/internal/telemetry"
	"github.com/platipy-io/d2s/server"
)

func Index(ctx *server.Context) error {
	defer ctx.LogWrapper("propagate endpoint")()
	url := "http://localhost:8081"
	req, _ := http.NewRequestWithContext(ctx.Context(), http.MethodGet, url, nil)
	_, err := telemetry.HTTPClient.Do(req)
	if err != nil {
		ctx.Err(err).Msg("failed to send http request")
	}
	return nil
}
