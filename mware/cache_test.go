package mware

import (
	"net/http"
	"testing"

	"zgo.at/zhttp"
	"zgo.at/zstd/ztest"
)

func TestNoCache(t *testing.T) {
	rr := ztest.HTTP(t, nil, zhttp.Wrap(NoCache()(handle{}.ServeHTTP)))
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
	rr := ztest.HTTP(t, nil, zhttp.Wrap(NoStore()(handle{}.ServeHTTP)))

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
