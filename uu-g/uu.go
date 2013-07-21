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
	"mime/multipart"
	"os"
	"strconv"
)

var globals = struct {
	pasteResolver *PasteResolver
	attnResolver  *AttachmentResolver
}{}

func init() {
	globals.pasteResolver = &PasteResolver{}
	globals.attnResolver = &AttachmentResolver{}
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

func savePost(params map[string]string) string {
	fname, mnem := getNextIdentifier(globals.pasteResolver)
	file, err := os.OpenFile(fname, os.O_EXCL|os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var count int
	var data []byte

	var paste = make(map[string]interface{})

	paste["content"] = params["content"]
	paste["attachments"] = params["attachments"]
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

	return mnem
}

func saveAttachment(attn multipart.File) string {

	content, err := ioutil.ReadAll(attn)

	if err != nil {
		panic(err)
	}

	fname, mnem := getNextIdentifier(globals.attnResolver)
	file, err := os.OpenFile(fname, os.O_EXCL|os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var count int
	count, err = file.Write(content)
	if err != nil {
		panic(err)
	}
	if count != len(content) {
		panic(fmt.Sprintf("Wrote only %d/%d in %s", count, len(content), fname))
	}
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

	fname := savePost(ctxt.Params)
	ctxt.WriteString(fmt.Sprintf("/p/%s", fname))
}

func fileHandler(ctxt *web.Context) {
	file, _, _ := ctxt.Request.FormFile("file")
	attachment_mnem := saveAttachment(file)
	ctxt.WriteString(fmt.Sprintf("%s", attachment_mnem))
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
	web.Get("/p/(.*)", viewHandler)
	web.Run(*hostAndPort)
}
