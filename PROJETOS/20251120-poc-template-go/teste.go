package main

import (
	"bytes"
	"fmt"
	"text/template"
)

type Result struct {
	Major string
}

type Input struct {
	Version string
}

func testTemplate(templ string) string {
	var buf bytes.Buffer

	tmpl := template.Must(template.New("teste").Funcs(
		template.FuncMap{
			"semver": func(args string) Result {
				return Result{Major: "2"}
			},
		},
	).Parse(templ))
	tmpl.Execute(&buf, Input{Version: "2.0.0"})
	return buf.String()
}

func main() {
	fmt.Printf("Result no space: '%s'\n", testTemplate(`{{(semver .Version).Major}}`))
	fmt.Printf("Result space: '%s'\n", testTemplate(`{{(semver .Version) .Major}}`))
}
