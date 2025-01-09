package mapp

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

var CurrentPath string

type Rule interface {
	FieldFullName() string
}

type IgnoreTarget struct {
	FullName string
}

type IgnoreCase struct {
	FullName string
}

type Qual struct {
	Target, Source string
}

type MethodSource struct {
	Target, Name, Path string
	hasErr             bool
}

func (ms MethodSource) validate(expected ExpectedSignature) error {
	path := cwd
	if ms.Path != "" {
		path = ms.Path
	}
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypesInfo,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]
	found := false
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcDecl.Recv != nil {
				continue
			}

			funcDef := pkg.TypesInfo.Defs[funcDecl.Name]
			if funcDecl == nil {
				continue
			}

			fn, ok := funcDef.(*types.Func)
			if !ok {
				continue
			}
			if fn.Name() != ms.Name {
				continue
			}
			found = true
			err := ms.compareSignature(fn, expected)
			if err != nil {
				return fmt.Errorf("%s: %w", ms.Name, err)
			}
		}
	}

	if !found {
		return fmt.Errorf("'%s' function not found", ms.Name)
	}
	return nil
}

// ExpectedSignature описывает ожидаемую сигнатуру метода
type ExpectedSignature struct {
	In  string
	Out string
}

func (ms MethodSource) compareSignature(obj *types.Func, expected ExpectedSignature) error {
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		panic("is not a signature")
	}

	// Проверяем параметры
	params := sig.Params()
	if params.Len() != 1 || params.At(0).Type().String() != expected.In {
		return fmt.Errorf("must receive only one argument with type '%s'", expected.In)
	}
	// Проверяем возвращаемые значения
	results := sig.Results()
	if results.Len() < 1 {
		return fmt.Errorf("must returns '%s' type", expected.Out)
	}

	if results.Len() == 1 && results.At(0).Type().String() != expected.Out {
		return fmt.Errorf("must returns '%s' type", expected.Out)
	}

	if results.Len() == 2 && results.At(1).Type().String() != "error" {
		return fmt.Errorf("second argument must be 'error'")
	}
	if results.Len() > 2 {
		return errors.New("must returns less or equal than 2 types")
	}

	return nil
}

func (ms *MethodSource) WithErr() bool {
	// Конфигурация загрузки пакетов
	cfg := &packages.Config{
		Mode: packages.NeedTypes | // Необходимы типы
			packages.NeedTypesInfo | // Необходима типовая информация
			packages.NeedSyntax | // Необходимы AST-файлы
			packages.NeedName, // Необходимо имя пакета
	}

	pkgPath := ms.Path
	if pkgPath == "" {
		pkgPath = cwd
	}
	// Загрузка пакетов по заданному пути
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		panic(err)
	}

	// Определение встроенного типа error для сравнения
	errorType := types.Universe.Lookup("error").Type().Underlying()

	// Итерация по загруженным пакетам
	for _, pkg := range pkgs {
		// Поиск объекта функции в области видимости пакета
		obj := pkg.Types.Scope().Lookup(ms.Name)
		if obj == nil {
			continue // Функция не найдена в этом пакете
		}

		// Проверка, что объект является функцией
		funcObj, ok := obj.(*types.Func)
		if !ok {
			continue // Объект не является функцией
		}

		// Получение сигнатуры функции
		sig, ok := funcObj.Type().(*types.Signature)
		if !ok {
			continue // Тип объекта не является сигнатурой функции
		}

		// Проверка возвращаемых типов
		results := sig.Results()
		for i := 0; i < results.Len(); i++ {
			retType := results.At(i).Type().Underlying()
			if types.Identical(retType, errorType) {
				return true // Функция возвращает error
			}
		}
	}

	return false // Функция не возвращает error
}

func (i IgnoreTarget) FieldFullName() string { return i.FullName }
func (i Qual) FieldFullName() string         { return i.Target }
func (i IgnoreCase) FieldFullName() string   { return i.FullName }
func (i MethodSource) FieldFullName() string { return i.Target }
