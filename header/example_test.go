package header_test

import (
	"net/http"

	"zgo.at/zhttp/header"
)

func ExampleSetCSP() {
	static := "static.example.com"
	headers := make(http.Header)
	header.SetCSP(headers, header.CSPArgs{
		header.CSPDefaultSrc: {header.CSPSourceNone},
		header.CSPScriptSrc:  {static},
		header.CSPStyleSrc:   {static, header.CSPSourceUnsafeInline},
		header.CSPFormAction: {header.CSPSourceSelf},
		header.CSPReportURI:  {"/csp"},
	})

	// Output:
}

func ExampleSetContentDisposition() {
	headers := make(http.Header)
	header.SetContentDisposition(headers, header.DispositionArgs{
		Type:     "image/png",
		Filename: "foo.png",
	})

	// Output:
}
