package mapp

import (
	"go/ast"
	"log"
	"slices"
	"strings"
)

type EnumMapper struct {
	spec    *ast.Field
	imports []Import
}

const (
	source = "-s"
	target = "-t"
)

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

func (em EnumMapper) EnumsMap() map[string]string {
	mapping := make(map[string]string, 0)
	sliceToMap := func(arr []string) map[string]bool {
		m := make(map[string]bool, len(arr))
		for _, v := range arr {
			m[v] = false
		}
		return m
	}

	ignores := em.ignores()
	checkIgnore := func(goal string, m map[string]bool) {
		for k := range m {
			if ignore, ok := ignores[goal]; ok && k == ignore {
				m[k] = true
			}
		}
	}
	sourceValMap := sliceToMap(em.Source().Values())
	targetValMap := sliceToMap(em.Target().Values())
	checkIgnore(source, sourceValMap)
	checkIgnore(target, targetValMap)

	var ignoreCase bool
	for _, c := range em.Comments() {
		if strings.HasPrefix(c.Value(), "@ignorecase") {
			ignoreCase = true
			continue
		}
	}

	mapRules := em.mapQual()
	hasMapped := func(sv, tv string) {
		sourceValMap[sv] = true
		targetValMap[tv] = true
		mapping[sv] = tv
	}

	for sourceEnm, mapped := range sourceValMap {
		if mapped {
			continue
		}

		targetEnm, hasQual := mapRules[sourceEnm]
		if hasQual {
			hasMapped(sourceEnm, targetEnm)
			continue
		}

		_, samename := targetValMap[sourceEnm]
		if samename {
			hasMapped(sourceEnm, sourceEnm)
			continue
		}

		if !ignoreCase {
			continue
		}

		for k := range targetValMap {
			if strings.EqualFold(sourceEnm, k) {
				hasMapped(sourceEnm, k)
				continue
			}
		}
	}

	unmapped := checkUnmapedEnms(sourceValMap)
	if len(unmapped) > 0 {
		log.Fatalf("%s has unmapped source enums: %v", em.Name(), unmapped)
	}


	unmapped = checkUnmapedEnms(targetValMap)
	if len(unmapped) > 0 {
		log.Fatalf("%s has unmapped target enums: %v", em.Name(), unmapped)
	}

	return mapping
}

func checkUnmapedEnms(m map[string]bool) []string {
	arr := make([]string, 0)
	for enm, mapped := range m {
		if mapped {
			continue
		}
		arr = append(arr, enm)
	}
	return arr
}

func (em EnumMapper) mapQual() map[string]string {
	mapRule := map[string]string{}
	for _, c := range em.Comments() {
		val := c.Value()
		if !strings.HasPrefix(val, "@enum") {
			continue
		}

		args := strings.Split(val, " ")
		for _, a := range args[1:] {
			kv := strings.Split(a, "=")
			if len(kv) != 2 {
				log.Fatalf("wrong @enum config: %s", a)
			}

			mapRule[kv[0]] = kv[1]
		}
	}

	return mapRule
}

func (em EnumMapper) ignores() map[string]string {
	ignores := map[string]string{}
	for _, c := range em.Comments() {
		val := c.Value()
		if !strings.HasPrefix(val, "@ignore") {
			continue
		}

		args := strings.Split(val, " ")
		for _, a := range args[1:] {
			kv := strings.Split(a, "=")
			if len(kv) != 2 {
				log.Fatalf("wrong @ignore config: %s", a)
			}
			if !slices.Contains([]string{source, target}, kv[0]) {
				log.Fatalf("wrong @ignore key '%s' please use '-s' ot '-t'.", kv[0])
			}

			ignores[kv[0]] = kv[1]
		}
	}

	return ignores
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

func (em EnumMapper) Source() SourceEnum {
	return SourceEnum{
		t: em.Params()[0],
	}
}

func (em EnumMapper) Target() TargetEnum {
	return TargetEnum{
		t: em.Results()[0],
	}
}
