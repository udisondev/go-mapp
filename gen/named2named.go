package gen

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

func namedToNamed(bl mapperBlock, s, t mapp.Field, opts ...genOpts) error {
	enmmap, ok := bl.enmMappers[fieldHash(s)][fieldHash(t)]
	if !ok {
		return fmt.Errorf("define @emapper please")
	}

	_, enmmapWithErr := enmmap.Errormsg()
	resVar := "enm" + t.Name()

	returns := []Code{}

	switch {
	case bl.mapper.WithError() && enmmapWithErr:
		errVar := "map" + t.Name() + "Err"
		bl.List(Id(resVar), Id(errVar)).Op(":=").Id(enmmap.Name()).Call(Id("src").Dot(s.Name()))
		bl.If(Id(errVar).Op("!=").Nil()).BlockFunc(func(g *Group) {
			if bl.mapperFunc.isRoot {
				returns = append(returns, Qual(bl.mapper.Target().Path(), bl.mapper.Target().TypeName()).Block())
			} else {
				returns  = append(returns, Qual(bl.mapperFunc.target.Type().Path(), bl.mapperFunc.target.Type().TypeName()).Block())
			}
			
			if enmmapWithErr {
				returns = append(returns, Id(errVar))
			}

			g.Return(returns...)
		})
		bl.Id("target").Dot(t.Name()).Op("=").Id(resVar)
	case enmmapWithErr:
		bl.List(Id(resVar), Id("err")).Op(":=").Id(enmmap.Name()).Call(Id("src").Dot(s.Name()))
		bl.If(Id("err").Op("!=").Nil()).Block(
			Panic(Qual("fmt", "Sprintf").Call(Lit(fmt.Sprintf("error map '%s%s'", bl.mapper.Target().TypeName(), t.FullName())+": %v"), Id("err").Dot("Error").Call())),
		)
		bl.Id("target").Dot(t.Name()).Op("=").Id(resVar)
	default:
		bl.Id("target").Dot(t.Name()).Op("=").Id(enmmap.Name()).Call(Id("src").Dot(s.Name()))
	}

	return nil
}
