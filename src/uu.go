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

func slashHandler(ctxt *web.Context) {
	// Main Router
	var buf bytes.Buffer
	var scope = make(map[string]interface{})
	scope["code"] = ""
	scope["snippet"] = "Copie Priv&eacute;e is a new kind of paste website. It will try to auto-detect the language you're pasting."
	content, err := ioutil.ReadFile("views/index.haml.template")
	if err == nil {
		tmpl, err := template.New("layout").Parse(string(content))
		if err != nil {
			panic(err)
		}
		var templated bytes.Buffer
		err = tmpl.Execute(&templated, scope)
		if err != nil {
			panic(err)
		}
		engine, _ := gohaml.NewEngine(templated.String())
		output := engine.Render(scope)
		buf.WriteString(output)
	} else {
		log.Fatal(err)
	}
	io.Copy(ctxt, &buf)
}

func main() {
	web.Config.StaticDir = "data"

	var hostAndPort = flag.String("-listen", ":8080", "IP and port to listen to")
	flag.Parse()

	web.Get("/", slashHandler)
	web.Run(*hostAndPort)
}
