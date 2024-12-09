package gen

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func namedToNamed(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	gp := gParams(opts...)

	enmmap, ok := gp.enmMappers[fieldHash(src)][fieldHash(tt)]
	if !ok {
		return fmt.Errorf("define @emapper please")
	}

	_, enmmapWithErr := enmmap.Errormsg()

	switch {
	case enmmapWithErr:
		assign(g).
			new(ttVar(tt.Name())).
			new(mapErrName(tt.Name())).
			from(func(stmnt *Statement) {
				methodSource(stmnt, enmmap.Name(), srcFldName(src.Name()))
			})

		ifErrNotNil(g, mapErrName(tt.Name()), func(g *Group) {

			switch {
			case gp.withErr && enmmapWithErr:
				retrn(g).
					emptyStruct(gp.ttPath, gp.ttType).
					defErr(src, tt, mapErrName(tt.Name())).
					build()
			case enmmapWithErr:
				defPanic(g, src, tt, mapErrName(tt.Name()))
			}
		})

		assign(g).toTarget(tt.Name(), func(stm *Statement) {
			basicSource(stm, ttVar(tt.Name()))
		})
	default:
		assign(g).toTarget(tt.Name(), func(stm *Statement) {
			methodSource(stm, enmmap.Name(), srcFldName(src.Name()))
		})
	}

	return nil
}
