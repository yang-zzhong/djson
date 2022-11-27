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
	return &CallableRegister{typ: typ, calls: map[string]Callback{
		"if": ifCall,
	}}
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
		err = fmt.Errorf("undefined method [%s] for %s", k, c.typ)
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

func ifCall(val Value, scanner TokenScanner, ctx Context) (ret Value, err error) {
	ctx.pushMe(val)
	defer ctx.popMe()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	expr := NewStmtExecutor(scanner, ctx)
	if err = expr.Execute(For(NullValue())); err != nil {
		return
	}
	if expr.Exited() {
		Exit()
	}
	if ret = expr.Value(); ret.Type != ValueNull {
		return
	}
	ret = val
	return
}
