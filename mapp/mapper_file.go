package mapp

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

type File struct {
	spec *ast.File
}

var modFiles []*ast.File
var cwd string

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
			mapper := Mapper{
				spec:    v,
				imports: imports,
			}
			mapperErr := mapper.validate()
			if mapperErr != nil {
				log.Fatalf("error generate '%s' mapper: %v", mapper.Name(), mapperErr)
			}

			methodList = append(methodList, mapper)
		}

		return false
	})

	return methodList
}

func MapperFile(cwd, filename string) File {
	cwd = cwd
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, cwd)
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]
	modFiles = pkg.Syntax
	for _, f := range modFiles {
		fullFileName := pkg.Fset.Position(f.Pos()).Filename
		if filepath.Base(fullFileName) == filename {
			return NewMapperFile(f)
		}
	}
	panic(fmt.Sprintf("file '%s' not found", filepath.Join(cwd, filename)))
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
