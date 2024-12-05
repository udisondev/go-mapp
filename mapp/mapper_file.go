package mapp

import (
	"fmt"
	"go/ast"
	"strings"
)

type File struct {
	spec *ast.File
}

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
			for _, d := range v.Doc.List {
				if strings.Contains(d.Text, "@emapper") {
					continue searchMethodLoop
				}
			}

			methodList = append(methodList, Mapper{
				spec:    v,
				imports: imports,
			})
		}

		return false
	})

	return methodList
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
			fmt.Printf("found method name: %s\n", v.Names[0].Name)
			var isEmapper bool
			for _, d := range v.Doc.List {
				if strings.Contains(d.Text, "@emapper") {
					isEmapper = true
				}
			}
			if !isEmapper {
				continue
			}
			fmt.Printf("found enum mapper: %s\n", v.Names[0].Name)

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
