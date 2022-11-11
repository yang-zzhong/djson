package djson

import (
	"bytes"
	"errors"
)

type object struct {
	pairs []*pair
	*callableImp
}

type pair struct {
	key []byte
	val value
}

func newObject() *object {
	obj := &object{pairs: []*pair{}, callableImp: newCallable("object")}
	obj.register("set", setObject)
	obj.register("replace", replaceObject)
	obj.register("del", delObject)
	obj.register("filter", filterObject)
	return obj
}

func setObject(val value, nexter *tokenScanner, vars *variables) (value, error) {
	o := val.value.(*object)
	return eachPairForSet(o, nexter, vars, func(k []byte, val value, idx int) error {
		o.pairs[idx].val = val
		return nil
	})
}

func replaceObject(caller value, nexter *tokenScanner, vars *variables) (value, error) {
	o := caller.value.(*object)
	return eachPairForSet(o, nexter, vars, func(k []byte, val value, idx int) error {
		if val.typ != valueObject {
			return errors.New("replace only support a object as value")
		}
		obj := val.value.(*object)
		o.pairs[idx] = obj.pairs[0]
		return nil
	})
}

func filterObject(caller value, nexter *tokenScanner, vars *variables) (ret value, err error) {
	o := caller.value.(*object)
	no := newObject()
	_, err = eachPairForFilter(o, nexter, vars, func(k []byte, val value, idx int) error {
		no.pairs = append(no.pairs, &pair{key: k, val: val})
		return nil
	})
	ret = value{value: no, typ: valueObject}
	return
}

func delObject(caller value, nexter *tokenScanner, vars *variables) (ret value, err error) {
	o := caller.value.(*object)
	_, err = eachPairForFilter(o, nexter, vars, func(k []byte, val value, idx int) error {
		o.pairs = append(o.pairs[:idx], o.pairs[idx+1:]...)
		return nil
	})
	return
}

func eachPairForSet(o *object, nexter *tokenScanner, vars *variables, handle func(k []byte, val value, idx int) error) (ret value, err error) {
	offset := nexter.offset()
	nexter.pushEnds(TokenParenthesesClose, TokenReduction)
	defer nexter.popEnds(2)
	for i, p := range o.pairs {
		nexter.setOffset(offset)
		vars.set([]byte{'k'}, value{typ: valueString, value: p.key})
		vars.set([]byte{'v'}, p.val)
		expr := newStmt(nexter, vars)
		if err = expr.execute(); err != nil {
			return
		}
		if nexter.endAt() == TokenParenthesesClose {
			if err = handle(p.key, expr.value, i); err != nil {
				return
			}
		}
		var bv bool
		if bv, err = expr.value.toBool(); err != nil {
			return
		}
		if !bv {
			continue
		}
		expr = newStmt(nexter, vars)
		if err = expr.execute(); err != nil {
			return
		}
		if err = handle(p.key, expr.value, i); err != nil {
			return
		}
	}
	return
}

func eachPairForFilter(o *object, nexter *tokenScanner, vars *variables, handle func(k []byte, val value, idx int) error) (ret value, err error) {
	offset := nexter.offset()
	nexter.pushEnds(TokenParenthesesClose)
	defer nexter.popEnds(1)
	for i, p := range o.pairs {
		nexter.setOffset(offset)
		vars.set([]byte{'k'}, value{typ: valueString, value: p.key})
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

func (obj object) copy() *object {
	r := newObject()
	r.pairs = obj.pairs
	return r
}

func (obj *object) get(k []byte) value {
	for _, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			return p.val
		}
	}
	return value{typ: valueNull}
}

func (obj *object) has(k []byte) bool {
	for _, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			return true
		}
	}
	return false
}

func (obj *object) set(k []byte, val value) {
	for i, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			obj.pairs[i] = &pair{key: k, val: val}
			return
		}
	}
	obj.pairs = append(obj.pairs, &pair{key: k, val: val})
}

func (obj *object) del(k []byte) {
	for i, p := range obj.pairs {
		if bytes.Equal(p.key, k) {
			obj.pairs = append(obj.pairs[:i], obj.pairs[i+1:]...)
			return
		}
	}
}

type objectExecutor struct {
	scanner *tokenScanner
	vars    *variables
	value   *object
}

func newObjectExecutor(scanner *tokenScanner, vars *variables) *objectExecutor {
	return &objectExecutor{scanner: scanner, vars: vars}
}

func (e *objectExecutor) execute() (err error) {
	e.value, err = e.pairs()
	return
}

func (e *objectExecutor) pairs() (val *object, err error) {
	val = newObject()
	e.vars.pushMe(value{typ: valueObject, value: val})
	defer e.vars.popMe()
	for {
		expr := newStmt(e.scanner, e.vars)
		func() {
			e.scanner.pushEnds(TokenColon)
			defer e.scanner.popEnds(1)
			if err = expr.execute(); err != nil {
				return
			}
			if expr.value.typ != valueString {
				err = errors.New("object key must be string")
				return
			}
		}()
		key := expr.value.value.([]byte)
		func() {
			e.scanner.pushEnds(TokenComma, TokenBraceClose)
			defer e.scanner.popEnds(2)
			expr = newStmt(e.scanner, e.vars)
			if err = expr.execute(); err != nil {
				return
			}
		}()
		val.set(key, expr.value)
		if e.scanner.endAt() == TokenBraceClose {
			return
		}
	}
}
