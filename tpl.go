package zhttp

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"zgo.at/zlog"
)

// TplPath is the path to template files.
var TplPath = "tpl"

// Add a lock to template.Template for reloading in dev without race
// conditions.
type lockedTpl struct {
	sync.RWMutex
	t *template.Template
}

func (t *lockedTpl) set(tp *template.Template) {
	t.Lock()
	t.t = tp
	t.Unlock()
}

func (t *lockedTpl) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	if t == nil || t.t == nil {
		return errors.New("zhttp.ExecuteTemplate: not initialized; call InitTpl")
	}

	t.RLock()
	defer t.RUnlock()
	return t.t.ExecuteTemplate(wr, name, data)
}

var tpl = new(lockedTpl)

// InitTpl sets up the templates.
func InitTpl(pack map[string][]byte) {
	if pack == nil {
		ReloadTpl()
		return
	}

	t := NewTpl()
	for k, v := range pack {
		k = strings.Trim(strings.TrimPrefix(k, TplPath), "/")
		t = template.Must(t.New(k).Parse(string(v)))
	}
	tpl.set(t)
}

func NewTpl() *template.Template {
	return template.New("").Option("missingkey=error").Funcs(FuncMap)
}

// ReloadTpl reloads the templates.
func ReloadTpl() {
	html, err := filepath.Glob(TplPath + "/*.gohtml")
	if err != nil {
		zlog.Print(err)
	}
	txt, err := filepath.Glob(TplPath + "/*.gotxt")
	if err != nil {
		zlog.Print(err)
	}

	t, err := NewTpl().ParseFiles(append(html, txt...)...)
	if err != nil {
		zlog.Error(err)
	}
	tpl.set(t)
}

// ExecuteTpl executes a named template.
func ExecuteTpl(name string, data interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	tpl.RLock()
	defer tpl.RUnlock()
	err := tpl.t.ExecuteTemplate(w, name, data)
	return w.Bytes(), err
}
