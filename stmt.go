package djson

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var (
	errExit = errors.New("__exit__")
)

type stmt struct {
	next    *stmt                                                              // next stmt should try, the priority of this stmt is always lower than next
	name    string                                                             // stmt name
	handle  func(val Value, token *Token) (handled bool, ret Value, err error) // match token and handle the stmt
	scanner TokenScanner                                                       // scanner
	opt     *option                                                            // opt
}

func (e *stmt) Value(val Value) (ret Value, err error) {
	terminal := e.next == nil
	var matched, end, nextTried bool
	var ht Value
	for {
		if end, err = e.scanner.Scan(); err != nil || end {
			return
		}
		if terminal {
			_, ret, err = e.handle(val, e.scanner.Token())
			if e.opt.debug {
				fmt.Printf("%s\n", e.name)
			}
			return
			// try this level
		} else if matched, ht, err = e.handle(val, e.scanner.Token()); err != nil {
			return
		} else if matched {
			nextTried = true
			val = ht
			ret = val
			if !e.opt.debug {
				continue
			}
			fmt.Printf("%s\n", e.name)
		} else if !nextTried {
			// try higher priorities
			if val, err = e.next.Value(val); err != nil {
				return
			}
			ret = val
			nextTried = true
		} else {
			return
		}
	}
}

func Exit() {
	panic(errExit)
}

func Assign(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "Assign"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if token.Type != TokenAssignation {
			return
		}
		matched = true
		e.scanner.Forward()
		var right Value
		if right, err = e.next.Value(NullValue()); err != nil {
			return
		}
		if val.Type != ValueIdentifier {
			err = errors.New("only identifier can assign to")
			return
		}
		err = val.Value.(Identifier).Assign(right)
		ret = right
		return
	}
	return e
}

func Reduction(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "Reduction"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if token.Type != TokenReduction {
			return
		}
		matched = true
		e.scanner.Forward()
		var right Value
		if right, err = e.next.Value(NullValue()); err != nil {
			return
		}
		if val.Bool() {
			ret = right
		}
		return
	}
	return e
}

func Or(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "Or"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if token.Type != TokenOr {
			return
		}
		matched = true
		e.scanner.Forward()
		if !val.Bool() {
			return
		}
		var right Value
		if right, err = e.next.Value(val); err != nil {
			return
		}
		ret = val.Or(right)
		return
	}
	return e
}

func And(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "And"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if token.Type != TokenAnd {
			return
		}
		matched = true
		e.scanner.Forward()
		if !val.Bool() {
			return
		}
		var right Value
		if right, err = e.next.Value(val); err != nil {
			return
		}
		ret = val.And(right)
		return
	}
	return e
}

func Compare(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "Compare"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		switch token.Type {
		case TokenEqual, TokenNotEqual:
			matched = true
			scanner.Forward()
			var right Value
			right, err = e.next.Value(val)
			if err != nil {
				return
			}
			boo := val.Equal(right)
			if TokenNotEqual == token.Type {
				boo = !boo
			}
			ret = BoolValue(boo)
			return
		case TokenGreateThan, TokenGreateThanEqual, TokenLessThan, TokenLessThanEqual:
			matched = true
			scanner.Forward()
			var right Value
			right, err = e.next.Value(val)
			if err != nil {
				return
			}
			var com int
			com, err = val.Compare(right)
			if err != nil {
				return
			}
			switch token.Type {
			case TokenGreateThan:
				ret = BoolValue(com > 0)
			case TokenGreateThanEqual:
				ret = BoolValue(com >= 0)
			case TokenLessThan:
				ret = BoolValue(com < 0)
			case TokenLessThanEqual:
				ret = BoolValue(com <= 0)
			}
			return
		}
		return
	}
	return e
}

func AddOrMinus(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "AddOrMinus"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		switch token.Type {
		case TokenAddition:
			matched = true
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Add(term)
			return
		case TokenMinus:
			matched = true
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Minus(term)
			return
		}
		return
	}
	return e
}

func MultiplyOrDevide(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "MultiplyOrDevide"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		switch token.Type {
		case TokenMultiplication:
			matched = true
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Multiply(term)
			return
		case TokenDevision:
			matched = true
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Devide(term)
			return
		}
		return
	}
	return e
}

func Mod(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "Mod"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if token.Type != TokenMod {
			return
		}
		matched = true
		scanner.Forward()
		var term Value
		if term, err = e.next.Value(val); err != nil {
			return
		}
		ret, err = val.Mod(term)
		return
	}
	return e
}

func Call(scanner TokenScanner, ctx Context) *stmt {
	e := &stmt{scanner: scanner, name: "Call"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if !(val.Type == ValueIdentifier && e.scanner.Token().Type == TokenParenthesesOpen) {
			return
		}
		matched = true
		scanner.Forward()
		identifier := val.Value.(Identifier)
		e.scanner.PushEnds(TokenParenthesesClose)
		defer e.scanner.PopEnds(TokenParenthesesClose)
		if ret, err = identifier.Call(e.scanner, ctx); err != nil {
			return
		}
		return
	}
	return e
}

