package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DataScanner struct {
	root string
}

func (d DataScanner) visit(path string, info os.FileInfo, err error) error {
	baseName := strings.TrimPrefix(path, d.root)
	// Keep only first level directories
	if info.IsDir() && len(baseName) > 0 && !strings.Contains(baseName, "/") {
		fmt.Printf("%s is a directory\n", baseName)
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
