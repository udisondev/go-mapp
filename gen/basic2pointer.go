package gen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func basicToPointer(bl mapperBlock, s, t mapp.Field) error {
	pt, ok := t.Type().(mapp.PointerType)
	if !ok {
		panic("is not a pointer")
	}

	if pt.Elem().TypeFamily() != mapp.FieldTypeBasic {
		panic("source refers to not basic")
	}

	if s.Type().TypeFamily() != pt.Elem().TypeFamily() {
		return fmt.Errorf(
			"could not mapp different types source: '%s' target: pointer to %s",
			s.Type().TypeFamily(),
			pt.Elem().TypeFamily())
	}

	bl.Id("target").Dot(t.Name()).Op("=").Add(jen.Op("&")).Id("src").Dot(s.Name())

	return nil
}
