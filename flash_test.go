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
	if out != "w00t" {
		t.Errorf("out: %#v", out)
	}
}
