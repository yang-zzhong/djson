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
	value     value
	scanner   *tokenScanner
}

func newStmt(scanner *tokenScanner, vars *variables) *stmt {
	return &stmt{
		scanner:   scanner,
		variables: vars,
		value:     value{typ: valueNull},
	}
}

func (e *stmt) endAt() TokenType {
	return e.scanner.endAt()
}

func (e *stmt) execute() (err error) {
	e.scanner.pushEnds(TokenSemicolon)
	defer e.scanner.popEnds(1)
	defer func() {
		if e.scanner.token.Type != TokenEOF {
			e.scanner.forward()
		}
	}()
	e.value, err = e.assign()
	return
}

func (e *stmt) assign() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenAssignation {
			e.useToken(func() {
				var right value
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

func (e *stmt) reduct() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenReduction {
			e.useToken(func() {
				var b bool
				b, err = val.toBool()
				if err != nil || !b {
					return
				}
				var right value
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

func (e *stmt) or() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenOr {
			e.useToken(func() {
				var right value
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

func (e *stmt) calc(handle func(input value) (val value, done bool, err error)) (val value, err error) {
	var end bool
	var done bool
	for {
		if end, err = e.scanner.scan(); err != nil || end {
			return
		}
		val, done, err = handle(val)
		if err != nil || done {
			return
		}
	}
}

func (e *stmt) and() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenAnd {
			e.useToken(func() {
				var right value
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

func (e *stmt) compare() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		switch e.scanner.token.Type {
		case TokenEqual, TokenNotEqual:
			e.useToken(func() {
				var right value
				right, err = e.expr()
				if err != nil {
					return
				}
				boo := val.equal(right)
				if TokenNotEqual == e.scanner.token.Type {
					boo = !boo
				}
				ret = value{typ: valueBool, value: boo}
			})
			return
		case TokenGreateThan, TokenGreateThanEqual, TokenLessThan, TokenLessThanEqual:
			var right value
			right, err = e.expr()
			if err != nil {
				return
			}
			var com int
			com, err = val.compare(right)
			if err != nil {
				return
			}
			switch e.scanner.token.Type {
			case TokenGreateThan:
				ret = value{typ: valueBool, value: com > 0}
			case TokenGreateThanEqual:
				ret = value{typ: valueBool, value: com >= 0}
			case TokenLessThan:
				ret = value{typ: valueBool, value: com < 0}
			case TokenLessThanEqual:
				ret = value{typ: valueBool, value: com <= 0}
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

func (e *stmt) expr() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		switch e.scanner.token.Type {
		case TokenAddition:
			e.useToken(func() {
				var term value
				term, err = e.term()
				if err != nil {
					return
				}
				ret, err = val.add(term)
			})
			return
		case TokenMinus:
			e.useToken(func() {
				var term value
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

func (e *stmt) term() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		switch e.scanner.token.Type {
		case TokenMultiplication:
			e.useToken(func() {
				var factor value
				factor, err = e.dot()
				ret, err = val.multiply(factor)
			})
			return
		case TokenDevision:
			e.useToken(func() {
				var factor value
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

func (e *stmt) dot() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenDot {
			e.useToken(func() {
				var right value
				right, err = e.call()
				if err != nil {
					return
				}
				if right.typ != valueIdentifier {
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

func (e *stmt) call() (value, error) {
	terminated := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if val.typ == valueIdentifier && e.scanner.token.Type == TokenParenthesesOpen {
			e.scanner.pushEnds(TokenParenthesesClose)
			defer e.scanner.popEnds(1)
			e.useToken(func() {
				ret, err = val.value.(*identifier).call(e.scanner, e.variables)
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

func (e *stmt) factor() (value, error) {
	return e.calc(func(val value) (ret value, done bool, err error) {
		token := e.scanner.token
		done = true
		switch token.Type {
		case TokenIdentifier:
			e.useToken(func() {
				ret = value{typ: valueIdentifier, value: &identifier{
					name:      token.Raw,
					variables: e.variables,
				}}
			})
			return
		case TokenTrue:
			e.useToken(func() {
				ret = value{value: true, typ: valueBool}
			})
			return
		case TokenFalse:
			e.useToken(func() {
				ret = value{value: false, typ: valueBool}
			})
			return
		case TokenString:
			e.useToken(func() {
				ret = value{value: token.Raw, typ: valueString}
			})
			return
		case TokenNumber:
			e.useToken(func() {
				ret = e.number(token.Raw)
			})
			return
		case TokenParenthesesOpen:
			e.useToken(func() {
				e.scanner.pushEnds(TokenParenthesesClose)
				defer e.scanner.popEnds(1)
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
					ret = value{typ: valueArray, value: sub.value}
				}
			})
			return
		case TokenBraceOpen:
			e.useToken(func() {
				sub := newObjectExecutor(e.scanner, e.variables)
				if err = sub.execute(); err == nil {
					ret = value{typ: valueObject, value: sub.value}
				}
			})
			return
		default:
			err = ErrFromToken(UnexpectedToken, token)
			return
		}
	})
}

func (e *stmt) number(bs []byte) value {
	if bytes.Contains(bs, []byte{'.'}) {
		v, _ := strconv.ParseFloat(string(bs), 64)
		return value{typ: valueFloat, value: v}
	}
	v, _ := strconv.ParseInt(string(bs), 10, 64)
	return value{typ: valueInt, value: v}
}

func (e *stmt) useToken(useToken func()) {
	e.scanner.forward()
	useToken()
}
