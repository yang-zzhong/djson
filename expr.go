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

const (
	logicAnd = iota
	logicOr
)

type expr struct {
	getter      lexer
	ends        [][]byte
	variables   *variables
	ahead       Token
	value       value
	tokenUnused bool
	ended       bool
}

func newExpr(getter lexer, ends [][]byte, vars *variables, ahead ...*Token) *expr {
	e := &expr{
		getter:    getter,
		ends:      ends,
		variables: vars,
		value:     value{typ: valueNull},
	}
	if len(ahead) > 0 {
		e.ahead = *ahead[0]
		e.tokenUnused = true
	}
	return e
}

func (e *expr) endAt() []byte {
	return e.ahead.Raw
}

func (e *expr) execute() (err error) {
	e.value, err = e.logicor()
	return
}

func (e *expr) logicor() (val value, err error) {
	var end bool
	for {
		if end, err = e.next(); err != nil || end {
			return
		}
		switch {
		case bytes.Equal([]byte{'|', '|'}, e.ahead.Raw):
			e.useToken(func() {
				var right value
				right, err = e.logicand()
				val, err = e.lor(val, right)
			})
		default:
			if val.typ != valueNull {
				return
			}
			if val, err = e.logicand(); err != nil {
				return
			}
		}
	}
}

func (e *expr) logicand() (val value, err error) {
	var end bool
	for {
		if end, err = e.next(); err != nil || end {
			return
		}
		switch {
		case bytes.Equal([]byte{'&', '&'}, e.ahead.Raw):
			e.useToken(func() {
				var right value
				right, err = e.expr()
				val, err = e.land(val, right)
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
		if end, err = e.next(); err != nil || end {
			return
		}
		switch {
		case bytes.Equal([]byte{'+'}, e.ahead.Raw):
			e.useToken(func() {
				var term value
				term, err = e.term()
				if err != nil {
					return
				}
				val, err = e.add(val, term)
			})
		case bytes.Equal([]byte{'-'}, e.ahead.Raw):
			e.useToken(func() {
				var term value
				term, err = e.term()
				if err != nil {
					return
				}
				val, err = e.minus(val, term)
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
		if end, err = e.next(); err != nil || end {
			return
		}
		switch {
		case bytes.Equal([]byte{'*'}, e.ahead.Raw):
			e.useToken(func() {
				var factor value
				factor, err = e.factor()
				val, err = e.multiply(val, factor)
			})
		case bytes.Equal([]byte{'/'}, e.ahead.Raw):
			e.useToken(func() {
				var factor value
				factor, err = e.factor()
				val, err = e.devide(val, factor)
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
	if end, err = e.next(); err != nil || end {
		return
	}
	token := &e.ahead
	switch token.Type {
	case TokenVariable:
		e.useToken(func() {
			val = e.variables.lookup(e.ahead.Raw)
			if val.typ == valueNull {
				err = ErrFromToken(UnexpectedToken, token)
			}
		})
		return
	case TokenString:
		e.useToken(func() {
			val = value{value: e.ahead.Raw[1 : len(e.ahead.Raw)-1], typ: valueString}
		})
		return
	case TokenNumber:
		e.useToken(func() {
			val = e.number(e.ahead.Raw)
		})
		return
	case TokenBlockStart:
		if bytes.Equal([]byte{'('}, e.ahead.Raw) {
			e.useToken(func() {
				sub := newExpr(e.getter, append(e.ends, []byte{')'}), e.variables)
				if err = sub.execute(); err == nil {
					val = sub.value
				}
			})
		} else if bytes.Equal([]byte{'['}, e.ahead.Raw) {
			e.useToken(func() {
				sub := newArrayExecutor(e.getter, e.variables)
				if err = sub.execute(); err == nil {
					val = value{typ: valueArray, value: sub.value}
				}
			})
		} else if bytes.Equal([]byte{'{'}, e.ahead.Raw) {
			e.useToken(func() {
				sub := newObjectExecutor(e.getter, e.variables)
				if err = sub.execute(); err == nil {
					val = value{typ: valueObject, value: sub.value}
				}
			})
		} else {
			err = ErrFromToken(UnexpectedToken, token)
			return
		}
		return
	}
	err = ErrFromToken(UnexpectedToken, token)
	return
}

func (e *expr) add(left, right value) (value, error) {
	return left.add(right)
}

func (e *expr) minus(left, right value) (value, error) {
	return left.minus(right)
}

func (e *expr) multiply(left, right value) (value, error) {
	return left.multiply(right)
}

func (e *expr) devide(left, right value) (value, error) {
	return left.devide(right)
}

func (e *expr) next() (end bool, err error) {
	if e.tokenUnused || e.ended {
		end = e.ended
		return
	}
	if err = e.getter.NextToken(&e.ahead); err != nil {
		return
	}
	if e.ahead.Type == TokenEOF {
		e.useToken(func() {
			end = true
			e.ended = end
		})
		return
	}
	e.tokenUnused = true
	for _, ed := range e.ends {
		if bytes.Equal(ed, e.ahead.Raw) {
			e.useToken(func() {
				end = true
				e.ended = end
			})
			return
		}
	}
	return
}

func (e *expr) land(left, right value) (val value, err error) {
	return e.logic(left, right, logicAnd)
}

func (e *expr) lor(left, right value) (val value, err error) {
	return e.logic(left, right, logicOr)
}

func (e *expr) logic(left, right value, operator int) (val value, err error) {
	leftVal := e.toBool(left)
	rightVal := e.toBool(right)
	switch operator {
	case logicOr:
		val = value{typ: valueBool, value: leftVal || rightVal}
	case logicAnd:
		val = value{typ: valueBool, value: leftVal && rightVal}
	}
	return
}

func (e *expr) toBool(val value) bool {
	switch val.typ {
	case valueInt:
		return val.value.(int64) != 0
	case valueFloat:
		return val.value.(float64) != 0
	case valueString:
		return len(val.value.([]byte)) > 0
	case valueArray:
		return len(val.value.(array)) > 0
	case valueObject:
		return len(val.value.(object)) > 0
	case valueBool:
		return val.value.(bool)
	default:
		return false
	}
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
	e.tokenUnused = false
	useToken()
}
