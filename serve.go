package zhttp

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"zgo.at/zlog"
)

// Serve a HTTP server with graceful shutdown.
func Serve(server *http.Server, wait func()) {
	consClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		shutdown := make(chan struct{})
		go func() {
			err := server.Shutdown(context.Background())
			if err != nil {
				zlog.Errorf("server.Shutdown: %s", err)
			}
			shutdown <- struct{}{}
		}()

		<-shutdown
		close(consClosed)
	}()

	// Start HTTP servers.
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			zlog.Errorf("http.ListenAndServe: %s", err)
		}
	}()

	<-consClosed
	if wait != nil {
		wait()
	}
}
