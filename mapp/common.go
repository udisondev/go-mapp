package mapp

import (
	"go/token"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

func deepFields(f Mappable) []Mappable {
	out := make([]Mappable, len(f.Fields()))
	for _, fld := range f.Fields() {
		out = append(out, fld)
		if fflds := f.Fields(); len(fflds) > 0 {
			out = append(out, deepFields(f)...)
		}
	}
	return out
}

func deepFieldSearch(f Mappable, fieldFullName string, ignoreCase bool) (Mappable, bool) {
	actualName := f.FullName()
	expectedName := fieldFullName
	if ignoreCase {
		actualName = strings.ToLower(actualName)
		expectedName = strings.ToLower(expectedName)
	}

	if ignoreCase {
		log.Printf("Actual: %s expected: %s", actualName, expectedName)
	}
	if actualName == expectedName {
		return f, true
	}

	fields := f.Fields()
	if len(fields) == 0 {
		return &Field{}, false
	}

	for _, ff := range fields {
		expF, found := deepFieldSearch(ff, fieldFullName, ignoreCase)
		if found {
			return expF, true
		}
	}

	return &Field{}, false
}

func extractFieldsFromStruct(filedPath, typePath, typeName string) []Mappable {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, typePath)
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]

	obj := pkg.Types.Scope().Lookup(typeName)
	str, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return nil
	}

	fields := make([]Mappable, 0, str.NumFields())
	for i := 0; i < str.NumFields(); i++ {
		fields = append(fields, New(
			str.Field(i),
			str,
			filedPath,
		))
	}

	return fields
}
