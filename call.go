package djson

import "fmt"

type callable interface {
	call(k string, caller value, scanner *tokenScanner, vars *variables) (value, error)
}

type callableImp struct {
	calls map[string]callback
	typ   string
}

type callback func(caller value, scanner *tokenScanner, vars *variables) (value, error)

func newCallable(typ string) *callableImp {
	return &callableImp{typ: typ}
}

func (c *callableImp) register(k string, ck callback) {
	if c.calls == nil {
		c.calls = make(map[string]callback)
	}
	c.calls[k] = ck
}

func (c *callableImp) call(k string, caller value, scanner *tokenScanner, vars *variables) (val value, err error) {
	call, ok := c.calls[k]
	if !ok {
		err = fmt.Errorf("undefined method for %s", c.typ)
		return
	}
	return call(val, scanner, vars)
}
