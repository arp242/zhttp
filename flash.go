package zhttp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"zgo.at/zlog"
)

// Level constants.
const (
	LevelInfo  = "i"
	LevelError = "e"
)

const cookieFlash = "flash"

// Flags to add to all cookies (login and flash).
var (
	CookieSecure   = false
	CookieSameSite = http.SameSiteLaxMode
)

// Flash sets a new flash message at the LevelInfo, overwriting any previous
// messages (if any).
func Flash(w http.ResponseWriter, msg string, v ...interface{}) {
	flash(w, LevelInfo, msg, v...)
}

// FlashError sets a new flash message at the LevelError, overwriting any
// previous messages (if any).
func FlashError(w http.ResponseWriter, msg string, v ...interface{}) {
	flash(w, LevelError, msg, v...)
}

// FlashMessage is a displayed flash message.
type FlashMessage struct {
	Level   string
	Message string
}

// ReadFlash reads any existing flash message, returning the severity level and
// the message itself.
func ReadFlash(w http.ResponseWriter, r *http.Request) *FlashMessage {
	c, err := r.Cookie(cookieFlash)
	if err != nil || c.Value == "" {
		// The value won't be read if we set the flash on the same request.
		c = readSetCookie(w)
		if c == nil {
			return nil
		}
	}

	b, err := base64.StdEncoding.DecodeString(c.Value[1:])
	if err != nil {
		zlog.FieldsRequest(r).Error(err)
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieFlash, Value: "", Path: "/",
		Expires: time.Now().Add(-24 * time.Hour),
	})
	return &FlashMessage{string(c.Value[0]), string(b)}
}

func flash(w http.ResponseWriter, lvl, msg string, v ...interface{}) {
	if f := ReadFlash(w, &http.Request{}); f != nil {
		fmt.Fprintf(os.Stderr, "double flash message while setting %q:\n\talready set: %q\n",
			msg, f.Message)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieFlash,
		Value:    lvl + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(msg, v...))),
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Minute),
		HttpOnly: true,
		Secure:   CookieSecure,
		SameSite: CookieSameSite,
	})
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
