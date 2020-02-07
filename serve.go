package zhttp

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"zgo.at/zlog"
)

// Serve a HTTP server with graceful shutdown.
//
// If tls is given, port 80 will be redirected.
func Serve(server *http.Server, tls string, wait func()) {
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

	var host, port string

	// Redirect port 80 to TLS.
	if tls != "" {
		var err error
		host, port, err = net.SplitHostPort(server.Addr)
		if err != nil {
			host = server.Addr
			port = "443"
		}

		go func() {
			err = http.ListenAndServe(host+":80", HandlerRedirectHTTP(port))
			if err != nil && err != http.ErrServerClosed {
				zlog.Errorf("zhttp.Serve: ListenAndServe redirect: %s", err)
			}
		}()
	}

	// Start HTTP server.
	go func() {
		var err error
		if tls == "" {
			err = server.ListenAndServe()
		} else {
			certAndKey := strings.SplitN(tls, ":", 2)
			err = server.ListenAndServeTLS(certAndKey[0], certAndKey[1])
		}
		if err != nil && err != http.ErrServerClosed {
			zlog.Errorf("zhttp.Serve: ListenAndServe: %s", err)
		}
	}()

	<-consClosed
	if wait != nil {
		wait()
	}
}
