package main

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/octplane/mnemo"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Resolver interface {
	Cleanup()
	GetNextIdentifier() (fname string, mnem string)
	GetNextIdentifierWithPrefix(prefix string) (fname string, mnem string)
	GetFilename(identifier string) string
	LoadItem(identifier string) (map[string]string, error)
}

type ExpireChecker interface {
	HasExpired(identifier string, res Resolver) bool
	LoadItem(identifier string, res Resolver) (map[string]string, error)
}

type FsResolver struct {
	baseFolder    string
	baseExtension string
	expireChecker ExpireChecker
}

type PasteResolver struct {
	FsResolver
}

type AttachmentResolver struct {
	FsResolver
}

type PasteChecker struct{}

type AttachmentChecker struct{}

func (at *AttachmentChecker) HasExpired(filename string, res Resolver) bool {
	return false
}

func (pc *AttachmentChecker) LoadItem(identifier string, res Resolver) (map[string]string, error) {
	return nil, errors.New("Not implemented")
}

func (pc *PasteChecker) HasExpired(identifier string, res Resolver) bool {
	content, err := pc.LoadItem(identifier, res)
	if err != nil {
		fmt.Printf("[EMERG] Error while loading paste %s:\n", identifier)
		fmt.Printf("[EMERG] %v\n", err)
		// Keep under the elbow for a while
		return false
	}
	if content["expire"] != "-1" {
		expire, _ := strconv.ParseInt(content["expire"], 10, 64)
		if time.Unix(expire, 0).Before(time.Now()) {
			return true
		}
	}
	return false
}

func (res *FsResolver) LoadItem(identifier string) (map[string]string, error) {
	return res.expireChecker.LoadItem(identifier, res)
}

func (pc *PasteChecker) LoadItem(identifier string, res Resolver) (map[string]string, error) {
	fname := res.GetFilename(identifier)
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

func (at *FsResolver) GetNextIdentifier() (fname string, mnem string) {
	return at.GetNextIdentifierWithPrefix("")
}

func (at *FsResolver) GetNextIdentifierWithPrefix(prefix string) (fname string, mnem string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := r.Int()
	for {
		basename := mnemo.FromInteger(id & 0xFFFFFF)
		if prefix != "" {
			basename += "-" + prefix
		}
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
		identifier, _ := e.Value.(string)
		if at.expireChecker.HasExpired(identifier, at) {

			fmt.Printf("[DEL] %s has expired\n", at.GetFilename(identifier))
			syscall.Unlink(at.GetFilename(identifier))
		}
	}
}

type DataScanner struct {
	root  string
	Items *list.List
}

func (d DataScanner) visit(pth string, info os.FileInfo, err error) error {
	baseName := strings.TrimPrefix(pth, d.root)
	// Keep only first level directories
	if info.Mode().IsRegular() && len(baseName) > 0 {
		d.Items.PushBack(baseName[0 : len(baseName)-len(path.Ext(baseName))])
	}
	return nil
}

func (d DataScanner) Scan() {
	_ = filepath.Walk(d.root, d.visit)
}
