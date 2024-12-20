package mapp

import (
	"fmt"
	"go/ast"
	"go/types"
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

		if !mapperRuleReg.MatchString(val) {
			panic(fmt.Sprintf("unsupported rule: %s", val))
		}

		buildRuleArg := func(arg string) (RuleArg, string) {
			kvPair := strings.Split(arg, "=")
			if len(kvPair) != 2 {
				panic(fmt.Sprintf("invalid argument format: %s", arg))
			}

			switch kvPair[0] {
			case "-s":
				return RuleArgSource, kvPair[1]
			case "-t":
				return RuleArgTarget, kvPair[1]
			case "-mn":
				return RuleArgMname, kvPair[1]
			case "-mp":
				return RuleArgMpath, kvPair[1]
			default:
				panic(fmt.Sprintf("unknown arg key: %s", kvPair[0]))
			}
		}

		args := strings.Split(val, " ")
		ruleArgs := make(map[RuleArg]string, len(args)-1)
		for _, a := range args[1:] {
			k, v := buildRuleArg(a)
			ruleArgs[k] = v
		}
		rules = append(rules, Rule{
			spec: val,
			args: ruleArgs,
		})

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
		if r.Type() != RuleTypeQual {
			continue
		}
		tname, ok := r.Arg(RuleArgTarget)
		if !ok {
			continue
		}
		if tname != targetFullName {
			continue
		}

		if r.Type() == RuleTypeIgnore {
			panic(fmt.Sprintf("target field: %s must be ignored", targetFullName))
		}

		sname, ok := r.Arg(RuleArgSource)
		if !ok {
			continue
		}
		if strings.Contains(sname, ".") {
			sourceFullName = sname
			break
		}
		lastElemStartPos := strings.LastIndexAny(targetFullName, ".")
		pref := targetFullName[:lastElemStartPos]
		sourceFullName = pref + "." + sname
		break
	}

	for _, f := range m.Source().Fields() {
		expF, found := deepFieldSearch(f, sourceFullName)
		if found {
			return expF, found
		}
	}

	return nil, false
}

func (m Mapper) FieldRules(fieldFullName string) []Rule {
	rules := make([]Rule, 0)
	for _, r := range m.Rules() {
		_, isTargetFieldRule := r.Arg(RuleArgTarget)
		_, isSourceFieldRule := r.Arg(RuleArgSource)
		if !isTargetFieldRule && !isSourceFieldRule {
			continue
		}

		rules = append(rules, r)
	}

	return rules
}

func (m Mapper) RulesBy(fieldFullName string, ruleType RuleType) (Rule, bool) {
	for _, r := range m.Rules() {
		if r.Type() != ruleType {
			continue
		}

		tflname, isTargetFieldRule := r.Arg(RuleArgTarget)
		sflname, isSourceFieldRule := r.Arg(RuleArgSource)
		if !isTargetFieldRule && !isSourceFieldRule {
			continue
		}

		if tflname != fieldFullName && sflname != fieldFullName {
			continue
		}

		return r, true
	}

	return Rule{}, false
}
