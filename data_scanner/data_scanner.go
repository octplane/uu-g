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

func (d DataScanner) visit(relpath string, info os.FileInfo, err error) error {
	rootAbs, _ := filepath.Abs(d.root)

	fullpath, _ := filepath.Abs(relpath)
	basename := strings.TrimPrefix(fullpath, rootAbs)
	binRoot, _ := filepath.Abs(path.Join(d.root, "bindata"))

	// Keep only first level directories
	if !strings.HasPrefix(basename, "/__") {
		return nil
	}
	fmt.Printf("- %s\n", basename)
	if info.IsDir() && len(basename) > 0 && !strings.Contains(basename, "/") {
		fmt.Printf("%s is a directory\n", basename)
	} else if info.Mode().IsRegular() {
		in, _ := filepath.Abs(fullpath)
		fs, err := os.Open(in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s\n", err)
			return err
		}
		out := path.Join(binRoot, basename+".go")
		fmt.Printf("---> %s\n", out)

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
		funcname := bindata.SafeFuncname(basename, "uu")
		bindata.Translate(fs, fd, "uu", funcname, false, false)
		fs.Close()
		err = bindata.CreateTOC(binRoot, "uu")

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
