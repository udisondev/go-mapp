package gen

import "github.com/udisondev/go-mapp/mapp"

type genOpts struct {
	ttIsPtr, srcIsPtr   bool
	ttFld, srcFld       mapp.Mappable
	tt, src             mapp.Mappable
	withErr             bool
	mapperName          string
	iterValName         func() string
	rulesBy             func(fieldFullName string) []mapp.Rule
	sourceFieldByTarget func(targetFullName string) (mapp.Mappable, bool)
}

type optFunc func(genOpts) genOpts

func WithRulesBy(fn func(fieldFullName string) []mapp.Rule) optFunc {
	return func(g genOpts) genOpts {
		g.rulesBy = fn
		return g
	}
}

func WithIterValName(fn func() string) optFunc {
	return func(g genOpts) genOpts {
		g.iterValName = fn
		return g
	}
}

func WithSourceFieldByTarget(fn func(targetFullName string) (mapp.Mappable, bool)) optFunc {
	return func(g genOpts) genOpts {
		g.sourceFieldByTarget = fn
		return g
	}
}

func WithTargetIsPointer(b bool) optFunc {
	return func(g genOpts) genOpts {
		g.ttIsPtr = b
		return g
	}
}

func WithErr(b bool) optFunc {
	return func(g genOpts) genOpts {
		g.withErr = b
		return g
	}
}

func WithMapperName(n string) optFunc {
	return func(g genOpts) genOpts {
		g.mapperName = n
		return g
	}
}

func WithSourceIsPointer(b bool) optFunc {
	return func(g genOpts) genOpts {
		g.srcIsPtr = b
		return g
	}
}

func gOpts(opts ...optFunc) genOpts {
	g := genOpts{}
	for _, o := range opts {
		g = o(g)
	}
	return g
}

func WithTargetField(ttFld mapp.Mappable) optFunc {
	return func(g genOpts) genOpts {
		g.ttFld = ttFld
		return g
	}
}

func WithSourceField(srcFld mapp.Mappable) optFunc {
	return func(g genOpts) genOpts {
		g.srcFld = srcFld
		return g
	}
}

func WithTarget(tt mapp.Mappable) optFunc {
	return func(g genOpts) genOpts {
		g.tt = tt
		return g
	}
}

func WithSource(src mapp.Mappable) optFunc {
	return func(g genOpts) genOpts {
		g.src = src
		return g
	}
}
