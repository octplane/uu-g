package uu

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

var templates map[string]*template.Template = make(map[string]*template.Template)
var defaultLayoutName string = "layout"

var DefaultTemplateFolder = "./views/"

func Yield(tmpl string) template.HTML {
	val, _ := raw_tmpl(tmpl, nil)
	return template.HTML(val)
}

type MissingTemplateError struct {
	Identifier string
}

func (f MissingTemplateError) Error() string {
	return fmt.Sprintf("uu: unable to load view \"%s\"", f.Identifier)
}

func get_or_load(templateName string) (*template.Template, error) {
	tmpl, exists := templates[templateName]

	if debug.InDebug() || !exists {
		var b []byte

		sourcePath := filepath.Join(DefaultTemplateFolder, templateName+".tmpl")
		b, err := ioutil.ReadFile(sourcePath)
		if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
			return nil, &MissingTemplateError{templateName}
		}

		if err != nil {
			fmt.Printf("Error while Loading view %s: %v\n", templateName, err)
			return nil, err
		}
		fMap := template.FuncMap{"yield": Yield}
		tmpl = template.Must(template.New(templateName).Funcs(fMap).Parse(string(b)))
		templates[templateName] = tmpl
	}

	return tmpl, nil
}

func raw_tmpl(templateName string, context map[string]interface{}) (string, error) {
	tmpl, err := get_or_load(templateName)
	if err != nil {
		return "", err
	}

	var templated bytes.Buffer
	err = tmpl.Execute(&templated, context)
	if err != nil {
		return "", err
	}

	return templated.String(), nil
}

func tmpl_with_layout(layoutName string, templateName string, context map[string]interface{}) (string, error) {
	// render content
	content, err := raw_tmpl(templateName, context)
	if err != nil {
		return "", err
	}
	context["content"] = template.HTML(content)

	output, err := raw_tmpl(layoutName, context)
	if err != nil {
		return "", err
	}
	return output, nil
}

func tmpl(templateName string, context map[string]interface{}) (string, error) {
	return tmpl_with_layout(defaultLayoutName, templateName, context)
}
