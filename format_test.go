package main

import (
	"bytes"
	"reflect"
	"testing"
	"text/template"
)

func TestFormat(t *testing.T) {
	tpl, _ := template.New("base").Parse("test {{.Foo}} {{if 1}}{{end}}")
	var buf bytes.Buffer
	s := state{
		tmpl:  tpl,
		wr:    &buf,
		node:  tpl.Tree.Root,
		vars:  make([]variable, 0),
		depth: 0,
	}
	s.walk(reflect.Value{}, tpl.Root)

	t.Logf("`%v`", buf.String())

	t.Error("failed")
}
