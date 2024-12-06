package gen

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func structToStruct(bl mapperBlock, s, t mapp.Field, opts ...genOpts) error {
	genParam := genParams{}

	returns := []Code{}
	for _, fn := range opts {
		genParam = fn(genParam)
	}

	hash := fieldsHash(s, t)
	submapperName, submapperExists := bl.submappers[hash]
	if !submapperExists {
		submapperName = genRandomName(10)
		bl.submappers[hash] = submapperName
	}
	
	errVar := "map" + t.Name() + "Err"
	resVar := "target" + t.Name()
	var srcVar Code
	if genParam.srcIsPointer {
		srcVar = Add(Op("*")).Id("src").Dot(s.Name())
	} else {
		srcVar = Id("src").Dot(s.Name())
	}
	bl.List(Id(resVar), Id(errVar)).Op(":=").Id(submapperName).Call(srcVar)
	bl.If(Id(errVar)).Op("!=").Nil().BlockFunc(func(g *Group) {
		withErr := bl.mapper.WithError()
		if !withErr {
			g.Panic(Qual("fmt", "Sprintf").Call(Lit(fmt.Sprintf("error map '%s%s'", bl.mapper.Target().TypeName(), t.FullName())+": %v"), Id(errVar).Dot("Error").Call()))
			return
		}

		if bl.mapperFunc.isRoot {
			returns = append(returns, Qual(bl.mapper.Target().Path(), bl.mapper.Target().TypeName()).Block())
			if withErr {
				returns = append(returns, Id(errVar))
			}
		} else {
			returns = append(returns, Qual(bl.mapperFunc.target.Type().Path(), bl.mapperFunc.target.Type().TypeName()))
			returns = append(returns, Id(errVar))
		}

		g.Return(returns...)
	})

	bl.Id("target").Dot(t.Name()).Op("=").Id(resVar)

	if !submapperExists {
		mfn := mapperFunc{
			generatedFn:        bl.file.Func().Id(submapperName),
			mapper:             bl.mapper,
			file:               bl.mapperFunc.file,
			source:             s,
			target:             t,
			submappers:         bl.submappers,
			fieldMapGenerators: bl.fieldMapGenerators,
		}
		bl.file.Line()
		mfn.generateSignature()
		err := mfn.generateBlock()
		if err != nil {
			return err
		}
	}

	return nil
}
