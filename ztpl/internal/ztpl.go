package internal

import (
	"errors"
	htmlTemplate "html/template"
	"io"
	"sort"
	"strings"
	"sync"
	textTemplate "text/template"
	"text/template/parse"
)

var Templates = new(LockedTpl)

type (
	Template interface {
		ExecuteTemplate(io.Writer, string, interface{}) error
		Lookup(string) Template
		GetTree(string) *parse.Tree
		TemplateNames() []string
	}

	// LockedTpl adds a lock to template.Template for reloading in dev without
	// race conditions.
	LockedTpl struct {
		sync.Mutex
		html Template
		text Template
	}

	htmlWrap struct{ *htmlTemplate.Template }
	textWrap struct{ *textTemplate.Template }
)

func (t htmlWrap) GetTree(n string) *parse.Tree { return t.Template.Lookup(n).Tree }
func (t textWrap) GetTree(n string) *parse.Tree { return t.Template.Lookup(n).Tree }
func (t htmlWrap) Lookup(n string) Template {
	tt := t.Template.Lookup(n)
	if tt == nil {
		return nil
	}
	return htmlWrap{tt}
}
func (t textWrap) Lookup(n string) Template {
	tt := t.Template.Lookup(n)
	if tt == nil {
		return nil
	}
	return textWrap{tt}
}
func (t htmlWrap) TemplateNames() []string {
	tpl := t.Template.Templates()
	names := make([]string, 0, len(tpl))
	for _, n := range tpl {
		names = append(names, n.Name())
	}
	return names
}
func (t textWrap) TemplateNames() []string {
	tpl := t.Template.Templates()
	names := make([]string, 0, len(tpl))
	for _, n := range tpl {
		names = append(names, n.Name())
	}
	return names
}

func (t *LockedTpl) Set(html *htmlTemplate.Template, text *textTemplate.Template) {
	t.Lock()
	defer t.Unlock()

	if html == nil {
		html = htmlTemplate.New("")
	}
	t.html = htmlWrap{html}

	if text == nil {
		text = textTemplate.New("")
	}
	t.text = textWrap{text}
}

func (t *LockedTpl) Has(name string) bool {
	t.Lock()
	defer t.Unlock()
	if t.html == nil && t.text == nil {
		return false
	}
	if t.html.Lookup(name) != nil {
		return true
	}
	return t.text.Lookup(name) != nil
}

func (t *LockedTpl) List() []string {
	t.Lock()
	defer t.Unlock()
	l := append(t.html.TemplateNames(), t.text.TemplateNames()...)
	sort.Strings(l)
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
	if t.html == nil && t.text == nil {
		return errors.New("ztpl.ExecuteTemplate: not initialized; call ztpl.Init()")
	}

	var tpl Template = t.html
	if strings.HasSuffix(name, ".gotxt") {
		tpl = t.text
	}

	if TraceEnabled {
		Ran[name] = append(Ran[name], data)

		invoks := make(map[string]struct{})
		for _, n := range tpl.GetTree(name).Root.Nodes {
			for _, k := range lsTpl(tpl, n) {
				invoks[k] = struct{}{}
			}
		}
		for k := range invoks {
			// TODO: need to process TemplateNode.Pipe to get the data
			Ran[k] = append(Ran[k], nil)
		}
	}

	return tpl.ExecuteTemplate(wr, name, data)
}

func lsTpl(t Template, n parse.Node) []string {
	var (
		tpls  []string
		nodes []parse.Node
	)
	switch nn := n.(type) {
	case *parse.TemplateNode:
		tpls = append(tpls, nn.Name)
		for _, n2 := range t.GetTree(nn.Name).Root.Nodes {
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
