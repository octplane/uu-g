package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hoisie/web"
	"github.com/octplane/mnemo"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
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

func savePost(id int, params map[string]string) string {
	basename := mnemo.FromInteger(id)
	fname := basename + ".uu"
	file, err := os.OpenFile(fname, os.O_EXCL|os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	var count int
	var data []byte

	data, err = json.Marshal(params)
	if err != nil {
		panic(err)
	}

	count, err = file.Write(data)
	if err != nil {
		panic(err)
	}
	if count != len(data) {
		panic(fmt.Sprintf("Wrote only %d/%d in %s", count, len(data), fname))
	}

	file.Close()
	return basename
}

func loadPost(basename string) (map[string]string, error) {
	fname := basename + ".uu"
	content, err := ioutil.ReadFile(basename)
	if err != nil {
		return nil, err
	}
	var data map[string]string
	data = json.Unmarshal(content, data)
	return data

}

func postHandler(ctxt *web.Context) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := r.Int() & 0xFFFFFFFF

	fname := savePost(id, ctxt.Params)

	ctxt.WriteString(fmt.Sprintf("/v/%s", fname))
}

func viewHandler(ctx *web.Context, basename string) {
	data, err := loadPost(basename)
	ctx.Write(data)
}

func main() {
	web.Config.StaticDir = "data"

	var hostAndPort = flag.String("l", ":8080", "IP and port to listen to")

	flag.Parse()
	web.Get("/", slashHandler)
	web.Post("/paste", postHandler)
	web.Get("/v/(.*)", viewHandler)
	web.Run(*hostAndPort)
}
