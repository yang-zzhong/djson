package djson

import (
	"bytes"
	"errors"
	"fmt"
)

type ValueType int

const (
	ValueNull = ValueType(iota)
	ValueObject
	ValueArray
	ValueString
	ValueFloat
	ValueInt
	ValueBool
	ValueIdentifier
)

var (
	valueNames = map[ValueType]string{
		ValueNull:       "null",
		ValueObject:     "object",
		ValueArray:      "array",
		ValueString:     "string",
		ValueFloat:      "float",
		ValueInt:        "int",
		ValueBool:       "bool",
		ValueIdentifier: "idenfitier",
	}
)

const (
	logicAnd = iota
	logicOr
)

type identifier struct {
	name      []byte
	p         *Value
	variables *variables
}

type Value struct {
	Value interface{}
	Type  ValueType
	p     *Value
}

type p struct {
	key []byte
	idx int
}

func (id identifier) value() Value {
	dots := id.name
	tmp := &id
	for tmp.p != nil && tmp.p.Type == ValueIdentifier {
		tmp = tmp.p.Value.(*identifier)
		n := append(tmp.name, '.')
		dots = append(n, dots...)
	}
	val := id.variables.lookup(dots)
	if val.Type == ValueIdentifier {
		return val.Value.(*identifier).value()
	}
	return val
}

func (id identifier) call(scanner TokenScanner, vars *variables) (val Value, err error) {
	name := id.name
	val = id.p.realValue()
	call, ok := val.Value.(callable)
	if !ok {
		err = fmt.Errorf("%s can't support call function", valueNames[val.Type])
	}
	return call.call(string(name), val, scanner, vars)
}

func (left Value) assign(right Value) (val Value, err error) {
	if left.Type != ValueIdentifier {
		err = errors.New("only identifier can assign to")
	}
	id := left.Value.(*identifier)
	id.variables.set(id.name, right)
	return
}

func (left Value) realValue() (val Value) {
	if left.Type == ValueIdentifier {
		val = left.Value.(*identifier).value()
		return
	}
	val = left
	return
}

func (left Value) add(right Value) (Value, error) {
	return left.arithmatic(right, '+')
}

func (left Value) minus(right Value) (Value, error) {
	return left.arithmatic(right, '-')
}

func (left Value) multiply(right Value) (Value, error) {
	return left.arithmatic(right, '*')
}

func (left Value) devide(right Value) (Value, error) {
	return left.arithmatic(right, '/')
}

