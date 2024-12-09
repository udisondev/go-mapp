package gen

import (
	"fmt"

	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func basicToPointer(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	pt, ok := tt.Type().(mapp.PointerType)
	if !ok {
		panic("is not a pointer")
	}

	if pt.Elem().TypeFamily() != mapp.FieldTypeBasic {
		panic("source refers to not basic")
	}

	if src.Type().TypeFamily() != pt.Elem().TypeFamily() {
		return fmt.Errorf(
			"could not mapp different types source: '%s' target: pointer to %s",
			src.Type().TypeFamily(),
			pt.Elem().TypeFamily())
	}

	assign(g).toTarget(tt.Name(), func(stm *Statement) {
		basicSource(stm, srcFldName(src.Name()), append(opts, ttIsPtr(true), srcIsPtr(false))...)
	})

	return nil
}
