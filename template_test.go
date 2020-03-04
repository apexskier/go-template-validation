package main

import (
	"bytes"
	"testing"
	textTemplate "text/template"
)

func TestParseSimple(t *testing.T) {
	tpl, errs := parse("hello world", textTemplate.New("base"))
	if len(errs) != 0 {
		t.Fatalf("errs found: %v", errs)
	}
	defined := tpl.DefinedTemplates()
	if defined != `; defined templates are: "base"` {
		t.Errorf("unexpected templates defined: %s", defined)
	}
}

func assertError(t *testing.T, expected templateError, actual templateError) {
	if expected.Char != actual.Char {
		t.Errorf("error Char doesn't match: expected `%d`, actual `%d`", expected.Char, actual.Char)
	}
	if expected.Line != actual.Line {
		t.Errorf("error Line doesn't match: expected `%d`, actual `%d`", expected.Line, actual.Line)
	}
	if expected.Level != actual.Level {
		t.Errorf("error Level doesn't match: expected `%s`, actual `%s`", expected.Level, actual.Level)
	}
	if expected.Description != actual.Description {
		t.Errorf("error Description doesn't match: expected `%s`, actual `%s`", expected.Description, actual.Description)
	}
}

func TestParseUnexpectedEOF(t *testing.T) {
	_, errs := parse("{{if .Value}}", textTemplate.New("base"))
	if len(errs) != 1 {
		t.Errorf("unexpected errors found: %v", errs)
	}
	assertError(t, templateError{
		Char:        -1,
		Line:        0,
		Level:       parseErrorLevel,
		Description: "unexpected EOF",
	}, errs[0])
}

func TestParseUnknownFunctions(t *testing.T) {
	_, errs := parse("{{foo}}{{bar}}", textTemplate.New("base"))
	if len(errs) != 2 {
		t.Errorf("unexpected errors found: %v", errs)
	}
	assertError(t, templateError{
		Char:        2,
		Line:        0,
		Level:       parseErrorLevel,
		Description: `function "foo" not defined`,
	}, errs[0])
	assertError(t, templateError{
		Char:        9,
		Line:        0,
		Level:       parseErrorLevel,
		Description: `function "bar" not defined`,
	}, errs[1])
}

func TestParseNoname(t *testing.T) {
	_, errs := parse("{{foo}}", textTemplate.New(""))
	if len(errs) != 1 {
		t.Errorf("unexpected errors found: %v", errs)
	}
	assertError(t, templateError{
		Char:        2,
		Line:        0,
		Level:       parseErrorLevel,
		Description: `function "foo" not defined`,
	}, errs[0])
}

func TestParseInvalidIf(t *testing.T) {
	_, errs := parse("{{if}}{{end}}", textTemplate.New("base"))
	if len(errs) != 1 {
		t.Errorf("unexpected errors found: %v", errs)
	}
	assertError(t, templateError{
		Char:        -1,
		Line:        0,
		Level:       parseErrorLevel,
		Description: `missing value for if`,
	}, errs[0])
}

func TestParseIndexSyntax(t *testing.T) {
	_, errs := parse("<{{.Foo[2]}}>", textTemplate.New("base"))
	if len(errs) != 1 {
		t.Errorf("unexpected errors found: %v", errs)
	}
	assertError(t, templateError{
		Char:        7,
		Line:        0,
		Level:       parseErrorLevel,
		Description: `unexpected bad character U+005B '[' in command`,
	}, errs[0])
}

func TestParseEmptyCommand(t *testing.T) {
	for _, testCase := range []string{"{{}}", "{{- }}", "{{  -}}"} {
		_, errs := parse(testCase, textTemplate.New("base"))
		if len(errs) != 1 {
			t.Errorf("unexpected errors found: %v", errs)
		}
		assertError(t, templateError{
			Char:        0,
			Line:        0,
			Level:       parseErrorLevel,
			Description: `missing value for command`,
		}, errs[0])
	}
}

func TestParseEmptyCommands(t *testing.T) {
	_, errs := parse("\n\n{{ }} hello world {{ }}", textTemplate.New("base"))
	if len(errs) != 2 {
		t.Errorf("unexpected errors found: %v", errs)
	}
	assertError(t, templateError{
		Char:        0,
		Line:        2,
		Level:       parseErrorLevel,
		Description: `missing value for command`,
	}, errs[0])
	assertError(t, templateError{
		Char:        18,
		Line:        2,
		Level:       parseErrorLevel,
		Description: `missing value for command`,
	}, errs[1])
}

func TestExecWorks(t *testing.T) {
	tpl, _ := textTemplate.New("base").Parse("<{{.Value}}>")
	var buf bytes.Buffer
	errs := exec(tpl, struct{ Value string }{Value: "foo"}, &buf)
	if len(errs) != 0 {
		t.Errorf("errs found: %v", errs)
	}
	if buf.String() != "<foo>" {
		t.Errorf("output doesn't match: `%s`", buf.String())
	}
}

func TestExecGenericStruct(t *testing.T) {
	tpl, _ := textTemplate.New("base").Parse("<{{.Foo.Bar}}>")
	var buf bytes.Buffer
	errs := exec(tpl, map[string]interface{}{}, &buf)
	if len(errs) != 0 {
		t.Errorf("errs found: %v", errs)
	}
}

func TestExecMissing(t *testing.T) {
	tpl, _ := textTemplate.New("base").Parse("<{{.Value}}>")
	var buf bytes.Buffer
	errs := exec(tpl, struct{}{}, &buf)
	if len(errs) != 1 {
		t.Errorf("unexpected errs: %v", errs)
	}
	assertError(t, templateError{
		Line:        0,
		Char:        3,
		Level:       execErrorLevel,
		Description: `executing "base" at <.Value>: can't evaluate field Value in type struct {}`,
	}, errs[0])
	if buf.String() != "<" {
		t.Errorf("output doesn't match: `%s`", buf.String())
	}
}

func TestExecIgnoresIncomplete(t *testing.T) {
	tpl := textTemplate.New("base")
	var buf bytes.Buffer
	errs := exec(tpl, nil, &buf)
	if len(errs) != 0 {
		t.Errorf("errs found: %v", errs)
	}
}
