package djson

type array struct {
	*callableImp
	items []value
}

func newArray(items ...value) *array {
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

func setArray(val value, scanner *tokenScanner, vars *variables) (value, error) {
	o := val.value.(*array)
	return eachItemForSet(o, scanner, vars, func(val value, idx int) error {
		o.items[idx] = val
		return nil
	})
}

func delArray(caller value, scanner *tokenScanner, vars *variables) (value, error) {
	o := caller.value.(*array)
	return eachItemForFilter(o, scanner, vars, func(_ value, idx int) error {
		o.items = append(o.items[:idx], o.items[idx+1:]...)
		return nil
	})
}

func filterArray(caller value, scanner *tokenScanner, vars *variables) (ret value, err error) {
	o := caller.value.(*array)
	no := newArray()
	_, err = eachItemForFilter(o, scanner, vars, func(val value, idx int) error {
		no.items = append(no.items, val)
		return nil
	})
	ret = value{value: no, typ: valueArray}
	return
}

func eachItemForSet(o *array, scanner *tokenScanner, vars *variables, handle func(val value, idx int) error) (ret value, err error) {
	offset := scanner.offset()
	for i, p := range o.items {
		scanner.setOffset(offset)
		vars.set([]byte{'i'}, value{typ: valueInt, value: int64(i)})
		vars.set([]byte{'v'}, p)
		var bv bool
		func() {
			scanner.pushEnds(TokenParenthesesClose, TokenReduction)
			defer scanner.popEnds(2)
			expr := newStmt(scanner, vars)
			if err = expr.execute(); err != nil {
				return
			}
			if scanner.endAt() == TokenParenthesesClose {
				if err = handle(expr.value, i); err != nil {
					return
				}
			}
			if bv, err = expr.value.toBool(); err != nil {
				return
			}
		}()
		if err != nil {
			return
		}
		if !bv {
			continue
		}
		func() {
			scanner.pushEnds(TokenParenthesesClose)
			defer scanner.popEnds(1)
			expr := newStmt(scanner, vars)
			if err = expr.execute(); err != nil {
				return
			}
			if err = handle(expr.value, i); err != nil {
				return
			}
		}()
		if err != nil {
			return
		}
	}
	return
}

func eachItemForFilter(o *array, scanner *tokenScanner, vars *variables, handle func(val value, idx int) error) (ret value, err error) {
	offset := scanner.offset()
	scanner.pushEnds(TokenParenthesesClose)
	defer scanner.popEnds(1)
	for i, p := range o.items {
		scanner.setOffset(offset)
		vars.set([]byte{'i'}, value{typ: valueInt, value: int64(i)})
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

func (arr *array) set(idx int, val value) {
	arr.items[idx] = val
}

func (arr *array) get(idx int) value {
	return arr.items[idx]
}

func (arr *array) delAt(idx int) {
	arr.items = append(arr.items[:idx], arr.items[idx+1:]...)
}

func (arr *array) del(val ...value) {
	for i := 0; i < len(arr.items); i++ {
		for _, v := range val {
			if arr.items[i].equal(v) {
				arr.delAt(i)
				i--
			}
		}
	}
}

func (arr *array) append(val ...value) {
	arr.items = append(arr.items, val...)
}

func (arr *array) insertAt(idx int, val value) {
	tmp := arr.items[idx:]
	arr.items = append(arr.items[:idx], val)
	arr.items = append(arr.items, tmp...)
}

type arrayExecutor struct {
	scanner *tokenScanner
	vars    *variables
	value   array
}

func newArrayExecutor(scanner *tokenScanner, vs *variables) *arrayExecutor {
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
	e.scanner.pushEnds(TokenBracketsClose, TokenComma)
	defer e.scanner.popEnds(2)
	e.vars.pushMe(value{typ: valueArray, value: &val})
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
