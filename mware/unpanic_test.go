package mware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"zgo.at/zhttp"
	"zgo.at/zstd/ztest"
)

func TestUnpanic(t *testing.T) {
	var capture error
	zhttp.ErrPage = func(w http.ResponseWriter, r *http.Request, reported error) {
		capture = reported
		zhttp.DefaultErrPage(w, r, reported)
	}
	defer func() { zhttp.ErrPage = zhttp.DefaultErrPage }()

	handler := Unpanic()(zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		panic("oh noes!")
		return nil
	}))

	r, _ := http.NewRequest("GET", "/panic", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, r)
	if rr.Code != 500 {
		t.Fatalf("wrong response code: %d; wanted 200", rr.Code)
	}

	if !ztest.ErrorContains(capture, "oh noes!") {
		t.Errorf("wrong error\n%s", capture)
	}
}
