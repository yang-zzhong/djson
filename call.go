package djson

import (
	"fmt"
	"strings"
)

type callable interface {
	call(k string, caller Value, scanner TokenScanner, vars *variables) (Value, error)
}

type callableImp struct {
	calls map[string]callback
	typ   string
}

type callback func(caller Value, scanner TokenScanner, vars *variables) (Value, error)

func newCallable(typ string) *callableImp {
	return &callableImp{typ: typ}
}

func (c *callableImp) register(k string, ck callback) {
	if c.calls == nil {
		c.calls = make(map[string]callback)
	}
	c.calls[k] = ck
}

func (c *callableImp) call(k string, caller Value, scanner TokenScanner, vars *variables) (val Value, err error) {
	call, ok := c.calls[k]
	if !ok {
		call, ok = c.caseInsensitiveCallback(k)
	}
	if !ok {
		err = fmt.Errorf("undefined method for %s", c.typ)
		return
	}
	return call(caller, scanner, vars)
}

func (c *callableImp) caseInsensitiveCallback(k string) (callback, bool) {
	for ck, c := range c.calls {
		if strings.EqualFold(ck, k) {
			return c, true
		}
	}
	return nil, false
}
