// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package mapp

import (
	"errors"
	"fmt"
)

const (
	// FieldTypeBasic is a FieldType of type Basic.
	FieldTypeBasic TypeFamily = iota
	// FieldTypeNamed is a FieldType of type Named.
	FieldTypeNamed
	// FieldTypeStruct is a FieldType of type Struct.
	FieldTypeStruct
	// FieldTypePointer is a FieldType of type Pointer.
	FieldTypePointer
	// FieldTypeSlice is a FieldType of type Slice.
	FieldTypeSlice
)

var ErrInvalidFieldType = errors.New("not a valid FieldType")

const _FieldTypeName = "basicnamedstructpointerslice"

var _FieldTypeMap = map[TypeFamily]string{
	FieldTypeBasic:   _FieldTypeName[0:5],
	FieldTypeNamed:   _FieldTypeName[5:10],
	FieldTypeStruct:  _FieldTypeName[10:16],
	FieldTypePointer: _FieldTypeName[16:23],
	FieldTypeSlice:   _FieldTypeName[23:28],
}

// String implements the Stringer interface.
func (x TypeFamily) String() string {
	if str, ok := _FieldTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("FieldType(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x TypeFamily) IsValid() bool {
	_, ok := _FieldTypeMap[x]
	return ok
}

var _FieldTypeValue = map[string]TypeFamily{
	_FieldTypeName[0:5]:   FieldTypeBasic,
	_FieldTypeName[5:10]:  FieldTypeNamed,
	_FieldTypeName[10:16]: FieldTypeStruct,
	_FieldTypeName[16:23]: FieldTypePointer,
	_FieldTypeName[23:28]: FieldTypeSlice,
}

// ParseFieldType attempts to convert a string to a FieldType.
func ParseFieldType(name string) (TypeFamily, error) {
	if x, ok := _FieldTypeValue[name]; ok {
		return x, nil
	}
	return TypeFamily(0), fmt.Errorf("%s is %w", name, ErrInvalidFieldType)
}
