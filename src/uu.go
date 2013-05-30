package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"log"
  "os"
//	"github.com/realistschuckle/gohaml"
)

const debug debugging = true // or flip to false

type debugging bool

type content [] byte

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



func mainHandler(w http.ResponseWriter, r *http.Request) {
  if !publicHandler(w, r) {
    // Main Router
    if r.URL.Path == "/" {
      fmt.Fprintf(w, "<h1>Home</h1>")
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


