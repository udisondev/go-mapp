package gen

import (
	"fmt"
	"math/rand/v2"
	"time"

	. "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
	"github.com/udisondev/go-mapp/mapp"
)

func genRandomName(length int) string {
	seed := time.Now().UnixNano()

	src := rand.NewPCG(uint64(seed), uint64(seed>>32))
	r := rand.New(src)

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[r.IntN(len(charset))]
	}
	return string(result)

}

func fieldsHash(fs ...mapp.Field) string {
	var hash string
	for _, f := range fs {
		hash += fieldHash(f)
	}

	return hash
}

func fieldHash(f mapp.Field) string {
	return f.Type().Path() + "." + f.Type().TypeName()
}

func enumHash(f mapp.Enum) string {
	_, typeName := f.Type()
	return f.Path() + "." + typeName
}

func defErrWrapMsg(src, tt mapp.Field) string {
	return fmt.Sprintf(
		"error mapping from '%s.%s' to '%s.%s'",
		src.Type().TypeName(),
		src.Name(),
		tt.Type().TypeName(),
		tt.Name())
}

func defPanic(g *Group, src, tt mapp.Field, errName string) {
	g.Panic(Qual("fmt", "Sprintf").Call(Lit(defErrWrapMsg(src, tt)+": %v"), Id(errName).Dot("Error").Call()))
}

type AssignOpt struct {
	isNew bool
	list  []string
	g     *Group
}

func assign(g *Group) AssignOpt {
	return AssignOpt{g: g}
}

func appnd(g *Group, sliceName, val string) {
	g.Id(sliceName).Op("=").Append(Id(sliceName), Id(val))
}

func (a AssignOpt) emptyStruct(path, typeName string) {
	a.g.Id(a.list[0]).Op(a.op()).Qual(path, typeName).Block()
}

func (a AssignOpt) op() string {
	if a.isNew {
		return ":="
	}
	return "="
}

func (a AssignOpt) new(n string) AssignOpt {
	a.isNew = true
	a.list = append(a.list, n)
	return a
}

func (a AssignOpt) toTarget(ttName string, fn func(*Statement)) {
	fn(a.g.Id(ttFldName(ttName)).Op("="))
}

func ttFldName(ttName string) string {
	return "target." + ttName
}

func srcFldName(ttName string) string {
	return "src." + ttName
}

func (a AssignOpt) from(fn func(*Statement)) {
	fn(a.g.List(lo.Map(a.list, func(it string, _ int) Code { return Id(it) })...).Op(a.op()))
}

func basicSource(stm *Statement, srcName string, opts ...genOptFunc) {
	if gParams(opts...).srcIsPtr {
		stm.Add(Op("*"))
	}
	if gParams(opts...).ttIsPtr {
		stm.Add(Op("&"))
	}
	stm.Id(srcName)
}

func methodSource(tt *Statement, methodName, srcName string, opts ...genOptFunc) {
	gp := gParams(opts...)
	fldName := "*" + srcName
	if !gp.srcIsPtr {
		fldName = srcName
	}
	tt.Id(methodName).Call(Id(fldName))
}

func gParams(opts ...genOptFunc) genParams {
	genParams := genParams{}
	for _, opt := range opts {
		genParams = opt(genParams)
	}
	return genParams
}

func ifErrNotNil(g *Group, errName string, fn func(g *Group)) {
	g.If(Id(errName).Op("!=").Nil()).BlockFunc(fn)
}

func ifSrcNotNil(g *Group, fname string, fn func(g *Group)) {
	g.If(Id(srcFldName(fname)).Op("!=").Nil()).BlockFunc(fn)
}

func makeSlice(stm *Statement, typePath, typeName, fldname string) {
	stm.Make(
		Index().Qual(typePath, typeName),
		Lit(0),
		Len(Id("src").Dot(fldname)))
}

type forOpts struct {
	vars []string
	g    *Group
}

func forr(g *Group) forOpts {
	return forOpts{g: g}
}

func (f forOpts) put(v string) forOpts {
	f.vars = append(f.vars, v)
	return f
}

func (f forOpts) rangForSlice(slcName string) func(fn func(*Group)) {
	return func(fn func(g *Group)) {
		f.g.For(List(lo.Map(f.vars, func(it string, _ int) Code { return Id(it) })...)).
			Op(":=").
			Range().Id("src").Dot(slcName).
			BlockFunc(fn)
	}
}

type ReturnOpts struct {
	pathTypes []struct {
		isEmpty        bool
		path, typeName string
	}
	errWrapMsg string
	errName    string
	g          *Group
}

func retrn(g *Group) ReturnOpts {
	return ReturnOpts{g: g}
}

func (r ReturnOpts) emptyStruct(path, typeName string) ReturnOpts {
	r.pathTypes = append(r.pathTypes, struct {
		isEmpty  bool
		path     string
		typeName string
	}{true, path, typeName})
	return r
}

func (r ReturnOpts) defErr(src, tt mapp.Field, errName string) ReturnOpts {
	r.errName = errName
	r.errWrapMsg = defErrWrapMsg(src, tt)
	return r
}

func (r ReturnOpts) Val(name string) ReturnOpts {
	r.pathTypes = append(r.pathTypes, struct {
		isEmpty  bool
		path     string
		typeName string
	}{false, "", name})
	return r
}

func (r ReturnOpts) Nil() {
	r.g.Return(
		append(
			lo.Map(r.pathTypes, func(it struct {
				isEmpty        bool
				path, typeName string
			}, _ int) Code {
				ret := Qual(it.path, it.typeName)
				if it.isEmpty {
					ret.Block()
				}

				return ret
			}),
			Nil(),
		)...)
}

func (r ReturnOpts) build() {
	returns := lo.Map(r.pathTypes, func(it struct {
		isEmpty        bool
		path, typeName string
	}, _ int) Code {
		ret := Qual(it.path, it.typeName)
		if it.isEmpty {
			ret.Block()
		}

		return ret
	})

	switch {
	case r.errName != "" && r.errWrapMsg != "":
		returns = append(returns, Qual("fmt", "Errorf").Call(Lit(r.errWrapMsg+": %w"), Id(r.errName)))
	case r.errName != "":
		returns = append(returns, Id(r.errName))
	}
	r.g.Return(returns...)
}

func mapErrName(fieldName string) string {
	return "map" + fieldName + "Err"
}

func ttSliceVar(fieldName string) string {
	return "tt" + fieldName + "Slice"
}

func ttVar(fieldName string) string {
	return "tt" + fieldName
}
