package djson

import "sync"

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
	rg.RegisterCall("parallel", mapRangeParallelly)
	rg.RegisterCall("each", eachRange)
	return rg
}

func mapRange(val Value, scanner TokenScanner, ctx Context) (ret Value, err error) {
	resetableScanner := NewCachedTokenScanner(scanner)
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	stmt := NewStmtExecutor(resetableScanner, ctx)
	ctx.pushMe(val)
	defer ctx.popMe()
	r := NewArrayWithLength(val.Value.(*range_).Total())
	val.Value.(ItemEachable).Each(func(i int, val Value) bool {
		resetableScanner.ResetRead()
		stmt.AssignVar([]byte{'i'}, IntValue(int64(i)))
		stmt.AssignVar([]byte{'v'}, val)
		if err = stmt.Execute(For(NullValue())); err != nil {
			return false
		}
		if stmt.Exited() {
			Exit()
		}
		r.Set(i, stmt.value)
		return true
	})
	ret = Value{Type: ValueArray, Value: r}
	return
}

func mapRangeParallelly(val Value, scanner TokenScanner, ctx Context) (ret Value, err error) {
	cachedScanner := NewCachedTokenScanner(scanner)
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	if err = cachedScanner.CacheToEnd(); err != nil {
		return
	}
	ctx.pushMe(val)
	defer ctx.popMe()
	total := val.Value.(*range_).Total()
	r := NewArrayWithLength(total)
	var wg sync.WaitGroup
	wg.Add(total)
	var lock sync.Mutex
	val.Value.(ItemEachable).Each(func(i int, val Value) bool {
		go func() {
			defer wg.Done()
			localCtx := ctx.Copy()
			stmt := NewStmtExecutor(cachedScanner.Copy(), localCtx)
			cachedScanner.ResetRead()
			stmt.AssignVar([]byte{'i'}, IntValue(int64(i)))
			stmt.AssignVar([]byte{'v'}, val)
			if e := stmt.Execute(For(NullValue())); e != nil {
				lock.Lock()
				err = e
				lock.Unlock()
				return
			}
			if stmt.Exited() {
				Exit()
			}
			lock.Lock()
			r.Set(i, stmt.value)
			ctx.Merge(localCtx)
			lock.Unlock()
		}()
		return true
	})
	if err != nil {
		return
	}
	scanner.Forward()
	wg.Wait()
	ret = Value{Type: ValueArray, Value: r}
	return
}

func eachRange(val Value, scanner TokenScanner, ctx Context) (ret Value, err error) {
	resetableScanner := NewCachedTokenScanner(scanner)
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(TokenParenthesesClose)
	stmt := NewStmtExecutor(resetableScanner, ctx)
	ctx.pushMe(val)
	defer ctx.popMe()
	val.Value.(ItemEachable).Each(func(i int, val Value) bool {
		resetableScanner.ResetRead()
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
	for i := arr.from; i < arr.to; i++ {
		if !handle(i-arr.from, IntValue(int64(i))) {
			break
		}
	}
}

func (arr *range_) Copy() *range_ {
	return &range_{from: arr.from, to: arr.to}
}

func (arr *range_) Total() int {
	return arr.to - arr.from
}
