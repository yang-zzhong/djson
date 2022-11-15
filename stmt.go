package djson

import (
	"bytes"
	"errors"
	"strconv"
)

// BNF (Context Free Grammar)
// stmt -> stmt = assignation | assignation
// assignation -> assignation = reduction | reduction
// reduction -> reduction => or | or
// or -> or && and | and
// and -> compare > expr | compare >= expr | compare < expr | compare <= expr | compare == expr | compare != expr | expr
// expr -> expr + term | expr - term | term
// term -> term * dot | term / dot | dot
// dot -> dot . identifier call | call
// call -> call identifier(stmt) | factor
// factor -> digit | identifier | string | null | (call)

type Stmt interface {
	Execute() error
	Value() Value
}

type stmt struct {
	ctx     Context
	value   Value
	scanner TokenScanner
}

func NewStmt(scanner TokenScanner, vars Context) *stmt {
	return &stmt{
		scanner: scanner,
		ctx:     vars,
		value:   Value{Type: ValueNull},
	}
}

func (e *stmt) endAt() TokenType {
	return e.scanner.EndAt()
}

func (e *stmt) Value() Value {
	return e.value
}

func (e *stmt) Execute() (err error) {
	e.scanner.PushEnds(TokenSemicolon)
	defer e.scanner.PopEnds(1)
	defer func() {
		if e.scanner.Token().Type != TokenEOF {
			e.scanner.Forward()
		}
	}()
	e.value, err = e.assign()
	return
}

func (e *stmt) assign() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		if e.scanner.Token().Type == TokenAssignation {
			e.useToken(func() {
				var right Value
				right, err = e.reduct()
				if err != nil {
					return
				}
				if val.Type != ValueIdentifier {
					err = errors.New("only identifier can assign to")
					return
				}
				err = val.Value.(Identifier).Assign(right)
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.reduct()
		terminated = true
		return
	})
}

func (e *stmt) reduct() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		if e.scanner.Token().Type == TokenReduction {
			e.useToken(func() {
				if !val.toBool() {
					return
				}
				var right Value
				if right, err = e.or(); err != nil {
					return
				}
				ret = right
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.or()
		terminated = true
		return
	})
}

func (e *stmt) or() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		if e.scanner.Token().Type == TokenOr {
			e.useToken(func() {
				var right Value
				if right, err = e.and(); err != nil {
					return
				}
				ret = val.or(right)
			})
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.and()
		terminated = true
		return
	})
}

func (e *stmt) calc(handle func(input Value) (val Value, done bool, err error)) (val Value, err error) {
	var end bool
	var done bool
	for {
		if end, err = e.scanner.Scan(); err != nil || end {
			return
		}
		val, done, err = handle(val)
		if err != nil || done {
			return
		}
	}
}

func (e *stmt) and() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		if e.scanner.Token().Type == TokenAnd {
			e.useToken(func() {
				var right Value
				if right, err = e.compare(); err != nil {
					return
				}
				ret = val.and(right)
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.compare()
		terminated = true
		return
	})
}

func (e *stmt) compare() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		switch e.scanner.Token().Type {
		case TokenEqual, TokenNotEqual:
			e.useToken(func() {
				var right Value
				var token = e.scanner.Token()
				right, err = e.expr()
				if err != nil {
					return
				}
				boo := val.equal(right)
				if TokenNotEqual == token.Type {
					boo = !boo
				}
				ret = Value{Type: ValueBool, Value: boo}
			})
			return
		case TokenGreateThan, TokenGreateThanEqual, TokenLessThan, TokenLessThanEqual:
			e.useToken(func() {
				var right Value
				var token = e.scanner.Token()
				right, err = e.expr()
				if err != nil {
					return
				}
				var com int
				com, err = val.compare(right)
				if err != nil {
					return
				}
				switch token.Type {
				case TokenGreateThan:
					ret = Value{Type: ValueBool, Value: com > 0}
				case TokenGreateThanEqual:
					ret = Value{Type: ValueBool, Value: com >= 0}
				case TokenLessThan:
					ret = Value{Type: ValueBool, Value: com < 0}
				case TokenLessThanEqual:
					ret = Value{Type: ValueBool, Value: com <= 0}
				}
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.expr()
		terminated = true
		return
	})
}

