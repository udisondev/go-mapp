package mapp

import (
	"go/ast"
	"go/types"
	"log"
	"regexp"
	"strings"
)

var mapperRuleReg = regexp.MustCompile(`^@(qual|enum|ignore) `)

type Mappable interface {
	Path() string
	Name() string
	FullName() string
	TypeName() string
	Type() types.Type
	FullType() []types.Type
	Fields() []Mappable
}

type Mapper struct {
	spec    *ast.Field
	imports []Import
}

func (m Mapper) Name() string {
	return m.spec.Names[0].Name
}

func (m Mapper) WithError() bool {
	if len(m.Results()) < 2 {
		return false
	}

	_, typeName := m.Results()[1].Type()
	return typeName == "error"
}

func (m Mapper) Comments() []Comment {
	comments := make([]Comment, 0, len(m.spec.Doc.List))
	for _, v := range m.spec.Doc.List {
		comments = append(comments, Comment{spec: v})
	}

	return comments
}

func (m Mapper) Rules() []Rule {
	rules := []Rule{}
	for _, c := range m.Comments() {
		val := c.Value()
		if !strings.HasPrefix(val, "@") {
			continue
		}

		commandElements := strings.Split(val, " ")
		switch commandElements[0] {
		case "@ql":
			q := Qual{}
			for _, el := range commandElements[1:] {
				args := strings.Split(el, "=")
				switch args[0] {
				case "-s":
					q.Source = args[1]
				case "-t":
					q.Target = args[1]
				default:
					log.Fatalf("Unknown argument: %s", args[0])
				}
			}
			rules = append(rules, q)
		case "@igt":
			rules = append(rules, IgnoreTarget{FullName: commandElements[1]})
		case "@mm":
			mm := MethodSource{}
			for _, el := range commandElements[1:] {
				args := strings.Split(el, "=")
				switch args[0] {
				case "-t":
					mm.Target = args[1]
				case "-n":
					mm.Name = args[1]
				case "-p":
					mm.Path = args[1]
				default:
					log.Fatalf("Unknown argument: %s", args[0])
				}
			}
			rules = append(rules, mm)
		default:
			log.Fatalf("Unknown command: %s", commandElements[0])
		}
	}

	return rules
}

func (m Mapper) Params() []Param {
	fnT, ok := m.spec.Type.(*ast.FuncType)
	if !ok {
		panic("is not a ast.FuncType")
	}

	params := make([]Param, 0, len(fnT.Params.List))
	for _, p := range fnT.Params.List {
		params = append(params, Param{spec: p, imports: m.imports})
	}

	return params
}

func (m Mapper) Results() []Result {
	fnT, ok := m.spec.Type.(*ast.FuncType)
	if !ok {
		panic("is not a ast.FuncType")
	}

	params := make([]Result, 0, len(fnT.Results.List))
	for _, p := range fnT.Results.List {
		params = append(params, Result{spec: p, imports: m.imports})
	}

	return params
}

func (m Mapper) Source() Mappable {
	return &Source{
		spec: m.Params()[0].spec,
		p:    m.Params()[0],
	}
}

func (m Mapper) Target() Mappable {
	return &Target{
		spec: m.Results()[0].spec,
		r:    m.Results()[0],
	}
}

func (m Mapper) SourceFieldByTarget(targetFullName string) (Mappable, bool) {
	sourceFullName := targetFullName
	for _, r := range m.Rules() {
		switch rule := r.(type) {
		case Qual:
			if rule.Target == targetFullName {
				lastElemStartPos := strings.LastIndexAny(targetFullName, ".")
				pref := targetFullName[:lastElemStartPos]
				sourceFullName = pref + "." + rule.Source
			}
		case IgnoreTarget:
			if rule.FullName == targetFullName {
				return nil, false
			}
		}
	}

	for _, f := range m.Source().Fields() {
		expF, found := deepFieldSearch(f, sourceFullName)
		if found {
			return expF, found
		}
	}

	return nil, false
}

func (m Mapper) RulesBy(fieldFullName string) []Rule {
	rules := make([]Rule, 0)
	for _, r := range m.Rules() {
		if r.FieldFullName() == fieldFullName {
			rules = append(rules, r)
		}
	}

	return rules
}
