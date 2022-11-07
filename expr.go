package djson

import (
	"bytes"
	"errors"
	"strconv"
)

// BNF (Context Free Grammar)
// call -> call callable( dot | dot
// dot -> dot . assign | assign
// assign -> assign = or | or
// or -> or || and | and
// and -> and && compare | compare
// compare -> compare > expr | compare >= expr | compare < expr | compare <= expr | compare == expr | compare != expr | expr
// expr -> expr + term | expr - term | term
// term -> term * factor | term / factor | factor
// factor -> digit | identifier | string | null | (call)

type expr struct {
	variables *variables
	value     value
	scanner   *tokenScanner
}

func newExpr(scanner *tokenScanner, vars *variables) *expr {
	return &expr{
		scanner:   scanner,
		variables: vars,
		value:     value{typ: valueNull},
	}
}

func (e *expr) endAt() TokenType {
	return e.scanner.endAt()
}

func (e *expr) execute() (err error) {
	e.value, err = e.call()
	return
}

func (e *expr) call() (value, error) {
	undered := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if val.typ == valueIdentifier && e.scanner.token.Type == TokenParenthesesOpen {
			e.scanner.pushEnds(TokenParenthesesClose)
			defer e.scanner.popEnds(1)
			e.useToken(func() {
				ret, err = val.value.(*identifier).call(e.scanner, e.variables)
			})
			return
		}
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.dot()
		undered = true
		return
	})
}

func (e *expr) dot() (value, error) {
	undered := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenDot {
			e.useToken(func() {
				var right value
				right, err = e.assign()
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
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.assign()
		undered = true
		return
	})
}

func (e *expr) assign() (value, error) {
	undered := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenAssignation {
			e.useToken(func() {
				var right value
				right, err = e.or()
				if err != nil {
					return
				}
				ret, err = val.assign(right)
			})
			return
		}
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.or()
		undered = true
		return
	})
}

func (e *expr) or() (value, error) {
	undered := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenOr {
			e.useToken(func() {
				var right value
				right, err = e.and()
				ret, err = val.or(right)
			})
		}
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.and()
		undered = true
		return
	})
}

func (e *expr) calc(handle func(input value) (val value, done bool, err error)) (val value, err error) {
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

func (e *expr) and() (value, error) {
	undered := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		if e.scanner.token.Type == TokenAnd {
			e.useToken(func() {
				var right value
				right, err = e.compare()
				ret, err = val.and(right)
			})
			return
		}
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.compare()
		undered = true
		return
	})
}

func (e *expr) compare() (value, error) {
	undered := false
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
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.expr()
		undered = true
		return
	})
}

func (e *expr) expr() (value, error) {
	undered := false
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
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.term()
		undered = true
		return
	})
}

func (e *expr) term() (value, error) {
	undered := false
	return e.calc(func(val value) (ret value, done bool, err error) {
		switch e.scanner.token.Type {
		case TokenMultiplication:
			e.useToken(func() {
				var factor value
				factor, err = e.factor()
				ret, err = val.multiply(factor)
			})
		case TokenDevision:
			e.useToken(func() {
				var factor value
				factor, err = e.factor()
				ret, err = val.devide(factor)
			})
		}
		if undered {
			done = true
			ret = val
			return
		}
		ret, err = e.factor()
		undered = true
		return
	})
}

func (e *expr) factor() (value, error) {
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
		case TokenString:
			e.useToken(func() {
				ret = value{value: token.Raw[1 : len(token.Raw)-1], typ: valueString}
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
				sub := newExpr(e.scanner, e.variables)
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

func (e *expr) number(bs []byte) value {
	if bytes.Contains(bs, []byte{'.'}) {
		v, _ := strconv.ParseFloat(string(bs), 64)
		return value{typ: valueFloat, value: v}
	}
	v, _ := strconv.ParseInt(string(bs), 10, 64)
	return value{typ: valueInt, value: v}
}

func (e *expr) useToken(useToken func()) {
	e.scanner.forward()
	useToken()
}
