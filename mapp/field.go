//go:generate go-enum
package mapp

import (
	"fmt"
	"go/types"
	"strings"
	"unicode"
)

const stdlib = "stdlib"

type Field struct {
	spec      *types.Var
	owner     *types.Struct
	fieldPath string
}

func New(spec *types.Var,
	owner *types.Struct,
	fieldPath string) Mappable {
	return &Field{
		spec:      spec,
		owner:     owner,
		fieldPath: fieldPath,
	}
}

func (f *Field) Name() string {
	return f.spec.Origin().Name()
}

func (f *Field) Path() string {
	typeString := f.spec.Type().String()
	fmt.Printf("Type.String() of '%s' is : %v\n", f.FullName(), f.spec)
	startTypeNamePos := strings.LastIndex(typeString, ".")
	if startTypeNamePos < 0 {
		return stdlib
	}
	subString := f.spec.Type().String()[:startTypeNamePos]
	for i, r := range subString {
		if unicode.IsLetter(r) {
			return subString[i:]
		}
	}
	return subString
}

func (f *Field) TypeName() string {
	var subType string
	for i, r := range f.spec.Type().String() {
		if !unicode.IsLetter(r) {
			subType = f.spec.Type().String()[i:]
			break
		}
	}
	startTypeNamePos := strings.LastIndex(subType, ".")
	if startTypeNamePos < 0 {
		return subType
	}

	return subType[startTypeNamePos+1:]
}

func (f *Field) FullType() []types.Type {
	out := make([]types.Type, 0)
	out = append(out, f.Type())
	if !isButtom(f.Type()) {
		undT := under(f.Type())
		out = append(out, undT)
		for !isButtom(undT) {
			undT = under(undT)
			out = append(out, undT)
		}
	}
	return out
}

func (f *Field) FullName() string {
	return f.fieldPath + "." + f.Name()
}

func (f *Field) Fields() []Mappable {
	typeChain := f.FullType()

	_, ok := typeChain[len(typeChain)-1].(*types.Struct)
	if !ok {
		return nil
	}

	return extractFieldsFromStruct(f.FullName(), f.Path(), f.TypeName())
}

func (f *Field) Type() types.Type {
	return f.spec.Type()
}

func isButtom(t types.Type) bool {
	switch t.(type) {
	case *types.Basic, *types.Struct:
		return true
	default:
		return false
	}
}

func under(t types.Type) types.Type {
	switch currentType := t.(type) {
	case *types.Pointer:
		return currentType.Elem()
	case *types.Named:
		return currentType.Underlying()
	case *types.Slice:
		return currentType.Elem()
	case *types.Map:
		return currentType.Elem()
	case *types.Array:
		return currentType.Elem()
	default:
		return t
	}
}
