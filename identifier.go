package djson

import (
	"errors"
	"fmt"
	"strconv"
)

type Identifier interface {
	Value() Value
	Assign(val Value) error
	SetParent(p Value)
	Call(scanner TokenScanner, vars Context) (val Value, err error)
}

type identifier struct {
	name []byte
	p    Value
	vars Context
}

func NewIdentifier(name []byte, ctx Context) *identifier {
	return &identifier{name: name, vars: ctx}
}

func (id *identifier) SetParent(p Value) {
	id.p = p
}

func (id *identifier) String() string {
	n := id.name
	tmp := id.p
	for tmp.Type == ValueIdentifier {
		t := append(tmp.Value.(*identifier).name, '.')
		n = append(t, n...)
	}
	if tmp.Type == ValueNull {
		return string(n)
	}
	return tmp.String() + "." + string(n)
}

func (id identifier) Value() Value {
	root, err := id.root()
	if err != nil {
		return Value{Type: ValueNull}
	}
	if lookuper, ok := root.(lookuper); ok {
		return lookuper.lookup(id.name)
	}
	return Value{Type: ValueNull}
}

func (id identifier) Assign(right Value) error {
	root, err := id.root()
	if err != nil {
		return err
	}
	if vars, ok := root.(*ctx); ok {
		vars.Assign(id.name, right)
		return nil
	} else if obj, ok := root.(Object); ok {
		obj.Set(id.name, right)
	} else if arr, ok := root.(Array); ok {
		idx, err := strconv.Atoi(string(id.name))
		if err != nil {
			return err
		}
		arr.Set(idx, right)
	} else {
		return errors.New("can't support assign")
	}
	return nil
}

func (id identifier) root() (root interface{}, err error) {
	dots := []byte{}
	tmp := &id
	for tmp.p.Type == ValueIdentifier {
		tmp = tmp.p.Value.(*identifier)
		n := tmp.name
		if len(dots) > 0 {
			n = append(n, '.')
		}
		dots = append(n, dots...)
	}
	root = id.vars
	if tmp.p.Type != ValueNull {
		root = tmp.p.Value
	}
	if len(dots) == 0 {
		return
	}
	if lookuper, ok := root.(lookuper); ok {
		val := lookuper.lookup(dots)
		if val.Type == ValueNull {
			err = errors.New("can't find root")
			return
		}
		root = val.Value
		return
	}
	err = errors.New("can't find root")
	return
}

func (id identifier) Call(scanner TokenScanner, ctx Context) (val Value, err error) {
	name := id.name
	if id.p.Type == ValueNull {
		err = fmt.Errorf("can't call function [%s] without caller", name)
		return
	}
	val = id.p.RealValue()
	call, ok := val.Value.(Callable)
	if !ok {
		err = fmt.Errorf("%s can't support call function", val.TypeName())
		return
	}
	return call.call(string(name), val, scanner, ctx)
}
