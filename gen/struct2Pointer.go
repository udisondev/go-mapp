package gen

import (
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func structToPointer(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	structToStruct(g, src, tt, append(opts, ttIsPtr(true), srcIsPtr(false))...)

	return nil
}
