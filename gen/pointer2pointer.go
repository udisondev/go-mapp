package gen

import (
	"fmt"

	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func pointerToPointer(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	fmt.Printf("%s to %s has no mapper\n", src.FullName(), tt.FullName())
	// TODO доделать
	// spt, ok := s.Type().(mapp.PointerType)
	// if !ok {
	// 	panic("is not a pointer")
	// }
	// tpt, ok := t.Type().(mapp.PointerType)
	// if !ok {
	// 	panic("is not a pointer")
	// }

	// switch {
	// case spt.Elem().TypeFamily() == mapp.FieldTypeBasic && tpt.Elem().TypeFamily() == mapp.FieldTypeBasic:
	// 	basicToBasic(bl, s, t)
	// }
	// bl.If(
	// 	jen.Id("src").Dot(s.Name()).Op("!=").Nil(),
	// ).BlockFunc(
	// 	func(g *jen.Group) {
	// 		bl.Group = g
	// 		structToStruct(bl, s, t)
	// 	},
	// )

	return nil
}
