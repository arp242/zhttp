package zhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRobots(t *testing.T) {
	// rr.Get("/robots.txt", zhttp.HandlerRobots([][]string{{"User-agent: *", "Disallow: /"}}))
	h := HandlerRobots([][]string{
		{"User-agent: Terminator", "Disallow: /sarah-connor"},
		{"User-agent: *", "Disallow: /foo", "Disallow: /bar"},
	})

	r, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	h(rr, r)

	want := `User-agent: Terminator
Disallow: /sarah-connor
User-agent: *
Disallow: /foo
`

	got := rr.Body.String()
	if got != want {
		t.Errorf("\ngot:  %q\nwant: %q", got, want)
	}

}
