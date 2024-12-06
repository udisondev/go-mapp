package gen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func namedToNamed(bl mapperBlock, s, t mapp.Field) error {
	enmmap, ok := bl.enmMappers[fieldHash(s)][fieldHash(t)]
	if !ok {
		return fmt.Errorf("define @emapper please")
	}

	bl.Id("target").Dot(t.Name()).Op("=").Id(enmmap.Name()).Call(jen.Id("src").Dot(s.Name()))

	return nil
}
