package mapp

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

type SourceEnum struct {
	name string
	t Param
}

type TargetEnum struct {
	name string
	t Result
}

func (se SourceEnum) Values() []string {
	_, typeName := se.t.Type()
	return enumValues(se.Path(), typeName)
}

func (te TargetEnum) Values() []string {
	_, typeName := te.t.Type()
	return enumValues(te.Path(), typeName)
}

func enumValues(path, typeName string) []string {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, path)
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

func (se SourceEnum) Path() string {
	return se.t.Path()
}

func (te TargetEnum) Path() string {
	return te.t.Path()
}

func (se SourceEnum) Type() string {
	_, t := se.t.Type()
	return t
}

func (te TargetEnum) Type() string {
	_, t := te.t.Type()
	return t
}