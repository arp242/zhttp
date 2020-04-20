// Package zmail is a simple mail sender.
package zmail

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"hash/fnv"
	"math/big"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"zgo.at/zhttp"
	"zgo.at/zlog"
)

var (
	SMTP  = ""   // SMTP server connection string.
	Print = true // Print emails to stdout if SMTP if empty.
)

// Send an email.
func Send(subject string, from mail.Address, to []mail.Address, body string) error {
	_ = zhttp.ExecuteTpl
	msg := format(subject, from, to, body)
	toList := make([]string, len(to))
	for i := range to {
		toList[i] = to[i].Address
	}

	var err error
	switch SMTP {
	case "stdout":
		if Print {
			l := strings.Repeat("═", 50)
			fmt.Println("╔═══ EMAIL " + l + "\n║ " +
				strings.Replace(strings.TrimSpace(string(msg)), "\r\n", "\r\n║ ", -1) +
				"\n╚══════════" + l + "\n")
		}
	case "":
		err = sendMail(subject, from, toList, msg)
	default:
		err = sendRelay(subject, from, toList, msg)
	}

	if err != nil {
		return fmt.Errorf("zmail.Send: %w", err)
	}
	return nil
}

var deHTML = strings.NewReplacer(
	"&lt;", "<",
	"&gt;", ">",
	"&amp;", "&",
	"&quot;", `"`)

func SendTemplate(subject string, from mail.Address, to []mail.Address, tplName string, tplArgs interface{}) error {
	body, err := zhttp.ExecuteTpl(tplName, tplArgs)
	if err != nil {
		return fmt.Errorf("zmail.SendTemplate: %q: %w", tplName, err)
	}

	err = Send(subject, from, to, deHTML.Replace(string(body)))
	if err != nil {
		return fmt.Errorf("zmail.SendTemplate: %w", err)
	}
	return nil

}

var hostname sync.Once

// Send direct.
func sendMail(subject string, from mail.Address, to []string, body []byte) error {
	hello := "localhost"
	hostname.Do(func() {
		var err error
		hello, err = os.Hostname()
		if err != nil {
			zlog.Error(err)
		}
	})

	go func() {
		for _, t := range to {
			zlog.Module("zmail").Fields(zlog.F{
				"subject": subject,
				"to":      t,
			}).Debugf("sendMail")
			domain := t[strings.LastIndex(t, "@")+1:]
			mxs, err := net.LookupMX(domain)

			var hosts []string
			if err != nil {
				hosts = []string{domain}
			} else {
				for _, mx := range mxs {
					hosts = append(hosts, mx.Host)
				}
			}

			for _, h := range hosts {
				logerr := func(err error) {
					zlog.Fields(zlog.F{
						"host": h,
						"from": from,
						"to":   to,
					}).Error(err)
				}

				c, err := smtp.Dial(h + ":25")
				if err != nil {
					logerr(err)
					if strings.Contains(err.Error(), " blocked ") {
						// 14:52:24 ERROR: 554 5.7.1 Service unavailable; Client host [xxx.xxx.xx.xx] blocked using
						// xbl.spamhaus.org.rbl.local; https://www.spamhaus.org/query/ip/xxx.xxx.xx.xx
						break
					}
					continue // Can't connect: try next MX
				}
				defer c.Close()

				err = c.Hello(hello)
				if err != nil {
					logerr(err)
					// Errors from here on are probably fatal error, so just
					// abort.
					// TODO: could improve by checking the status code, but
					// net/smtp doesn't provide them in a good way. This is fine
					// for now as it's intended as a simple backup solution.
					break
				}

				if ok, _ := c.Extension("STARTTLS"); ok {
					err := c.StartTLS(&tls.Config{ServerName: h})
					if err != nil {
						logerr(err)
						break
					}
				}

				err = c.Mail(from.Address)
				if err != nil {
					logerr(err)
					break
				}
				for _, addr := range to {
					err = c.Rcpt(addr)
					if err != nil {
						logerr(err)
						break
					}
				}

				w, err := c.Data()
				if err != nil {
					logerr(err)
					break
				}
				_, err = w.Write(body)
				if err != nil {
					logerr(err)
					break
				}

				err = w.Close()
				if err != nil {
					logerr(err)
					break
				}

				err = c.Quit()
				if err != nil {
					logerr(err)
					break
				}

				break
			}
		}
	}()
	return nil
}

// Send via relay.
func sendRelay(subject string, from mail.Address, to []string, body []byte) error {
	srv, err := url.Parse(SMTP)
	if err != nil {
		return err
	}

	user := srv.User.Username()
	pw, _ := srv.User.Password()
	host := srv.Host
	if h, _, err := net.SplitHostPort(srv.Host); err == nil {
		host = h
	}

	go func() {
		var auth smtp.Auth
		if user != "" {
			auth = smtp.PlainAuth("", user, pw, host)
		}

		zlog.Module("zmail").Fields(zlog.F{
			"subject": subject,
			"to":      to,
		}).Debugf("sendRelay")
		err := smtp.SendMail(srv.Host, auth, from.Address, to, body)
		if err != nil {
			zlog.Fields(zlog.F{
				"host": srv.Host,
				"from": from,
				"to":   to,
			}).Errorf("smtp.SendMail: %w", err)
		}
	}()
	return nil
}

var reNL = regexp.MustCompile(`(\r\n){2,}`)

// format a message.
func format(subject string, from mail.Address, to []mail.Address, body string) []byte {
	var msg strings.Builder
	t := time.Now()

	fmt.Fprintf(&msg, "From: %s\r\n", from.String())

	tos := make([]string, len(to))
	for i := range to {
		tos[i] = to[i].String()
	}
	fmt.Fprintf(&msg, "To: %s\r\n", strings.Join(tos, ","))

	// Create Message-ID
	domain := from.Address[strings.Index(from.Address, "@")+1:]
	hash := fnv.New64a()
	hash.Write([]byte(body))
	rnd, _ := rand.Int(rand.Reader, big.NewInt(0).SetUint64(18446744073709551615))
	msgid := fmt.Sprintf("zmail-%s-%s@%s", strconv.FormatUint(hash.Sum64(), 36),
		strconv.FormatUint(rnd.Uint64(), 36), domain)

	fmt.Fprintf(&msg, "Date: %s\r\n", t.Format(time.RFC1123Z))
	fmt.Fprintf(&msg, "Content-Type: text/plain;charset=utf-8\r\n")
	fmt.Fprintf(&msg, "Content-Transfer-Encoding: quoted-printable\r\n")
	fmt.Fprintf(&msg, "Message-ID: <%s>\r\n", msgid)
	fmt.Fprintf(&msg, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", reNL.ReplaceAllString(subject, "")))
	msg.WriteString("\r\n")

	w := quotedprintable.NewWriter(&msg)
	w.Write([]byte(body))
	w.Close()

	return []byte(msg.String())
}