func Dot(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "Dot"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if token.Type != TokenDot {
			return
		}
		matched = true
		scanner.Forward()
		var right Value
		right, err = e.next.Value(NullValue())
		if err != nil {
			return
		}
		if right.Type != ValueIdentifier {
			err = errors.New("an identifier must be followed by dot")
		}
		right.Value.(Identifier).SetParent(val)
		ret = right
		return
	}
	return e
}

func Range(scanner TokenScanner) *stmt {
	e := &stmt{scanner: scanner, name: "Range"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		if token.Type != TokenRange {
			return
		}
		matched = true
		scanner.Forward()
		var begin, end int64
		if inter, ok := val.Value.(Inter); ok {
			if begin, err = inter.Int(); err != nil {
				err = fmt.Errorf("can't convert to int for range begin: %w", err)
				return
			}
		} else {
			err = errors.New("range ... must follow an int and be followed by an int too")
			return
		}
		var right Value
		right, err = e.next.Value(val)
		if err != nil {
			return
		}
		if inter, ok := right.Value.(Inter); ok {
			if end, err = inter.Int(); err != nil {
				err = fmt.Errorf("can't convert to int for range end: %w", err)
				return
			}
		} else {
			err = errors.New("range ... must follow an int and be followed by an int too")
			return
		}
		ret = RangeValue(int(begin), int(end))
		return
	}
	return e
}

func Factor(scanner TokenScanner, ctx Context) *stmt {
	e := &stmt{scanner: scanner, name: "Factor"}
	e.handle = func(val Value, token *Token) (matched bool, ret Value, err error) {
		scanner.Forward()
		switch token.Type {
		case TokenIdentifier:
			ret = Value{Type: ValueIdentifier, Value: &identifier{
				name: token.Raw,
				vars: ctx,
			}}
		case TokenExit:
			Exit()
		case TokenTrue:
			ret = BoolValue(true)
		case TokenFalse:
			ret = BoolValue(false)
		case TokenString:
			ret = StringValue(token.Raw...)
		case TokenNumber:
			if bytes.Contains(token.Raw, []byte{'.'}) {
				v, _ := strconv.ParseFloat(string(token.Raw), 64)
				ret = FloatValue(v)
				return
			}
			v, _ := strconv.ParseInt(string(token.Raw), 10, 64)
			ret = IntValue(v)
		case TokenParenthesesOpen:
			e.scanner.PushEnds(TokenParenthesesClose)
			defer e.scanner.PopEnds(TokenParenthesesClose)
			sub := NewStmtExecutor(e.scanner, ctx)
			if err = sub.Execute(); err == nil {
				ret = sub.value
			}
		case TokenBracketsOpen:
			sub := newArrayExecutor(e.scanner, ctx)
			if err = sub.execute(); err == nil {
				ret = ArrayValue(sub.value)
			}
		case TokenBraceOpen:
			sub := newObjectExecutor(e.scanner, ctx)
			if err = sub.execute(); err == nil {
				ret = ObjectValue(sub.value)
			}
		default:
			err = fmt.Errorf("unexpected token [%s] at %d, %d", token.Name(), token.Row, token.Col)
		}
		return
	}
	return e
}

type stmts []*stmt

func (es stmts) init(opt *option) *stmt {
	for i := 0; i < len(es); i++ {
		es[i].opt = opt
		if i < len(es)-1 {
			es[i].next = es[i+1]
		}
	}
	return es[0]
}

type stmtExecutor struct {
	scanner TokenScanner
	expr    *stmt
	value   Value
	opt     *option
	exited  bool
}

type option struct {
	debug bool
}

type StmtOption func(opt *option)

func Debug() StmtOption {
	return func(opt *option) {
		opt.debug = true
	}
}

func NewStmtExecutor(scanner TokenScanner, ctx Context, opts ...StmtOption) *stmtExecutor {
	es := stmts([]*stmt{
		Assign(scanner),
		Reduction(scanner),
		Or(scanner),
		And(scanner),
		Compare(scanner),
		AddOrMinus(scanner),
		MultiplyOrDevide(scanner),
		Mod(scanner),
		Call(scanner, ctx),
		Dot(scanner),
		Range(scanner),
		Factor(scanner, ctx),
	})
	opt := &option{}
	for _, apply := range opts {
		apply(opt)
	}
	return &stmtExecutor{expr: es.init(opt), scanner: scanner, opt: opt}
}

func (ns *stmtExecutor) Execute() (err error) {
	ns.scanner.PushEnds(TokenSemicolon)
	defer ns.scanner.PopEnds(TokenSemicolon)
	defer func() {
		if ns.scanner.Token().Type != TokenEOF {
			ns.scanner.Forward()
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && errors.Is(err, errExit) {
				ns.exited = true
				return
			}
			panic(r)
		}
	}()
	var end bool
	for {
		if end, err = ns.scanner.Scan(); end || err != nil {
			return
		}
		if ns.value, err = ns.expr.Value(ns.value); err != nil {
			return
		}
	}
}

func (ns *stmtExecutor) Exited() bool {
	return ns.exited
}

func (ns *stmtExecutor) Value() Value {
	return ns.value
}
