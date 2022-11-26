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

func mapRange(val Value, scanner TokenScanner, ctx Context) (ret Value, err error) {
	resetableScanner := NewTokenRecordScanner(scanner)
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	stmt := NewStmtExecutor(resetableScanner, ctx)
	ctx.pushMe(val)
	defer ctx.popMe()
	r := NewArray()
	val.Value.(ItemEachable).Each(func(i int, val Value) bool {
		resetableScanner.Reset()
		stmt.AssignVar([]byte{'i'}, IntValue(int64(i)))
		stmt.AssignVar([]byte{'v'}, val)
		if err = stmt.Execute(For(NullValue())); err != nil {
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

func eachRange(val Value, scanner TokenScanner, ctx Context) (ret Value, err error) {
	resetableScanner := NewTokenRecordScanner(scanner)
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	stmt := NewStmtExecutor(resetableScanner, ctx)
	ctx.pushMe(val)
	defer ctx.popMe()
	val.Value.(ItemEachable).Each(func(i int, val Value) bool {
		resetableScanner.Reset()
		stmt.AssignVar([]byte{'i'}, IntValue(int64(i)))
		stmt.AssignVar([]byte{'v'}, val)
		if err = stmt.Execute(For(NullValue())); err != nil {
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
