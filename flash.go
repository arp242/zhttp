package zhttp

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
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
		return ""
	}

	b, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		zlog.Error(err)
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieFlash, Value: "", Path: "/",
		Expires: time.Now().Add(-24 * time.Hour),
	})
	return template.HTML(b)
}
