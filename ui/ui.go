// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/go/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ui

import (
	"context"
	"io"
	"net/http"
)

// UI Package Architecture

// Process
// 1. Intake user html template
// 2. Build CSS
//   - parse html classes from template
//   - build and return css from classes

// style, component, page, template
// <ComponentName prop1="value" prop2="value" class="Class1 Class2" />
// how to push data to component

// The ui package is responsible for the user interface of the application. It contains the following sub-packages:
// - css: contains the color palette for the user interface
// - components: contains the reusable components for the user interface
// - pages: contains the different pages of the application
// - templates: contains the HTML templates for the different pages of the application
// - ui.go: contains the main function for the user interface

/*
	package main

	import (
		"net/http"
		"github.com/jcdotter/go/ui"
	)

	func main() {
		http.Handle("/", ui.Page(...))
		http.ListenAndServe(":3000", nil)
	}

	ui.El(`div`, `id="div-001" class="container mx-auto" visible`).Inner(
		ui.El(`p`, "text-center").Inner(
			ui.Text("Hello, World!"),
		),
	)


*/

// TODO:
// build out components
// capabilities for custom components
// add css rendering
// page rendering
// think about how to include js

/* type class interface {
	Name() string
	Render(ctx context.Context, w io.Writer) error
} */

type Component interface {
	Render(ctx context.Context, w io.Writer) error
}

type Page interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// ----------------------------------------------------------------------------
// Implementation of raw text

type text string

func (t text) Render(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(t))
	return err
}

// Text creates a new HTML text.
func Text(contents string) text {
	return text(contents)
}

// Comment creates a new HTML comment.
func Comment(contents string) text {
	return text("<!-- " + contents + " -->")
}
