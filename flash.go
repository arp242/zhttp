package zhttp

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"zgo.at/zlog"
)

var cookieFlash = "flash"

func Flash(w http.ResponseWriter, msg string, v ...interface{}) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieFlash,
		Value:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(msg, v...))),
		Path:    "/",
		Expires: time.Now().Add(1 * time.Minute),
	})
}

func ReadFlash(w http.ResponseWriter, r *http.Request) template.HTML {
	c, err := r.Cookie(cookieFlash)
	if err != nil || c.Value == "" {
		// The value won't be read if we set the flash on the same
		// request.
		c = readSetCookie(w)
		if c == nil {
			return ""
		}
	}

	b, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		zlog.Request(r).Error(err)
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieFlash, Value: "", Path: "/",
		Expires: time.Now().Add(-24 * time.Hour),
	})
	return template.HTML(b)
}

func readSetCookie(w http.ResponseWriter) *http.Cookie {
	sk := w.Header().Get("Set-Cookie")
	if sk == "" {
		return nil
	}

	e := strings.Index(sk, "=")
	if e == -1 || sk[:e] != cookieFlash {
		return nil
	}
	s := strings.Index(sk, ";")
	if s == -1 {
		return nil
	}

	return &http.Cookie{Value: sk[e+1 : s]}
}
