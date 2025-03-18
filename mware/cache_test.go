package mware

import (
	"net/http"
	"testing"

	"zgo.at/zstd/ztest"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

func TestNoCache(t *testing.T) {
	rr := ztest.HTTP(t, nil, NoCache()(handle{}))
	if rr.Code != http.StatusOK {
		t.Errorf("want code 200, got %v", rr.Code)
	}
	if b := rr.Body.String(); b != "handler" {
		t.Errorf("body wrong: %#v", b)
	}
	if h := rr.Header().Get("Cache-Control"); h != "no-cache" {
		t.Errorf("header wrong: %#v", h)
	}
}

func TestNoStore(t *testing.T) {
	rr := ztest.HTTP(t, nil, NoStore()(handle{}))

	if rr.Code != http.StatusOK {
		t.Errorf("want code 200, got %v", rr.Code)
	}
	if b := rr.Body.String(); b != "handler" {
		t.Errorf("body wrong: %#v", b)
	}
	if h := rr.Header().Get("Cache-Control"); h != "no-store,no-cache" {
		t.Errorf("header wrong: %#v", h)
	}
}
