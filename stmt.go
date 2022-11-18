package djson

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

type expr struct {
	next    *expr
	name    string
	handle  func(val Value, token *Token) (ret Value, leftValue bool, err error)
	scanner TokenScanner
}

func (e *expr) Value(val Value) (ret Value, err error) {
	var leftValue, end, leftValueGetted bool
	var i int
	for {
		if i > 50 {
			err = fmt.Errorf("[%s] calls a lot", e.name)
			return
		}
		i++
		if end, err = e.scanner.Scan(); err != nil || end {
			return
		}
		if val, leftValue, err = e.handle(val, e.scanner.Token()); err != nil {
			return
		}
		if leftValue && leftValueGetted {
			return
		}
		if !leftValue {
			// fmt.Printf("%s\n", e.name)
			ret = val
			continue
		}
		if e.next == nil {
			ret = val
			return
		}
		if val, err = e.next.Value(val); err != nil {
			return
		}
		ret = val
		leftValueGetted = true
	}
}

func (e *expr) WithNext(next *expr) {
	e.next = next
}

func Assign(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "Assign"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		if token.Type != TokenAssignation {
			leftValue = true
			return
		}
		e.scanner.Forward()
		var right Value
		if right, err = e.next.Value(val); err != nil {
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

func Reduction(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "Reduction"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		if token.Type != TokenReduction {
			leftValue = true
			return
		}
		e.scanner.Forward()
		if !val.Bool() {
			return
		}
		var right Value
		if right, err = e.next.Value(val); err != nil {
			return
		}
		ret = right
		return
	}
	return e
}

func Or(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "Or"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		if token.Type != TokenOr {
			leftValue = true
			return
		}
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

func And(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "And"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		if token.Type != TokenAnd {
			leftValue = true
			return
		}
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

func Compare(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "Compare"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		switch token.Type {
		case TokenEqual, TokenNotEqual:
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
		leftValue = true
		return
	}
	return e
}

func AddOrMinus(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "AddOrMinus"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		switch token.Type {
		case TokenAddition:
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Add(term)
			return
		case TokenMinus:
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Minus(term)
			return
		}
		leftValue = true
		return
	}
	return e
}

func MultiplyOrDevide(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "MultiplyOrDevide"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		switch token.Type {
		case TokenMultiplication:
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Multiply(term)
			return
		case TokenDevision:
			scanner.Forward()
			var term Value
			if term, err = e.next.Value(val); err != nil {
				return
			}
			ret, err = val.Devide(term)
			return
		}
		leftValue = true
		return
	}
	return e
}

func Call(scanner TokenScanner, ctx Context) *expr {
	e := &expr{scanner: scanner, name: "Call"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		if !(val.Type == ValueIdentifier && e.scanner.Token().Type == TokenParenthesesOpen) {
			leftValue = true
			return
		}
		scanner.Forward()
		identifier := val.Value.(Identifier)
		e.scanner.PushEnds(TokenParenthesesClose)
		defer e.scanner.PopEnds(1)
		ret, err = identifier.Call(e.scanner, ctx)
		return
	}
	return e
}

func Dot(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "Dot"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		if token.Type != TokenDot {
			leftValue = true
			return
		}
		scanner.Forward()
		var right Value
		right, err = e.next.Value(val)
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

func Range(scanner TokenScanner) *expr {
	e := &expr{scanner: scanner, name: "Range"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		if token.Type != TokenRange {
			leftValue = true
			return
		}
		scanner.Forward()
		if val.Type != ValueInt {
			err = errors.New("range ... must follow an int and be followed by an int too")
			return
		}
		begin, _ := val.Value.(Inter).Int()
		var right Value
		right, err = e.next.Value(val)
		if err != nil {
			return
		}
		if right.Type != ValueInt {
			err = errors.New("range ... must follow an int and be followed by an int too")
			return
		}
		end, _ := right.Value.(Inter).Int()
		ret = RangeValue(int(begin), int(end))
		return
	}
	return e
}

func Factor(scanner TokenScanner, ctx Context) *expr {
	e := &expr{scanner: scanner, name: "Factor"}
	e.handle = func(val Value, token *Token) (ret Value, leftValue bool, err error) {
		leftValue = true
		scanner.Forward()
		switch token.Type {
		case TokenIdentifier:
			ret = Value{Type: ValueIdentifier, Value: &identifier{
				name: token.Raw,
				vars: ctx,
			}}
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
			defer e.scanner.PopEnds(1)
			sub := NewStmt(e.scanner, ctx)
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

type exprs []*expr

func (es exprs) init() *expr {
	for i := 0; i < len(es); i++ {
		if i < len(es)-1 {
			es[i].WithNext(es[i+1])
		}
	}
	return es[0]
}

type stmt struct {
	scanner TokenScanner
	expr    *expr
	value   Value
}

func NewStmt(scanner TokenScanner, ctx Context) *stmt {
	es := exprs([]*expr{
		Assign(scanner),
		Reduction(scanner),
		Or(scanner),
		And(scanner),
		Compare(scanner),
		AddOrMinus(scanner),
		MultiplyOrDevide(scanner),
		Call(scanner, ctx),
		Dot(scanner),
		Range(scanner),
		Factor(scanner, ctx),
	})
	return &stmt{expr: es.init(), scanner: scanner}
}

func (ns *stmt) Execute() (err error) {
	ns.scanner.PushEnds(TokenSemicolon)
	defer ns.scanner.PopEnds(1)
	defer func() {
		if ns.scanner.Token().Type != TokenEOF {
			ns.scanner.Forward()
		}
	}()
	ns.value, err = ns.expr.Value(NullValue())
	return
}
