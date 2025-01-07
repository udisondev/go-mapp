package mapp

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Enum struct {
	spec    *ast.Field
	imports []Import
}

func (e Enum) Values() []string {
	_, typeName := e.Type()
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, e.Path())
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]

	vals := []string{}
	for _, s := range pkg.Syntax {
		ast.Inspect(s, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				return true
			}

			for _, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				if typeIdent, ok := valueSpec.Type.(*ast.Ident); ok {
					if typeIdent.Name != typeName {
						break
					}
				}

				for _, n := range valueSpec.Names {
					vals = append(vals, n.Name)
				}
			}

			return false
		})
	}

	return vals
}

func (e Enum) BaseType() string {
	_, typeName := e.Type()
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, e.Path())
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]

	var baseType string
	for _, s := range pkg.Syntax {
		ast.Inspect(s, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				return true
			}

			for _, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				typeIdent, ok := valueSpec.Type.(*ast.Ident)
				if !ok {
					continue
				}

				if typeIdent.Name != typeName {
					continue
				}

				typeSpec, ok := typeIdent.Obj.Decl.(*ast.TypeSpec)
				if !ok {
					continue
				}

				baseIdent, ok := typeSpec.Type.(*ast.Ident)
				if !ok {
					continue
				}

				baseType = baseIdent.Name
				return false

			}

			return false
		})
	}

	return baseType
}

func (e Enum) Type() (string, string) {
	switch tt := e.spec.Type.(type) {
	case *ast.Ident:
		return "", tt.Name
	case *ast.SelectorExpr:
		return tt.X.(*ast.Ident).Name, tt.Sel.Name

	default:
		panic(fmt.Sprintf("could not extract type from: %T", tt))
	}
}

func (e Enum) Path() string {
	alias, _ := e.Type()
	if alias == "" {
		return ""
	}

	for _, i := range e.imports {
		if i.Alias() == alias {
			return strings.ReplaceAll(i.Path(), "\"", "")
		}
	}

	return ""
}

func (p Enum) Name() string {
	if len(p.spec.Names) == 0 {
		return ""
	}

	return p.spec.Names[0].Name
}
