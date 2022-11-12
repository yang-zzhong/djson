package djson

type array struct {
	*callableImp
	items []Value
}

func newArray(items ...Value) *array {
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
	o := val.Value.(*array)
	return eachItemForSet(o, scanner, vars, func(val Value, idx int) error {
		o.items[idx] = val
		return nil
	})
}

func delArray(caller Value, scanner TokenScanner, vars *variables) (Value, error) {
	o := caller.Value.(*array)
	return eachItemForFilter(o, scanner, vars, func(_ Value, idx int) error {
		o.items = append(o.items[:idx], o.items[idx+1:]...)
		return nil
	})
}

func filterArray(caller Value, scanner TokenScanner, vars *variables) (ret Value, err error) {
	o := caller.Value.(*array)
	no := newArray()
	_, err = eachItemForFilter(o, scanner, vars, func(val Value, idx int) error {
		no.items = append(no.items, val)
		return nil
	})
	ret = Value{Value: no, Type: ValueArray}
	return
}

func eachItemForSet(o *array, scanner TokenScanner, vars *variables, handle func(val Value, idx int) error) (ret Value, err error) {
	offset := scanner.Offset()
	for i, p := range o.items {
		scanner.SetOffset(offset)
		vars.set([]byte{'i'}, Value{Type: ValueInt, Value: int64(i)})
		vars.set([]byte{'v'}, p)
		func() {
			scanner.PushEnds(TokenParenthesesClose)
			defer scanner.PopEnds(1)
			expr := newStmt(scanner, vars)
			if err = expr.execute(); err != nil {
				return
			}
			err = handle(expr.value, i)
		}()
	}
	return
}

func eachItemForFilter(o *array, scanner TokenScanner, vars *variables, handle func(val Value, idx int) error) (ret Value, err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(1)
	for i, p := range o.items {
		scanner.SetOffset(offset)
		vars.set([]byte{'i'}, Value{Type: ValueInt, Value: int64(i)})
		vars.set([]byte{'v'}, p)
		expr := newStmt(scanner, vars)
		if err = expr.execute(); err != nil {
			return
		}
		var bv bool
		if bv, err = expr.value.toBool(); err != nil {
			return
		}
		if !bv {
			continue
		}
		handle(p, i)
	}
	return
}

func (arr *array) set(idx int, val Value) {
	arr.items[idx] = val
}

func (arr *array) get(idx int) Value {
	return arr.items[idx]
}

func (arr *array) delAt(idx int) {
	arr.items = append(arr.items[:idx], arr.items[idx+1:]...)
}

func (arr *array) del(val ...Value) {
	for i := 0; i < len(arr.items); i++ {
		for _, v := range val {
			if arr.items[i].equal(v) {
				arr.delAt(i)
				i--
			}
		}
	}
}

func (arr *array) append(val ...Value) {
	arr.items = append(arr.items, val...)
}

func (arr *array) insertAt(idx int, val Value) {
	tmp := arr.items[idx:]
	arr.items = append(arr.items[:idx], val)
	arr.items = append(arr.items, tmp...)
}

type arrayExecutor struct {
	scanner TokenScanner
	vars    *variables
	value   array
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

func (e *arrayExecutor) items() (val array, err error) {
	e.scanner.PushEnds(TokenBracketsClose, TokenComma)
	defer e.scanner.PopEnds(2)
	e.vars.pushMe(Value{Type: ValueArray, Value: &val})
	defer e.vars.popMe()
	for {
		expr := newStmt(e.scanner, e.vars)
		if err = expr.execute(); err != nil {
			return
		}
		val.append(expr.value)
		if expr.endAt() == TokenBracketsClose {
			return
		}
	}
}
