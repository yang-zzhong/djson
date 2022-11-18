package djson

import (
	"errors"
	"fmt"
)

type TypeConverter interface {
	Stringer
	Bytesable
	Booler
	Inter
	Floater
}

type Comparable interface {
	Compare(Value) (int, error)
}

type Arithmacable interface {
	Add(Value) (Value, error)
	Minus(Value) (Value, error)
	Devide(Value) (Value, error)
	Multiply(Value) (Value, error)
}

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
	ValueRange
)

type Value struct {
	Value interface{}
	Type  ValueType
	p     *Value
}

func (val Value) TypeName() string {
	return map[ValueType]string{
		ValueNull:       "null",
		ValueObject:     "object",
		ValueArray:      "array",
		ValueString:     "string",
		ValueFloat:      "float",
		ValueInt:        "int",
		ValueBool:       "bool",
		ValueIdentifier: "idenfitier",
	}[val.Type]
}

func IntValue(v int64) Value {
	return Value{Type: ValueInt, Value: Int(v)}
}

func FloatValue(v float64) Value {
	return Value{Type: ValueFloat, Value: Float(v)}
}

func StringValue(v ...byte) Value {
	return Value{Type: ValueString, Value: NewString(v...)}
}

func ObjectValue(o Object) Value {
	return Value{Type: ValueObject, Value: o}
}

func ArrayValue(a Array) Value {
	return Value{Type: ValueArray, Value: a}
}

func BoolValue(b bool) Value {
	return Value{Type: ValueBool, Value: Bool(b)}
}

func RangeValue(begin, end int) Value {
	return Value{Type: ValueRange, Value: NewRange(begin, end)}
}

func NullValue() Value {
	return Value{Type: ValueNull}
}

func (val Value) Copy() Value {
	switch val.Type {
	case ValueFloat, ValueInt, ValueBool, ValueNull:
		return Value{Type: val.Type, Value: val.Value}
	case ValueString:
		return Value{Type: ValueString, Value: val.Value.(String).Copy()}
	case ValueObject:
		return Value{Type: ValueObject, Value: val.Value.(Object).Copy()}
	case ValueArray:
		return Value{Type: ValueObject, Value: val.Value.(Array).Copy()}
	}
	return val
}

func (left Value) realValue() (val Value) {
	if left.Type == ValueIdentifier {
		val = left.Value.(Identifier).Value()
		return
	}
	val = left
	return
}

func (left Value) Add(right Value) (val Value, err error) {
	left = left.realValue()
	right = right.realValue()
	addable, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't + [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return addable.Add(right)
}

func (left Value) Minus(right Value) (val Value, err error) {
	left = left.realValue()
	right = right.realValue()
	minusable, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't - [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return minusable.Minus(right)
}

func (left Value) Multiply(right Value) (val Value, err error) {
	left = left.realValue()
	right = right.realValue()
	multi, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't * [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return multi.Multiply(right)
}

func (left Value) Devide(right Value) (val Value, err error) {
	left = left.realValue()
	right = right.realValue()
	devi, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't * [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return devi.Devide(right)
}

func (left Value) Compare(right Value) (ret int, err error) {
	rlv := left.realValue()
	rrv := right.realValue()
	if rlv.Type != rrv.Type {
		return 0, errors.New("type not match")
	}
	com, ok := rlv.Value.(Comparable)
	if !ok {
		err = fmt.Errorf("can't * [%s] with [%s]", rlv.TypeName(), rrv.TypeName())
		return
	}
	return com.Compare(right)
}

func (val Value) String() string {
	val = val.realValue()
	if val.Value == nil {
		return "nil"
	}
	if stringer, ok := val.Value.(Stringer); ok {
		return stringer.String()
	}
	return val.TypeName()
}

func (left Value) Equal(right Value) bool {
	c, err := left.Compare(right)
	if err != nil {
		return false
	}
	return c == 0
}

func (left Value) And(right Value) Value {
	return Value{Type: ValueBool, Value: left.Bool() && right.Bool()}
}

func (left Value) Or(right Value) Value {
	return Value{Type: ValueBool, Value: left.Bool() || right.Bool()}
}

func (val Value) Bool() (ret bool) {
	val = val.realValue()
	if b, ok := val.Value.(Booler); ok {
		return b.Bool()
	}
	return false
}
