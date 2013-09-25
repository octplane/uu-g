package main

import (
	"flag"
	"github.com/hoisie/web"
	"github.com/octplane/uu-g/lib"
)

func main() {
	web.Config.StaticDir = "__public"
	uu.DefaultTemplateFolder = "./__views"

	hostAndPort := flag.String("l", ":8080", "IP and port to listen to")
	dataPath := flag.String("d", ".", "Relative path to attachments and pastes")

	flag.Parse()

	uu.InitResolvers(*dataPath)
	web.Get("/", uu.SlashHandler)
	web.Post("/paste", uu.PostHandler)
	web.Post("/file-upload", uu.FileHandler)
	web.Get("/p/(.*)", uu.ViewHandler)
	web.Get("/a/(.*)", uu.AttachmentHandler)
	web.Get("/(.*)", uu.AnyHandler)
	web.Run(*hostAndPort)
}
