package gen

import (
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

func panicWithCause(g *Group, msg, errName string) {
	g.Panic(Qual("fmt", "Sprintf").Call(Lit(msg+": %v"), Id(errName).Dot("Error").Call()))
}

func assignToTarget(g *Group, fname string, fn func(*Statement)) {
	fn(g.Id("target").Dot(fname).Op("="))
}

func assignTo(g *Group, fn func(*Statement), names ...string) {
	fn(g.List(lo.Map(names, func(it string, _ int) Code { return Id(it) })...).Op(":="))
}

type AssignOpt struct {
	list []string
	g    *Group
}

func assign(g *Group) AssignOpt {
	return AssignOpt{g: g}
}

func (a AssignOpt) to(n string) AssignOpt {
	a.list = append(a.list, n)
	return a
}

func (a AssignOpt) from(fn func(*Statement)) {
	fn(a.g.List(lo.Map(a.list, func(it string, _ int) Code { return Id(it) })...).Op(":="))
}

func basicSource(trgt *Statement, sourceName string) {
	trgt.Id(sourceName)
}

func methodSource(trgt *Statement, methodName, sname string) {
	trgt.Id(methodName).Call(Id("src").Dot(sname))
}

func gParams(opts ...genOpts) genParams {
	genParams := genParams{}
	for _, opt := range opts {
		genParams = opt(genParams)
	}
	return genParams
}

func ifErrNotNil(g *Group, errName string, fn func(g *Group)) {
	g.If(Id(errName).Op("!=").Nil()).BlockFunc(fn)
}

func makeSlice(fn *Statement, typePath, typeName, fieldName string) {
	Make(
		Index().Qual(typePath, typeName),
		Lit(0),
		Len(Id("src").Dot(fieldName)))
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
			Range().Id("src").Dot(slcName).
			BlockFunc(fn)
	}
}

type ReturnOpts struct {
	pts     []struct{ path, typeName string }
	errName string
	g       *Group
}

func ret(g *Group) ReturnOpts {
	return ReturnOpts{g: g}
}

func (r ReturnOpts) DefaultVal(path, typeName string) ReturnOpts {
	r.pts = append(r.pts, struct {
		path     string
		typeName string
	}{path, typeName})
	return r
}

func (r ReturnOpts) Err(errName string) ReturnOpts {
	r.errName = errName
	return r
}

func (r ReturnOpts) build() {
	r.g.Return(lo.Map(r.pts, func(it struct{ path, typeName string }, _ int) Code { return Qual(it.path, it.typeName) })...)
}

func returnMapResult(g *Group, path, typeName, errName string, opts ...genOpts) {
	genParams := gParams(opts...)
	returns := []Code{Qual(path, typeName).Block()}
	switch {
	case genParams.withErr:
		returns = append(returns, Id(errName))
		g.Return(returns...)
	case genParams.withPanic:
		panicWithCause(g, genParams.panicMsg, errName)
	}
}

func mapErrName(fieldName string) string {
	return "map" + fieldName + "Err"
}

func targetSliceName(fieldName string) string {
	return "tt" + fieldName + "Slice"
}

func targetFieldName(fieldName string) string {
	return "tt" + fieldName
}
