package propagate

import (
	"net/http"

	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
)

func index(w http.ResponseWriter, r *http.Request) {
	defer log.FnWrapper(r.Context(), "propagate endpoint")()
	logger := log.Ctx(r.Context())
	url := "http://localhost:8081"
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	_, err := telemetry.HTTPClient.Do(req)
	if err != nil {
		logger.Err(err).Msg("failed to send http request")
	}

}

var Index = http.HandlerFunc(index)
