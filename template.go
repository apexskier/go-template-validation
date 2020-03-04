package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	textTemplate "text/template"
)

var (
	maxFixes                    = 10
	templateErrorRegex          = regexp.MustCompile(`template: (.*?):((\d+):)?(\d+): (.*)`)
	findTokenRegex              = regexp.MustCompile(`['"](.+)['"]`)
	findExprRegex               = regexp.MustCompile(`<(\.Value)>`)
	functionNotFoundRegex       = regexp.MustCompile(`function "(.+)" not defined`)
	missingValueForCommandRegex = regexp.MustCompile(`missing value for command`)
	firstEmptyCommandRegex      = regexp.MustCompile(`{{((-?\s*?)|(\s*?-?))}}`)
)

func createTemplateError(err error, level ErrorLevel) templateError {
	matches := templateErrorRegex.FindStringSubmatch(err.Error())
	if len(matches) == 6 {
		// tplName := matches[1]

		// 2 is line + : group if char is found
		// line is in pos 4, unless a char is found in which case it's 3 and char is 4

		lineIndex := 4
		char := -1
		if matches[3] != "" {
			lineIndex = 3
			char, err = strconv.Atoi(matches[4])
			if err != nil {
				char = -1
			}
		}

		line, err := strconv.Atoi(matches[lineIndex])
		if err != nil {
			line = -1
		} else {
			line = line - 1
		}

		description := matches[5]

		return templateError{
			Line:        line,
			Char:        char,
			Description: description,
			Level:       level,
		}
	}
	return templateError{
		Line:        -1,
		Char:        -1,
		Description: err.Error(),
		Level:       misunderstoodError,
	}
}

func parse(text string, baseTpl *textTemplate.Template) (*textTemplate.Template, []templateError) {
	return parseInternal(text, baseTpl, 0)
}

func parseInternal(text string, baseTpl *textTemplate.Template, depth int) (t *textTemplate.Template, tplErrs []templateError) {
	lines := SplitLines(text)

	if depth >= maxFixes {
		return baseTpl, tplErrs
	}

	t, err := baseTpl.Parse(text)
	if err != nil {
		tplErrs = append(tplErrs, createTemplateError(err, parseErrorLevel))
		// make this mutable
		tplErr := &tplErrs[len(tplErrs)-1]

		if tplErr.Level != misunderstoodError {
			if tplErr.Char == -1 {
				// try to find a character to line up with
				tokenLoc := findTokenRegex.FindStringIndex(tplErr.Description)
				if tokenLoc != nil {
					token := string(tplErr.Description[tokenLoc[0]+1 : tokenLoc[1]-1])
					lastChar := strings.LastIndex(lines[tplErr.Line], token)
					firstChar := strings.Index(lines[tplErr.Line], token)
					// if it's not the only match, we don't know which character is the one the error occured on
					if lastChar == firstChar {
						tplErr.Char = firstChar
					}
				}
			}

			badFunctionMatch := functionNotFoundRegex.FindStringSubmatch(tplErr.Description)
			if badFunctionMatch != nil {
				token := badFunctionMatch[1]
				t, parseTplErrs := parseInternal(text, baseTpl.Funcs(textTemplate.FuncMap{
					token: func() error {
						return nil
					},
				}), depth+1)
				return t, append(tplErrs, parseTplErrs...)
			}

			if missingValueForCommandRegex.MatchString(tplErr.Description) {
				matches := firstEmptyCommandRegex.FindStringSubmatch(text)
				if matches != nil {
					line := SplitLines(text)[tplErr.Line]
					indexes := firstEmptyCommandRegex.FindStringIndex(line)
					if indexes != nil {
						tplErr.Char = indexes[0]
					}
					replacement := fmt.Sprintf(fmt.Sprintf("%%%ds", len(matches[0])), "")
					t, parseTplErrs := parseInternal(
						strings.Replace(text, matches[0], replacement, 1),
						baseTpl,
						depth+1,
					)
					return t, append(tplErrs, parseTplErrs...)
				}
			}
		}

		return baseTpl, tplErrs
	}
	return t, tplErrs
}

func exec(t *textTemplate.Template, data interface{}, buf *bytes.Buffer) []templateError {
	tplErrs := make([]templateError, 0)
	err := t.Execute(buf, data)
	if err != nil {
		if err.Error() == fmt.Sprintf(`template: %s: "%s" is an incomplete or empty template`, t.Name(), t.Name()) {
			return tplErrs
		}
		tplErr := createTemplateError(err, execErrorLevel)
		matches := findExprRegex.FindStringSubmatch(tplErr.Description)
		if len(matches) == 2 {
			value := matches[1]
			fmt.Println(value)
		}
		tplErrs = append(tplErrs, tplErr)
	}
	return tplErrs
}
