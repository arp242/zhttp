package internal

import (
	"errors"
	"html/template"
	"io"
	"sort"
	"sync"
	"text/template/parse"
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

func (t *LockedTpl) List() []string {
	t.Lock()
	defer t.Unlock()
	var (
		tpl = t.t.Templates()
		l   = make([]string, 0, len(tpl))
	)
	for _, a := range tpl {
		l = append(l, a.Name())
	}
	return l
}

type Trace map[string][]interface{}

func (t Trace) Names() []string {
	names := make([]string, 0, len(t))
	for tt := range t {
		names = append(names, tt)
	}
	sort.Strings(names)
	return names
}

var (
	TraceEnabled bool
	Ran          Trace
)

func (t *LockedTpl) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	t.Lock()
	defer t.Unlock()
	if t.t == nil {
		return errors.New("ztpl.ExecuteTemplate: not initialized; call ztpl.Init()")
	}

	if TraceEnabled {
		Ran[name] = append(Ran[name], data)

		invoks := make(map[string]struct{})
		for _, n := range t.t.Lookup(name).Tree.Root.Nodes {
			for _, k := range lsTpl(t.t, n) {
				invoks[k] = struct{}{}
			}
		}
		for k := range invoks {
			// TODO: need to process TemplateNode.Pipe to get the data
			Ran[k] = append(Ran[k], nil)
		}
	}

	return t.t.ExecuteTemplate(wr, name, data)
}

func lsTpl(t *template.Template, n parse.Node) []string {
	var (
		tpls  []string
		nodes []parse.Node
	)
	switch nn := n.(type) {
	case *parse.TemplateNode:
		tpls = append(tpls, nn.Name)
		for _, n2 := range t.Lookup(nn.Name).Tree.Root.Nodes {
			tpls = append(tpls, lsTpl(t, n2)...)
		}
	case *parse.IfNode:
		if nn.List != nil {
			nodes = append(nodes, nn.List.Nodes...)
		}
		if nn.ElseList != nil {
			nodes = append(nodes, nn.ElseList.Nodes...)
		}
	case *parse.ListNode:
		nodes = nn.Nodes
	case *parse.CommandNode:
		nodes = nn.Args
	case *parse.PipeNode:
		for _, n2 := range nn.Cmds {
			nodes = append(nodes, n2.Args...)
		}
	}
	for _, n2 := range nodes {
		tpls = append(tpls, lsTpl(t, n2)...)
	}
	return tpls
}
