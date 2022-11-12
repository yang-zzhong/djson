package djson

import (
	"bytes"
	"errors"
)

type Object struct {
	pairs []*pair
	*callableImp
}

type pair struct {
	key []byte
	val Value
}

func newObject(pairs ...*pair) *Object {
	obj := &Object{pairs: pairs, callableImp: newCallable("object")}
	obj.register("set", setObject)
	obj.register("replace", replaceObject)
	obj.register("del", delObject)
	obj.register("filter", filterObject)
	return obj
}

func setObject(val Value, nexter TokenScanner, vars *variables) (Value, error) {
	o := val.Value.(*Object)
	return eachPairForSet(o, nexter, vars, func(k []byte, val Value, idx int) error {
		o.pairs[idx].val = val
		return nil
	})
}

func replaceObject(caller Value, nexter TokenScanner, vars *variables) (Value, error) {
	o := caller.Value.(*Object)
	return eachPairForSet(o, nexter, vars, func(k []byte, val Value, idx int) error {
		if val.Type != ValueObject {
			return errors.New("replace only support a object as Value")
		}
		obj := val.Value.(*Object)
		o.pairs[idx] = obj.pairs[0]
		return nil
	})
}

func filterObject(caller Value, nexter TokenScanner, vars *variables) (ret Value, err error) {
	o := caller.Value.(*Object)
	no := newObject()
	_, err = eachPairForFilter(o, nexter, vars, func(k []byte, val Value, idx int) error {
		no.pairs = append(no.pairs, &pair{key: k, val: val})
		return nil
	})
	ret = Value{Value: no, Type: ValueObject}
	return
}

func delObject(caller Value, nexter TokenScanner, vars *variables) (ret Value, err error) {
	o := caller.Value.(*Object)
	_, err = eachPairForFilter(o, nexter, vars, func(k []byte, val Value, idx int) error {
		o.pairs = append(o.pairs[:idx], o.pairs[idx+1:]...)
		return nil
	})
	return
}

func eachPairForSet(o *Object, nexter TokenScanner, vars *variables, handle func(k []byte, val Value, idx int) error) (ret Value, err error) {
	offset := nexter.Offset()
	nexter.PushEnds(TokenParenthesesClose)
	defer nexter.PopEnds(1)
	for i, p := range o.pairs {
		nexter.SetOffset(offset)
		vars.set([]byte{'k'}, Value{Type: ValueString, Value: p.key})
		vars.set([]byte{'v'}, p.val)
		expr := newStmt(nexter, vars)
		if err = expr.execute(); err != nil {
			return
		}
		if err = handle(p.key, expr.value, i); err != nil {
			return
		}
	}
	return
}

func eachPairForFilter(o *Object, nexter TokenScanner, vars *variables, handle func(k []byte, val Value, idx int) error) (ret Value, err error) {
	offset := nexter.Offset()
	nexter.PushEnds(TokenParenthesesClose)
	defer nexter.PopEnds(1)
	for i, p := range o.pairs {
		nexter.SetOffset(offset)
		vars.set([]byte{'k'}, Value{Type: ValueString, Value: p.key})
		vars.set([]byte{'v'}, p.val)
		expr := newStmt(nexter, vars)
		if err = expr.execute(); err != nil {
			return
		}
		var bv bool
		if bv, err = expr.value.toBool(); err != nil {
			return
		}
		if !bv {
			continue
		}
		handle(p.key, expr.value, i)
	}
	return
}

func (obj Object) copy() *Object {
	r := newObject()
	r.pairs = obj.pairs
	return r
}

func (obj *Object) get(k []byte) Value {
	for _, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			return p.val
		}
	}
	return Value{Type: ValueNull}
}

func (obj *Object) has(k []byte) bool {
	for _, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			return true
		}
	}
	return false
}

func (obj *Object) set(k []byte, val Value) {
	for i, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			obj.pairs[i] = &pair{key: k, val: val}
			return
		}
	}
	obj.pairs = append(obj.pairs, &pair{key: k, val: val})
}

func (obj *Object) del(k []byte) {
	for i, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			obj.pairs = append(obj.pairs[:i], obj.pairs[i+1:]...)
			return
		}
	}
}

type objectExecutor struct {
	scanner TokenScanner
	vars    *variables
	value   *Object
}

func newObjectExecutor(scanner TokenScanner, vars *variables) *objectExecutor {
	return &objectExecutor{scanner: scanner, vars: vars}
}

func (e *objectExecutor) execute() (err error) {
	e.value, err = e.pairs()
	return
}

func (e *objectExecutor) pairs() (val *Object, err error) {
	val = newObject()
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
		val.set(key, expr.value)
		if e.scanner.EndAt() == TokenBraceClose {
			return
		}
	}
}
