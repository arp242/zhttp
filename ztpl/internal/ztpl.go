package internal

import (
	"errors"
	"html/template"
	"io"
	"sync"
)

var Templates = new(LockedTpl)

// LockedTpl adds a lock to template.Template for reloading in dev without race
// conditions.
type LockedTpl struct {
	sync.Mutex
	t *template.Template
}

func (t *LockedTpl) Set(tp *template.Template) {
	t.Lock()
	defer t.Unlock()
	t.t = tp
}

func (t *LockedTpl) Has(name string) bool {
	t.Lock()
	defer t.Unlock()
	if t.t == nil {
		return false
	}
	return t.t.Lookup(name) != nil
}

func (t *LockedTpl) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	t.Lock()
	defer t.Unlock()
	if t.t == nil {
		return errors.New("ztpl.ExecuteTemplate: not initialized; call ztpl.Init()")
	}
	return t.t.ExecuteTemplate(wr, name, data)
}
