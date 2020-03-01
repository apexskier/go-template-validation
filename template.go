package main

import (
	"regexp"
	"strconv"
	"strings"
	textTemplate "text/template"
)

var (
	templateParseErrorRegex = regexp.MustCompile(`template: (.*):(\d+): (.*)`)
	templateExecErrorRegex  = regexp.MustCompile(`template: (.*):(\d+):(\d+): (.*)`)
	findTokenRegex          = regexp.MustCompile(`"(.+)"`)
	functionNotFoundRegex   = regexp.MustCompile(`function "(.+)" not defined`)
)

func parse(text string, baseTpl *textTemplate.Template, depth int) (*textTemplate.Template, []templateError) {
	lines := strings.Split(strings.Replace(text, "\r\n", "\n", -1), "\n")
	tplErrs := make([]templateError, 0)

	if depth > 10 {
		return baseTpl, tplErrs
	}

	t, err := baseTpl.Parse(text)
	if err != nil {
		errStr := err.Error()
		matches := templateParseErrorRegex.FindStringSubmatch(errStr)
		if len(matches) == 4 {
			description := matches[3]
			line, err := strconv.Atoi(matches[2])
			if err != nil {
				line = -1
			} else {
				line = line - 1
			}
			char := -1
			// try to find a character to line up with
			var token string
			tokenLoc := findTokenRegex.FindStringIndex(description)
			if tokenLoc != nil {
				token = string(description[tokenLoc[0]+1 : tokenLoc[1]-1])
				lastChar := strings.LastIndex(lines[line], token)
				firstChar := strings.Index(lines[line], token)
				// if it's not the only match, we don't know which character is the one the error occured on
				if lastChar == firstChar {
					char = firstChar
				}
			}
			tplErrs = append(tplErrs, templateError{
				Line:        line,
				Char:        char,
				Description: description,
				Level:       parseErrorLevel,
			})
			isBadFunction := functionNotFoundRegex.MatchString(description)
			if isBadFunction {
				t, parseTplErrs := parse(text, baseTpl.Funcs(textTemplate.FuncMap{
					token: func() error {
						return nil
					},
				}), depth+1)
				return t, append(tplErrs, parseTplErrs...)
			}
		} else {
			tplErrs = append(tplErrs, templateError{
				Line:        -1,
				Char:        -1,
				Description: errStr,
				Level:       misunderstoodError,
			})
		}
		return baseTpl, tplErrs
	}
	return t, tplErrs
}
