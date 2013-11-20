package uu

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

type resolvers struct {
	pasteResolver *PasteResolver
	attnResolver  *AttachmentResolver
}

func (r *resolvers) cleanup() {
	fmt.Print("[DEL] Waking up\n")
	r.pasteResolver.cleanup()
	r.attnResolver.cleanup()
}

var res = resolvers{}

func InitResolvers(baseFolder string) {
	fmt.Printf("internal is %v\n", internalAssets)
	res.pasteResolver = &PasteResolver{FsResolver{path.Join(baseFolder, "/pastes/"), ".uu", &PasteChecker{}}}
	res.attnResolver = &AttachmentResolver{FsResolver{path.Join(baseFolder, "/attn/"), ".data", &AttachmentChecker{}}}

	go func() {
		for true {
			res.cleanup()
			time.Sleep(60 * time.Second)
		}
	}()
}

func init() {
}

type Resolver interface {
	cleanup()
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
	dashPos := strings.Index(filename, "-")
	if dashPos == -1 {
		fmt.Printf("[ERR] Unable to find timestamp in attachment %s\n", filename)
		return false
	}
	expire, err := strconv.ParseInt(filename[dashPos+1:len(filename)], 10, 64)
	if err != nil {
		fmt.Printf("[ERR] error while parsing timestamp %s for %s\n", expire, filename)
		fmt.Print(err)
	}
	if expire == -1 {
		return false
	}
	if time.Unix(expire, 0).Before(time.Now()) {
		return true
	}
	fmt.Printf("expire is %d\n", expire)
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

type MissingPasteError struct {
	identifier string
}

func (f MissingPasteError) Error() string {
	return fmt.Sprintf("[ERR] Unable to load paste \"%s\"", f.identifier)
}

func (pc *PasteChecker) LoadItem(identifier string, res Resolver) (map[string]string, error) {
	fname := res.GetFilename(identifier)
	content, err := ioutil.ReadFile(fname)
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		return nil, &MissingPasteError{identifier}
	}
	if err != nil {
		fmt.Printf("[ERR] Unable to load paste %v\n", err)
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
		basename := mnemo.FromInteger(id & 0xFFFF)
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
	return path.Join(at.baseFolder, identifier+at.baseExtension)
}

func (at *FsResolver) cleanup() {
	ds := DataScanner{at.baseFolder, list.New()}
	ds.Scan()
	for e := ds.Items.Front(); e != nil; e = e.Next() {
		identifier, _ := e.Value.(string)
		debug.Printf("[INF] Checking %s\n", identifier)
		if at.expireChecker.HasExpired(identifier, at) {

			fmt.Printf("[DEL] %s has expired, deleting\n", at.GetFilename(identifier))
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
	debug.Printf("[INF] Scanning %s\n", d.root)
	_ = filepath.Walk(d.root, d.visit)
}
