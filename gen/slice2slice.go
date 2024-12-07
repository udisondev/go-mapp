package gen

import (
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func sliceToSlice(bl mapperBlock, s, t mapp.Field, opts ...genOpts) error {
	gParams := gParams(opts...)
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

		assign(bl.Group).
			to(targetSliceName(t.Name())).
			from(func(stm *Statement) { makeSlice(stm, t.Type().Path(), t.Type().TypeName(), t.Name()) })

		forr(bl.Group).put("_").put("it").rangForSlice(s.Name())(func(g *Group) {
			assign(g).
				to(targetFieldName(t.Name())).
				to(mapErrName(t.Name())).
				from(func(stm *Statement) { methodSource(stm, submapperName, "it") })
			ifErrNotNil(g, mapErrName(t.Name()), func(g *Group) {
				ret(g).
					DefaultVal(gParams.strPath, gParams.strType).
					Err(mapErrName(t.Name())).build()
			})
		})

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
