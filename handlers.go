package zhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"zgo.at/zlog"
)

// HandlerRobots writes a simple robots.txt.
func HandlerRobots(rules [][]string) func(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	for _, r := range rules {
		buf.WriteString(fmt.Sprintf("%s\n%s\n\n", r[0], r[1]))
	}
	text := buf.Bytes()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Cache-Control", "public,max-age=31536000")
		w.WriteHeader(200)
		w.Write(text)
	}
}

// CSP errors.
type (
	CSPError struct {
		Report Report `json:"csp-report"`
	}

	Report struct {
		ColumnNumber int    `json:"column-number"`
		LineNumber   int    `json:"line-number"`
		BlockedURI   string `json:"blocked-uri"`
		Violated     string `json:"violated-directive"`
		SourceFile   string `json:"source-file"`
	}
)

// HandlerCSP handles CSP errors.
func HandlerCSP() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		d, _ := ioutil.ReadAll(r.Body)
		var csp CSPError
		err := json.Unmarshal(d, &csp)

		// Probably an extension or something.
		if err == nil && !noise(csp.Report) {
			zlog.Errorf("CSP error: %s", string(d))
		}

		w.WriteHeader(202)
	}
}

func noise(r Report) bool {
	// Probably some extension or whatnot that injected a script.
	if r.ColumnNumber == 1 && r.LineNumber == 1 && r.BlockedURI == "inline" && r.Violated == "script-src" {
		return true
	}

	if strings.HasPrefix(r.SourceFile, "safari-extension://") {
		return true
	}

	return false
}
