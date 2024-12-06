package gen

import (
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func sliceToSlice(bl mapperBlock, s, t mapp.Field, opts ...genOpts) error {

	sslice, ok := s.Type().(mapp.SliceType)
	if !ok {
		panic("is not a slice")
	}

	tslice, ok := t.Type().(mapp.SliceType)
	if !ok {
		panic("is not a slice")
	}

	switch {
	case sslice.Elem().TypeFamily() == mapp.FieldTypeBasic &&
		tslice.Elem().TypeFamily() == mapp.FieldTypeBasic:
		basicToBasic(bl, s, t)
	case sslice.Elem().TypeFamily() == mapp.FieldTypeStruct &&
		tslice.Elem().TypeFamily() == mapp.FieldTypeStruct:
		hash := fieldsHash(s, t)
		submapperName, submapperExists := bl.submappers[hash]
		if !submapperExists {
			submapperName = genRandomName(10)
			bl.submappers[hash] = submapperName
		}

		targetSliceName := "target" + t.Name() + "Slice"
		targetTypePath := t.Type().Path()
		bl.Id(targetSliceName).
			Op(":=").
			Make(
				Index().Qual(targetTypePath, t.Type().TypeName()),
				Lit(0),
				Len(Id("src").Dot(t.Name())))
		bl.
			For(
				List(Id("_"), Id("it")).Op(":=").Range().Id("src").Dot(s.Name()),
			).
			BlockFunc(func(g *Group) {
				errVar := "map" + t.Name() + "err"
				resVar := "target" + t.Name()
				g.List(Id(resVar), Id(errVar)).Op(":=").Id(submapperName).Call(Id("it"))
				g.If(Id(errVar)).Op("!=").Nil().
					BlockFunc(func(g *Group) {
						returns := []Code{}
						if bl.mapperFunc.isRoot {
							returns = append(returns, Qual(bl.mapper.Target().Path(), bl.mapper.Target().TypeName()).Block())
							if bl.mapper.WithError() {
								returns = append(returns, Id(errVar))
							}
						} else {
							returns = append(returns, Qual(bl.mapperFunc.target.Type().Path(), bl.mapperFunc.target.Type().TypeName()))
							returns = append(returns, Id(errVar))
						}
						g.Return(returns...)

					})
				g.Id(targetSliceName).Op("=").Append(Id(targetSliceName), Id(resVar))
			})

		bl.Id("target").Dot(t.Name()).Op("=").Id(targetSliceName)

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
	}

	return nil

}