func (left Value) compare(right Value) (int, error) {
	rlv := left.realValue()
	rrv := right.realValue()
	if rlv.Type != rrv.Type {
		return 0, errors.New("type not match")
	}
	switch rlv.Type {
	case ValueNull:
		if rrv.Type == ValueNull {
			return 0, nil
		} else if rrv.Type != ValueNull {
			return -1, nil
		}
	case ValueInt:
		lr := rlv.Value.(int64)
		rr := rrv.Value.(int64)
		if lr > rr {
			return 1, nil
		} else if lr == rr {
			return 0, nil
		} else {
			return -1, nil
		}
	case ValueFloat:
		lr := rlv.Value.(float64)
		rr := rrv.Value.(float64)
		if lr > rr {
			return 1, nil
		} else if lr == rr {
			return 0, nil
		} else {
			return -1, nil
		}
	case ValueString:
		return bytes.Compare(rlv.Value.([]byte), rrv.Value.([]byte)), nil
	case ValueObject:
		lr := rlv.Value.(*Object)
		rr := rrv.Value.(*Object)
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
	case ValueArray:
		lr := rlv.Value.(*array)
		rr := rrv.Value.(*array)
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

func (left Value) equal(right Value) bool {
	c, err := left.compare(right)
	if err != nil {
		return false
	}
	return c == 0
}

func (left Value) arithmatic(right Value, operator byte) (val Value, err error) {
	left = left.realValue()
	right = right.realValue()
	switch left.Type {
	case ValueNull:
		return right, nil
	case ValueInt, ValueFloat:
		if right.Type != left.Type {
			err = errors.New("type not match")
			return
		}
		switch operator {
		case '+':
			val.Type = left.Type
			if left.Type == ValueInt {
				val.Value = left.Value.(int64) + right.Value.(int64)
			} else if left.Type == ValueFloat {
				val.Value = left.Value.(float64) + right.Value.(float64)
			}
		case '-':
			val.Type = left.Type
			if left.Type == ValueInt {
				val.Value = left.Value.(int64) - right.Value.(int64)
			} else if left.Type == ValueFloat {
				val.Value = left.Value.(float64) - right.Value.(float64)
			}
		case '*':
			val.Type = left.Type
			if left.Type == ValueInt {
				val.Value = left.Value.(int64) * right.Value.(int64)
			} else if left.Type == ValueFloat {
				val.Value = left.Value.(float64) * right.Value.(float64)
			}
		case '/':
			val.Type = left.Type
			if left.Type == ValueInt {
				val.Value = left.Value.(int64) / right.Value.(int64)
			} else if left.Type == ValueFloat {
				val.Value = left.Value.(float64) / right.Value.(float64)
			}
		}
	case ValueString:
		if operator != '+' {
			err = fmt.Errorf("unsupported string operator [%s]", []byte{operator})
			return
		}
		if right.Type != ValueString {
			err = errors.New("type not match")
			return
		}
		val.Type = ValueString
		val.Value = append(left.Value.([]byte), right.Value.([]byte)...)
	case ValueArray:
		switch operator {
		case '+':
			arr := left.Value.(*array)
			if right.Type == ValueArray {
				arr.append(right.Value.(*array).items...)
			} else {
				arr.append(right)
			}
			val = Value{Type: ValueArray, Value: arr}
		case '-':
			arr := left.Value.(*array)
			if right.Type == ValueArray {
				arr.del(right.Value.(*array).items...)
			} else {
				arr.del(right)
			}
			val = Value{Type: ValueArray, Value: arr}
		default:
			err = fmt.Errorf("unsupported arithmatic for array as left value: %s", []byte{operator})
		}
	case ValueObject:
		switch operator {
		case '+':
			if right.Type != ValueObject {
				err = fmt.Errorf("unsupported arithmatic for object as right value")
			}
			obj := left.Value.(*Object)
			for _, p := range right.Value.(*Object).pairs {
				obj.set(p.key, p.val)
			}
			val = Value{Type: ValueObject, Value: obj}
		case '-':
			if right.Type != ValueObject {
				err = fmt.Errorf("unsupported arithmatic for object as right value")
			}
			obj := left.Value.(*Object)
			for _, p := range right.Value.(*Object).pairs {
				obj.del(p.key)
			}
			val = Value{Type: ValueObject, Value: obj}
		}
	default:
		err = errors.New("unsupported type to arithmatic")
	}
	return
}

func (left Value) and(right Value) (val Value, err error) {
	return left.logic(right, logicAnd)
}

func (left Value) or(right Value) (val Value, err error) {
	return left.logic(right, logicOr)
}

func (left Value) logic(right Value, operator int) (val Value, err error) {
	var lv, rv bool
	if lv, err = left.toBool(); err != nil {
		return
	}
	if rv, err = right.toBool(); err != nil {
		return
	}
	switch operator {
	case logicOr:
		val = Value{Type: ValueBool, Value: lv || rv}
	case logicAnd:
		val = Value{Type: ValueBool, Value: lv && rv}
	}
	return
}

func (val Value) toBool() (ret bool, err error) {
	val = val.realValue()
	switch val.Type {
	case ValueInt:
		ret = val.Value.(int64) != 0
	case ValueFloat:
		ret = int64(val.Value.(float64)) != 0
	case ValueString:
		ret = len(val.Value.([]byte)) > 0
	case ValueArray:
		ret = len(val.Value.(*array).items) > 0
	case ValueObject:
		ret = len(val.Value.(*Object).pairs) > 0
	case ValueBool:
		ret = val.Value.(bool)
	}
	return
}
