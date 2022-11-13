package djson

type Array interface {
	Set(i int, val Value)
	Get(i int) Value
	Append(val ...Value)
	Copy() Array
	Del(i int)
	Total() int
	Each(func(i int, val Value) bool)
}

type array struct {
	*callableImp
	items []Value
}

var _ Array = &array{}

func NewArray(items ...Value) *array {
	arr := &array{
		callableImp: newCallable("array"),
		items:       items,
	}
	arr.register("set", setArray)
	arr.register("map", setArray)
	arr.register("del", delArray)
	arr.register("filter", filterArray)
	return arr
}

func setArray(val Value, scanner TokenScanner, vars *variables) (Value, error) {
	o := val.Value.(Array)
	return eachItemForSet(o, scanner, vars, func(val Value, idx int) error {
		o.Set(idx, val)
		return nil
	})
}

func delArray(caller Value, scanner TokenScanner, vars *variables) (Value, error) {
	o := caller.Value.(Array)
	return eachItemForFilter(o, scanner, vars, func(_ Value, idx int) error {
		o.Del(idx)
		return nil
	})
}

func filterArray(caller Value, scanner TokenScanner, vars *variables) (ret Value, err error) {
	o := caller.Value.(*array)
	no := NewArray()
	_, err = eachItemForFilter(o, scanner, vars, func(val Value, idx int) error {
		no.items = append(no.items, val)
		return nil
	})
	ret = Value{Value: no, Type: ValueArray}
	return
}

func eachItemForSet(o Array, scanner TokenScanner, vars *variables, handle func(val Value, idx int) error) (ret Value, err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(1)
	o.Each(func(i int, val Value) bool {
		scanner.SetOffset(offset)
		vars.set([]byte{'i'}, Value{Type: ValueInt, Value: int64(i)})
		vars.set([]byte{'v'}, val)
		scanner.PushEnds(TokenParenthesesClose)
		defer scanner.PopEnds(1)
		expr := newStmt(scanner, vars)
		if err = expr.execute(); err != nil {
			return false
		}
		err = handle(expr.value, i)
		return err == nil
	})
	return
}

func eachItemForFilter(o Array, scanner TokenScanner, vars *variables, handle func(val Value, idx int) error) (ret Value, err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(1)
	o.Each(func(i int, val Value) bool {
		scanner.SetOffset(offset)
		vars.set([]byte{'i'}, Value{Type: ValueInt, Value: int64(i)})
		vars.set([]byte{'v'}, val)
		expr := newStmt(scanner, vars)
		if err = expr.execute(); err != nil {
			return false
		}
		var bv bool
		if bv, err = expr.value.toBool(); err != nil {
			return false
		}
		if !bv {
			return true
		}
		handle(val, i)
		return true
	})
	return
}

func arrayAdd(arr Array, val Value) Array {
	ret := arr.Copy()
	switch val.Type {
	case ValueArray:
		val.Value.(Array).Each(func(i int, val Value) bool {
			ret.Append(val)
			return true
		})
	default:
		ret.Append(val)
	}
	return ret
}

func arrayDel(arr Array, val Value) Array {
	ret := arr.Copy()
	for i := 0; i < arr.Total(); i++ {
		v := arr.Get(i)
		var eq bool
		switch val.Type {
		case ValueArray:
			val.Value.(Array).Each(func(_ int, right Value) bool {
				eq = v.equal(right)
				return !eq
			})
		default:
			eq = v.equal(val)
		}
		if eq {
			ret.Del(i)
			i--
		}
	}
	return ret
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

func (arr *array) del(val ...Value) {

}

type arrayExecutor struct {
	scanner TokenScanner
	vars    *variables
	value   Array
}

func newArrayExecutor(scanner TokenScanner, vs *variables) *arrayExecutor {
	return &arrayExecutor{
		scanner: scanner,
		vars:    vs,
	}
}

func (e *arrayExecutor) execute() (err error) {
	e.value, err = e.items()
	return
}

func (e *arrayExecutor) items() (val Array, err error) {
	val = NewArray()
	e.scanner.PushEnds(TokenBracketsClose, TokenComma)
	defer e.scanner.PopEnds(2)
	e.vars.pushMe(Value{Type: ValueArray, Value: &val})
	defer e.vars.popMe()
	for {
		expr := newStmt(e.scanner, e.vars)
		if err = expr.execute(); err != nil {
			return
		}
		val.Append(expr.value)
		if expr.endAt() == TokenBracketsClose {
			return
		}
	}
}
