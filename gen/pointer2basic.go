package gen

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func pointerToBasic(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	pt, ok := src.Type().(mapp.PointerType)
	if !ok {
		panic("is not a pointer")
	}

	if tt.Type().TypeName() != src.Type().TypeName() {
		return fmt.Errorf(
			"could not mapp different types source: '*%s' target: %s",
			pt.Elem().TypeFamily(),
			tt.Type().TypeFamily())
	}

	ifSrcNotNil(g, src.Name(), func(g *Group) {
		assign(g).toTarget(tt.Name(), func(stm *Statement) {
			basicSource(stm, srcFldName(src.Name()), append(opts, srcIsPtr(true), ttIsPtr(false))...)
		})
	})

	return nil
}
