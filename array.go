package djson

import (
	"bytes"
	"fmt"
	"strconv"
)

type Array interface {
	ItemEachable
	Set(i int, val Value)
	Get(i int) Value
	Append(val ...Value)
	Copy() Array
	Del(i int)
	Total() int
}

type array struct {
	*CallableRegister
	items []Value
}

var _ Array = &array{}

func NewArray(items ...Value) *array {
	arr := &array{
		CallableRegister: NewCallableRegister("array"),
		items:            items,
	}
	arr.RegisterCall("map", setArray)
	arr.RegisterCall("del", delArray)
	arr.RegisterCall("filter", filterArray)
	return arr
}

func NewArrayWithLength(length int) *array {
	arr := &array{
		CallableRegister: NewCallableRegister("array"),
		items:            make([]Value, length),
	}
	arr.RegisterCall("map", setArray)
	arr.RegisterCall("del", delArray)
	arr.RegisterCall("filter", filterArray)
	return arr
}

func setArray(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	o := val.Value.(Array).Copy()
	err = eachItemForSet(o, scanner, vars, func(val Value, idx int) error {
		o.Set(idx, val)
		return nil
	})
	ret = Value{Type: ValueArray, Value: o}
	return
}

func delArray(caller Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	o := caller.Value.(Array)
	r := o.Copy()
	err = eachArrayItem(o, scanner, vars, func(_ Value, idx int) error {
		r.Del(idx)
		return nil
	})
	ret = Value{Type: ValueArray, Value: r}
	return
}

func filterArray(caller Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	o := caller.Value.(*array)
	no := NewArray()
	err = eachArrayItem(o, scanner, vars, func(val Value, idx int) error {
		no.items = append(no.items, val)
		return nil
	})
	ret = Value{Value: no, Type: ValueArray}
	return
}

func eachItemForSet(o Array, scanner TokenScanner, vars Context, handle func(val Value, idx int) error) (err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	o.Each(func(i int, val Value) bool {
		scanner.SetOffset(offset)
		vars.Assign([]byte{'i'}, IntValue(int64(i)))
		vars.Assign([]byte{'v'}, val)
		expr := NewStmtExecutor(scanner, vars)
		if err = expr.Execute(); err != nil {
			return false
		}
		if expr.Exited() {
			Exit()
		}
		p := expr.Value()
		if p.Type == ValueNull {
			return true
		}
		return handle(p, i) == nil
	})
	return
}

func eachArrayItem(o Array, scanner TokenScanner, vars Context, handle func(val Value, idx int) error) (err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	o.Each(func(i int, val Value) bool {
		scanner.SetOffset(offset)
		vars.Assign([]byte{'i'}, IntValue(int64(i)))
		vars.Assign([]byte{'v'}, val)
		expr := NewStmtExecutor(scanner, vars)
		if err = expr.Execute(); err != nil {
			return false
		}
		if expr.Exited() {
			Exit()
		}
		if !expr.value.Bool() {
			return true
		}
		handle(val, i)
		return true
	})
	return
}

func (arr *array) Add(val Value) (ret Value, err error) {
	r := arr.Copy()
	if val.Type != ValueArray {
		err = fmt.Errorf("array can't + a [%s]", val.TypeName())
	}
	val.Value.(Array).Each(func(i int, val Value) bool {
		r.Append(val)
		return true
	})
	ret = Value{Type: ValueArray, Value: r}
	return
}

func (arr *array) Minus(val Value) (ret Value, err error) {
	r := arr.Copy()
	if val.Type != ValueArray {
		err = fmt.Errorf("array can't + a [%s]", val.TypeName())
	}
	for i := 0; i < arr.Total(); i++ {
		v := arr.Get(i)
		var eq bool
		switch val.Type {
		case ValueArray:
			val.Value.(Array).Each(func(_ int, right Value) bool {
				eq = v.Equal(right)
				return !eq
			})
		default:
			eq = v.Equal(val)
		}
		if eq {
			r.Del(i)
			i--
		}
	}
	ret = Value{Type: ValueArray, Value: r}
	return
}

func (arr *array) Multiply(val Value) (ret Value, err error) {
	err = fmt.Errorf("array can't * a [%s]", val.TypeName())
	return
}

func (arr *array) Devide(val Value) (ret Value, err error) {
	err = fmt.Errorf("array can't / a [%s]", val.TypeName())
	return
}

func (arr *array) Compare(val Value) (ret int, err error) {
	if val.Type != ValueArray {
		err = fmt.Errorf("array can't compare with [%s]", val.TypeName())
		return
	}
	rr := val.Value.(Array)
	if arr.Total() > rr.Total() {
		return 1, nil
	} else if arr.Total() < rr.Total() {
		return -1, nil
	}
	var c int
	arr.Each(func(i int, val Value) bool {
		c, err = val.Compare(rr.Get(i))
		return err == nil && c != 0
	})
	return c, err

}
func (arr *array) Set(idx int, val Value) {
	arr.items[idx] = val
}

func (arr *array) Get(idx int) Value {
	return arr.items[idx]
}

func (arr *array) Del(idx int) {
	arr.items = append(arr.items[:idx], arr.items[idx+1:]...)
}

func (arr *array) Each(handle func(i int, val Value) bool) {
	for i, item := range arr.items {
		if !handle(i, item) {
			break
		}
	}
}

func (arr *array) Copy() Array {
	return NewArray(arr.items...)
}

func (arr *array) Total() int {
	return len(arr.items)
}

func (arr *array) Append(val ...Value) {
	arr.items = append(arr.items, val...)
}

type arrayExecutor struct {
	scanner TokenScanner
	vars    Context
	value   Value
}

func newArrayExecutor(scanner TokenScanner, vs Context) *arrayExecutor {
	return &arrayExecutor{
		scanner: scanner,
		vars:    vs,
	}
}

func (e *arrayExecutor) execute() (err error) {
	e.value, err = e.items()
	return
}

func (e *arrayExecutor) items() (ret Value, err error) {
	arr := NewArray()
	e.scanner.PushEnds(TokenBracketsClose, TokenComma)
	defer e.scanner.PopEnds(TokenBracketsClose, TokenComma)
	e.vars.pushMe(ArrayValue(arr))
	defer e.vars.popMe()
	for {
		expr := NewStmtExecutor(e.scanner, e.vars)
		if err = expr.Execute(); err != nil {
			return
		}
		if expr.Exited() {
			Exit()
		}
		if expr.value.Type == ValueRange {
			ret = expr.value
			return
		}
		arr.Append(expr.value)
		if e.scanner.EndAt() == TokenBracketsClose {
			break
		}
	}
	ret = ArrayValue(arr)
	return
}

func (arr *array) lookup(k []byte) Value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		idx, err := strconv.Atoi(string(i))
		if err != nil {
			return Value{Type: ValueNull}
		}
		if idx > len(arr.items) {
			return Value{Type: ValueNull}
		}
		if len(r) == 0 {
			return arr.items[idx]
		}
		return arr.items[idx].lookup(r)
	}
	if len(r) == 0 {
		return Value{Type: ValueArray, Value: arr}
	}
	ret := NewArray()
	for _, item := range arr.items {
		v := item.lookup(r)
		if v.Type != ValueNull {
			ret.items = append(ret.items, v)
		}
	}
	return Value{Type: ValueArray, Value: ret}
}
