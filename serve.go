package zhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
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
func Serve(flags uint8, server *http.Server, done func()) {
	abs, err := exec.LookPath(os.Args[0])
	if err != nil {
		abs = os.Args[0]
	}

	if strings.HasPrefix(server.Addr, "*:") {
		// Go uses :80 to listen on all addresses, this makes sure that the
		// common-ish "*:80" is also accepted.
		server.Addr = server.Addr[1:]
	}
	host, port, err := net.SplitHostPort(server.Addr)
	if err != nil {
		host = server.Addr
		port = "443"
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		zlog.Errorf("zhttp.Serve: %s", err)
		if errors.Is(err, os.ErrPermission) {
			fmt.Fprintf(os.Stderr, "\nPermission denied to bind to port %s; on Linux, try:\n", port)
			fmt.Fprintf(os.Stderr, "    sudo setcap 'cap_net_bind_service=+ep' %s\n", abs)
		}
		os.Exit(1)
	}
	defer ln.Close()

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
			err = server.ServeTLS(ln, "", "")
		} else {
			err = server.Serve(ln)
		}
		if err != nil && err != http.ErrServerClosed {
			zlog.Errorf("zhttp.Serve: %s", err)
			os.Exit(1)
		}
	}()

	if flags&ServeRedirect != 0 {
		go func() {
			err := http.ListenAndServe(host+":80", HandlerRedirectHTTP(port))
			if err != nil && err != http.ErrServerClosed {
				zlog.Errorf("zhttp.Serve: ListenAndServe redirect 80: %s", err)
				if errors.Is(err, os.ErrPermission) {
					fmt.Fprintf(os.Stderr, "WARNING: No permission to bind to port 80, not setting up port 80 → %s redirect\n", port)
					fmt.Fprintf(os.Stderr, "WARNING: On Linux, try: sudo setcap 'cap_net_bind_service=+ep' %s\n", abs)
				}
			}
		}()
	}

	if done != nil {
		done()
	}

	<-consClosed
}
