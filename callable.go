package djson

import (
	"fmt"
	"strings"
)

type Callable interface {
	call(k string, caller Value, scanner TokenScanner, vars Context) (Value, error)
}

type CallableRegister struct {
	calls map[string]Callback
	typ   string
}

type Callback func(caller Value, scanner TokenScanner, vars Context) (Value, error)

func NewCallableRegister(typ string) *CallableRegister {
	return &CallableRegister{typ: typ}
}

func (c *CallableRegister) RegisterCall(k string, ck Callback) {
	if c.calls == nil {
		c.calls = make(map[string]Callback)
	}
	c.calls[k] = ck
}

func (c *CallableRegister) call(k string, caller Value, scanner TokenScanner, vars Context) (val Value, err error) {
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

func (c *CallableRegister) caseInsensitiveCallback(k string) (Callback, bool) {
	for ck, c := range c.calls {
		if strings.EqualFold(ck, k) {
			return c, true
		}
	}
	return nil, false
}
