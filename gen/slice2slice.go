package gen

import (
	"fmt"
	"log"

	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func sliceToSlice(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	srcSlc, ok := src.Type().(mapp.SliceType)
	if !ok {
		panic("is not a slice")
	}

	ttSlc, ok := tt.Type().(mapp.SliceType)
	if !ok {
		panic("is not a slice")
	}

	switch {
	case srcSlc.Elem().TypeFamily() == mapp.FieldTypeNamed &&
		ttSlc.Elem().TypeFamily() == mapp.FieldTypeNamed:
		namedToNamed(g, src, tt, opts...)
	case srcSlc.Elem().TypeFamily() == mapp.FieldTypeBasic &&
		ttSlc.Elem().TypeFamily() == mapp.FieldTypeBasic:
		basicToBasic(g, src, tt, opts...)
	case srcSlc.Elem().TypeFamily() == mapp.FieldTypeStruct &&
		ttSlc.Elem().TypeFamily() == mapp.FieldTypeStruct:
		gp := gParams(opts...)

		hash := fieldsHash(src, tt)
		submapperName, submapperExists := gp.submappers[hash]
		if !submapperExists {
			submapperName = genRandomName(10)
			gp.submappers[hash] = submapperName
		}

		assign(g).
			new(ttSliceVar(tt.Name())).
			from(func(stm *Statement) { makeSlice(stm, tt.Path(), tt.TypeName(), tt.Name()) })

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
				sourcePath(src.Path()),
				sourceType(src.TypeName()),
				targetPath(tt.Path()),
				targetType(tt.TypeName()),
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
	default:
		panic(fmt.Sprintf("unsupported case: src '%s' tt '%s'", srcSlc.Elem().TypeFamily(), ttSlc.Elem().TypeFamily()))
	}

	return nil

}
