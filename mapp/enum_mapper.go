package mapp

import (
	"errors"
	"fmt"
	"go/ast"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var strreg = regexp.MustCompile(`^".*"$`)

type EnumMapper struct {
	spec    *ast.Field
	imports []Import
}

type Default struct {
	Type     string
	Value    string
	IsConst  bool
	IsString bool
}

const (
	source = "-s"
	target = "-t"
)

func (em EnumMapper) Name() string {
	return em.spec.Names[0].Name
}

func (em EnumMapper) WithError() bool {
	if len(em.Results()) < 2 {
		return false
	}

	_, typeName := em.Results()[1].Type()
	return typeName == "error"
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
		if strings.HasPrefix(c.Value(), "@igcase") {
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

	checkUnmapedEnms := func(m map[string]bool) []string {
		arr := make([]string, 0)
		for enm, mapped := range m {
			if mapped {
				continue
			}
			arr = append(arr, enm)
		}
		return arr
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
		if !strings.HasPrefix(val, "@ig") {
			continue
		}
		if strings.HasPrefix(val, "@igcase") {
			continue
		}

		args := strings.Split(val, " ")
		for _, a := range args[1:] {
			kv := strings.Split(a, "=")
			if len(kv) != 2 {
				log.Fatalf("wrong ig config: %s", a)
			}
			if !slices.Contains([]string{source, target}, kv[0]) {
				log.Fatalf("wrong ig key '%s' please use '-s' ot '-t'.", kv[0])
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

func (em EnumMapper) Source() Enum {
	return Enum{
		spec:    em.Params()[0].spec,
		imports: em.imports,
	}
}

func (em EnumMapper) Target() Enum {
	return Enum{
		spec:    em.Results()[0].spec,
		imports: em.imports,
	}
}

func (em EnumMapper) Errormsg() []string {
	for _, c := range em.Comments() {
		if !strings.HasPrefix(c.Value(), "@err") {
			continue
		}

		elems, err := parseComment(c.Value())
		if err != nil {
			log.Fatal(err)
		}
		return elems
	}

	return []string{"unknown source enum: %v", em.Source().Name()}
}

// parseComment принимает строку комментария и возвращает массив строк или ошибку
func parseComment(comment string) ([]string, error) {
	comment = strings.TrimPrefix(comment, "//")

	out := make([]string, 0)
	stargMsg, endMsg := -1, -1

	cur := 0
	for cur < len(comment)-1 {
		if comment[cur] == '\\' {
			cur++
			continue
		}
		if comment[cur] == '"' {
			if stargMsg < 0 {
				stargMsg = cur+1
			} else {
				endMsg = cur
			}
		}
		cur++
	}
	if endMsg < 0 {
		return nil, errors.New("message must be wrapped by \"some message\"")
	}

	out = append(out, comment[stargMsg:endMsg])
	if !strings.HasPrefix(comment, "@errf") {
		return out, nil
	}

	argsPart := comment[endMsg+1:]
	argsPart = strings.TrimPrefix(argsPart, ",")
	argsPart = strings.TrimSuffix(argsPart, ")")
	args := strings.Split(argsPart, ",")
	for _, a := range args {
		out = append(out, strings.TrimSpace(a))
	}
	return out, nil
}

func (em EnumMapper) Default() (Default, bool) {
	tt := em.Target().BaseType()
	for _, c := range em.Comments() {
		if !strings.HasPrefix(c.Value(), "@def") {
			continue
		}

		defConf := strings.Split(c.Value(), " ")
		if len(defConf) != 2 {
			log.Fatalf(`
			%s has invalid @def format '%s'.
			Examples:
			1) If target is string -> '@def "any value" or '@def ""'
			2) If target is integer -> '@def 0' or '@def 5' etc
			3) If target is float -> '@def 4.7' or '@def 0' etc
			4) If target is bool -> '@def true' or '@def false
			5) Or any target enum value without type specifying -> if target has AnyType.Value '@def Value'`, em.Name(), c.Value())
		}
		defVal := defConf[1]

		if slices.Contains(em.Target().Values(), defVal) {
			return Default{
				Type:     tt,
				Value:    defVal,
				IsString: tt == "string",
				IsConst:  true,
			}, true
		}

		suffixMsg := fmt.Sprintf("Define value with correct type or use any of: %v", em.Target().Values())

		switch tt {
		case "int", "int8", "int16", "int32", "int64":
			bitSizeVal := strings.ReplaceAll(tt, "int", "")
			bitSize := 32
			if bitSizeVal != "" && bitSizeVal != "32" {
				var bitSizeErr error
				bitSize, bitSizeErr = strconv.Atoi(bitSizeVal)
				if bitSizeErr != nil {
					panic(bitSizeErr.Error())
				}
			}
			_, err := strconv.ParseInt(defVal, 10, bitSize)
			if err != nil {
				log.Fatalf("%s has invalid @def value '%s'. Target type is int%s. %s", em.Name(), defVal, bitSizeVal, suffixMsg)
			}
			return Default{
				Type:     tt,
				Value:    defVal,
				IsString: tt == "string",
				IsConst:  false,
			}, true
		case "uint", "uint8", "uint16", "uint32", "uint64":
			bitSizeVal := strings.ReplaceAll(tt, "uint", "")
			bitSize := 32
			if bitSizeVal != "" && bitSizeVal != "32" {
				var bitSizeErr error
				bitSize, bitSizeErr = strconv.Atoi(bitSizeVal)
				if bitSizeErr != nil {
					panic(bitSizeErr.Error())
				}
			}
			_, err := strconv.ParseUint(defVal, 10, bitSize)
			if err != nil {
				log.Fatalf("%s has invalid @def value '%s'. Target type is uint%s. %s", em.Name(), defVal, bitSizeVal, suffixMsg)
			}
			return Default{
				Type:     tt,
				Value:    defVal,
				IsString: tt == "string",
				IsConst:  false,
			}, true
		case "float32", "float64":
			bitSizeVal := strings.ReplaceAll(tt, "float", "")
			bitSize := 32
			if bitSizeVal != "" && bitSizeVal != "32" {
				var bitSizeErr error
				bitSize, bitSizeErr = strconv.Atoi(bitSizeVal)
				if bitSizeErr != nil {
					panic(bitSizeErr.Error())
				}
			}
			_, err := strconv.ParseFloat(defVal, bitSize)
			if err != nil {
				log.Fatalf("%s has invalid @def value '%s'. Target type is float%s. %s", em.Name(), defVal, bitSizeVal, suffixMsg)
			}
			return Default{
				Type:     tt,
				Value:    defVal,
				IsString: tt == "string",
				IsConst:  false,
			}, true
		case "string":
			if !strreg.MatchString(defVal) {
				log.Fatalf("%s has invelid @def value: %s. Must be wrapped by quotes.", em.Name(), defVal)
			}
			return Default{
				Type:     tt,
				Value:    strings.ReplaceAll(defVal, "\"", ""),
				IsString: tt == "string",
				IsConst:  false,
			}, true
		case "bool":
			if _, err := strconv.ParseBool(defVal); err != nil {
				log.Fatalf("%s has invalid @def value '%s'. Target type bool. %s", em.Name(), defVal, suffixMsg)
			}
			return Default{
				Type:     tt,
				Value:    defVal,
				IsString: tt == "string",
				IsConst:  false,
			}, true
		}

		log.Fatalf("%s could not resolve @def value '%s'", em.Name(), defVal)
	}

	switch tt {
	case "int", "int8", "int16", "int32", "int64":
		return Default{
			Type:     tt,
			Value:    "0",
			IsString: tt == "string",
			IsConst:  false,
		}, false
	case "uint", "uint8", "uint16", "uint32", "uint64":

		return Default{
			Type:     tt,
			Value:    "0",
			IsString: tt == "string",
			IsConst:  false,
		}, false
	case "float32", "float64":
		return Default{
			Type:     tt,
			Value:    "0",
			IsString: tt == "string",
			IsConst:  false,
		}, false
	case "string":
		return Default{
			Type:     tt,
			Value:    `""`,
			IsString: tt == "string",
			IsConst:  false,
		}, false
	case "bool":
		return Default{
			Type:     tt,
			Value:    "false",
			IsString: tt == "string",
			IsConst:  false,
		}, false
	}

	return Default{}, false
}
