package djson

import (
	"bytes"
	"errors"
	"fmt"
)

type PairEachable interface {
	Each(func(k []byte, val Value) bool)
}

type Object interface {
	PairEachable
	Set(k []byte, val Value)
	Get(k []byte) Value
	Has(k []byte) bool
	Del(k []byte)
	Copy() Object
	Total() int
}

type object struct {
	pairs []*pair
	*CallableRegister
}

type pair struct {
	key []byte
	val Value
}

func (p *pair) copy() *pair {
	r := &pair{}
	r.key = make([]byte, len(p.key))
	copy(r.key, p.key)
	r.val = p.val.Copy()
	return r
}

var _ Object = &object{}

func NewObject(pairs ...*pair) *object {
	obj := &object{pairs: pairs, CallableRegister: NewCallableRegister("object")}
	obj.RegisterCall("map", setObject)
	obj.RegisterCall("trans", transObject)
	obj.RegisterCall("replace", replaceObject)
	obj.RegisterCall("del", delObject)
	obj.RegisterCall("filter", getObject)
	return obj
}

func setObject(val Value, nexter TokenScanner, vars Context) (ret Value, err error) {
	r := val.Value.(Object).Copy()
	err = eachObjectItemForSet(val, nexter, vars, func(k []byte, val Value) error {
		r.Set(k, val)
		return nil
	})
	ret = Value{Type: ValueObject, Value: r}
	return
}

