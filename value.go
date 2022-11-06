package djson

import (
	"bytes"
	"errors"
	"fmt"
)

type valueType int

const (
	valueNull = valueType(iota)
	valueObject
	valueArray
	valueString
	valueFloat
	valueInt
	valueBool
	valueIdentifier
)

var (
	valueNames = map[valueType]string{
		valueNull:       "null",
		valueObject:     "object",
		valueArray:      "array",
		valueString:     "string",
		valueFloat:      "float",
		valueInt:        "int",
		valueBool:       "bool",
		valueIdentifier: "idenfitier",
	}
)

const (
	logicAnd = iota
	logicOr
)

type identifier struct {
	name      []byte
	variables *variables
}

type value struct {
	value   interface{}
	typ     valueType
	p       *value
	dotPath []byte
}

type p struct {
	key []byte
	idx int
}

func (id identifier) value() value {
	return id.variables.lookup(id.name)
}

func (left value) call(nexter *tokenScanner, vars *variables) (val value, err error) {
	call, ok := left.value.(callable)
	if !ok {
		err = fmt.Errorf("%s can't support call function", valueNames[left.typ])
		return
	}
	dotPath := left.dotPath
	idx := bytes.LastIndex(left.dotPath, []byte{'.'})
	if idx > -1 {
		left.dotPath = dotPath[:idx]
	}
	val, err = left.realValue()
	if err != nil {
		return
	}
	return call.call(string(dotPath[idx+1:]), val, nexter, vars)
}

func (left value) merge(right value) (val value, err error) {
	if right.typ != valueIdentifier {
		err = errors.New("only identifier as right value can merge")
		return
	}
	val = left
	if len(val.dotPath) > 0 {
		val.dotPath = append(val.dotPath, '.')
	}
	val.dotPath = append(val.dotPath, right.value.(*identifier).name...)
	return
}

func (left value) assign(right value) (val value, err error) {
	if left.typ != valueIdentifier {
		err = errors.New("only identifier can assign to")
	}
	return
}

func (left value) realValue() (val value, err error) {
	if left.typ == valueIdentifier {
		left = left.value.(*identifier).value()
	}
	if len(left.dotPath) == 0 {
		val = left
		return
	}
	lookuper, ok := left.value.(lookuper)
	if !ok {
		err = fmt.Errorf("%s can't support dot path search", valueNames[left.typ])
	}
	val = lookuper.lookup(left.dotPath)
	return
}

func (left value) add(right value) (value, error) {
	return left.arithmatic(right, '+')
}

func (left value) minus(right value) (value, error) {
	return left.arithmatic(right, '-')
}

func (left value) multiply(right value) (value, error) {
	return left.arithmatic(right, '*')
}

func (left value) devide(right value) (value, error) {
	return left.arithmatic(right, '/')
}

func (left value) compare(right value) (int, error) {
	var err error
	if left, err = left.realValue(); err != nil {
		return 0, err
	}
	if right, err = right.realValue(); err != nil {
		return 0, err
	}
	if left.typ != right.typ {
		return 0, errors.New("type not match")
	}
	switch left.typ {
	case valueNull:
		if right.typ == valueNull {
			return 0, nil
		} else if right.typ != valueNull {
			return -1, nil
		}
	case valueInt:
		lr := left.value.(int64)
		rr := right.value.(int64)
		if lr > rr {
			return 1, nil
		} else if lr == rr {
			return 0, nil
		} else {
			return -1, nil
		}
	case valueFloat:
		lr := left.value.(float64)
		rr := right.value.(float64)
		if lr > rr {
			return 1, nil
		} else if lr == rr {
			return 0, nil
		} else {
			return -1, nil
		}
	case valueString:
		return bytes.Compare(left.value.([]byte), right.value.([]byte)), nil
	case valueObject:
		lr := left.value.(*object)
		rr := right.value.(*object)
		if len(lr.pairs) > len(rr.pairs) {
			return 1, nil
		} else if len(lr.pairs) < len(rr.pairs) {
			return -1, nil
		}
		for _, p := range lr.pairs {
			c, err := p.val.compare(rr.get(p.key))
			if err != nil {
				return 0, err
			} else if c != 0 {
				return c, nil
			}
		}
		return 0, nil
	case valueArray:
		lr := left.value.(*array)
		rr := right.value.(*array)
		if len(lr.items) > len(rr.items) {
			return 1, nil
		} else if len(lr.items) < len(rr.items) {
			return -1, nil
		}
		for i, p := range lr.items {
			c, err := p.compare(rr.items[i])
			if err != nil {
				return 0, err
			}
			if c != 0 {
				return c, nil
			}
		}
		return 0, nil
	}
	return 0, errors.New("not supported type")
}

func (left value) equal(right value) bool {
	c, err := left.compare(right)
	if err != nil {
		return false
	}
	return c == 0
}

func (left value) arithmatic(right value, operator byte) (val value, err error) {
	if left, err = left.realValue(); err != nil {
		return
	}
	if right, err = right.realValue(); err != nil {
		return
	}
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
	case valueArray:
		switch operator {
		case '+':
			arr := left.value.(*array)
			if right.typ == valueArray {
				arr.append(right.value.(*array).items...)
			} else {
				arr.append(right)
			}
			val = value{typ: valueArray, value: arr}
		case '-':
			arr := left.value.(*array)
			if right.typ == valueArray {
				arr.del(right.value.(*array).items...)
			} else {
				arr.del(right)
			}
			val = value{typ: valueArray, value: arr}
		default:
			err = fmt.Errorf("unsupported arithmatic for array as left value: %s", []byte{operator})
		}
	case valueObject:
		switch operator {
		case '+':
			if right.typ != valueObject {
				err = fmt.Errorf("unsupported arithmatic for object as right value")
			}
			obj := left.value.(*object)
			for _, p := range right.value.(*object).pairs {
				obj.set(p.key, p.val)
			}
			val = value{typ: valueObject, value: obj}
		case '-':
			if right.typ != valueObject {
				err = fmt.Errorf("unsupported arithmatic for object as right value")
			}
			obj := left.value.(*object)
			for _, p := range right.value.(*object).pairs {
				obj.del(p.key)
			}
			val = value{typ: valueObject, value: obj}
		}
	default:
		err = errors.New("unsupported type to arithmatic")
	}
	return
}

func (left value) and(right value) (val value, err error) {
	return left.logic(right, logicAnd)
}

func (left value) or(right value) (val value, err error) {
	return left.logic(right, logicOr)
}

func (left value) logic(right value, operator int) (val value, err error) {
	var lv, rv bool
	if lv, err = left.toBool(); err != nil {
		return
	}
	if rv, err = right.toBool(); err != nil {
		return
	}
	switch operator {
	case logicOr:
		val = value{typ: valueBool, value: lv || rv}
	case logicAnd:
		val = value{typ: valueBool, value: lv && rv}
	}
	return
}

func (val value) toBool() (ret bool, err error) {
	if val, err = val.realValue(); err != nil {
		return
	}
	switch val.typ {
	case valueInt:
		ret = val.value.(int64) != 0
	case valueFloat:
		ret = int64(val.value.(float64)) != 0
	case valueString:
		ret = len(val.value.([]byte)) > 0
	case valueArray:
		ret = len(val.value.(*array).items) > 0
	case valueObject:
		ret = len(val.value.(*object).pairs) > 0
	case valueBool:
		ret = val.value.(bool)
	}
	return
}
