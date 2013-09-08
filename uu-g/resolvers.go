package main

import (
	"container/list"
	"fmt"
	"github.com/octplane/mnemo"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FsResolver struct {
	baseFolder    string
	baseExtension string
}

type PasteResolver struct{}

func (pr PasteResolver) GetFilename(identifier string) string {
	return "pastes/" + identifier + ".uu"
}

type AttachmentResolver struct{}

func (at *AttachmentResolver) GetFilename(identifier string) string {
	return "attn/" + identifier + ".data"
}

func (at *FsResolver) GetNextIdentifier() (fname string, mnem string) {
	return at.GetNextIdentifierWithPrefix("")
}

func (at *FsResolver) GetNextIdentifierWithPrefix(prefix string) (fname string, mnem string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := r.Int()
	for {
		basename := mnemo.FromInteger(id&0xFFFFFF) + "-" + prefix
		inc := 1
		_, err := os.Stat(at.GetFilename(basename))
		if err != nil && os.IsNotExist(err) {
			return at.GetFilename(basename), basename
		}
		id += inc
		inc *= 2
	}

	return "", ""
}

func (at *FsResolver) GetFilename(identifier string) string {
	return at.baseFolder + identifier + at.baseExtension
}

func (at *FsResolver) Cleanup() {
	ds := DataScanner{at.baseFolder, list.New()}
	ds.Scan()
	for e := ds.Items.Front(); e != nil; e = e.Next() {
		fmt.Printf("- %s\n", e.Value)
	}
}

type DataScanner struct {
	root  string
	Items *list.List
}

func (d DataScanner) visit(path string, info os.FileInfo, err error) error {
	baseName := strings.TrimPrefix(path, d.root)
	// Keep only first level directories
	if info.Mode().IsRegular() && len(baseName) > 0 {
		d.Items.PushBack(baseName)
	}
	return nil
}

func (d DataScanner) Scan() {
	_ = filepath.Walk(d.root, d.visit)
}