func replaceObject(caller Value, nexter TokenScanner, vars Context) (ret Value, err error) {
	r := caller.Value.(Object).Copy()
	err = eachObjectItemForSet(caller, nexter, vars, func(k []byte, val Value) error {
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

func transObject(caller Value, scanner TokenScanner, ctx Context) (ret Value, err error) {
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	o := caller.Copy()
	ctx.pushMe(o)
	defer ctx.popMe()
	stmt := NewStmtExecutor(scanner, ctx)
	if err = stmt.Execute(); err != nil {
		return
	}
	ret = o
	return
}

func getObject(caller Value, nexter TokenScanner, vars Context) (ret Value, err error) {
	no := NewObject()
	err = eachObjectItem(caller, nexter, vars, func(k []byte, val Value) error {
		no.Set(k, val)
		return nil
	})
	ret = Value{Value: no, Type: ValueObject}
	return
}

func delObject(caller Value, nexter TokenScanner, vars Context) (ret Value, err error) {
	r := caller.Value.(Object).Copy()
	err = eachObjectItem(caller, nexter, vars, func(k []byte, val Value) error {
		r.Del(k)
		return nil
	})
	ret = ObjectValue(r)
	return
}

func eachObjectItemForSet(caller Value, nexter TokenScanner, ctx Context, handle func(k []byte, val Value) error) (err error) {
	scanner := NewCachedTokenScanner(nexter)
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	stmt := NewStmtExecutor(scanner, ctx)
	ctx.pushMe(caller)
	defer ctx.popMe()
	caller.Value.(PairEachable).Each(func(k []byte, val Value) bool {
		scanner.ResetRead()
		ctx.Assign([]byte{'k'}, StringValue(k...))
		ctx.Assign([]byte{'v'}, val)
		if err = stmt.Execute(); err != nil {
			return false
		}
		if stmt.Exited() {
			Exit()
		}
		p := stmt.Value()
		if p.Type == ValueNull {
			return true
		}
		return handle(k, p) == nil
	})
	return
}

func eachObjectItem(caller Value, nexter TokenScanner, ctx Context, handle func(k []byte, val Value) error) (err error) {
	scanner := NewCachedTokenScanner(nexter)
	nexter.PushEnds(TokenParenthesesClose)
	defer nexter.PopEnds(TokenParenthesesClose)
	stmt := NewStmtExecutor(scanner, ctx)
	ctx.pushMe(caller)
	defer ctx.popMe()
	caller.Value.(PairEachable).Each(func(k []byte, val Value) bool {
		scanner.ResetRead()
		ctx.Assign([]byte{'k'}, StringValue(k...))
		ctx.Assign([]byte{'v'}, val)
		if err = stmt.Execute(); err != nil {
			return false
		}
		if stmt.Exited() {
			Exit()
		}
		if !stmt.value.Bool() {
			return true
		}
		return handle(k, val) == nil
	})
	return
}

func (obj *object) Add(val Value) (ret Value, err error) {
	if val.Type != ValueObject {
		err = fmt.Errorf("object can't + a [%s]", val.TypeName())
		return
	}
	r := obj.Copy()
	val.Value.(Object).Each(func(k []byte, v Value) bool {
		r.Set(k, v)
		return true
	})
	ret = Value{Type: ValueObject, Value: r}
	return
}

func (obj *object) Minus(val Value) (ret Value, err error) {
	r := obj.Copy()
	switch val.Type {
	case ValueObject:
		val.Value.(Object).Each(func(k []byte, val Value) bool {
			it := r.Get(k)
			if val.Equal(it) {
				r.Del(k)
			}
			return true
		})
	case ValueArray:
		val.Value.(Array).Each(func(i int, val Value) bool {
			switch val.Type {
			case ValueString:
				it := r.Get(val.Value.(String).Bytes())
				if it.Type != ValueNull {
					r.Del(val.Value.(String).Bytes())
				}
			}
			return true
		})
	default:
		err = fmt.Errorf("object can't - a [%s]", val.TypeName())
		return
	}
	ret = ObjectValue(r)
	return
}

func (obj *object) Multiply(val Value) (ret Value, err error) {
	err = fmt.Errorf("object can't * a [%s]", val.TypeName())
	return
}

func (obj *object) Devide(val Value) (ret Value, err error) {
	err = fmt.Errorf("object can't / a [%s]", val.TypeName())
	return
}

func (obj *object) Compare(val Value) (ret int, err error) {
	if val.Type != ValueObject {
		err = fmt.Errorf("object can't compare with [%s]", val.TypeName())
		return
	}
	rr := val.Value.(Object)
	if obj.Total() > rr.Total() {
		return 1, nil
	} else if obj.Total() < rr.Total() {
		return -1, nil
	}
	var c int
	obj.Each(func(k []byte, val Value) bool {
		c, err = val.Compare(rr.Get(k))
		return err == nil && c != 0
	})
	return c, err
}

func (obj *object) Copy() Object {
	r := NewObject()
	r.pairs = make([]*pair, len(obj.pairs))
	for i, p := range obj.pairs {
		r.pairs[i] = p.copy()
	}
	return r
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
	ctx     Context
	value   Value
}

func newObjectExecutor(scanner TokenScanner, vars Context) *objectExecutor {
	return &objectExecutor{scanner: scanner, ctx: vars}
}

func (e *objectExecutor) execute() (err error) {
	e.value, err = e.pairs()
	return
}

func (e *objectExecutor) pairs() (ret Value, err error) {
	val := NewObject()
	ret = ObjectValue(val)
	stmt := NewStmtExecutor(e.scanner, e.ctx)
	e.ctx.pushMe(ObjectValue(val))
	defer e.ctx.popMe()
	for {
		err = stmt.Execute(For(NullValue()), EndWhen(TokenColon))
		if stmt.Exited() {
			Exit()
		}
		if err != nil || stmt.value.Type == ValueNull {
			return
		}
		if stmt.value.Type != ValueString {
			err = fmt.Errorf("object key [%v] must be string", stmt.value.Value)
			return
		}
		key := stmt.value.Value.(String).Bytes()
		stmt.Execute(For(NullValue()), EndWhen(TokenComma, TokenBraceClose))
		if stmt.Exited() {
			Exit()
		}
		if err != nil {
			return
		}
		val.Set(key, stmt.value)
		if e.scanner.EndAt() == TokenBraceClose {
			return
		}
	}
}

func (obj *object) lookup(k []byte) Value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		val := obj.Get(i)
		if val.Type == ValueNull || len(r) == 0 {
			return val
		}
		return val.lookup(r)
	}
	arr := NewArray()
	for _, p := range obj.pairs {
		if len(r) == 0 {
			arr.items = append(arr.items, p.val)
			continue
		}
		item := p.val.lookup(r)
		if item.Type != ValueNull {
			arr.items = append(arr.items, item)
		}
	}
	return Value{Type: ValueArray, Value: arr}
}
