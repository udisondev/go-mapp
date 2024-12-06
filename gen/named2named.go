package gen

import (
	"log"

	"github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func namedToNamed(bl mapperBlock, s, t mapp.Field) {
	enmmap, ok := bl.enmMappers[fieldHash(s)][fieldHash(t)]
	if !ok {
		log.Fatalf("Cold not map from %s to %s. Define @emapper please", s.FullName(), t.FullName())
	}

	bl.Id("target").Dot(t.Name()).Op("=").Id(enmmap.Name()).Call(jen.Id("src").Dot(s.Name()))
}
