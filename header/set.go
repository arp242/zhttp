package header

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

// Constants for DispositionArgs.
const (
	TypeInline     = "inline"
	TypeAttachment = "attachment"
)

// DispositionArgs are arguments for SetContentDisposition().
type DispositionArgs struct {
	Type     string // disposition-type
	Filename string // filename-parm
	//CreationDate     time.Time // creation-date-parm
	//ModificationDate time.Time // modification-date-parm
	//ReadDate         time.Time // read-date-parm
	//Size             int       // size-parm
}

// SetContentDisposition sets the Content-Disposition header. Any previous value
// will be overwritten.
//
// https://tools.ietf.org/html/rfc2183
// https://tools.ietf.org/html/rfc6266
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Disposition
func SetContentDisposition(header http.Header, args DispositionArgs) error {
	if header == nil {
		return errors.New("header is nil map")
	}

	if args.Type == "" {
		return errors.New("the Type field is mandatory")
	}
	if args.Type != TypeInline && args.Type != TypeAttachment {
		return fmt.Errorf("the Type field must be %#v or %#v", TypeInline, TypeAttachment)
	}
	v := args.Type

	if args.Filename != "" {
		// Format filename= according to <quoted-string> as defined in RFC822.
		// We don't don't allow \, and % though. Replacing \ is a slightly lazy
		// way to prevent certain injections in case of user-provided strings
		// (ending the quoting and injecting their own values or even headers).
		// % because some user agents interpret percent-encodings, and others do
		// not (according to the RFC anyway). Finally escape " with \".
		r := strings.NewReplacer("\\", "", "%", "", `"`, `\"`)
		args.Filename = r.Replace(args.Filename)

		// Don't allow unicode in the "filename" attribute; instead, add that to
		// the filename* one.
		filename, ascii, hasUni := formatFilename(args.Filename)
		v += fmt.Sprintf(`; filename="%v"`, ascii)

		// Add filename* for unicode, encoded according to
		// https://tools.ietf.org/html/rfc5987
		if hasUni {
			v += fmt.Sprintf("; filename*=UTF-8''%v",
				url.QueryEscape(filename))
		}
	}

	header.Set("Content-Disposition", v)
	return nil
}

func formatFilename(s string) (string, string, bool) {
	uni := make([]rune, len(s))
	ascii := make([]rune, len(s))
	has := false
	asciiIdx := 0
	uniIdx := 0
	for _, c := range s {
		if unicode.IsControl(c) {
			continue
		}

		switch {
		case c > 255:
			has = true
		default:
			ascii[asciiIdx] = c
			asciiIdx++
		}

		uni[uniIdx] = c
		uniIdx++
	}

	return strings.TrimRight(string(uni), "\x00"), strings.TrimRight(string(ascii), "\x00"), has
}

// CSP Directives.
const (
	// Fetch directives
	CSPChildSrc    = "child-src"    // Web workers and nested contexts such as frames
	CSPConnectSrc  = "connect-src"  // Script interfaces: Ajax, WebSocket, Fetch API, etc
	CSPDefaultSrc  = "default-src"  // Fallback for the other directives
	CSPFontSrc     = "font-src"     // Custom fonts
	CSPFrameSrc    = "frame-src"    // <frame> and <iframe>
	CSPImgSrc      = "img-src"      // Images (HTML and CSS), favicon
	CSPManifestSrc = "manifest-src" // Web app manifest
	CSPMediaSrc    = "media-src"    // <audio> and <video>
	CSPObjectSrc   = "object-src"   // <object>, <embed>, and <applet>
	CSPScriptSrc   = "script-src"   // JavaScript
	CSPStyleSrc    = "style-src"    // CSS

	// Document directives govern the properties of a document
	CSPBaseURI     = "base-uri"     // Restrict what can be used in <base>
	CSPPluginTypes = "plugin-types" // Whitelist MIME types for <object>, <embed>, <applet>
	CSPSandbox     = "sandbox"      // Enable sandbox for the page

	// Navigation directives govern whereto a user can navigate
	CSPFormAction     = "form-action"     // Restrict targets for form submissions
	CSPFrameAncestors = "frame-ancestors" // Valid parents for embedding with frames, <object>, etc.

	// Reporting directives control the reporting process of CSP violations; see
	// also the Content-Security-Policy-Report-Only header
	CSPReportURI = "report-uri"

	// Other directives
	CSPBlockAllMixedContent = "block-all-mixed-content" // Don't load any HTTP content when using https
)

// Content-Security-Policy values
const (
	CSPSourceSelf         = "'self'"          // Exact origin of the document
	CSPSourceNone         = "'none'"          // Nothing matches
	CSPSourceUnsafeInline = "'unsafe-inline'" // Inline <script>/<style>, onevent="", etc.
	CSPSourceUnsafeEval   = "'unsafe-eval'"   // eval()
	CSPSourceStar         = "*"               // Everything

	CSPSourceHTTP        = "http:"
	CSPSourceHTTPS       = "https:"
	CSPSourceData        = "data:"
	CSPSourceMediastream = "mediastream:"
	CSPSourceBlob        = "blob:"
	CSPSourceFilesystem  = "filesystem:"
)

// CSPArgs are arguments for SetCSP().
type CSPArgs map[string][]string

// SetCSP sets a Content-Security-Policy header.
//
// Most directives require a value. The exceptions are CSPSandbox and
// CSPBlockAllMixedContent.
//
// Only special values (CSPSource* constants) need to be quoted. Don't add
// quotes around hosts.
//
// Valid sources:
//
//	CSPSource*
//	Hosts               example.com, *.example.com, https://example.com
//	Schema              data:, blob:, etc.
//	nonce-<val>         inline scripts using a cryptographic nonce
//	<hash_algo>-<val>   hash of specific script.
//
// Also see: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP and
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
func SetCSP(header http.Header, args CSPArgs) {
	if header == nil {
		panic("SetCSP: header is nil map")
	}

	var b strings.Builder
	i := 1
	for k, v := range args {
		b.WriteString(k)
		b.WriteString(" ")

		for j := range v {
			b.WriteString(v[j])
			if j != len(v)-1 {
				b.WriteString(" ")
			}
		}

		if i != len(args) {
			b.WriteString("; ")
		}
		i++
	}

	header["Content-Security-Policy"] = []string{b.String()}
}
