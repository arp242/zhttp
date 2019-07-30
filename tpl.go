package zhttp

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"strings"
	"sync"
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
	t, err := NewTpl().ParseGlob(TplPath + "/*.gohtml")
	if err != nil {
		log.Print(err)
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
