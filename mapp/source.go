package mapp

import (
	"go/ast"
)

type Source struct {
	spec *ast.Field
	p Param
}

func (s Source) Fields() []Field {
	_, name := s.p.Type()
	return extractFieldsFromStruct(".", s.p.Path(), name)
}

func (s Source) Path() string {
	return s.p.Path()
}

func (s Source) TypeName() string {
	_, typeName := s.p.Type()
	return typeName
}

func (s Source) FieldByFullName(fullName string) (Field, bool) {
	fields := s.Fields()
	for _, f := range fields {	
		expF, found := deepFieldSearch(f, fullName)
		if found {
			return expF, found
		}
	}
	return Field{}, false
}
