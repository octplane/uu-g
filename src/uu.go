package main

import (
	"bytes"
	"flag"
	"github.com/octplane/web"
	"github.com/realistschuckle/gohaml"
	"html/template"
	"io"
	"io/ioutil"
	"log"
)

const debug debugging = true // or flip to false

type debugging bool

func (d debugging) Print(content string) {
	if d {
		log.Print(content)
	}
}

var templates map[string]*template.Template = make(map[string]*template.Template)
var DefaultLayout string = "layout"

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		log.Printf(format, args...)
	}
}

type TimeSpan struct {
	name     string
	duration int
	selected bool
}

func (t TimeSpan) HtmlProperties() map[string]interface{} {
	var ret = make(map[string]interface{})
	if t.selected {
		ret["selected"] = "selected"
	}
	return ret
}

func haml(source string, scope map[string]interface{}) string {
	return haml_with_layout(DefaultLayout, source, scope)
}

func haml_with_layout(layout string, source string, scope map[string]interface{}) string {
	// render content
	content, err := raw_haml(source, scope)
	if err != nil {
		panic(err)
	}
	scope["content"] = content

	output, err := raw_haml(layout, scope)
	if err != nil {
		panic(err)
	}
	return output
}

func haml_helper(source string) (string, error) {
	return raw_haml(source, nil)
}

func raw_haml(source string, scope map[string]interface{}) (string, error) {
	var tmpl *template.Template
	var err error
	tmpl, exists := templates[source]

	if !exists {
		var b []byte
		b, err = ioutil.ReadFile("views/" + source + ".template")
		debug.Printf("Reading %s from disk", source)
		tmpl := template.New(string(b))
		fMap := template.FuncMap{"haml": haml_helper}
		tmpl.Funcs(fMap)
		templates[source] = tmpl
	}

	var templated bytes.Buffer

	err = tmpl.Execute(&templated, scope)
	if err != nil {
		return "", err
	}

	engine, err := gohaml.NewEngine(templated.String())
	if err != nil {
		return "", err
	}

	output := engine.Render(scope)
	return output, nil
}

func slashHandler(ctxt *web.Context) {
	// Main Router
	var buf bytes.Buffer
	var scope = make(map[string]interface{})
	scope["code"] = ""
	scope["snippet"] = "Copie Priv&eacute;e is a new kind of paste website. It will try to auto-detect the language you're pasting."
	output := haml("index", scope)
	buf.WriteString(output)
	io.Copy(ctxt, &buf)
}

func main() {
	web.Config.StaticDir = "data"

	var hostAndPort = flag.String("l", ":8080", "IP and port to listen to")

	flag.Parse()
	web.Get("/", slashHandler)
	web.Run(*hostAndPort)
}
