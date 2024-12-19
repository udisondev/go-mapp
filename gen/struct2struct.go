package gen

import (
	"log"

	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func structToStruct(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	gp := gParams(opts...)

	hash := fieldsHash(src, tt)
	submapperName, submapperExists := gp.submappers[hash]
	if !submapperExists {
		submapperName = genRandomName(10)
		gp.submappers[hash] = submapperName
	}

	assign(g).
		new(ttVar(tt.Name())).
		new(mapErrName(tt.Name())).
		from(func(stm *Statement) { methodSource(stm, submapperName, srcFldName(src.Name()), opts...) })
	opts = append(opts, srcIsPtr(false))
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

	assign(g).toTarget(tt.Name(), func(stm *Statement) { basicSource(stm, ttVar(tt.Name()), opts...) })

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

	return nil
}
