package gsp

import "reflect"

type Atom interface {
	Assign(reflect.Type, reflect.Value)
}
type IndexAtom struct {
	Field int
	Data  any
}

func (i IndexAtom) Assign(_ reflect.Type, v reflect.Value) {
	v.Field(i.Field).Set(reflect.ValueOf(i.Data))
}

type NameAtom struct {
	FieldName string
	Data      any
}

func (n NameAtom) Assign(t reflect.Type, v reflect.Value) {
	v.FieldByName(n.FieldName).Set(reflect.ValueOf(n.Data))
}
