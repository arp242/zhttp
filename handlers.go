package zhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"zgo.at/json"
	"zgo.at/zlog"
)

// HandlerRobots writes a simple robots.txt.
func HandlerRobots(rules [][]string) func(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	for _, r := range rules {
		buf.WriteString(fmt.Sprintf("%s\n%s\n", r[0], r[1]))
	}
	text := buf.Bytes()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Cache-Control", "public,max-age=31536000")
		w.WriteHeader(200)
		w.Write(text)
	}
}

// HandlerJSErr logs JavaScript errors from window.onerror.
func HandlerJSErr() func(w http.ResponseWriter, r *http.Request) {
	type Error struct {
		Msg       string `json:"msg"`
		URL       string `json:"url"`
		Loc       string `json:"loc"`
		Line      string `json:"line"`
		Column    string `json:"column"`
		Stack     string `json:"stack"`
		UserAgent string `json:"ua"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var args Error
		_, err := Decode(r, &args)
		if err != nil {
			zlog.Error(err)
			return
		}

		zlog.Fields(zlog.F{
			"url":       args.URL,
			"loc":       args.Loc,
			"line":      args.Line,
			"column":    args.Column,
			"stack":     args.Stack,
			"userAgent": args.UserAgent,
		}).Errorf("JavaScript error: %s", args.Msg)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNoContent)
	}
}

// CSP errors.
type (
	CSPError struct {
		Report Report `json:"csp-report"`
	}

	Report struct {
		BlockedURI   string `json:"blocked-uri"`
		ColumnNumber int    `json:"column-number"`
		DocumentURI  string `json:"document-uri"`
		LineNumber   int    `json:"line-number"`
		Referrer     string `json:"referrer"`
		SourceFile   string `json:"source-file"`
		Violated     string `json:"violated-directive"`
	}
)

// HandlerCSP handles CSP errors.
func HandlerCSP() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		d, _ := io.ReadAll(r.Body)
		var csp CSPError
		err := json.Unmarshal(d, &csp)

		// Probably an extension or something.
		if err == nil && !noise(csp.Report) {
			fmt.Println(string(d))
			zlog.Fields(zlog.F{
				"BlockedURI":   csp.Report.BlockedURI,
				"ColumnNumber": csp.Report.ColumnNumber,
				"DocumentURI":  csp.Report.DocumentURI,
				"LineNumber":   csp.Report.LineNumber,
				"Referrer":     csp.Report.Referrer,
				"SourceFile":   csp.Report.SourceFile,
				"Violated":     csp.Report.Violated,
			}).Errorf("CSP error")
		}

		w.WriteHeader(202)
	}
}

func noise(r Report) bool {
	// Probably some extension or whatnot that injected a script.
	if r.ColumnNumber == 1 && r.LineNumber == 1 &&
		r.Violated == "script-src" &&
		(r.BlockedURI == "inline" || r.BlockedURI == "eval" || r.BlockedURI == "data") {
		return true
	}

	if strings.HasPrefix(r.SourceFile, "safari-extension://") || strings.HasPrefix(r.SourceFile, "moz-extension://") {
		return true
	}

	return false
}

// HandlerRedirectHTTP redirects all HTTP requests to HTTPS.
func HandlerRedirectHTTP(port string) http.HandlerFunc {
	if port == "" {
		port = "443"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Encode()
		if len(q) > 0 {
			q = "?" + q
		}
		w.Header().Set("Location", fmt.Sprintf("https://%s:%s%s%s", r.Host, port, r.URL.Path, q))
		w.WriteHeader(301)
	}
}
