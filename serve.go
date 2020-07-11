package zhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"zgo.at/zlog"
)

const (
	ServeTLS      = uint8(0b0001)
	ServeRedirect = uint8(0b0010)
)

// Serve a HTTP server with graceful shutdown.
//
// The server will use TLS if the http.Server has a valid TLSConfig; it will
// also try to redirect port 80 to the TLS server, but will gracefully fail if
// the permission for this is denied.
func Serve(flags uint8, server *http.Server) {
	// Go uses :80 to listen on all addresses, this makes sure that the
	// common-ish "*:80" is also accepted.
	if strings.HasPrefix(server.Addr, "*:") {
		server.Addr = server.Addr[1:]
	}

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

	go func() {
		var err error
		if flags&ServeTLS != 0 {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			zlog.Errorf("zhttp.Serve: %s", err)
			if errors.Is(err, os.ErrPermission) {
				os.Exit(1)
			}
		}
	}()

	if flags&ServeRedirect != 0 {
		go func() {
			host, port, err := net.SplitHostPort(server.Addr)
			if err != nil {
				host = server.Addr
				port = "443"
			}

			go func() {
				err := http.ListenAndServe(host+":80", HandlerRedirectHTTP(port))
				if err != nil && err != http.ErrServerClosed {
					if errors.Is(err, os.ErrPermission) {
						abs, _ := filepath.Abs(os.Args[0])
						fmt.Fprintf(os.Stderr, "WARNING: No permission to bind to port 80, not setting up port 80 → %s redirect\n", port)
						fmt.Fprintf(os.Stderr, "WARNING: On Linux, try: setcap 'cap_net_bind_service=+ep' %s\n",
							abs)
					}

					zlog.Errorf("zhttp.Serve: ListenAndServe redirect 80: %s", err)
				}
			}()
		}()
	}

	<-consClosed
}
