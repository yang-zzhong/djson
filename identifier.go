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
	Call(scanner TokenScanner, vars *variables) (val Value, err error)
}

type identifier struct {
	name []byte
	p    Value
	vars *variables
}

func NewIdentifier(name []byte, vars *variables) *identifier {
	return &identifier{name: name, vars: vars}
}

func (id *identifier) SetParent(p Value) {
	id.p = p
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
	if vars, ok := root.(*variables); ok {
		vars.set(id.name, right)
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

func (id identifier) Call(scanner TokenScanner, vars *variables) (val Value, err error) {
	name := id.name
	if id.p.Type == ValueNull {
		err = fmt.Errorf("can't call function [%s] without caller", name)
		return
	}
	val = id.p.realValue()
	call, ok := val.Value.(callable)
	if !ok {
		err = fmt.Errorf("%s can't support call function", valueNames[val.Type])
		return
	}
	return call.call(string(name), val, scanner, vars)
}
