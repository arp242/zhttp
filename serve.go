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
	"syscall"
	"time"

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
//
// This will return a channel which will send a value after the server is set up
// and ready to accept connections, and send another valiue after the server is
// shut down and stopped accepting connections:
//
//   ch := zhttp.Serve(..)
//   <-ch
//   fmt.Println("Ready to serve!")
//
//   <-ch
//   cleanup()
func Serve(flags uint8, testMode bool, server *http.Server) chan (struct{}) {
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

	ch := make(chan struct{}, 1)
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGHUP, syscall.SIGTERM, os.Interrupt /*SIGINT*/)

		// TODO: might be even better to expose s so tests can send their own
		// signals whenever they're ready.
		if testMode {
			fmt.Println("TEST MODE, SHUTTING DOWN IN A SECOND")
			go func() {
				for {
					time.Sleep(1 * time.Second)
					s <- os.Interrupt
				}
			}()
		}
		<-s

		err := server.Shutdown(context.Background())
		if err != nil {
			zlog.Errorf("server.Shutdown: %s", err)
		}
		ln.Close()
		ch <- struct{}{}
		close(ch)
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
					fmt.Fprintf(os.Stderr,
						"\x1b[1mWARNING: No permission to bind to port 80, not setting up port 80 → %s redirect\x1b[0m\n",
						port)
					fmt.Fprintf(os.Stderr, "WARNING: On Linux, try: sudo setcap 'cap_net_bind_service=+ep' %s\n", abs)
				}
			}
		}()
	}

	ch <- struct{}{}
	return ch
}
