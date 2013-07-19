package main

import (
	"github.com/octplane/mnemo"
	"math/rand"
	"os"
	"time"
)

type FsResolver interface {
	GetFilename(identifier string) string
}

type PasteResolver struct{}

func (pr PasteResolver) GetFilename(identifier string) string {
	return "pastes/" + identifier + ".uu"
}

type AttnResolver struct{}

func (at *AttnResolver) GetFilename(identifier string) string {
	return "attn/" + identifier + ".data"
}

func getNextIdentifier(resolver FsResolver) (fname string, mnem string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := r.Int()
	for {
		basename := mnemo.FromInteger(id & 0xFFFF)
		inc := 1
		_, err := os.Stat(resolver.GetFilename(basename))
		if err != nil && os.IsNotExist(err) {
			return resolver.GetFilename(basename), basename
		}
		id += inc
		inc *= 2
	}

	return "", ""
}