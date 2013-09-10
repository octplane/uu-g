package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

var templates map[string]*template.Template = make(map[string]*template.Template)
var defaultLayoutName string = "layout"

func Yield(tmpl string) template.HTML {
	val, _ := raw_tmpl(tmpl, nil)
	return template.HTML(val)
}

func get_or_load(templateName string) *template.Template {
	tmpl, exists := templates[templateName]

	if debug.InDebug() || !exists {
		var b []byte
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		sourcePath := filepath.Join(cwd, "./views/"+templateName+".tmpl")
		b, err = ioutil.ReadFile(sourcePath)
		if err != nil {
			panic(err)
		}
		fMap := template.FuncMap{"yield": Yield}
		tmpl = template.Must(template.New(templateName).Funcs(fMap).Parse(string(b)))
		templates[templateName] = tmpl
	}

	return tmpl
}

func raw_tmpl(templateName string, context map[string]interface{}) (string, error) {
	tmpl := get_or_load(templateName)
	var templated bytes.Buffer
	err := tmpl.Execute(&templated, context)
	if err != nil {
		return "", err
	}

	return templated.String(), nil
}

func tmpl_with_layout(layoutName string, templateName string, context map[string]interface{}) string {
	// render content
	content, err := raw_tmpl(templateName, context)
	if err != nil {
		panic(err)
	}
	context["content"] = template.HTML(content)

	output, err := raw_tmpl(layoutName, context)
	if err != nil {
		panic(err)
	}
	return output
}

func tmpl(templateName string, context map[string]interface{}) string {
	return tmpl_with_layout(defaultLayoutName, templateName, context)
}
