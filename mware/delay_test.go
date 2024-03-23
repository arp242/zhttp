package mware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"zgo.at/zhttp"
)

func TestDelay(t *testing.T) {
	tests := []struct {
		delay  time.Duration
		cookie string
	}{
		{0, ""},
		{100 * time.Millisecond, ""},

		{0, "1"},
		{100 * time.Millisecond, "0"},
		{100 * time.Millisecond, "1"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			handler := Delay(tt.delay)(zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				w.Write([]byte("Hello World"))
				return nil
			}))

			r, _ := http.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()
			if tt.cookie != "" {
				r.AddCookie(&http.Cookie{
					Name:  "debug-delay",
					Value: tt.cookie,
				})
			}

			handler.ServeHTTP(rr, r)
			if rr.Code != 200 {
				t.Fatalf("wrong response code: %d; wanted 200", rr.Code)
			}

			// TODO: test correct delay was used.
		})
	}
}
