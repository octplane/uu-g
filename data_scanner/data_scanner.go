package main

import (
	"flag"
	"fmt"
	"github.com/jteeuwen/go-bindata/lib"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type DataScanner struct {
	root string
}

func (d DataScanner) visit(fullpath string, info os.FileInfo, err error) error {
	rootAbs, _ := filepath.Abs(d.root)

	baseName := strings.TrimPrefix(fullpath, d.root)
	// Keep only first level directories
	if info.IsDir() && len(baseName) > 0 && !strings.Contains(baseName, "/") {
		fmt.Printf("%s is a directory\n", baseName)
	} else if info.Mode().IsRegular() {
		in, _ := filepath.Abs(fullpath)
		fs, err := os.Open(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s\n", err)
			return err
		}
		out, _ := filepath.Abs(path.Join(d.root, "../bindata", baseName+".go"))

		// Create missing folder if needed
		dir, _ := filepath.Split(out)
		_, err = os.Lstat(dir)
		if err != nil {
			os.MkdirAll(dir, 0755)
		}

		fd, err := os.Create(out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s\n", err)
			return err
		}
		// Translate binary to Go code.
		funcname := bindata.SafeFuncname(baseName, "uu")
		bindata.Translate(fs, fd, "uu", funcname, false, false)
		fs.Close()
		tocRoot, _ := filepath.Abs(path.Join(d.root, "../bindata"))
		err = bindata.CreateTOC(tocRoot, "uu")

		if err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s\n", err)
			return err
		}

		bindata.WriteTOCInit(fd, in, rootAbs, funcname)
		fd.Close()

	}
	return nil
}

func (d DataScanner) Scan() {
	filepath.Walk(d.root, d.visit)

}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	d := DataScanner{root: root}
	d.Scan()
}
