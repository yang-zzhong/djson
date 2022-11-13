package djson

import (
	"bytes"
	"errors"
)

type Object interface {
	Set(k []byte, val Value)
	Get(k []byte) Value
	Has(k []byte) bool
	Del(k []byte)
	Copy() Object
	Each(func(k []byte, val Value) bool)
	Total() int
}

type object struct {
	pairs []*pair
	*callableImp
}

type pair struct {
	key []byte
	val Value
}

var _ Object = &object{}

func NewObject(pairs ...*pair) Object {
	obj := &object{pairs: pairs, callableImp: newCallable("object")}
	obj.register("set", setObject)
	obj.register("replace", replaceObject)
	obj.register("del", delObject)
	obj.register("get", getObject)
	return obj
}

func setObject(val Value, nexter TokenScanner, vars *variables) (ret Value, err error) {
	o := val.Value.(Object)
	r := NewObject()
	err = eachObjectItemForSet(o, nexter, vars, func(k []byte, val Value) error {
		r.Set(k, val)
		return nil
	})
	ret = Value{Type: ValueObject, Value: r}
	return
}

func replaceObject(caller Value, nexter TokenScanner, vars *variables) (ret Value, err error) {
	o := caller.Value.(Object)
	r := NewObject()
	err = eachObjectItemForSet(o, nexter, vars, func(k []byte, val Value) error {
		if val.Type != ValueObject {
			return errors.New("replace only support a object as Value")
		}
		r.Del(k)
		obj := val.Value.(Object)
		obj.Each(func(k []byte, val Value) bool {
			r.Set(k, val)
			return true
		})
		return nil
	})
	ret = Value{Type: ValueObject, Value: r}
	return
}

func getObject(caller Value, nexter TokenScanner, vars *variables) (ret Value, err error) {
	o := caller.Value.(*object)
	no := NewObject()
	err = eachObjectItem(o, nexter, vars, func(k []byte, val Value) error {
		no.Set(k, val)
		return nil
	})
	ret = Value{Value: no, Type: ValueObject}
	return
}

func delObject(caller Value, nexter TokenScanner, vars *variables) (ret Value, err error) {
	o := caller.Value.(Object)
	err = eachObjectItem(o, nexter, vars, func(k []byte, val Value) error {
		o.Del(k)
		return nil
	})
	return
}

func eachObjectItemForSet(o Object, nexter TokenScanner, vars *variables, handle func(k []byte, val Value) error) (err error) {
	offset := nexter.Offset()
	nexter.PushEnds(TokenParenthesesClose)
	defer nexter.PopEnds(1)
	o.Each(func(k []byte, val Value) bool {
		nexter.SetOffset(offset)
		vars.set([]byte{'k'}, Value{Type: ValueString, Value: k})
		vars.set([]byte{'v'}, val)
		expr := newStmt(nexter, vars)
		if err = expr.execute(); err != nil {
			return false
		}
		p := expr.value
		if p.Type == ValueNull {
			p = val
		}
		if err = handle(k, p); err != nil {
			return false
		}
		return true
	})
	return
}

func eachObjectItem(o Object, nexter TokenScanner, vars *variables, handle func(k []byte, val Value) error) (err error) {
	offset := nexter.Offset()
	nexter.PushEnds(TokenParenthesesClose)
	defer nexter.PopEnds(1)
	o.Each(func(k []byte, val Value) bool {
		nexter.SetOffset(offset)
		vars.set([]byte{'k'}, Value{Type: ValueString, Value: k})
		vars.set([]byte{'v'}, val)
		expr := newStmt(nexter, vars)
		if err = expr.execute(); err != nil {
			return false
		}
		var bv bool
		if bv, err = expr.value.toBool(); err != nil {
			return false
		}
		if !bv {
			return true
		}
		return handle(k, val) == nil
	})
	return
}

func objectAdd(obj Object, val Value) Object {
	ret := obj.Copy()
	switch val.Type {
	case ValueObject:
		val.Value.(Object).Each(func(k []byte, v Value) bool {
			ret.Set(k, v)
			return true
		})
	}
	return ret
}

func objectDel(obj Object, val Value) Object {
	ret := obj.Copy()
	switch val.Type {
	case ValueObject:
		val.Value.(Object).Each(func(k []byte, val Value) bool {
			it := ret.Get(k)
			if val.equal(it) {
				ret.Del(k)
			}
			return true
		})
	case ValueArray:
		val.Value.(Array).Each(func(i int, val Value) bool {
			switch val.Type {
			case ValueString:
				it := ret.Get(val.Value.([]byte))
				if it.Type != ValueNull {
					ret.Del(val.Value.([]byte))
				}
			}
			return true
		})
	}
	return ret
}

func (obj *object) Copy() Object {
	return NewObject(obj.pairs...)
}

func (obj *object) Get(k []byte) Value {
	for _, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			return p.val
		}
	}
	return Value{Type: ValueNull}
}

func (obj *object) Total() int {
	return len(obj.pairs)
}

func (obj *object) Has(k []byte) bool {
	for _, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			return true
		}
	}
	return false
}

func (obj *object) Set(k []byte, val Value) {
	for i, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			obj.pairs[i] = &pair{key: k, val: val}
			return
		}
	}
	obj.pairs = append(obj.pairs, &pair{key: k, val: val})
}

func (obj *object) Del(k []byte) {
	for i, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			obj.pairs = append(obj.pairs[:i], obj.pairs[i+1:]...)
			return
		}
	}
}

func (obj *object) Each(handle func(k []byte, val Value) bool) {
	for _, p := range obj.pairs {
		if !handle(p.key, p.val) {
			break
		}
	}
}

type objectExecutor struct {
	scanner TokenScanner
	vars    *variables
	value   Object
}

func newObjectExecutor(scanner TokenScanner, vars *variables) *objectExecutor {
	return &objectExecutor{scanner: scanner, vars: vars}
}

func (e *objectExecutor) execute() (err error) {
	e.value, err = e.pairs()
	return
}

func (e *objectExecutor) pairs() (val Object, err error) {
	val = NewObject()
	e.vars.pushMe(Value{Type: ValueObject, Value: val})
	defer e.vars.popMe()
	for {
		expr := newStmt(e.scanner, e.vars)
		func() {
			e.scanner.PushEnds(TokenColon)
			defer e.scanner.PopEnds(1)
			if err = expr.execute(); err != nil {
				return
			}
			if expr.value.Type != ValueString {
				err = errors.New("object key must be string")
				return
			}
		}()
		key := expr.value.Value.([]byte)
		func() {
			e.scanner.PushEnds(TokenComma, TokenBraceClose)
			defer e.scanner.PopEnds(2)
			expr = newStmt(e.scanner, e.vars)
			if err = expr.execute(); err != nil {
				return
			}
		}()
		val.Set(key, expr.value)
		if e.scanner.EndAt() == TokenBraceClose {
			return
		}
	}
}
