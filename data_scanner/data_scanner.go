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
		out, _ := filepath.Abs(path.Join(d.root, "../bindata", baseName))

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
		bindata.Translate(fs, fd, "uu", bindata.SafeFuncname(baseName, "uu"), false, false)
		fd.Close()
		fs.Close()
	}
	return nil
}

func (d DataScanner) Scan() {
	err := filepath.Walk(d.root, d.visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)

}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	d := DataScanner{root: root}
	d.Scan()
}