func (e *stmt) expr() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		switch e.scanner.Token().Type {
		case TokenAddition:
			e.useToken(func() {
				var term Value
				term, err = e.term()
				if err != nil {
					return
				}
				ret, err = val.add(term)
			})
			return
		case TokenMinus:
			e.useToken(func() {
				var term Value
				term, err = e.term()
				if err != nil {
					return
				}
				ret, err = val.minus(term)
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.term()
		terminated = true
		return
	})
}

func (e *stmt) term() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		switch e.scanner.Token().Type {
		case TokenMultiplication:
			e.useToken(func() {
				var factor Value
				factor, err = e.dot()
				ret, err = val.multiply(factor)
			})
			return
		case TokenDevision:
			e.useToken(func() {
				var factor Value
				factor, err = e.dot()
				ret, err = val.devide(factor)
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.dot()
		terminated = true
		return
	})
}

func (e *stmt) dot() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		if e.scanner.Token().Type == TokenDot {
			e.useToken(func() {
				var right Value
				right, err = e.call(val)
				if err != nil {
					return
				}
				ret = right
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.call(val)
		terminated = true
		return
	})
}

func (e *stmt) call(left Value) (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		if val.Type == ValueIdentifier && e.scanner.Token().Type == TokenParenthesesOpen {
			identifier := val.Value.(Identifier)
			identifier.SetParent(left)
			e.scanner.PushEnds(TokenParenthesesClose)
			defer e.scanner.PopEnds(1)
			e.useToken(func() {
				ret, err = identifier.Call(e.scanner, e.ctx)
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.ranges()
		terminated = true
		return
	})
}

func (e *stmt) ranges() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		token := e.scanner.Token()
		if token.Type == TokenRange {
			if val.Type != ValueInt {
				err = errors.New("range ... must follow an int and be followed by an int too")
				return
			}
			e.useToken(func() {
				begin := val.Value.(int64)
				var right Value
				right, err = e.factor()
				if err != nil {
					return
				}
				if right.Type != ValueInt {
					err = errors.New("range ... must follow an int and be followed by an int too")
					return
				}
				ret = Value{Type: ValueRange, Value: NewRange(int(begin), int(right.Value.(int64)))}
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.factor()
		terminated = true
		return
	})
}

func (e *stmt) factor() (Value, error) {
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		token := e.scanner.Token()
		done = true
		switch token.Type {
		case TokenIdentifier:
			e.useToken(func() {
				ret = Value{Type: ValueIdentifier, Value: &identifier{
					name: token.Raw,
					vars: e.ctx,
				}}
			})
			return
		case TokenTrue:
			e.useToken(func() {
				ret = Value{Value: true, Type: ValueBool}
			})
			return
		case TokenFalse:
			e.useToken(func() {
				ret = Value{Value: false, Type: ValueBool}
			})
			return
		case TokenString:
			e.useToken(func() {
				ret = Value{Value: NewString(token.Raw...), Type: ValueString}
			})
			return
		case TokenNumber:
			e.useToken(func() {
				ret = e.number(token.Raw)
			})
			return
		case TokenParenthesesOpen:
			e.useToken(func() {
				e.scanner.PushEnds(TokenParenthesesClose)
				defer e.scanner.PopEnds(1)
				sub := NewStmt(e.scanner, e.ctx)
				if err = sub.Execute(); err == nil {
					ret = sub.value
				}
			})
			return
		case TokenBracketsOpen:
			e.useToken(func() {
				sub := newArrayExecutor(e.scanner, e.ctx)
				if err = sub.execute(); err == nil {
					ret = Value{Type: ValueArray, Value: sub.value}
				}
			})
			return
		case TokenBraceOpen:
			e.useToken(func() {
				sub := newObjectExecutor(e.scanner, e.ctx)
				if err = sub.execute(); err == nil {
					ret = Value{Type: ValueObject, Value: sub.value}
				}
			})
			return
		default:
			err = ErrFromToken(UnexpectedToken, token)
			return
		}
	})
}

func (e *stmt) number(bs []byte) Value {
	if bytes.Contains(bs, []byte{'.'}) {
		v, _ := strconv.ParseFloat(string(bs), 64)
		return Value{Type: ValueFloat, Value: v}
	}
	v, _ := strconv.ParseInt(string(bs), 10, 64)
	return Value{Type: ValueInt, Value: v}
}

func (e *stmt) useToken(useToken func()) {
	e.scanner.Forward()
	useToken()
}
