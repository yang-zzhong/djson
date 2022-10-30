package djson

import (
	"bytes"
	"errors"
)

type object []pair

type pair struct {
	key []byte
	val value
}

func (obj *object) set(key []byte, val value) {
	for i, p := range *obj {
		if !bytes.Equal(p.key, key) {
			continue
		}
		(*obj)[i] = pair{key: key, val: val}
		return
	}
	*obj = append(*obj, pair{key: key, val: val})
}

func (obj *object) get(key []byte) value {
	for _, p := range *obj {
		if !bytes.Equal(p.key, key) {
			continue
		}
		return p.val
	}
	return value{typ: valueNull}
}

func (obj *object) del(key []byte) {
	for i := range *obj {
		if !bytes.Equal((*obj)[i].key, key) {
			continue
		}
		*obj = append((*obj)[0:i], (*obj)[i+1:]...)
	}
}

func (obj *object) setConditionally(condition func(k []byte, val value) bool, val value) {
	for i, pair := range *obj {
		if condition(pair.key, pair.val) {
			pair.val = val
			(*obj)[i] = pair
		}
	}
}

func (obj *object) delConditionally(condition func(k []byte, val value) bool) {
	for i := 0; i < len(*obj); i++ {
		p := (*obj)[i]
		if condition(p.key, p.val) {
			*obj = append((*obj)[0:i], (*obj)[i+1:]...)
			i--
		}
	}
}

func (obj *object) replace(condition func(k []byte, val value) bool, val func(k []byte, val value) pair) {
	for i, p := range *obj {
		if condition(p.key, p.val) {
			(*obj)[i] = val(p.key, p.val)
		}
	}
}

func (obj *object) merge(n *object) {
	for _, p := range *n {
		(*obj) = append(*obj, p)
	}
}

func (obj *object) keys(condition func(k []byte, val value) bool) [][]byte {
	ret := [][]byte{}
	for _, p := range *obj {
		if condition(p.key, p.val) {
			ret = append(ret, p.key)
		}
	}
	return ret
}

func (obj *object) values(condition func(k []byte, val value) bool) []value {
	ret := []value{}
	for _, p := range *obj {
		if condition(p.key, p.val) {
			ret = append(ret, p.val)
		}
	}
	return ret
}

type objectExecutor struct {
	getter    lexer
	variables *variables
	value     object
}

func newObjectExecutor(getter lexer, vs *variables) *objectExecutor {
	return &objectExecutor{
		getter:    getter,
		variables: vs,
	}
}

func (e *objectExecutor) execute() (err error) {
	e.value, err = e.pairs()
	return
}

func (e *objectExecutor) pairs() (val object, err error) {
	for {
		expr := newExpr(e.getter, [][]byte{{':'}}, e.variables)
		if err = expr.execute(); err != nil {
			return
		}
		if expr.value.typ != valueString {
			err = errors.New("object key must be string")
			return
		}
		key := expr.value.value.([]byte)
		expr = newExpr(e.getter, [][]byte{{','}, {'}'}}, e.variables)
		if err = expr.execute(); err != nil {
			return
		}
		val.set(key, expr.value)
		if bytes.Equal(expr.endAt(), []byte{'}'}) {
			return
		}
	}
}
