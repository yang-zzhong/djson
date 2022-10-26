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
)

type value struct {
	value interface{}
	typ   valueType
}

type variable struct {
	name  []byte
	value value
}

type variables []variable

type p struct {
	key []byte
	idx int
}

func (val value) each(handle func(p p, val value)) {
	switch val.typ {
	case valueObject:
		for _, v := range val.value.(object) {
			handle(p{key: v.key}, v.val)
		}
	case valueInt:
		for i, v := range val.value.(array) {
			handle(p{idx: i}, v)
		}
	default:
		handle(p{}, val)
	}
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
		lr := left.value.(object)
		rr := right.value.(object)
		if len(lr) > len(rr) {
			return 1, nil
		} else if len(lr) < len(rr) {
			return -1, nil
		}
		for _, p := range lr {
			c, err := p.val.compare(rr.get(p.key))
			if err != nil {
				return 0, err
			} else if c != 0 {
				return c, nil
			}
		}
		return 0, nil
	case valueArray:
		lr := left.value.(array)
		rr := right.value.(array)
		if len(lr) > len(rr) {
			return 1, nil
		} else if len(lr) < len(rr) {
			return -1, nil
		}
		for i, p := range lr {
			c, err := p.compare(rr[i])
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
			arr := left.value.(array)
			if right.typ == valueArray {
				(&arr).append(right.value.(array)...)
			} else {
				(&arr).append(right)
			}
			val = value{typ: valueArray, value: arr}
		case '-':
			arr := left.value.(array)
			if right.typ == valueArray {
				(&arr).del(right.value.(array)...)
			} else {
				(&arr).del(right)
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
			obj := left.value.(object)
			for _, p := range right.value.(object) {
				(&obj).set(p.key, p.val)
			}
			val = value{typ: valueObject, value: obj}
		case '-':
			if right.typ != valueObject {
				err = fmt.Errorf("unsupported arithmatic for object as right value")
			}
			obj := left.value.(object)
			for _, p := range right.value.(object) {
				(&obj).del(p.key)
			}
			val = value{typ: valueObject, value: obj}
		}
	default:
		err = errors.New("unsupported type to arithmatic")
	}
	return
}
