package uu

import (
	"bytes"
	"fmt"
	"github.com/hoisie/web"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type resolvers struct {
	pasteResolver *PasteResolver
	attnResolver  *AttachmentResolver
}

func (r *resolvers) cleanup() {
	r.pasteResolver.Cleanup()
	r.attnResolver.Cleanup()
}

var res = resolvers{}

func init() {
	res.pasteResolver = &PasteResolver{FsResolver{"pastes/", ".uu", &PasteChecker{}}}
	res.attnResolver = &AttachmentResolver{FsResolver{"attn/", ".data", &AttachmentChecker{}}}
}

func fileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func SlashHandler(ctxt *web.Context) {
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

func PostHandler(ctxt *web.Context) {

	fname := savePost(ctxt.Params)
	ctxt.WriteString(fmt.Sprintf("/p/%s", fname))
}

func FileHandler(ctxt *web.Context) {
	file, info, _ := ctxt.Request.FormFile("file")
	debug.Printf("%s\n", ctxt.Request.FormValue("expiry_delay"))
	debug.Printf("%s\n", ctxt.Request.FormValue("never_expire"))

	expire := makeExpiryFromPost(ctxt.Request.FormValue("expiry_delay"), ctxt.Request.FormValue("never_expire") == "true")
	var ext string
	if strings.LastIndex(info.Filename, ".") == -1 {
		ext = ".data"
	} else {
		ext = info.Filename[strings.LastIndex(info.Filename, "."):]
	}
	attachment_mnem := saveAttachment(file, expire) + ext
	ctxt.WriteString(fmt.Sprintf("%s", attachment_mnem))
}

func AttachmentHandler(ctx *web.Context, attachmentName string) {
	var baseName = attachmentName[0 : len(attachmentName)-len(path.Ext(attachmentName))]
	staticFile := res.attnResolver.GetFilename(baseName)
	if fileExists(staticFile) {
		// Cargo culted from http.ServeContent
		sdir, file := filepath.Split(staticFile)
		dir := http.Dir(sdir)
		f, err := dir.Open(file)
		if err != nil {
			return
		}
		var buf [1024]byte
		n, _ := io.ReadFull(f, buf[:])
		b := buf[:n]
		ctype := http.DetectContentType(b)
		_, err = f.Seek(0, os.SEEK_SET) // rewind to output whole file
		if err != nil {
			http.Error(ctx, "seeker can't seek", http.StatusInternalServerError)
			return
		}
		ctx.Header().Set("Content-Type", ctype)

		http.ServeFile(ctx, ctx.Request, staticFile)
		return
	}
	ctx.NotFound(fmt.Sprintf("%s was not found.", baseName))
}

func ViewHandler(ctx *web.Context, basename string) {
	res.cleanup()

	data, err := res.pasteResolver.LoadItem(basename)
	if _, ok := err.(*MissingPasteError); ok {
		ctx.NotFound(fmt.Sprintf("%s is no longer available.", basename))
		return
	}
	if err != nil {
		ctx.Abort(500, err.Error())
		return
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
