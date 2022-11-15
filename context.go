package djson

import "bytes"

type Context interface {
	Assign(varName []byte, val Value)
	ValueOf(name []byte) Value
	popMe()
	pushMe(val Value)
}

type Variable struct {
	Name  []byte
	Value Value
}

type ctx []Variable

var _ Context = &ctx{}

func NewContext(vars ...Variable) *ctx {
	ret := ctx(vars)
	return &ret
}

func (v *ctx) Assign(name []byte, val Value) {
	for i := range *v {
		if bytes.Equal((*v)[i].Name, name) {
			(*v)[i] = Variable{Name: name, Value: val}
			return
		}
	}
	*v = append(*v, Variable{Name: name, Value: val})
}

func (v *ctx) ValueOf(name []byte) Value {
	for i := range *v {
		if bytes.Equal((*v)[i].Name, name) {
			return (*v)[i].Value
		}
	}
	return Value{Type: ValueNull}
}

func (v *ctx) pushMe(val Value) {
	mk := []byte{'_', 'm', 'e'}
	me := v.ValueOf(mk)
	if me.Type != ValueNull {
		val.p = &me
	}
	v.Assign(mk, val)
}

func (v *ctx) popMe() {
	mk := []byte{'_'}
	me := v.ValueOf(mk)
	if me.p != nil {
		v.Assign(mk, *me.p)
	}
}
