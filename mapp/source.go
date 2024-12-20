package mapp

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Source struct {
	spec *ast.Field
	pkg  *packages.Package
	p    Param
}

func (s *Source) Fields() []Mappable {
	_, name := s.p.Type()
	return extractFieldsFromStruct("", s.p.Path(), name)
}

func (s *Source) Path() string {
	return s.p.Path()
}

func (s *Source) TypeName() string {
	_, typeName := s.p.Type()
	return typeName
}

func (s *Source) Name() string {
	return s.p.Name()
}

func (s *Source) FullName() string {
	return s.p.Name()
}

func (s *Source) Type() types.Type {
	if s.pkg == nil {
		cfg := &packages.Config{
			Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
			Fset: token.NewFileSet(),
		}
		pkgs, err := packages.Load(cfg, s.Path())
		if err != nil {
			panic(err)
		}
		s.pkg = pkgs[0]
	}

	obj := s.pkg.Types.Scope().Lookup(s.TypeName())
	return obj.Type()
}

func (s *Source) FullType() []types.Type {
	return []types.Type{s.Type()}
}
