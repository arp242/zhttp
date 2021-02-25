// Package ztpl implements the loading and reloading of templates.
package ztpl

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"os"

	"zgo.at/zhttp/ztpl/internal"
	"zgo.at/zhttp/ztpl/tplfunc"
)

// Init sets up the templates.
func Init(files fs.FS) {
	internal.Templates.Set(template.Must(New().ParseFS(files, "*.gohtml", "*.gotxt")))
}

// New creates a new empty template instance.
func New() *template.Template {
	return template.New("").Option("missingkey=error").Funcs(tplfunc.FuncMap)
}

// Reload the templates from the filesystem.
//
// Errors are logged but not fatal! This is intentional as you really don't want
// a simple typo to crash your app.
func Reload(path string) {
	Init(os.DirFS(path))
}

// IsLoaded reports if templates have been loaded.
func IsLoaded() bool { return internal.Templates != nil }

// HasTemplate reports if this template is loaded.
func HasTemplate(name string) bool { return internal.Templates != nil && internal.Templates.Has(name) }

// Execute a named template.
func Execute(w io.Writer, name string, data interface{}) error {
	return internal.Templates.ExecuteTemplate(w, name, data)
}

// ExecuteBytes a named template and return the data as bytes.
func ExecuteBytes(name string, data interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	err := internal.Templates.ExecuteTemplate(w, name, data)
	return w.Bytes(), err
}

// ExecuteString a named template and return the data as a string.
func ExecuteString(name string, data interface{}) (string, error) {
	w := new(bytes.Buffer)
	err := internal.Templates.ExecuteTemplate(w, name, data)
	return w.String(), err
}
