package zhttp

import (
	"net/http"
	"testing"
)

func TestServe(t *testing.T) {
	stop := make(chan struct{})
	s := http.Server{Addr: "127.0.0.1:0"}
	ch, err := Serve(0, stop, &s)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	stop <- struct{}{}
	<-ch
}
