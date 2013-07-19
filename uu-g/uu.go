package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hoisie/web"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var globals = struct {
	pasteResolver *PasteResolver
	attnResolver  *AttnResolver
}{}

func init() {
	globals.pasteResolver = &PasteResolver{}
	globals.attnResolver = &AttnResolver{}
}

const debug debugging = false // or flip to false

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

func makeExpiryFromPost(expiry_key string, never bool) string {

	if never {
		return "-1"
	}
	for _, exp := range expiries {
		if expiry_key == exp.Name {
			return strconv.FormatInt(time.Now().Add(time.Duration(exp.duration)*time.Second).Unix(), 10)
		}
	}
	panic(fmt.Sprintf("Unknown duration \"%s\"", expiry_key))
}

func expiryStringFromTime(when int64) string {
	if when == -1 {
		return "never"
	}
	expire := time.Unix(when, 0)
	rest := int64(expire.Sub(time.Now()) / time.Second)
	if rest > 86400*2 {
		return fmt.Sprintf("in %d days", rest/86400)
	}
	if rest > 3600*2 {
		return fmt.Sprintf("in %d hours", rest/3600)
	}
	return fmt.Sprintf("in %d minutes", rest/60)
}

func savePost(id int, params map[string]string) string {
	fname, mnem := getNextIdentifier(globals.pasteResolver)
	file, err := os.OpenFile(fname, os.O_EXCL|os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	var count int
	var data []byte

	var paste = make(map[string]interface{})

	paste["content"] = params["content"]
	paste["attachments"] = params["attachements"]
	paste["expire"] = makeExpiryFromPost(params["expiry_delay"], params["never_expire"] == "true")

	data, err = json.Marshal(paste)
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
	return mnem
}

func loadPost(basename string) (map[string]string, error) {
	fname := globals.pasteResolver.GetFilename(basename)
	content, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	var data map[string]string
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}
	return data, nil

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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := r.Int() & 0xFFFFFFFF

	fname := savePost(id, ctxt.Params)

	ctxt.WriteString(fmt.Sprintf("/v/%s", fname))
}

func fileHandler(ctxt *web.Context) {
	file, header, _ := ctxt.Request.FormFile("file")
	fmt.Printf("%v %v", file, header)
}

func viewHandler(ctx *web.Context, basename string) {
	data, err := loadPost(basename)
	if err != nil {
		panic(err)
	}
	// Main Router
	var buf bytes.Buffer
	var scope = make(map[string]interface{})
	scope["encrypted_content"] = template.HTML(data["content"])
	scope["attachments"] = template.HTML(data["attachments"])
	expire, _ := strconv.ParseInt(data["expire"], 10, 64)
	scope["never"] = false
	if expire == -1 {
		scope["never"] = true
	}
	scope["expire"] = expiryStringFromTime(expire)

	output := tmpl("index", scope)
	buf.WriteString(output)
	io.Copy(ctx, &buf)
}

func main() {
	web.Config.StaticDir = "data"

	var hostAndPort = flag.String("l", ":8080", "IP and port to listen to")

	flag.Parse()
	web.Get("/", slashHandler)
	web.Post("/paste", postHandler)
	web.Post("/file-upload", fileHandler)
	web.Get("/v/(.*)", viewHandler)
	web.Run(*hostAndPort)
}
