package zhttp

import (
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
	"zgo.at/zstd/zstring"
	"zgo.at/zstd/zsync"
)

const (
	ServeRedirect = uint8(0b0001)
)

type st struct{}

func (st) String() string { return "signalStop" }
func (st) Signal()        {}

var signalStop os.Signal = st{}

// Serve a HTTP server with graceful shutdown and reasonable timeouts.
//
// The ReadHeader, Read, Write, or Idle timeouts are set to 10, 60, 60, and 120
// seconds respectively if they are 0.
//
// Errors from the net/http package are logged via zgo.at/zlog instead of the
// default log package. "TLS handshake error" are silenced, since there's rarely
// anything that can be done with that. Define server.ErrorLog if you want the
// old behaviour back.
//
// The server will use TLS if the http.Server has a valid TLSConfig, and will to
// redirect port 80 to the TLS server if ServeRedirect is in flags, but will
// fail gracefully with a warning to stderr if the permission for this is
// denied.
//
// The returned channel sends a value after the server is set up and ready to
// accept connections, and another one after the server is shut down and stopped
// accepting connections:
//
//	ch, _ := zhttp.Serve(..)
//	<-ch
//	fmt.Println("Ready to serve!")
//
//	<-ch
//	cleanup()
//
// The stop channel can be used to tell the server to shut down gracefully; this
// is especially useful for tests:
//
//	stop := make(chan struct{})
//	go zhttp.Serve(0, stop, [..])
//	<-ch // Ready to serve
//
//	time.Sleep(1 * time.Second)
//	stop <- struct{}{}
func Serve(flags uint8, stop chan struct{}, server *http.Server) (chan (struct{}), error) {
	argv0, err := exec.LookPath(os.Args[0])
	if err != nil {
		argv0 = os.Args[0]
	}

	// Go uses ":80" to listen on all addresses, but also accept "*:80".
	if strings.HasPrefix(server.Addr, "*:") {
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
		// Don't log errors we don't care about:
		//
		//   http: TLS handshake error from %s: %s
		//   http2: received GOAWAY [FrameHeader GOAWAY len=20], starting graceful shutdown
		//   http2: server: error reading preface from client %s: %s
		//   http2: timeout waiting for SETTINGS frames from %v
		//   http: URL query contains semicolon, which is no longer a supported separator; parts of the query may be stripped when parsed; see golang.org/issue/25192
		//
		// This is people sending wrong data; not much we can do about that.
		server.ErrorLog = LogWrap(
			"http: TLS handshake",
			"http2: received GOAWAY",
			"http2: server: error reading preface",
			"http2: timeout waiting for SETTINGS",
			"http: URL query contains semicolon",
			"write tcp ")
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			fmt.Fprintf(os.Stderr, "\nPermission denied to bind to port %s; on Linux, try:\n", port)
			fmt.Fprintf(os.Stderr, "    "+suCmd("setcap 'cap_net_bind_service=+ep' "+argv0)+"\n")
		}
		return nil, fmt.Errorf("zhttp.Serve: %w", err)
	}
	// Set back the address; useful when using ":0" for a random port.
	server.Addr = ln.Addr().String()

	// Listen on signal to stop the server and gracefully shut the lot down.
	ch := make(chan struct{}, 1)
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGHUP, syscall.SIGTERM, os.Interrupt /*SIGINT*/)
		go func() {
			<-stop
			s <- signalStop
		}()
		<-s

		err := server.Shutdown(context.Background())
		if err != nil {
			zlog.Errorf("zhttp.Serve shutdown: %s", err)
		}
		ln.Close()

		signal.Stop(s)
		ch <- struct{}{}
		close(ch)
	}()

	// Set up main server.
	go func() {
		var err error
		if server.TLSConfig != nil {
			err = server.ServeTLS(ln, "", "")
		} else {
			err = server.Serve(ln)
		}
		if err != nil && err != http.ErrServerClosed {
			zlog.Errorf("zhttp.Serve: %s", err)
			os.Exit(66)
		}
	}()

	// Set up http → https redirect.
	if flags&ServeRedirect != 0 {
		go func() {
			err := http.ListenAndServe(host+":80", HandlerRedirectHTTP(port))
			if err != nil && err != http.ErrServerClosed {
				zlog.Errorf("zhttp.Serve: ListenAndServe redirect 80: %s", err)
				if errors.Is(err, os.ErrPermission) {
					fmt.Fprintf(os.Stderr,
						"\x1b[1mWARNING: No permission to bind to port 80, not setting up port 80 → %s redirect\x1b[0m\n",
						port)
					fmt.Fprintf(os.Stderr, "WARNING: On Linux, try:\n")
					fmt.Fprintf(os.Stderr, "    "+suCmd("setcap 'cap_net_bind_service=+ep' "+argv0)+"\n")
				}
			}
		}()
	}

	ch <- struct{}{} // Ready to accept connections.
	return ch, nil
}

func suCmd(cmd string) string {
	if _, err := exec.LookPath("doas"); err == nil {
		return "doas " + cmd
	} else if _, err := exec.LookPath("sudo"); err == nil {
		return "sudo " + cmd
	} else {
		if strings.Contains(cmd, " ") {
			return `su -c "` + cmd + `"`
		}
		return `su -c ` + cmd
	}
}

// LogWrap returns a log.Logger which ignores any lines starting with prefixes.
func LogWrap(prefixes ...string) *log.Logger {
	b := zsync.NewBuffer(nil)
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

			if zstring.HasPrefixes(strings.TrimRight(l, "\n"), prefixes...) {
				continue
			}

			zlog.Module("http").Errorf(l)
		}
	}()

	return ll
}
