package mapp

import "go/ast"

type Mappable interface {
	Path() string
	Name() string
	TypeName() string
	Type() any
	Underlying() *Mappable
	Fields() []*Mappable
}

type mapped struct {
	spec *ast.Field
}

type Mapper struct {
	spec    *ast.Field
	imports []Import
}

func (m *mapped) 
