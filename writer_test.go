package zhttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	_ http.Flusher  = &flushWriter{}
	_ http.Flusher  = &http2FancyWriter{}
	_ http.Flusher  = &httpFancyWriter{}
	_ http.Hijacker = &httpFancyWriter{}
	_ http.Pusher   = &http2FancyWriter{}
	_ io.ReaderFrom = &httpFancyWriter{}
)

func TestFlushWriterRemembersWroteHeaderWhenFlushed(t *testing.T) {
	f := &flushWriter{basicWriter{ResponseWriter: httptest.NewRecorder()}}
	f.Flush()

	if !f.wroteHeader {
		t.Fatal("want Flush to have set wroteHeader=true")
	}
}

func TestHttpFancyWriterRemembersWroteHeaderWhenFlushed(t *testing.T) {
	f := &httpFancyWriter{basicWriter{ResponseWriter: httptest.NewRecorder()}}
	f.Flush()

	if !f.wroteHeader {
		t.Fatal("want Flush to have set wroteHeader=true")
	}
}

func TestHttp2FancyWriterRemembersWroteHeaderWhenFlushed(t *testing.T) {
	f := &http2FancyWriter{basicWriter{ResponseWriter: httptest.NewRecorder()}}
	f.Flush()

	if !f.wroteHeader {
		t.Fatal("want Flush to have set wroteHeader=true")
	}
}
