package mapp

import (
	"errors"
	"fmt"
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
	if m.spec.Doc == nil {
		return nil
	}
	comments := make([]Comment, 0, len(m.spec.Doc.List))
	for _, v := range m.spec.Doc.List {
		comments = append(comments, Comment{spec: v})
	}

	return comments
}

func (m Mapper) validate() error {
	params := m.Params()
	if len(params) > 1 {
		return errors.New("more than 1 argument")
	}

	results := m.Results()
	if len(results) > 2 {
		return errors.New("more than 2 returning types")
	}

	target := m.Target()
	rules := m.Rules()
	rulesMap := map[string][]Rule{}
	for _, r := range rules {
		fldRules, ok := rulesMap[r.FieldFullName()]
		if !ok {
			fldRules = []Rule{}
		}
		fldRules = append(fldRules, r)
		rulesMap[r.FieldFullName()] = fldRules
	}

fldCheck:
	for _, f := range target.Fields() {
		fldRules := rulesMap[f.FullName()]
		for _, r := range fldRules {
			_, isIgnored := r.(IgnoreTarget)
			if isIgnored {
				continue fldCheck
			}
		}

		src, exists := m.SourceFieldByTarget(f.FullName())
		if !exists {
			return fmt.Errorf("'%s' field has no source. Use '@ql -t=%s -s=<SourceName>' to define or @igt to ignore", f.FullName(), f.FullName())
		}
		for _, r := range fldRules {
			mr, hasMethodSource := r.(MethodSource)
			if !hasMethodSource {
				continue fldCheck
			}
			err := mr.validate(ExpectedSignature{
				In:  src.Type().String(),
				Out: f.Type().String(),
			})
			if err != nil {
				return fmt.Errorf("@fn rule validation error: %w", err)
			}
			continue fldCheck
		}

		for _, t := range f.FullType() {
			_, isMap := t.(*types.Map)
			if isMap {
				return fmt.Errorf("'%s' has map type. \nPlease provide custom method which will map it by @fn -t=%s -n=<functionName> -p=<import/path if neccessory>. \nOr use @igt %s to ignore the field", f.FullName(), f.FullName(), f.FullName())
			}
		}

	}
	return nil
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
		case "@igc":
			igc := IgnoreCase{}
			if len(commandElements) > 1 {
				igc.FullName = commandElements[1]
			}
			rules = append(rules, igc)
		case "@fn":
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
	ignoreCase := false
	for _, r := range m.RulesBy(targetFullName) {
		switch rule := r.(type) {
		case IgnoreCase:
			ignoreCase = true
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
		expF, found := deepFieldSearch(f, sourceFullName, ignoreCase)
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
			continue
		}
		if igc, ok := r.(IgnoreCase); ok && igc.FieldFullName() == "" {
			igc.FullName = fieldFullName
			rules = append(rules, igc)
		}
	}

	return rules
}
