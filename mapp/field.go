//go:generate go-enum
package mapp

import (
	// "fmt"
	"go/types"
	"strings"
	"unicode"
)

// TypeFamily ENUM(basic, named, struct, pointer, slice)
type TypeFamily uint8

type TypedField interface {
	TypeFamily() TypeFamily
	Path() string
	TypeName() string
}

const stdlib = "stdlib"

func extractType(s string) string {
	toReturn := s
	if strings.Contains(s, ".") {
		splitedT := strings.Split(s, ".")
		toReturn = splitedT[len(splitedT)-1]
	}

	toReturn = strings.ReplaceAll(toReturn, "*", "")
	toReturn = strings.ReplaceAll(toReturn, "[]", "")

	return toReturn
}

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

// Path implements Mappable.
func (f *Field) Path() string {
	typeString := f.spec.Type().String()
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

// TypeName implements Mappable.
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

// Underlying implements Mappable.
func (f *Field) DeepType() func() (types.Type, bool) {
	level := -1
	currentType := f.Type()
	return func() (types.Type, bool) {
		level++
		if level == 0 {
			currentType = under(currentType)
			return f.Type(), isButtom(f.Type())
		}

		if isButtom(currentType) {
			return currentType, true
		}

		defer func() { currentType = under(currentType) }()

		return currentType, false
	}
}

func (f *Field) FullName() string {
	return f.fieldPath + "." + f.Name()
}

func (f *Field) Fields() []Mappable {
	deepTypeFn := f.DeepType()
	ft, bottom := deepTypeFn()
	for !bottom {
		ft, bottom = deepTypeFn()
	}

	_, ok := ft.(*types.Struct)
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
