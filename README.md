# Go template validation

When working with the [`"text/template"`](https://golang.org/pkg/text/template/) and [`"html/template"`](https://golang.org/pkg/html/template/) packages, I often have a hard time understanding go's errors, especially when they're inline in code. This is a simple tool to visually show where validation errors are happening.

To use, choose a file or insert your template code directly. You can add mock data in the form of JSON.

## Features

* Show errors at the relavent line/character
* Recovery from unknown function errors
* Recovery from missing value for command errors
* Some auto-handling of required data
* Discover character position of misunderstood tokens

## Goals

* Find as many issues as possible. The default package bails out at the first error (which makes sense at runtime), but often you'll fix one error only to have to track down the next.
* (at this point) Don't rewrite/maintain a fork of [`"text/template/parse"`](https://golang.org/pkg/text/template/parse/)
