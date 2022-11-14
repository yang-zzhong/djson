package djson

type range_ struct {
	from int
	to   int
	*callableImp
}

func NewRange(from, to int) *range_ {
	rg := &range_{
		from:        from,
		to:          to,
		callableImp: newCallable("range"),
	}
	rg.register("map", mapRange)
	return rg
}

func mapRange(val Value, scanner TokenScanner, vars *variables) (ret Value, err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(1)
	r := NewArray()
	val.Value.(Array).Each(func(i int, val Value) bool {
		scanner.SetOffset(offset)
		vars.set([]byte{'i'}, Value{Type: ValueInt, Value: int64(i)})
		vars.set([]byte{'v'}, val)
		expr := NewStmt(scanner, vars)
		if err = expr.Execute(); err != nil {
			return false
		}
		r.Append(expr.value)
		return true
	})
	ret = Value{Type: ValueArray, Value: r}
	return
}

func (arr *range_) Set(idx int, val Value) {
	// range not support set
}

func (arr *range_) Get(idx int) Value {
	// range not support set
	return Value{Type: ValueNull}
}

func (arr *range_) Del(idx int) {
	// range not support del
}

func (arr *range_) Each(handle func(i int, val Value) bool) {
	for i := arr.from; i <= arr.to; i++ {
		if !handle(i-arr.from, Value{Type: ValueInt, Value: int64(i)}) {
			break
		}
	}
}

func (arr *range_) Copy() Array {
	return &range_{from: arr.from, to: arr.to}
}

func (arr *range_) Total() int {
	return arr.to - arr.from + 1
}

func (arr *range_) Append(val ...Value) {
	// range not support append
}
