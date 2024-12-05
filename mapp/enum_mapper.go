package mapp

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

var emapperRuleReg = regexp.MustCompile(`^@(enum|ignore|eqname|ignorecase) `)

type EnumMapper struct {
	spec    *ast.Field
	imports []Import
}

func (em EnumMapper) Name() string {
	return em.spec.Names[0].Name
}

func (em EnumMapper) Comments() []Comment {
	comments := make([]Comment, 0, len(em.spec.Doc.List))
	for _, v := range em.spec.Doc.List {
		comments = append(comments, Comment{spec: v})
	}

	return comments
}

func (em EnumMapper) EnumPairs() []EnumPair {
	enumPairs := make([]EnumPair, 0)
	for _, c := range em.Comments() {
		val := c.Value()
		if !strings.HasPrefix(val, "@") {
			continue
		}

		if strings.HasPrefix(val, "@emapper") {
			continue
		}

		if !emapperRuleReg.MatchString(val) {
			panic(fmt.Sprintf("unsupported rule: %s", val))

		}

		args := strings.Split(val, " ")
		source := em.Params()[0]
		target := em.Results()[0]
		sourcesEnums := make([]string, 0, len(args[1:]))
		targetEnums := make([]string, 0, len(args[1:]))
		for _, a := range args[1:] {
			kv := strings.Split(a, "=")
			sourcesEnums = append(sourcesEnums, kv[0])
			targetEnums = append(targetEnums, kv[1])
			enumPairs = append(enumPairs, EnumPair{
				source: SourceEnum{
					name: kv[0],
					t:    source,
				},
				target: TargetEnum{
					name: kv[1],
					t:    target,
				},
			})
		}

		_, sourceType := source.Type()
		_, targetType := target.Type()
		notMapped, notExistingEnum := checkSource(source.Path(), sourceType, sourcesEnums)
		if len(notExistingEnum) > 0 {
			panic(fmt.Sprintf("%s declares not existing source enums: %v", em.Name(), notExistingEnum))
		}
		if len(notMapped) > 0 {
			panic(fmt.Sprintf("%s found not mapped source enums: %v", em.Name(), notMapped))
		}

		notMapped, notExistingEnum = checkSource(target.Path(), targetType, targetEnums)
		if len(notExistingEnum) > 0 {
			panic(fmt.Sprintf("%s declares not existing target enums: %v", em.Name(), notExistingEnum))
		}
		if len(notMapped) > 0 {
			panic(fmt.Sprintf("%s found not mapped target enums: %v", em.Name(), notMapped))
		}

	}

	return enumPairs
}

func checkSource(path, typeName string, defMapping []string) ([]string, []string) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax,
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]

	actualEnms := []string{}
	for _, s := range pkg.Syntax {
		ast.Inspect(s, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				return true
			}

			for _, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				if typeIdent, ok := valueSpec.Type.(*ast.Ident); ok {
					if typeIdent.Name != typeName {
						break
					}
				}

				for _, n := range valueSpec.Names {
					actualEnms = append(actualEnms, n.Name)
				}

			}

			return false
		})
	}
	sliceToMap := func(arr []string) map[string]struct{} {
		m := make(map[string]struct{})
		for _, s := range arr {
			m[s] = struct{}{}
		}
		return m
	}
	givenMap := sliceToMap(defMapping)
	actualMap := sliceToMap(actualEnms)

	notExistingEnum := make([]string, 0)
	for k := range givenMap {
		if _, ok := actualMap[k]; !ok {
			notExistingEnum = append(notExistingEnum, k)
		}
	}

	notMapped := make([]string, 0)
	for k := range actualMap {
		if _, ok := givenMap[k]; !ok {
			notMapped = append(notMapped, k)
		}
	}

	return notMapped, notExistingEnum

}

func (em EnumMapper) Params() []Param {
	fnT, ok := em.spec.Type.(*ast.FuncType)
	if !ok {
		panic("is not a ast.FuncType")
	}

	params := make([]Param, 0, len(fnT.Params.List))
	for _, p := range fnT.Params.List {
		params = append(params, Param{spec: p, imports: em.imports})
	}

	return params
}

func (em EnumMapper) Results() []Result {
	fnT, ok := em.spec.Type.(*ast.FuncType)
	if !ok {
		panic("is not a ast.FuncType")
	}

	params := make([]Result, 0, len(fnT.Results.List))
	for _, p := range fnT.Results.List {
		params = append(params, Result{spec: p, imports: em.imports})
	}

	return params
}

func (em EnumMapper) Source() Source {
	return Source{
		spec: em.Params()[0].spec,
		p:    em.Params()[0],
	}
}

func (em EnumMapper) Target() Target {
	return Target{
		spec: em.Results()[0].spec,
		r:    em.Results()[0],
	}
}
