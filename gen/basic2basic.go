package gen

import "github.com/udisondev/go-mapp/mapp"

func basicToBasic(bl mapperBlock, s, t mapp.Field) error {
	bl.Id("target").Dot(t.Name()).Op("=").Id("src").Dot(s.Name())

	return nil
}
