package mapp

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Target struct {
	spec *ast.Field
	pkg  *packages.Package
	r    Result
}

func (s *Target) Fields() []Mappable {
	_, name := s.r.Type()
	return extractFieldsFromStruct("", s.r.Path(), name)
}

func (s *Target) Path() string {
	return s.r.Path()
}

func (s *Target) TypeName() string {
	_, typeName := s.r.Type()
	return typeName
}

func (s *Target) Name() string {
	return s.r.Name()
}

func (s *Target) FullName() string {
	return s.r.Name()
}

func (s *Target) Type() types.Type {
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

func (t *Target) FullType() []types.Type {
	return []types.Type{t.Type()}
}
