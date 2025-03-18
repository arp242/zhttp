package zhttp

import (
	"log/slog"
	"net/http"
)

func withRequest(r *http.Request) *slog.Logger {
	if r == nil {
		panic("fieldsRequest: *http.Request is nil")
	}

	return slog.With(
		"http_method", r.Method,
		"http_url", r.URL.String(),
		"http_form", r.Form.Encode(),
		"http_host", r.Host,
		"http_user_agent", r.UserAgent())
}
