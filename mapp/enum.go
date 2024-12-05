package mapp

type EnumPair struct {
	source SourceEnum
	target TargetEnum
}

type SourceEnum struct {
	name string
	t Param
}

type TargetEnum struct {
	name string
	t Result
}

func (ep EnumPair) Source() SourceEnum {
	return ep.source
}

func (ep EnumPair) Target() TargetEnum {
	return ep.target
}

func (se SourceEnum) Name() string {
	return se.name
}

func (te TargetEnum) Name() string {
	return te.name
}

func (se SourceEnum) Path() string {
	return se.t.Path()
}

func (te TargetEnum) Path() string {
	return te.t.Path()
}

func (se SourceEnum) Type() string {
	_, t := se.t.Type()
	return t
}

func (te TargetEnum) Type() string {
	_, t := te.t.Type()
	return t
}