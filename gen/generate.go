package gen

import (
	"fmt"
	"go/types"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
	"golang.org/x/exp/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyz"

var enmMappers = make(map[string]mapp.EnumMapper)
var submappers = make(map[string]string)
var f *File

func generateEnumMapper(f *File, em mapp.EnumMapper, include func(key string, em mapp.EnumMapper)) {
	sourcePath := em.Source().Path()
	targetPath := em.Target().Path()
	_, sourceT := em.Source().Type()
	_, targetT := em.Target().Type()

	include(emapperMapKey(em.Source(), em.Target()), em)

	errElems := em.Errormsg()
	withErr := em.WithError()
	sign := f.Func().Id(em.Name()).Params(Id(em.Source().Name()).Qual(sourcePath, sourceT))
	if withErr {
		sign.Params(Qual(targetPath, targetT), Id("error"))
	} else {
		sign.Qual(targetPath, targetT)
	}
	sign.
		BlockFunc(func(g *Group) {
			errCall := []Code{}
			errCall = append(errCall, Lit(errElems[0]))
			if len(errElems) > 1 {
				for _, v := range errElems[1:] {
					errCall = append(errCall, Id(v))
				}
			}
			g.Switch(Id(em.Source().Name())).BlockFunc(func(g *Group) {
				for s, t := range em.EnumsMap() {
					returns := []Code{Qual(em.Target().Path(), t)}
					if withErr {
						returns = append(returns, Nil())
					}
					g.Case(Qual(em.Source().Path(), s)).Block(
						Return(returns...),
					)
				}
				g.Default().BlockFunc(func(g *Group) {
					def, ok := em.Default()
					switch {
					case ok && def.IsConst && withErr:
						g.Return(Qual(targetPath, def.Value), Nil())
					case ok && def.IsConst:
						g.Return(Qual(targetPath, def.Value))
					case ok && withErr:
						if def.IsString {
							g.Return(Lit(def.Value), Nil())
						} else {
							g.Return(Id(def.Value), Nil())
						}
					case ok:
						if def.IsString {
							g.Return(Lit(def.Value))
						} else {
							g.Return(Id(def.Value))
						}
					case withErr:
						if def.IsString {
							g.Return(Lit(def.Value), Qual("fmt", "Errorf").Call(errCall...))
						} else {
							g.Return(Id(def.Value), Qual("fmt", "Errorf").Call(errCall...))
						}
					default:
						g.Panic(Qual("fmt", "Sprintf").Call(errCall...))
					}
				})
			})
		})
	f.Line()
}

func Generate(mf mapp.File, packageName, outputFile string) {
	f = NewFile(packageName)

	for _, em := range mf.EnumsMappers() {
		generateEnumMapper(f, em, func(key string, em mapp.EnumMapper) { enmMappers[key] = em })
	}
	for _, m := range mf.Mappers() {
		err := generateMapper(
			m.Target(),
			m.Source(),
			WithMapperName(m.Name()),
			WithRulesBy(m.RulesBy),
			WithSourceFieldByTarget(m.SourceFieldByTarget),
			WithErr(m.WithError()),
		)
		if err != nil {
			log.Fatalf("Error generate mapper '%s': %v", m.Name(), err)
		}
	}

	err := f.Save(outputFile)
	if err != nil {
		log.Fatalf("failed to save file: %v", err)
	}
}

func emapperMapKey(from, to mapp.Enum) string {
	_, fromT := from.Type()
	_, toT := to.Type()
	return from.Path() + "." + fromT + "|" + to.Path() + "." + toT
}

func isBasic(t types.Type) (types.BasicKind, bool) {
	b, isBasic := t.(*types.Basic)
	if isBasic {
		return b.Kind(), true
	}
	return types.Invalid, false
}

func isPointer(t types.Type) bool {
	_, isPtr := t.(*types.Pointer)
	return isPtr
}

func isEnum(ft []types.Type) bool {
	cur := 0
	for cur < len(ft) {
		_, isNamed := ft[cur].(*types.Named)
		if isNamed {
			_, isBasic := isBasic(ft[cur+1])
			return isBasic
		}
		cur++
	}
	return false
}

func isStruct(ft []types.Type) bool {
	cur := 0
	for cur < len(ft) {
		if isSlice(ft[cur]) {
			return false
		}

		if isPointer(ft[cur]) {
			return false
		}

		_, isNamed := ft[cur].(*types.Named)
		if isNamed {
			_, isStruct := ft[cur+1].(*types.Struct)
			return isStruct
		}
		cur++
	}
	return false
}

func isSlice(t types.Type) bool {
	_, isSlice := t.(*types.Slice)
	return isSlice
}

func generateMapper(target, source mapp.Mappable, optFuncs ...optFunc) error {
	opts := gOpts(optFuncs...)
	stm := f.Func().Id(opts.mapperName).Params(Id("src").Qual(source.Path(), source.TypeName()))
	returns := []Code{Qual(target.Path(), target.TypeName())}
	if opts.withErr {
		returns = append(returns, Error())
	}
	stm.Params(returns...)

	stm.BlockFunc(func(g *Group) {
		if opts.withErr {
			g.Var().Id("err").Error()
		}
		type fldEq struct {
			target string
			source func(*Statement) *Statement
			ignore bool
		}
		fields := make([]fldEq, 0)
		addPair := func(tt string, src func(*Statement)) {
			fields = append(fields, fldEq{target: tt, source: func(s *Statement) *Statement {
				src(s)
				s.Op(",")
				return s
			}})
		}
	mapfieldloop:
		for _, ttFld := range target.Fields() {
			rules := opts.rulesBy(ttFld.FullName())
			for _, r := range rules {
				if _, isIgnored := r.(mapp.IgnoreTarget); isIgnored {
					fields = append(fields, fldEq{target: ttFld.Name(), ignore: true})
					continue mapfieldloop
				}
			}

			srcFld, exists := opts.sourceFieldByTarget(ttFld.FullName())
			if !exists {
				log.Fatalf("'%s' has no source field. Use '@igt %s' or '@ql -t=%s -s=<.Path.To.The.Source.FieldName>'", ttFld.FullName(), ttFld.FullName(), ttFld.FullName())
			}
			err := checkTypesChain(ttFld, srcFld)
			if err != nil {
				log.Fatal(err.Error())
			}
			for _, r := range rules {
				if r, hasMethod := r.(mapp.MethodSource); hasMethod {
					if !r.WithErr() {
						addPair(ttFld.Name(), func(s *Statement) { s.Qual(r.Path, r.Name).Call(Id("src." + srcFld.Name())) })
						continue mapfieldloop
					}
					errName := "err" + ttFld.Name() + "Mapping"
					g.List(Id("mapped"+ttFld.Name()), Id(errName)).Op(":=").Qual(r.Path, r.Name).Call(Id("src." + srcFld.Name()))
					addPair(ttFld.Name(), func(s *Statement) { s.Id("mapped" + ttFld.Name()) })
					g.If(Id(errName).Op("!=").Nil()).BlockFunc(func(g *Group) {
						errMessage := fmt.Sprintf("'%s' -> '%s'", srcFld.Name(), ttFld.Name())
						if opts.withErr {
							g.Return(List(Qual(target.Path(), target.TypeName()).Block(), Qual("fmt", "Errorf").Call(List(Lit(errMessage+": %w")), Id(errName))))
						} else {
							g.Panic(Qual("fmt", "Sprintf").Call(List(Lit(errMessage+": %v")), Id(errName)))
						}
					})

					continue mapfieldloop
				}
			}

			if strings.ReplaceAll(ttFld.Type().String(), "*", "") == srcFld.Type().String() {
				srcName := "src." + srcFld.Name()
				if isPointer(ttFld.Type()) {
					srcName = "&" + srcName
				}
				addPair(ttFld.Name(), func(s *Statement) { s.Id(srcName) })
				continue
			}

			if ttFld.Path() == "stdlib" && srcFld.Path() == "stdlib" && isPointer(ttFld.Type()) && !isPointer(srcFld.Type()) {
				addPair(ttFld.Name(), func(s *Statement) { s.Id("&src." + srcFld.Name()) })
				continue
			}

			ttFt := ttFld.FullType()
			srcFt := srcFld.FullType()
			truncatedTtFt, truncated := trucatePointer(ttFt)
			if truncated && truncatedTtFt[0].String() == srcFt[0].String() {
				addPair(ttFld.Name(), func(s *Statement) { s.Id("src." + srcFld.Name()) })
				continue
			}
			mappedFieldName := "mapped" + ttFld.Name()
			if truncated {
				mappedFieldName = "&" + mappedFieldName
			}
			addPair(ttFld.Name(), func(s *Statement) { s.Id(mappedFieldName) })
			if ttFld.Path() == "stdlib" {
				g.Var().Id("mapped" + ttFld.Name()).Op(truncatedTtFt[0].String())
			} else {
				g.Var().Id("mapped"+ttFld.Name()).Op(trimTypeString(truncatedTtFt[0])).Qual(ttFld.Path(), ttFld.TypeName())
			}

			mapFld("mapped"+ttFld.Name(), "src."+srcFld.Name(), truncatedTtFt, srcFt, g, append(optFuncs, WithTarget(target), WithSource(source), WithTargetField(ttFld), WithSourceField(srcFld))...)
		}

		returns := make([]Code, 0)
		returns = append(returns, Qual(target.Path(), target.TypeName()).BlockFunc(func(g *Group) {
			for _, pair := range fields {
				if pair.ignore {
					g.Comment(pair.target + ": ignored")
					continue
				}

				pair.source(g.Id(pair.target).Op(":"))
			}
		}))
		if opts.withErr {
			returns = append(returns, Id("err"))
		}
		g.Line()
		g.Return(returns...)
	})

	return nil
}

func mapFld(ttName, srcName string, ttTypes, srcTypes []types.Type, g *Group, optFns ...optFunc) {
	if len(ttTypes) < 1 || len(srcTypes) < 1 {
		return
	}

	opts := gOpts(optFns...)
	if opts.iterValName == nil {
		iterValsFunc := WithIterValName(func() func() string {
			cur := -1
			return func() string {
				cur++
				if cur == 0 {
					return "i"
				}

				return "i" + strconv.Itoa(cur)
			}
		}())
		optFns = append(optFns, iterValsFunc)
		opts = iterValsFunc(opts)
	}

	ttIsPtr := isPointer(ttTypes[0])
	srcIsPtr := isPointer(srcTypes[0])
	if ttIsPtr && srcIsPtr {
		mapFld(ttName, srcName, ttTypes[1:], srcTypes[1:], g, optFns...)
		return
	}
	if ttIsPtr && !srcIsPtr {
		mapFld(ttName, srcName, ttTypes[1:], srcTypes, g, append(optFns, WithTargetIsPointer(true))...)
		return
	}
	if !ttIsPtr && srcIsPtr {
		g.If(Id(srcName).Op("!=").Nil()).BlockFunc(func(g *Group) {
			mapFld(ttName, srcName, ttTypes, srcTypes[1:], g, append(optFns, WithSourceIsPointer(true))...)
		})
		return
	}

	basicTt, ttIsBasic := isBasic(ttTypes[0])
	basicSrc, srcIsBasic := isBasic(srcTypes[0])

	switch {
	case isSlice(ttTypes[0]) && isSlice(srcTypes[0]):
		if opts.srcIsPtr {
			srcName = "*" + srcName
		} else if opts.ttIsPtr {
			srcName = "&" + srcName
		}
		if opts.ttIsPtr {
			g.Var().Id(ttName+"Tmp").Op(trimTypeString(ttTypes[0])).Qual(opts.ttFld.Path(), opts.ttFld.TypeName())
		}
		iterName := opts.iterValName()
		g.For(List(Id("_"), Id(iterName))).Op(":=").Range().Id(srcName).BlockFunc(func(g *Group) {
			assignTo := opts.ttFld.Name() + strings.Join(buildTmpNames(ttTypes[1:]), "")
			if opts.ttFld.Path() == "stdlib" {
				g.Var().Id(assignTo).Op(opts.ttFld.Type().String())
			} else {
				g.Var().Id(assignTo).Op(trimTypeString(ttTypes[1])).Qual(opts.ttFld.Path(), opts.ttFld.TypeName())
			}
			mapFld(assignTo, iterName, ttTypes[1:], srcTypes[1:], g, append(optFns, WithTargetIsPointer(isPointer(ttTypes[1])), WithSourceIsPointer(isPointer(srcTypes[1])))...)
			if opts.ttIsPtr {
				g.Id(ttName+"Tmp").Op("=").Append(Id("*"+ttName), Id(assignTo))
			} else {
				g.Id(ttName).Op("=").Append(Id(ttName), Id(assignTo))
			}
		})
		if opts.ttIsPtr {
			g.Id(ttName).Op("=").Id("&" + ttName + "Tmp")
		}
	case isStruct(ttTypes) && isStruct(srcTypes):
		if opts.srcIsPtr {
			srcName = "*" + srcName
		} else if opts.ttIsPtr {
			srcName = "&" + srcName
		}
		if ttTypes[0].String() == srcTypes[0].String() {
			g.Id(ttName).Op("=").Id(srcName)
			return
		}

		errName := ttName + "MappingErr"
		if opts.withErr {
			errName = "err"
		} else {
			g.Var().Id(errName).Error()
		}

		typesKey := typePairKey(srcTypes[0], ttTypes[0])
		submapperName, ok := submappers[typesKey]
		if !ok {
			submapperName = randString(5)
			submappers[typesKey] = submapperName
			generateMapper(opts.ttFld, opts.srcFld, append(optFns, WithErr(true), WithMapperName(submapperName))...)
		}
		assignTo := ttName
		if opts.ttIsPtr {
			assignTo = ttName + "Tmp"
			g.Var().Id(assignTo).Op(trimTypeString(ttTypes[1])).Qual(opts.ttFld.Path(), opts.ttFld.TypeName())
		}
		g.List(Id(assignTo), Id(errName)).Op("=").Id(submapperName).Call(Id(srcName))
		g.If(Id(errName).Op("!=").Nil()).BlockFunc(func(g *Group) {
			errMessage := fmt.Sprintf("'%s' -> '%s'", opts.srcFld.Name(), opts.ttFld.Name())
			if opts.withErr {
				g.Return(List(Qual(opts.tt.Path(), opts.tt.TypeName()).Block(), Qual("fmt", "Errorf").Call(List(Lit(errMessage+": %w")), Id(errName))))
			} else {
				g.Panic(Qual("fmt", "Sprintf").Call(List(Lit(errMessage+": %v")), Id(errName)))
			}
		})
		if opts.ttIsPtr {
			assignTo = ttName + "Tmp"
			g.Id(ttName).Op("=").Id("&" + assignTo)
		}
	case isEnum(ttTypes) && isEnum(srcTypes):
		enmMapper, ok := enmMappers[typePairKey(srcTypes[0], ttTypes[0])]
		if !ok {
			log.Fatalf("has no enum mapper from: %s to: %s", srcTypes[0].String(), ttTypes[0].String())
		}
		withErr := enmMapper.WithError()
		assign := []Code{}
		errName := ttName + "MappingErr"
		if opts.withErr {
			errName = "err"
		}

		switch {
		case !withErr && !opts.ttIsPtr:
			g.Id(ttName).Op("=").Id(enmMapper.Name()).Call(Id(srcName))
			return
		case withErr && !opts.ttIsPtr:
			assign = append(assign, Id(ttName), Id(errName))
		default:
			assign = append(assign, Id("mapped"+ttName+"Enum"), Id(errName))
		}
		g.List(assign...).Op("=").Id(enmMapper.Name()).Call(Id(srcName))
		if withErr {
			g.If(Id(errName).Op("!=").Nil()).BlockFunc(func(g *Group) {
				errMessage := fmt.Sprintf("'%s' -> '%s'", opts.srcFld.Name(), opts.ttFld.Name())
				if opts.withErr {
					g.Return(List(Qual(opts.tt.Path(), opts.tt.TypeName()).Block(), Qual("fmt", "Errorf").Call(List(Lit(errMessage+": %w")), Id(errName))))
				} else {
					g.Panic(Qual("fmt", "Sprintf").Call(List(Lit(errMessage+": %v")), Id(errName)))
				}
			})
		}
		ttAssign := Id(ttName).Op("=")
		if opts.ttIsPtr {
			ttAssign.Add(Op("&"))
		}
		ttAssign.Id("mapped" + ttName + "Enum")
	case ttIsBasic && srcIsBasic:
		if basicTt != basicSrc {
			log.Fatalf("different basic types source: %s target: %s", srcTypes[0].String(), ttTypes[0].String())
		}
		if opts.ttIsPtr {
			srcName = "&" + srcName
		} else if opts.srcIsPtr {
			srcName = "*" + srcName
		}
		g.Id(ttName).Op("=").Id(srcName)
	}
}

func typePairKey(s, t types.Type) string {
	return s.String() + "|" + t.String()
}

func buildTmpNames(tps []types.Type) []string {
	out := []string{}
	for _, t := range tps {
		if isPointer(t) {
			out = append(out, "Pointer")
			continue
		}
		if isSlice(t) {
			out = append(out, "Slice")
			continue
		}
		if isEnum(tps) {
			out = append(out, "Enum")
			break
		}
		if isStruct(tps) {
			out = append(out, "Struct")
			break
		}
	}
	return out
}

func trimTypeString(t types.Type) string {
	raw := t.String()
	for i, r := range raw {
		if unicode.IsLetter(r) {
			return raw[:i]
		}
	}
	return ""
}

func checkTypesChain(tt, src mapp.Mappable) error {
	ttChain := dropPointers(tt.FullType())
	srcChain := dropPointers(src.FullType())
	if len(ttChain) != len(srcChain) {
		return fmt.Errorf("chains length is different for target: %s and source: %s", tt.FullName(), src.FullName())
	}

	for i := 0; i < len(ttChain); i++ {
		if ttChain[i] != srcChain[i] {
			return fmt.Errorf("different types chain between target: %s and source: %s", tt.FullName(), src.FullName())
		}
	}

	return nil
}

func trucatePointer(tps []types.Type) ([]types.Type, bool) {
	if isPointer(tps[0]) {
		return tps[1:], true
	}
	return tps, false
}

func dropPointers(arr []types.Type) []types.Type {
	out := make([]types.Type, 0)
	for _, it := range arr {
		t, ok := it.(*types.Pointer)
		if ok {
			continue
		}
		out = append(out, t)
	}

	return out
}

func randString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"

	// Создаем срез байтов нужной длины
	b := make([]byte, length)

	// Инициализируем генератор случайных чисел текущим временем
	rand.Seed(uint64(time.Now().UnixNano()))

	// Заполняем срез случайными буквами из letters
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
