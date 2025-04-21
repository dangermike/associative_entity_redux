package load

import (
	"embed"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

type InsertNameSpec struct {
	Table   string
	StartIx int
	Names   []string
}

type InsertAssocSpec struct {
	Table  string
	Assocs [][2]int
}

var templates = func() *template.Template {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"mod": func(a, b int) int {
			return a % b
		},
		"clean": func(a string) string {
			return strings.Replace(a, "'", "\\'", -1)
		},
	}

	templates, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.gtmpl")
	if err != nil {
		panic(err)
	}
	return templates
}()
