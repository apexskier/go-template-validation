package main

import (
	"testing"
	textTemplate "text/template"
)

func TestSimpleParse(t *testing.T) {
	tpl, errs := parse("hello world", textTemplate.New("base"), 0)
	if len(errs) != 0 {
		t.Fatal("errs found")
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
	_, errs := parse("{{if .Value}}", textTemplate.New("base"), 0)
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

func TestUnknownFunctions(t *testing.T) {
	_, errs := parse("{{foo}}{{bar}}", textTemplate.New("base"), 0)
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
	_, errs := parse("{{foo}}", textTemplate.New(""), 0)
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

func TestInvalidIf(t *testing.T) {
	_, errs := parse("{{if}}{{end}}", textTemplate.New("base"), 0)
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
