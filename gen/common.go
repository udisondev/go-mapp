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

func panicWithCause(msg, errName string) Code {
	return Panic(Qual("fmt", "Sprintf").Call(Lit(msg+": %v"), Id(errName).Dot("Error").Call()))
}

func assignToTarget(g *Group, fname string) *Statement {
	return g.Id("target").Dot(fname).Op("=")
}

func assignTo(g *Group, names ...string) *Statement {
	return g.List(lo.Map(names, func(it string, _ int) Code { return Id(it) })...).Op(":=")
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

func ifErrNotNil(g *Group, errName string) *Statement {
	return g.If(Id(errName).Op("!=").Nil())
}

func returnMapResult(stmnt *Statement, path, typeName, errName string, opts ...genOpts) {
	genParams := gParams(opts...)
	returns := []Code{Qual(path, typeName).Block()}
	switch {
	case genParams.withErr:
		returns = append(returns, Id(errName))
		stmnt.Block(Return(returns...))
	case genParams.withPanic:
		stmnt.Block(panicWithCause(genParams.panicMsg, errName))
	}
}
