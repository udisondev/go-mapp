package gen

import (

	//lint:ignore ST1001 it's ok
	. "github.com/dave/jennifer/jen"
	"github.com/udisondev/go-mapp/mapp"
)

type genParams struct {
	srcIsPtr, ttIsPtr bool
	file              *File
	withErr           bool
	srcPath, ttPath   string
	srcType, ttType   string
	submappers        map[string]string
	enmMappers        map[string]map[string]mapp.EnumMapper
	fldMapFuncs       map[mapp.TypeFamily]map[mapp.TypeFamily]fldMapFunc
	ttFields          []mapp.Mappable
	mapper            mapp.Mapper
}

type genOptFunc func(genParams) genParams

func file(f *File) genOptFunc {
	return func(gp genParams) genParams {
		gp.file = f
		return gp
	}
}

func enumMappers(m map[string]map[string]mapp.EnumMapper) genOptFunc {
	return func(gp genParams) genParams {
		gp.enmMappers = m
		return gp
	}
}

func srcIsPtr(b bool) genOptFunc {
	return func(gp genParams) genParams {
		gp.srcIsPtr = b
		return gp
	}
}

func ttIsPtr(b bool) genOptFunc {
	return func(gp genParams) genParams {
		gp.ttIsPtr = b
		return gp
	}
}

func submappers(m map[string]string) genOptFunc {
	return func(gp genParams) genParams {
		gp.submappers = m
		return gp
	}
}

func fldMapFuncs(m map[mapp.TypeFamily]map[mapp.TypeFamily]fldMapFunc) genOptFunc {
	return func(gp genParams) genParams {
		gp.fldMapFuncs = m
		return gp
	}
}

func withErr(b bool) genOptFunc {
	return func(gp genParams) genParams {
		gp.withErr = b
		return gp
	}
}

func targetPath(path string) genOptFunc {
	return func(gp genParams) genParams {
		gp.ttPath = path
		return gp
	}
}

func sourcePath(path string) genOptFunc {
	return func(gp genParams) genParams {
		gp.srcPath = path
		return gp
	}
}

func ttFields(fields []mapp.Mappable) genOptFunc {
	return func(gp genParams) genParams {
		gp.ttFields = fields
		return gp
	}
}

func mapper(m mapp.Mapper) genOptFunc {
	return func(gp genParams) genParams {
		gp.mapper = m
		return gp
	}
}

func targetType(ttType string) genOptFunc {
	return func(gp genParams) genParams {
		gp.ttType = ttType
		return gp
	}
}

func sourceType(t string) genOptFunc {
	return func(gp genParams) genParams {
		gp.srcType = t
		return gp
	}
}
