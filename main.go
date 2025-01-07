package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/udisondev/go-mapp/gen"
	"github.com/udisondev/go-mapp/mapp"
)

func main() {
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		fmt.Println("GOFILE not set")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	fpath := filepath.Join(cwd, goFile)
	fmt.Printf("generating '%s'...\n", fpath)

	pkgName := os.Getenv("GOPACKAGE")
	if goFile == "" {
		fmt.Println("GOPACKAGE not set")
		os.Exit(1)
	}
	filenameWithoutExtension, _ := strings.CutSuffix(goFile, ".go")
	gen.Generate(mapp.MapperFile(fpath), pkgName, filepath.Join(cwd, filenameWithoutExtension+"_impl.go"))
}
