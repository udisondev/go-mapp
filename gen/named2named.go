package gen

import (
	"fmt"

	"github.com/udisondev/go-mapp/mapp"
)

func namedToNamed(bl mapperBlock, s, t mapp.Field, opts ...genOpts) error {
	genParams := gParams(opts...)

	enmmap, ok := bl.enmMappers[fieldHash(s)][fieldHash(t)]
	if !ok {
		return fmt.Errorf("define @emapper please")
	}

	enmOpts := []genOpts(opts)
	_, enmmapWithErr := enmmap.Errormsg()
	if enmmapWithErr && !genParams.withErr {
		enmOpts = append(enmOpts, withPanic(fmt.Sprintf("error map '%s%s'", bl.mapper.Target().TypeName(), t.FullName())))
	}

	path := bl.mapper.Target().Path()
	typeName := bl.mapper.Target().TypeName()
	resVar := "enm" + t.Name()
	errVar := "map" + t.Name() + "Err"

	if !genParams.isTargetStrct {
		path = bl.mapperFunc.target.Type().Path()
		typeName = bl.mapperFunc.target.Type().TypeName()
	}

	methodSource(assignTo(bl.Group, resVar, errVar), enmmap.Name(), s.Name())
	returnMapResult(ifErrNotNil(bl.Group, errVar), path, typeName, errVar, enmOpts...)
	basicSource(assignToTarget(bl.Group, t.Name()), resVar)

	return nil
}
