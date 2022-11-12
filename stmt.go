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

type stmt struct {
	variables *variables
	value     Value
	scanner   TokenScanner
}

func newStmt(scanner TokenScanner, vars *variables) *stmt {
	return &stmt{
		scanner:   scanner,
		variables: vars,
		value:     Value{Type: ValueNull},
	}
}

func (e *stmt) endAt() TokenType {
	return e.scanner.EndAt()
}

func (e *stmt) execute() (err error) {
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
				ret, err = val.assign(right)
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
				var b bool
				b, err = val.toBool()
				if err != nil || !b {
					return
				}
				var right Value
				right, err = e.or()
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
				right, err = e.and()
				ret, err = val.or(right)
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
				right, err = e.compare()
				ret, err = val.and(right)
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
				right, err = e.expr()
				if err != nil {
					return
				}
				boo := val.equal(right)
				if TokenNotEqual == e.scanner.Token().Type {
					boo = !boo
				}
				ret = Value{Type: ValueBool, Value: boo}
			})
			return
		case TokenGreateThan, TokenGreateThanEqual, TokenLessThan, TokenLessThanEqual:
			var right Value
			right, err = e.expr()
			if err != nil {
				return
			}
			var com int
			com, err = val.compare(right)
			if err != nil {
				return
			}
			switch e.scanner.Token().Type {
			case TokenGreateThan:
				ret = Value{Type: ValueBool, Value: com > 0}
			case TokenGreateThanEqual:
				ret = Value{Type: ValueBool, Value: com >= 0}
			case TokenLessThan:
				ret = Value{Type: ValueBool, Value: com < 0}
			case TokenLessThanEqual:
				ret = Value{Type: ValueBool, Value: com <= 0}
			}
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
				right, err = e.call()
				if err != nil {
					return
				}
				if right.Type != ValueIdentifier {
					err = errors.New("dot must follow an identifier")
					return
				}
				right.p = &val
				ret = right
			})
			return
		}
		if terminated {
			done = true
			ret = val
			return
		}
		ret, err = e.call()
		terminated = true
		return
	})
}

func (e *stmt) call() (Value, error) {
	terminated := false
	return e.calc(func(val Value) (ret Value, done bool, err error) {
		if val.Type == ValueIdentifier && e.scanner.Token().Type == TokenParenthesesOpen {
			e.scanner.PushEnds(TokenParenthesesClose)
			defer e.scanner.PopEnds(1)
			e.useToken(func() {
				ret, err = val.Value.(*identifier).call(e.scanner, e.variables)
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
					name:      token.Raw,
					variables: e.variables,
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
				ret = Value{Value: token.Raw, Type: ValueString}
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
				sub := newStmt(e.scanner, e.variables)
				if err = sub.execute(); err == nil {
					ret = sub.value
				}
			})
			return
		case TokenBracketsOpen:
			e.useToken(func() {
				sub := newArrayExecutor(e.scanner, e.variables)
				if err = sub.execute(); err == nil {
					ret = Value{Type: ValueArray, Value: sub.value}
				}
			})
			return
		case TokenBraceOpen:
			e.useToken(func() {
				sub := newObjectExecutor(e.scanner, e.variables)
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
