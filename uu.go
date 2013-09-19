package main

import (
	"flag"
	"github.com/hoisie/web"
	"github.com/octplane/uu-g/lib"
)

func main() {
	web.Config.StaticDir = "data"

	var hostAndPort = flag.String("l", ":8080", "IP and port to listen to")

	flag.Parse()
	web.Get("/", uu.SlashHandler)
	web.Post("/paste", uu.PostHandler)
	web.Post("/file-upload", uu.FileHandler)
	web.Get("/p/(.*)", uu.ViewHandler)
	web.Get("/a/(.*)", uu.AttachmentHandler)
	web.Run(*hostAndPort)
}
