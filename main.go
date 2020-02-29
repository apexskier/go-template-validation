package main

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	textTemplate "text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	port = 8080
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

type templateError struct {
	Line        int
	Char        int
	Description string
}
type indexData struct {
	Text      string
	TextLines []string
	Errors    []templateError
}

func main() {
	indexTemplate, err := htmlTemplate.New("index.html").Funcs(htmlTemplate.FuncMap{
		"intRange": func(start, end int) []int {
			n := end - start + 1
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = start + i
			}
			return result
		},
	}).ParseFiles("index.html")
	if err != nil {
		panic(err)
	}
	templateParseErrorRegex, err := regexp.Compile(`template: :(\d+): (.*)`)
	if err != nil {
		panic(err)
	}
	templateExecErrorRegex, err := regexp.Compile(`template: :(\d+):(\d+): (.*)`)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		text := r.FormValue("from-raw-text")
		fromTextarea := true
		tplErrs := []templateError{}
		if text == "" {
			fromTextarea = false
			file, _, err := r.FormFile("from-file")
			if err == http.ErrMissingFile {
				tplErrs = append(tplErrs, templateError{
					Line:        -1,
					Char:        -1,
					Description: "couldn't accept file",
				})
			} else if err != nil {
				panic(err)
			} else {
				defer file.Close()
				var buf bytes.Buffer
				io.Copy(&buf, file)
				text = buf.String()
				defer buf.Reset()
			}
		}
		t, err := textTemplate.New("").Parse(text)
		if err != nil {
			errStr := err.Error()
			matches := templateParseErrorRegex.FindStringSubmatch(errStr)
			if len(matches) == 3 {
				line, err := strconv.Atoi(matches[1])
				if err != nil {
					line = -1
				}
				tplErrs = append(tplErrs, templateError{
					Line:        line - 1,
					Char:        -1,
					Description: matches[2],
				})
			}
		} else {
			var buff bytes.Buffer
			err := t.Execute(&buff, struct{}{})
			if err != nil {
				errStr := err.Error()
				matches := templateExecErrorRegex.FindStringSubmatch(errStr)
				if len(matches) == 4 {
					line, err := strconv.Atoi(matches[1])
					if err != nil {
						line = -1
					} else {
						line = line - 1
					}
					char, err := strconv.Atoi(matches[2])
					if err != nil {
						char = -1
					}
					fmt.Println(matches)
					tplErrs = append(tplErrs, templateError{
						Line:        line,
						Char:        char,
						Description: matches[3],
					})
				}
			}
		}
		lines := strings.Split(strings.Replace(text, "\r\n", "\n", -1), "\n")

		// outputs html into the textarea, so chrome gets worried
		// https://stackoverflow.com/a/17815577/2178159
		w.Header().Add("X-XSS-Protection", "0")
		var outputText string
		if fromTextarea {
			outputText = text
		}
		indexTemplate.Execute(w, indexData{
			Text:      outputText,
			Errors:    tplErrs,
			TextLines: lines,
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		indexTemplate.Execute(w, indexData{})
	})

	log.Printf("starting on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
