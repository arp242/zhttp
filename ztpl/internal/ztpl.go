package internal

import (
	"errors"
	"html/template"
	"io"
	"sync"
)

var Templates = new(LockedTpl)

// Add a lock to template.Template for reloading in dev without race conditions.
type LockedTpl struct {
	sync.RWMutex
	t    *template.Template
	Path string
}

func (t *LockedTpl) Set(path string, tp *template.Template) {
	t.Lock()
	t.Path = path
	t.t = tp
	t.Unlock()
}

func (t *LockedTpl) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	if t == nil || t.t == nil {
		return errors.New("ztpl.ExecuteTemplate: not initialized; call ztpl.Init()")
	}

	t.RLock()
	defer t.RUnlock()
	return t.t.ExecuteTemplate(wr, name, data)
}
