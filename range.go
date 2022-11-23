package djson

type ItemEachable interface {
	Each(func(i int, val Value) bool)
}

type range_ struct {
	from int
	to   int
	*CallableRegister
}

func NewRange(from, to int) *range_ {
	rg := &range_{
		from:             from,
		to:               to,
		CallableRegister: NewCallableRegister("range"),
	}
	rg.RegisterCall("map", mapRange)
	rg.RegisterCall("each", eachRange)
	return rg
}

func mapRange(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	r := NewArray()
	val.Value.(ItemEachable).Each(func(i int, val Value) bool {
		scanner.SetOffset(offset)
		vars.Assign([]byte{'i'}, IntValue(int64(i)))
		vars.Assign([]byte{'v'}, val)
		stmt := NewStmtExecutor(scanner, vars)
		if err = stmt.Execute(); err != nil {
			return false
		}
		if stmt.Exited() {
			Exit()
		}
		r.Append(stmt.value)
		return true
	})
	ret = Value{Type: ValueArray, Value: r}
	return
}

func eachRange(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	offset := scanner.Offset()
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	val.Value.(ItemEachable).Each(func(i int, val Value) bool {
		scanner.SetOffset(offset)
		vars.Assign([]byte{'i'}, IntValue(int64(i)))
		vars.Assign([]byte{'v'}, val)
		stmt := NewStmtExecutor(scanner, vars)
		if err = stmt.Execute(); err != nil {
			return false
		}
		if stmt.Exited() {
			Exit()
		}
		return true
	})
	return
}

func (arr *range_) Each(handle func(i int, val Value) bool) {
	for i := arr.from; i <= arr.to; i++ {
		if !handle(i-arr.from, IntValue(int64(i))) {
			break
		}
	}
}

func (arr *range_) Copy() *range_ {
	return &range_{from: arr.from, to: arr.to}
}

func (arr *range_) Total() int {
	return arr.to - arr.from + 1
}
