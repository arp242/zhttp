package zhttp

import (
	"net/http/httptest"
	"testing"
)

func TestFlash(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	Flash(rr, "w00t")

	out := ReadFlash(rr, r)
	if out == nil {
		t.Fatal("out is nil")
	}
	if out.Message != "w00t" {
		t.Errorf("wrong message: %#v", out)
	}
	if out.Level != "i" {
		t.Errorf("wrong level: %#v", out)
	}
}
