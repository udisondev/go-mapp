package mapp

import (
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/packages"
)

var CurrentPath string

type Rule interface {
	FieldFullName() string
}

type IgnoreTarget struct {
	FullName string
}

type Qual struct {
	Target, Source string
}

type MethodSource struct {
	Target, Name, Path string
}

func (ms MethodSource) WithErr() bool {
	nodes := []*ast.File{file}
	if ms.Path != "" {
		cfg := &packages.Config{
			Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
			Fset: token.NewFileSet(),
		}
		pkgs, err := packages.Load(cfg, ms.Path)
		if err != nil {
			panic(err)
		}
		pkg := pkgs[0]
		nodes = pkg.Syntax
	}

	var methodFound bool
	var returningErr bool
	for _, node := range nodes {
		ast.Inspect(node, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if funcDecl.Name.Name != ms.Name {
				return true
			}

			methodFound = true

			returns := funcDecl.Type.Results.List
			if len(returns) > 2 {
				log.Fatal("Too much returned value")
			}

			for _, r := range returns {
				id, ok := r.Type.(*ast.Ident)
				if !ok {
					continue
				}

				if id.Name == "error" {
					returningErr = true
					return false
				}

			}

			return false
		})
	}

	if !methodFound {
		log.Fatalf("Method '%s' not found", ms.Name)
	}

	return returningErr
}

func (i IgnoreTarget) FieldFullName() string { return i.FullName }
func (i Qual) FieldFullName() string         { return i.Target }
func (i MethodSource) FieldFullName() string { return i.Target }
