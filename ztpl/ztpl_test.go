package ztpl

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// func TestMain(m *testing.M) {
// 	os.Exit(TestTemplateExecution(m, ""))
// }

func TestTemplate(t *testing.T) {
	err := Init(fstest.MapFS{
		"a.gohtml": &fstest.MapFile{Data: []byte("html: {{.}}")},
		"a.gotxt":  &fstest.MapFile{Data: []byte("text: {{.}}")},
	})
	if err != nil {
		t.Fatal(err)
	}

	{
		tt := Trace(true)
		if len(tt) > 0 {
			t.Error()
		}
	}

	if !IsLoaded() {
		t.Error()
	}
	if !HasTemplate("a.gohtml") {
		t.Error()
	}
	if !HasTemplate("a.gotxt") {
		t.Error()
	}
	if HasTemplate("aad") {
		t.Error()
	}

	{
		got := List()
		if !reflect.DeepEqual(got, []string{"", "a.gohtml", "a.gotxt"}) {
			t.Error(got)
		}
	}

	{
		got, err := ExecuteString("a.gohtml", "<")
		if err != nil {
			t.Fatal(err)
		}
		if got != "html: &lt;" {
			t.Errorf("%q", got)
		}
	}

	{
		got, err := ExecuteString("a.gotxt", "<")
		if err != nil {
			t.Fatal(err)
		}
		if got != "text: <" {
			t.Errorf("%q", got)
		}
	}

	{
		got := fmt.Sprintf("%s", Trace(false))
		want := `map[a.gohtml:[<] a.gotxt:[<]]`
		if got != want {
			t.Error(got)
		}
		Trace(true) // So TestTemplateExecution() works
	}

	{
		err := Init(fstest.MapFS{
			"a.gohtml": &fstest.MapFile{Data: []byte("html: {{. | unsafe}}")},
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := ExecuteString("a.gohtml", "<")
		if err != nil {
			t.Fatal(err)
		}
		if got != "html: <" {
			t.Errorf("%q", got)
		}
	}

	{
		err := Init(fstest.MapFS{
			"a.gotxt": &fstest.MapFile{Data: []byte("text: {{. | unsafe}}")},
		})
		if err == nil || !strings.Contains(err.Error(), `function "unsafe" not defined`) {
			t.Fatal(err)
		}

		err = Init(fstest.MapFS{
			"a.gotxt": &fstest.MapFile{Data: []byte("text: {{.}}")},
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := ExecuteString("a.gotxt", "<")
		if err != nil {
			t.Fatal(err)
		}
		if got != "text: <" {
			t.Errorf("%q", got)
		}
	}

	func() {
		buf := new(bytes.Buffer)
		stderr = buf
		defer func() { stderr = os.Stderr }()
		c := TestTemplateExecution(&r{1, nil}, "")
		if c != 1 {
			t.Error(c)
		}
		if buf.String() != "" {
			t.Error(buf.String())
		}
	}()

	func() {
		err := Init(fstest.MapFS{
			"a.gohtml": &fstest.MapFile{Data: []byte("html: {{.}}")},
			"a.gotxt":  &fstest.MapFile{Data: []byte("text: {{.}}")},
		})
		if err != nil {
			t.Fatal(err)
		}

		buf := new(bytes.Buffer)
		stderr = buf
		defer func() { stderr = os.Stderr }()
		c := TestTemplateExecution(&r{0, nil}, "")
		if c != 1 {
			t.Error(c)
		}

		want := `
		Didn't execute the templates:
			a.gohtml
			a.gotxt`

		if !strings.Contains(buf.String(), want) {
			t.Error(buf.String())
		}
	}()

	func() {
		err := Init(fstest.MapFS{
			"a.gohtml": &fstest.MapFile{Data: []byte("html: {{.}}")},
			"a.gotxt":  &fstest.MapFile{Data: []byte("text: {{.}}")},
		})
		if err != nil {
			t.Fatal(err)
		}

		buf := new(bytes.Buffer)
		stderr = buf
		defer func() { stderr = os.Stderr }()
		c := TestTemplateExecution(&r{0, func() {
			_, err = ExecuteString("a.gotxt", "")
			if err != nil {
				t.Fatal(err)
			}
			_, err = ExecuteString("a.gohtml", "")
			if err != nil {
				t.Fatal(err)
			}
		}}, "")
		if c != 0 {
			t.Error(c)
		}
		if buf.String() != "" {
			t.Error(buf.String())
		}
	}()
}

type r struct {
	c int
	f func()
}

func (r r) Run() int {
	if r.f != nil {
		r.f()
	}
	return r.c
}
