package djson

import (
	"bytes"
	"errors"
	"fmt"
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

func newExpr(getter lexer, ends [][]byte, vars *variables) *expr {
	return &expr{
		getter:    getter,
		ends:      ends,
		variables: vars,
		value:     value{typ: valueNull},
	}
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
		e.useToken(func() {
			if !bytes.Equal(e.ahead.Raw, []byte{'('}) {
				err = ErrFromToken(UnexpectedToken, token)
				return
			}
			sub := newExpr(e.getter, [][]byte{{')'}}, e.variables)
			if err = sub.execute(); err == nil {
				val = sub.value
			}
		})
		return
	}
	err = ErrFromToken(UnexpectedToken, token)
	return
}

func (e *expr) add(left, right value) (value, error) {
	return e.arithmatic(left, right, '+')
}

func (e *expr) minus(left, right value) (value, error) {
	return e.arithmatic(left, right, '-')
}

func (e *expr) multiply(left, right value) (value, error) {
	return e.arithmatic(left, right, '*')
}

func (e *expr) devide(left, right value) (value, error) {
	return e.arithmatic(left, right, '/')
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

func (e *expr) arithmatic(left, right value, operator byte) (val value, err error) {
	switch left.typ {
	case valueNull:
		return right, nil
	case valueInt, valueFloat:
		if right.typ != left.typ {
			err = errors.New("type not match")
			return
		}
		switch operator {
		case '+':
			val.typ = left.typ
			if left.typ == valueInt {
				val.value = left.value.(int64) + right.value.(int64)
			} else if left.typ == valueFloat {
				val.value = left.value.(float64) + right.value.(float64)
			}
		case '-':
			val.typ = left.typ
			if left.typ == valueInt {
				val.value = left.value.(int64) - right.value.(int64)
			} else if left.typ == valueFloat {
				val.value = left.value.(float64) - right.value.(float64)
			}
		case '*':
			val.typ = left.typ
			if left.typ == valueInt {
				val.value = left.value.(int64) * right.value.(int64)
			} else if left.typ == valueFloat {
				val.value = left.value.(float64) * right.value.(float64)
			}
		case '/':
			val.typ = left.typ
			if left.typ == valueInt {
				val.value = left.value.(int64) / right.value.(int64)
			} else if left.typ == valueFloat {
				val.value = left.value.(float64) / right.value.(float64)
			}
		}
	case valueString:
		if operator != '+' {
			err = fmt.Errorf("unsupported string operator [%s]", []byte{operator})
			return
		}
		if right.typ != valueString {
			err = errors.New("type not match")
			return
		}
		val.typ = valueString
		val.value = append(left.value.([]byte), right.value.([]byte)...)
	default:
		err = errors.New("unsupported type to arithmatic")
	}
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
	e.tokenUnused = false
	useToken()
}
