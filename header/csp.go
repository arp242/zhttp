package header

import (
	"net/http"
	"strings"
)

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

func ParseCSP(h string) CSPArgs {
	var (
		sections = make([][]string, 0, 4)
		sect     = make([]string, 0, 4)
		cur      = make([]rune, 0, 16)
		appsect  = func() {
			s := strings.TrimSpace(string(cur))
			if len(s) > 0 {
				sect = append(sect, s)
			}
			cur = make([]rune, 0, 16)
		}
		app = func() {
			appsect()
			if len(sect) > 0 {
				sections = append(sections, sect)
			}
			sect = make([]string, 0, 4)
		}
	)
	for _, c := range h {
		switch c {
		case ' ', '\t':
			appsect()
		case ';':
			app()
		default:
			cur = append(cur, c)
		}
	}
	app()

	args := make(CSPArgs)
	for _, s := range sections {
		args[s[0]] = s[1:]
	}
	return args
}

func (c CSPArgs) Add(sect string, vals ...string) {
	if _, ok := c[sect]; ok {
		// Have this section: append.
		c[sect] = append(c[sect], vals...)
	} else {
		// Copy default-src (if any) and append.
		c[sect] = append(append([]string{}, c["default-src"]...), vals...)
	}
}

func (c CSPArgs) String() string {
	if len(c) == 0 {
		return ""
	}

	b := new(strings.Builder)
	b.Grow(128)
	for k, v := range c {
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		b.WriteString(k)
		for _, vv := range v {
			b.WriteByte(' ')
			b.WriteString(vv)
		}
	}
	b.WriteString(";")
	return b.String()
}
