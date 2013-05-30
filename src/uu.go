package main

import (
	"fmt"
	"github.com/realistschuckle/gohaml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const debug debugging = true // or flip to false

type debugging bool

type content []byte

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

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if !publicHandler(w, r) {
		// Main Router
		if r.URL.Path == "/" {
			var scope = make(map[string]interface{})
			scope["code"] = ""
			scope["snippet"] = "Copie Priv&eacute;e is a new kind of paste website. It will try to auto-detect the language you're pasting."
			content, err := ioutil.ReadFile("views/index.haml")
			if err == nil {
				engine, _ := gohaml.NewEngine(string(content))
				output := engine.Render(scope)
				fmt.Fprintf(w, output) // Prints "I love HAML!"
			} else {
				log.Fatal(err)
			}
		}
	}
}

func publicHandler(w http.ResponseWriter, r *http.Request) bool {
	path := "data" + r.URL.Path
	debug.Print(path)
	_, err := os.Stat(path)
	if err == nil {
		body, err := ioutil.ReadFile(path)
		if err != nil {
			return false
		}
		w.Write(body)
		return true
	}
	return false
}

func main() {
	http.HandleFunc("/", mainHandler)
	http.ListenAndServe(":8080", nil)
}
