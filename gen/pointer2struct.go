package gen

import (

	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func pointerToStruct(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	ifSrcNotNil(g, src.Name(), func(g *Group) {
		structToStruct(g, src, tt, append(opts, srcIsPtr(true), ttIsPtr(false))...)
	})

	return nil
}
