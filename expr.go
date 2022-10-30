package djson

import (
	"bytes"
	"strconv"
)

// arithmatic logic
// logicor -> logicor || logicand | logicand
// logicand -> logicand && expr | expr
// expr -> expr + term | expr - term | term
// term -> term * factor | term / factor | factor
// factor -> digit | (logicor)

type expr struct {
	variables *variables
	value     value
	n         *tokenNexter
}

func newExpr(l lexer, ends [][]byte, vars *variables, ahead ...Token) *expr {
	return &expr{
		n:         newTokenNexter(l, ends, ahead...),
		variables: vars,
		value:     value{typ: valueNull},
	}
}

func (e *expr) endAt() []byte {
	return e.n.endAt()
}

func (e *expr) execute() (err error) {
	e.value, err = e.or()
	return
}

func (e *expr) or() (val value, err error) {
	var end bool
	for {
		if end, err = e.n.next(); err != nil || end {
			return
		}
		switch {
		case bytes.Equal([]byte{'|', '|'}, e.n.token.Raw):
			e.useToken(func() {
				var right value
				right, err = e.and()
				val, err = val.or(right)
			})
		default:
			if val.typ != valueNull {
				return
			}
			if val, err = e.and(); err != nil {
				return
			}
		}
	}
}

func (e *expr) and() (val value, err error) {
	var end bool
	for {
		if end, err = e.n.next(); err != nil || end {
			return
		}
		switch {
		case bytes.Equal([]byte{'&', '&'}, e.n.token.Raw):
			e.useToken(func() {
				var right value
				right, err = e.expr()
				val, err = val.and(right)
			})
		default:
			if val.typ != valueNull {
				return
			}
			if val, err = e.expr(); err != nil {
				return
			}
		}
	}
}

func (e *expr) expr() (val value, err error) {
	var end bool
	for {
		if end, err = e.n.next(); err != nil || end {
			return
		}
		switch e.n.token.Type {
		case TokenAddition:
			e.useToken(func() {
				var term value
				term, err = e.term()
				if err != nil {
					return
				}
				val, err = val.add(term)
			})
		case TokenMinus:
			e.useToken(func() {
				var term value
				term, err = e.term()
				if err != nil {
					return
				}
				val, err = val.minus(term)
			})
		default:
			if val.typ != valueNull {
				return
			}
			if val, err = e.term(); err != nil {
				return
			}
		}
	}
}

func (e *expr) term() (val value, err error) {
	var end bool
	for {
		if end, err = e.n.next(); err != nil || end {
			return
		}
		switch e.n.token.Type {
		case TokenMultiplication:
			e.useToken(func() {
				var factor value
				factor, err = e.factor()
				val, err = val.multiply(factor)
			})
		case TokenDevision:
			e.useToken(func() {
				var factor value
				factor, err = e.factor()
				val, err = val.devide(factor)
			})
		default:
			if val.typ != valueNull {
				return
			}
			if val, err = e.factor(); err != nil {
				return
			}
		}
	}
}

func (e *expr) factor() (val value, err error) {
	var end bool
	if end, err = e.n.next(); err != nil || end {
		return
	}
	token := &e.n.token
	switch token.Type {
	case TokenIdentifier:
		e.useToken(func() {
			val = e.variables.lookup(token.Raw)
			if val.typ == valueNull {
				err = ErrFromToken(UnexpectedToken, token)
			}
		})
		return
	case TokenString:
		e.useToken(func() {
			val = value{value: token.Raw[1 : len(token.Raw)-1], typ: valueString}
		})
		return
	case TokenNumber:
		e.useToken(func() {
			val = e.number(token.Raw)
		})
		return
	case TokenParenthesesOpen:
		e.useToken(func() {
			sub := newExpr(e.n.lexer, append(e.n.ends, []byte{')'}), e.variables)
			if err = sub.execute(); err == nil {
				val = sub.value
			}
		})
	case TokenBracketsOpen:
		e.useToken(func() {
			sub := newArrayExecutor(e.n.lexer, e.variables)
			if err = sub.execute(); err == nil {
				val = value{typ: valueArray, value: sub.value}
			}
		})
	case TokenBraceOpen:
		e.useToken(func() {
			sub := newObjectExecutor(e.n.lexer, e.variables)
			if err = sub.execute(); err == nil {
				val = value{typ: valueObject, value: sub.value}
			}
		})
	default:
		err = ErrFromToken(UnexpectedToken, token)
		return
	}
	err = ErrFromToken(UnexpectedToken, token)
	return
}

func (e *expr) number(bs []byte) value {
	isFloat := bytes.Index(bs, []byte{'.'}) > -1
	if isFloat {
		v, _ := strconv.ParseFloat(string(bs), 64)
		return value{typ: valueFloat, value: v}
	}
	v, _ := strconv.ParseInt(string(bs), 10, 64)
	return value{typ: valueInt, value: v}
}

func (e *expr) useToken(useToken func()) {
	e.n.useToken(func(_ Token) {
		useToken()
	})
}
