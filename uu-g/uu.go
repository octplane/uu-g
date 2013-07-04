package main

import (
	"bytes"
	"flag"
	"github.com/hoisie/web"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const debug debugging = true // or flip to false

type debugging bool

func (d debugging) Print(content string) {
	if d {
		log.Print(content)
	}
}

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		log.Printf(format, args...)
	}
}

func (d debugging) InDebug() bool {
	if d {
		return true
	}
	return false
}

var templates map[string]*template.Template = make(map[string]*template.Template)
var defaultLayoutName string = "layout"

type TimeSpan struct {
	Name     string
	duration int
	selected bool
}

func (t TimeSpan) SelectedAttribute() template.HTML {
	if t.selected {
		return template.HTML("selected")
	} else {
		return template.HTML("")
	}

}

func Yield(tmpl string) template.HTML {
	val, _ := raw_tmpl(tmpl, nil)
	return template.HTML(val)
}

var expiries = [...]TimeSpan{TimeSpan{"30 min", 1800, true},
	TimeSpan{"1 day", 86400, false}}

func get_or_load(templateName string) *template.Template {
	tmpl, exists := templates[templateName]

	if debug.InDebug() || !exists {
		var b []byte
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		sourcePath := filepath.Join(cwd, "./views/"+templateName+".tmpl")
		debug.Printf("Reading %s from disk", sourcePath)
		b, err = ioutil.ReadFile(sourcePath)
		if err != nil {
			panic(err)
		}
		fMap := template.FuncMap{"yield": Yield}
		tmpl = template.Must(template.New(templateName).Funcs(fMap).Parse(string(b)))
		debug.Printf("Template %s is ready %x", templateName, tmpl)
		templates[templateName] = tmpl
	}
	debug.Printf("Returning: %s -> %x\n", templateName, tmpl)

	return tmpl
}

func raw_tmpl(templateName string, context map[string]interface{}) (string, error) {
	tmpl := get_or_load(templateName)
	var templated bytes.Buffer
	debug.Printf("Applying template %X to context %v\n", tmpl, context)
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

func slashHandler(ctxt *web.Context) {
	// Main Router
	var buf bytes.Buffer
	var scope = make(map[string]interface{})
	scope["code"] = ""
	scope["snippet"] = "Copie Priv&eacute;e is a new kind of paste website. It will try to auto-detect the language you're pasting."
	scope["expiries"] = expiries
	output := tmpl("index", scope)
	buf.WriteString(output)
	io.Copy(ctxt, &buf)
}

func postHandler(ctxt *web.Context) {
	println("Post !")
	for k, v := range ctxt.Params {
		println(k, v)
	}
}

func main() {
	web.Config.StaticDir = "data"

	var hostAndPort = flag.String("l", ":8080", "IP and port to listen to")

	flag.Parse()
	web.Get("/", slashHandler)
	web.Post("/paste", postHandler)
	web.Run(*hostAndPort)
}
