package zhttp

import (
	"net/http"
	"testing"
)

func TestServe(t *testing.T) {
	ch := Serve(0, 1, &http.Server{
		Addr: "127.0.0.1:0",
	})
	<-ch // Ready to serve
	<-ch // Stopped serving
}
