package gen

import (
	"fmt"

	"github.com/udisondev/go-mapp/mapp"
)

func structToPointer(bl mapperBlock, s, t mapp.Field) error{
	fmt.Printf("%s to %s has no mapper", s.FullName(), t.FullName())

	return nil
}
