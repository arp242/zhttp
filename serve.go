package zhttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
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

// Serve a HTTP server with graceful shutdown and reasonable timeouts.
//
// If the ReadHeader, Read, Write, or Idle timeouts are 0, then they will be set
// to 10, 60, 60, and 120 seconds respectively.
//
// Errors from the net/http package are logged via zgo.at/zlog instead of the
// default log package. "TLS handshake error" are silenced, since there's rarely
// anything that can be done with that. Define server.ErrorLog if you want the
// old behaviour back.
//
// The server will use TLS if the http.Server has a valid TLSConfig; it will
// also try to redirect port 80 to the TLS server, but will gracefully fail if
// the permission for this is denied.
//
// This will return a channel which will send a value after the server is set up
// and ready to accept connections, and send another value after the server is
// shut down and stopped accepting connections:
//
//   ch := zhttp.Serve(..)
//   <-ch
//   fmt.Println("Ready to serve!")
//
//   <-ch
//   cleanup()
func Serve(flags uint8, testMode int, server *http.Server) chan (struct{}) {
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

	// Set some sane-ish defaults.
	if server.ReadHeaderTimeout == 0 {
		server.ReadHeaderTimeout = 10 * time.Second
	}
	if server.ReadTimeout == 0 {
		server.ReadTimeout = 60 * time.Second
	}
	if server.WriteTimeout == 0 {
		server.WriteTimeout = 60 * time.Second
	}
	if server.IdleTimeout == 0 {
		server.IdleTimeout = 120 * time.Second
	}
	if server.ErrorLog == nil {
		server.ErrorLog = logwrap()
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
		if testMode > 0 {
			fmt.Printf("TEST MODE, SHUTTING DOWN IN %d SECONDS\n", testMode)
			go func() {
				for {
					time.Sleep(time.Duration(testMode) * time.Second)
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

func logwrap() *log.Logger {
	b := new(bytes.Buffer)
	ll := log.New(b, "", 0)

	go func() {
		for {
			l, err := b.ReadString('\n')
			if err != nil {
				if l != "" {
					fmt.Print(l)
				}
				time.Sleep(200 * time.Millisecond)
				continue
			}

			l = strings.TrimRight(l, "\n")

			// Don't log errors we don't care about:
			//
			//   http: TLS handshake error from %s: %s
			//   http2: received GOAWAY [FrameHeader GOAWAY len=20], starting graceful shutdown
			//   http2: server: error reading preface from client %s: %s
			//   http2: timeout waiting for SETTINGS frames from %v
			//
			// This is people sending wrong data; not much we can do about that.
			if strings.HasPrefix(l, "http: TLS handshake") ||
				strings.HasPrefix(l, "http2: received GOAWAY") ||
				strings.HasPrefix(l, "http2: server: error reading preface") ||
				strings.HasPrefix(l, "http2: timeout waiting for SETTINGS") ||
				strings.HasPrefix(l, "write tcp ") {
				continue
			}

			zlog.Module("http").Errorf(l)
		}
	}()

	return ll
}
