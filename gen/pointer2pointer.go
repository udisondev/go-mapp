package gen

import (
	// "fmt"
	
	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func pointerToPointer(g *Group, src, tt mapp.Field, opts ...genOptFunc) error {
	// srcPtr, ok := src.Type().(mapp.PointerType)
	// if !ok {
	// 	panic("is not a pointer")
	// }

	// ttPtr, ok := tt.Type().(mapp.PointerType)
	// if !ok {
	// 	panic("is not a pointer")
	// }

	// switch {
	// case srcPtr.Elem().TypeFamily() == mapp.FieldTypeNamed &&
	// 	ttPtr.Elem().TypeFamily() == mapp.FieldTypeNamed:
	// 	namedToNamed(g, src, tt, opts...)
	// case srcPtr.Elem().TypeFamily() == mapp.FieldTypeBasic &&
	// 	ttPtr.Elem().TypeFamily() == mapp.FieldTypeBasic:
	// 	basicToBasic(g, src, tt, opts...)
	// case srcPtr.Elem().TypeFamily() == mapp.FieldTypeStruct &&
	// 	ttPtr.Elem().TypeFamily() == mapp.FieldTypeStruct:
	// 	structToStruct(g, src, tt, opts...)
	// default:
	// 	panic(fmt.Sprintf("unsupported case: src '%s' tt '%s'", srcPtr.Elem().TypeFamily(), ttPtr.Elem().TypeFamily()))
	// }

	return nil
}
