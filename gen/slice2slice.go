package gen

import (
	"log"

	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func sliceToSlice(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	sslice, ok := src.Type().(mapp.SliceType)
	if !ok {
		panic("is not a slice")
	}

	tslice, ok := tt.Type().(mapp.SliceType)
	if !ok {
		panic("is not a slice")
	}

	switch {
	case sslice.Elem().TypeFamily() == mapp.FieldTypeBasic &&
		tslice.Elem().TypeFamily() == mapp.FieldTypeBasic:
		basicToBasic(g, src, tt, opts...)
	case sslice.Elem().TypeFamily() == mapp.FieldTypeStruct &&
		tslice.Elem().TypeFamily() == mapp.FieldTypeStruct:
		gp := gParams(opts...)

		hash := fieldsHash(src, tt)
		submapperName, submapperExists := gp.submappers[hash]
		if !submapperExists {
			submapperName = genRandomName(10)
			gp.submappers[hash] = submapperName
		}

		assign(g).
			new(ttSliceVar(tt.Name())).
			from(func(stm *Statement) { makeSlice(stm, tt.Type().Path(), tt.Type().TypeName(), tt.Name()) })

		forr(g).put("_").put("it").rangForSlice(src.Name())(func(g *Group) {
			assign(g).
				new(ttVar(tt.Name())).
				new(mapErrName(tt.Name())).
				from(func(stm *Statement) { methodSource(stm, submapperName, "it") })
			ifErrNotNil(g, mapErrName(tt.Name()), func(g *Group) {
				switch {
				case gp.withErr:
					retrn(g).
						emptyStruct(gp.ttPath, gp.ttType).
						defErr(src, tt, mapErrName(tt.Name())).
						build()
				default:
					defPanic(g, src, tt, mapErrName(tt.Name()))
				}

			})
			appnd(g, ttSliceVar(tt.Name()), ttVar(tt.Name()))
		})

		assign(g).toTarget(tt.Name(), func(stm *Statement) { basicSource(stm, ttSliceVar(tt.Name())) })

		if !submapperExists {
			opts = append(opts,
				sourcePath(src.Type().Path()),
				sourceType(src.Type().TypeName()),
				targetPath(tt.Type().Path()),
				targetType(tt.Type().TypeName()),
				ttFields(tt.Fields()),
				withErr(true),
			)
			err := generateMapper(
				submapperName,
				opts...,
			)
			if err != nil {
				log.Fatalf("Error generate mapper '%s': %v", submapperName, err)
			}
		}
	}

	return nil

}
