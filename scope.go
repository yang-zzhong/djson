package djson

import (
	"fmt"
)

type scopeType int
type variableType int

const (
	scopeGlobal = scopeType(iota)
	scopeObject
	scopeArray
	scopeTemplate
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
		return fmt.Errorf("undefined variable [%s]", name)
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
