package zhttp

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Level constants.
const (
	LevelInfo  = "i"
	LevelError = "e"
)

const cookieFlash = "flash"

// CookieSameSiteHelper can be used to set the SameSite attribute on cookies
// (auth and flash).
//
// The default is to use [http.SameSiteLaxMode].
var CookieSameSiteHelper func(*http.Request) http.SameSite

// Flags for auth cookie.
var (
	CookieAuthName   = "key"
	CookieAuthExpire = 24 * 365 * time.Hour
)

// Flash sets a new flash message at the LevelInfo, overwriting any previous
// messages (if any).
func Flash(w http.ResponseWriter, r *http.Request, msg string) {
	flash(w, r, LevelInfo, msg)
}

// FlashError sets a new flash message at the LevelError, overwriting any
// previous messages (if any).
func FlashError(w http.ResponseWriter, r *http.Request, msg string) {
	flash(w, r, LevelError, msg)
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
		// Simply ignore the error; this is always someone trying to inject
		// values:
		//
		//   flash=iTmVlZCB0byBsb2cgaW4='and(select'1'from/**/cast(md5(1229511929)as/**/int))>'0
		//
		// A second case is the Iframely bot; they seem to send some binary
		// value(?) For example:
		//
		//   flash=iTmVlZCB0byBsb2cgaW4%3D
		//
		// This is always the same value, accross multiple hosts... Not entirely
		// sure what's up with that, but doesn't look like anything we can fix,
		// so 🤷
		//
		// TODO: make a generic "ignore things" feature, or something; we
		// already have LogUnknownFields, and sometimes you might want to log
		// this too. Should switch to slog first though.
		// slog.Error(err)
		return nil
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieFlash, Value: "", Path: CookiePath(),
		Expires: time.Now().Add(-24 * time.Hour),
	})
	return &FlashMessage{string(c.Value[0]), string(b)}
}

func flash(w http.ResponseWriter, r *http.Request, lvl, msg string) {
	if f := ReadFlash(w, &http.Request{}); f != nil {
		slog.Debug("zhttp.flash: double flash message", "msg", msg, "f.Message", f.Message)
	}

	sameSite := http.SameSiteLaxMode
	if CookieSameSiteHelper != nil {
		sameSite = CookieSameSiteHelper(r)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieFlash,
		Value:    lvl + base64.StdEncoding.EncodeToString([]byte(msg)),
		Path:     CookiePath(),
		Expires:  time.Now().Add(1 * time.Minute),
		HttpOnly: true,
		Secure:   IsSecure(r),
		SameSite: sameSite,
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
