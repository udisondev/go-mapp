package gen

import (
	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func basicToBasic(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	assign(g).toTarget(tt.Name(), func(stm *Statement) {
		basicSource(stm, srcFldName(src.Name()))
	})

	return nil
}
