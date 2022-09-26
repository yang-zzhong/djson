package djson

import (
	"bytes"
	"fmt"
)

type scopeType int
type variableType int

const (
	scopeGlobal = scopeType(iota)
	scopeObject
	scopeArray
	scopeTemplate
	scopeKey
	scopeValue
	scopeExpr
	scopeUnknown
)

const (
	variableInt = variableType(iota)
	variableFloat
	variableString
	variableArray
	variableObject
)

var (
	errUndefinedVariable = func(name []byte) error {
		return fmt.Errorf("undefined variable [%s], name")
	}
)

type variable struct {
	typ  variableType
	name []byte
	val  []byte
}

type pair struct {
	key []byte
	val variable
}

type scope struct {
	typ      scopeType
	vars     []variable
	pairs    []pair
	stash    []*Token
	p        *scope
	children []*scope
}

func newScope(typ scopeType) *scope {
	return &scope{typ: typ}
}

func (scope *scope) addChild(c *scope) {
	c.p = scope
	scope.children = append(scope.children, c)
}

func (scope *scope) addVar(v *variable) {
	if i := scope.indexLocalVar(v.name); i > -1 {
		scope.vars[i] = *v
		return
	}
	scope.vars = append(scope.vars, *v)
}

func (scope *scope) getVar(name []byte, v *variable) bool {
	if i := scope.indexLocalVar(name); i > -1 {
		*v = scope.vars[i]
		return true
	}
	if scope.p == nil {
		return false
	}
	return scope.getVar(name, v)
}

func (scope *scope) indexLocalVar(name []byte) int {
	for i, iv := range scope.vars {
		if !bytes.Equal(name, iv.name) {
			continue
		}
		return i
	}
	return -1
}

func (scope *scope) indexPair(key []byte) int {
	for i, p := range scope.pairs {
		if bytes.Equal(p.key, key) {
			return i
		}
	}
	return -1
}

func (scope *scope) setPair(p pair) {
	if i := scope.indexPair(p.key); i > -1 {
		scope.pairs[i] = p
		return
	}
	scope.pairs = append(scope.pairs, p)
}
