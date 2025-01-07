package mapp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type File struct {
	spec *ast.File
}

var file *ast.File

func NewMapperFile(node *ast.File) File {
	return File{
		spec: node,
	}
}

func (f File) Imports() []Import {
	imports := make([]Import, 0, len(f.spec.Imports))
	for _, i := range f.spec.Imports {
		imp := Import{
			spec: i,
		}
		imports = append(imports, imp)
	}

	return imports
}

func (f File) Mappers() []Mapper {
	methodList := make([]Mapper, 0)
	ast.Inspect(f.spec, func(n ast.Node) bool {
		iface, ok := n.(*ast.InterfaceType)
		if !ok {
			return true
		}

		imports := f.Imports()
	searchMethodLoop:
		for _, v := range iface.Methods.List {
			if v.Doc == nil {
				goto addMapper
			}

			for _, d := range v.Doc.List {
				if strings.Contains(d.Text, "@emapper") {
					continue searchMethodLoop
				}
			}

			addMapper:
			methodList = append(methodList, Mapper{
				spec:    v,
				imports: imports,
			})
		}

		return false
	})

	return methodList
}

func MapperFile(filePath string) File {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse file: %v", err))
	}
	file = node

	return NewMapperFile(node)
}

func (f File) EnumsMappers() []EnumMapper {
	methodList := make([]EnumMapper, 0)
	ast.Inspect(f.spec, func(n ast.Node) bool {
		iface, ok := n.(*ast.InterfaceType)
		if !ok {
			return true
		}

		imports := f.Imports()
		for _, v := range iface.Methods.List {
			var isEmapper bool
			if v.Doc == nil {
				continue
			}
			for _, d := range v.Doc.List {
				if strings.Contains(d.Text, "@emapper") {
					isEmapper = true
				}
			}
			if !isEmapper {
				continue
			}

			emapper := EnumMapper{
				spec:    v,
				imports: imports,
			}

			methodList = append(methodList, emapper)
		}

		return false
	})

	return methodList
}
