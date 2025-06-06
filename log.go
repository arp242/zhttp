package zhttp

import (
	"log/slog"
	"net/http"
)

func withRequest(r *http.Request) *slog.Logger {
	if r == nil {
		panic("fieldsRequest: *http.Request is nil")
	}

	return slog.With(slog.Group("http",
		"method", r.Method,
		"url", r.URL.String(),
		"form", r.Form.Encode(),
		"host", r.Host,
		"user_agent", r.UserAgent()))
}
