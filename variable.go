package djson

import "bytes"

type variable struct {
	name  []byte
	value value
}

type variables []variable

func newVariables() *variables {
	return &variables{}
}

func (v *variables) set(name []byte, val value) {
	for i := range *v {
		if bytes.Equal((*v)[i].name, name) {
			(*v)[i] = variable{name: name, value: val}
			return
		}
	}
	*v = append(*v, variable{name: name, value: val})
}

func (v *variables) get(name []byte) *value {
	for i := range *v {
		if bytes.Equal((*v)[i].name, name) {
			return &(*v)[i].value
		}
	}
	return nil
}

func (v *variables) pushMe(val value) {
	mk := []byte{'_', 'm', 'e'}
	me := v.get(mk)
	if me == nil {
		*v = append(*v, variable{name: mk, value: val})
		return
	}
	val.p = me
	v.set(mk, val)
}

func (v *variables) popMe() {
	mk := []byte{'_'}
	me := v.get(mk)
	if me == nil {
		return
	}
	if me.p == nil {
		return
	}
	v.set(mk, *me.p)
}
