package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"log"
	"net/http"
	"strings"
	textTemplate "text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	port = 8080
)

type ErrorLevel string

const (
	misunderstoodError ErrorLevel = "misunderstood"
	parseErrorLevel    ErrorLevel = "parse"
	execErrorLevel     ErrorLevel = "exec"
)

type templateError struct {
	Line        int
	Char        int
	Description string
	Level       ErrorLevel
}
type indexData struct {
	RawText        string
	RawData        string
	RawFunctions   string
	TextLines      []string
	Output         string
	Errors         []templateError
	LineNumSpacing int
}

func getText(r *http.Request) (string, error) {
	file, _, err := r.FormFile("from-file")
	if err != nil {
		return r.FormValue("from-raw-text"), nil
	}
	defer file.Close()
	var buf bytes.Buffer
	defer buf.Reset()
	io.Copy(&buf, file)
	return buf.String(), nil
}

func main() {
	funcs := htmlTemplate.FuncMap{
		"intRange": func(start, end int) []int {
			n := end - start + 1
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = start + i
			}
			return result
		},
		"nl": func() string {
			return "\n"
		},
	}
	indexTemplate, err := htmlTemplate.New("index.html").Funcs(funcs).ParseFiles("index.html")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		maxRequestSize := int64(32 << 20)
		r.ParseMultipartForm(maxRequestSize)

		tplErrs := []templateError{}

		text, err := getText(r)
		if err == http.ErrMissingFile {
			tplErrs = append(tplErrs, templateError{
				Line:        -1,
				Char:        -1,
				Description: "couldn't accept file",
			})
		} else if err != nil {
			panic(err)
		}

		var data interface{}
		rawData := r.FormValue("data")
		err = json.Unmarshal([]byte(rawData), &data)
		if err != nil {
			tplErrs = append(tplErrs, templateError{
				Line:        -1,
				Char:        -1,
				Description: fmt.Sprintf("failed to understand data: %v", err),
				Level:       misunderstoodError,
			})
		}

		t := textTemplate.New("")

		// mock template functions - this'll happen automatically as they're found, but errors will be output and there's a max limit
		rawFunctions := r.FormValue("functions")
		var functions []string
		if rawFunctions != "" {
			functions = strings.Split(rawFunctions, ",")
		}
		for _, function := range functions {
			functionName := strings.TrimSpace(function)
			// wrap in func so we can catch panics on bad function names
			func() {
				defer func() {
					if r := recover(); r != nil {
						tplErrs = append(tplErrs, templateError{
							Line:        -1,
							Char:        -1,
							Description: fmt.Sprintf(`bad function name provided: "%s"`, functionName),
							Level:       misunderstoodError,
						})
					}
				}()
				t = t.Funcs(textTemplate.FuncMap{functionName: func() error { return nil }})
			}()
		}

		t, parseTplErrs := parse(text, t, 0)
		tplErrs = append(tplErrs, parseTplErrs...)

		var buf bytes.Buffer
		defer buf.Reset()
		execTplErrs := exec(t, data, &buf)
		tplErrs = append(tplErrs, execTplErrs...)

		// outputs html into the textarea, so chrome gets worried
		// https://stackoverflow.com/a/17815577/2178159
		w.Header().Add("X-XSS-Protection", "0")

		lines := SplitLines(text)
		indexTemplate.Execute(w, indexData{
			RawText:        text,
			RawData:        rawData,
			RawFunctions:   rawFunctions,
			Output:         buf.String(),
			Errors:         tplErrs,
			TextLines:      lines,
			LineNumSpacing: CountDigits(len(lines)),
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		indexTemplate.Execute(w, indexData{})
	})

	log.Printf("starting on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
