// Package ztpl implements the loading and reloading of templates.
package ztpl

// TODO: move this to zstd/ztemplate

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	textTemplate "text/template"

	"zgo.at/zhttp/ztpl/internal"
	"zgo.at/zhttp/ztpl/tplfunc"
	"zgo.at/zstd/zruntime"
	"zgo.at/zstd/zstring"
)

// Init sets up the templates.
//
// This assumes that HTML templates have the .gohtml extension, and that text
// templates have the .gotxt extension.
func Init(files fs.FS) error {
	html, err := htmlTemplate.New("").Option("missingkey=error").
		Funcs(tplfunc.FuncMap).Funcs(tplfunc.FuncMapHTML).
		ParseFS(files, "*.gohtml")
	if err != nil && !strings.Contains(err.Error(), "matches no files: `*.gohtml`") {
		return fmt.Errorf("ztpl.Init: %w", err)
	}

	text, err := textTemplate.New("").Option("missingkey=error").
		Funcs(tplfunc.FuncMap).
		ParseFS(files, "*.gotxt")
	if err != nil && !strings.Contains(err.Error(), "matches no files: `*.gotxt`") {
		return fmt.Errorf("ztpl.Init: %w", err)
	}

	internal.Templates.Set(html, text)
	return nil
}

// Reload the templates from the filesystem.
func Reload(path string) error { return Init(os.DirFS(path)) }

// IsLoaded reports if templates have been loaded.
func IsLoaded() bool { return internal.Templates != nil }

// List all template names.
func List() []string {
	if internal.Templates == nil {
		return nil
	}
	return internal.Templates.List()
}

// HasTemplate reports if this template is loaded.
func HasTemplate(name string) bool { return internal.Templates != nil && internal.Templates.Has(name) }

// Execute a named template.
func Execute(w io.Writer, name string, data interface{}) error {
	return internal.Templates.ExecuteTemplate(w, name, data)
}

// ExecuteString a named template and return the data as a string.
func ExecuteString(name string, data interface{}) (string, error) {
	w := new(bytes.Buffer)
	err := Execute(w, name, data)
	return w.String(), err
}

// ExecuteBytes a named template and return the data as a byte slice.
func ExecuteBytes(name string, data interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	err := Execute(w, name, data)
	return w.Bytes(), err
}

// Trace enables tracking of all template executions. When this is disabled the
// stored list is emptied and the previous value returned.
//
// This is mostly useful for tracking which templates are run from tests, e.g.:
//
//   func TestMain(m *testing.M) {
//       ztpl.Trace(true)
//       c := m.Run()
//
//       ran := ztpl.Trace(false)
//       for _, t := range ztpl.List() {
//           if _, ok := ran[t]; !ok {
//               fmt.Println("didn't execute template", t)
//           }
//       }
//       os.Exit(c)
//   }
//
// Also see TestTemplateExecution(), which wraps this for convenient use.
func Trace(on bool) internal.Trace {
	r := internal.Ran
	internal.TraceEnabled = on
	internal.Ran = make(internal.Trace)
	return r
}

var stderr io.Writer = os.Stderr

// TestTemplateExecution tests if all templates loaded through ztpl are
// executed.
//
// Go templates are dynamically typed and not counted in code coverage tools;
// this is a simple way to ensure all templates are executed at least once.
//
// Typical usage would be:
//
//   func TestMain(m *testing.M) {
//       os.Exit(ztpl.TestTemplateExecution(m, "ignore_this.gohtml"))
//   }
//func TestTemplateExecution(m *testing.M, ignore ...string) int {
func TestTemplateExecution(m interface{ Run() int }, ignore ...string) int {
	Trace(true)
	c := m.Run()

	// Don't run if tests failed or if -run was given; it's likely to cause
	// failures anyway so isn't all that useful.
	//
	// TODO: provide option to list which templates were executed when -run is
	// given; maybe through env? Can be useful for testing and seeing which
	// templates are being run.
	if c >= 1 {
		return c
	}
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.run") || strings.HasPrefix(a, "-test.bench") {
			return c
		}
	}

	ran := Trace(false).Names()
	unrun := zstring.Difference(List(), ran, ignore)
	sort.Strings(unrun)
	if len(unrun) == 0 {
		return 0
	}

	fmt.Fprintln(stderr, "    --- FAIL: ztpl.TestTemplateExecution")
	fmt.Fprintf(stderr, "\t\tDidn't execute the templates:\n\t\t\t%s\n", strings.Join(unrun, "\n\t\t\t"))
	if zruntime.TestVerbose() {
		fmt.Fprintf(stderr, "\n\t\tTemplates executed: %s\n\n", ran)
	}
	return 1
}
