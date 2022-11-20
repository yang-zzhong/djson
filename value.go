package djson

import (
	"bytes"
	"errors"
	"fmt"
)

// TypeConverter a type converter which converts the type to string,[]byte,bool,int,flaot
type TypeConverter interface {
	Stringer
	Byter
	Booler
	Inter
	Floater
}

// Comparable compare with a Value
type Comparable interface {
	// Compare to a Value
	// if -1 returned, it indicates less than the Value
	// if 0 returned, it indicates equal to the Value
	// if 1 returned, it indicates great than the Value
	Compare(Value) (int, error)
}

// Arithmacable a simple arithmatic able interface
type Arithmacable interface {
	Add(Value) (Value, error)
	Minus(Value) (Value, error)
	Devide(Value) (Value, error)
	Multiply(Value) (Value, error)
}

type lookuper interface {
	lookup([]byte) Value
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
	ValueCallable
	ValueExit
)

type Value struct {
	Value interface{}
	Type  ValueType
	p     *Value
}

// TypeName get a type name for a value
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

// IntValue return a Int Value
func IntValue(v int64) Value {
	return Value{Type: ValueInt, Value: Int(v)}
}

// FloatValue return a Float Value
func FloatValue(v float64) Value {
	return Value{Type: ValueFloat, Value: Float(v)}
}

// StringValue return a String Value
func StringValue(v ...byte) Value {
	return Value{Type: ValueString, Value: NewString(v...)}
}

// ObjectValue return a Object Value
func ObjectValue(o Object) Value {
	return Value{Type: ValueObject, Value: o}
}

// ArrayValue return a Array Value
func ArrayValue(a Array) Value {
	return Value{Type: ValueArray, Value: a}
}

// BoolValue return a Bool Value
func BoolValue(b bool) Value {
	return Value{Type: ValueBool, Value: Bool(b)}
}

func CallableValue(c Callable) Value {
	return Value{Type: ValueCallable, Value: c}
}

// RangeValue return a Range Value
func RangeValue(begin, end int) Value {
	return Value{Type: ValueRange, Value: NewRange(begin, end)}
}

// NullValue return a Null Value
func NullValue() Value {
	return Value{Type: ValueNull}
}

func ExitValue() Value {
	return Value{Type: ValueExit}
}

// Int convert the value to int64
func (val Value) Int() (ret int64, err error) {
	val = val.RealValue()
	if inter, ok := val.Value.(Inter); ok {
		return inter.Int()
	}
	err = fmt.Errorf("value of type [%s] can't cast to int64", val.TypeName())
	return
}

// Float convert the value to float64
func (val Value) Float() (ret float64, err error) {
	val = val.RealValue()
	if floater, ok := val.Value.(Floater); ok {
		return floater.Float()
	}
	err = fmt.Errorf("value of type [%s] can't cast to float64", val.TypeName())
	return
}

// String convert the value to string
func (val Value) String() string {
	val = val.RealValue()
	if val.Value == nil {
		return "nil"
	}
	if stringer, ok := val.Value.(Stringer); ok {
		return stringer.String()
	}
	return val.TypeName()
}

// Bytes convert the value to []byte
func (val Value) Bytes() []byte {
	val = val.RealValue()
	if val.Value == nil {
		return []byte{'n', 'i', 'l'}
	}
	if stringer, ok := val.Value.(Byter); ok {
		return stringer.Bytes()
	}
	return []byte(val.TypeName())
}

// MustInt convert the value to int64, if can't, will panic
func (val Value) MustInt() int64 {
	v, err := val.Int()
	if err != nil {
		panic(err)
	}
	return v
}

// MustFloat convert the value to float64, if can't, will panic
func (val Value) MustFloat() float64 {
	v, err := val.Float()
	if err != nil {
		panic(err)
	}
	return v
}

// Copy a Value
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

func (left Value) RealValue() (val Value) {
	if left.Type == ValueIdentifier {
		val = left.Value.(Identifier).Value()
		return
	}
	val = left
	return
}

func (left Value) Add(right Value) (val Value, err error) {
	left = left.RealValue()
	right = right.RealValue()
	addable, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't + [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return addable.Add(right)
}

func (left Value) Minus(right Value) (val Value, err error) {
	left = left.RealValue()
	right = right.RealValue()
	minusable, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't - [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return minusable.Minus(right)
}

func (left Value) Multiply(right Value) (val Value, err error) {
	left = left.RealValue()
	right = right.RealValue()
	multi, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't * [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return multi.Multiply(right)
}

func (left Value) Devide(right Value) (val Value, err error) {
	left = left.RealValue()
	right = right.RealValue()
	devi, ok := left.Value.(Arithmacable)
	if !ok {
		err = fmt.Errorf("can't * [%s] with [%s]", left.TypeName(), right.TypeName())
		return
	}
	return devi.Devide(right)
}

func (left Value) Compare(right Value) (ret int, err error) {
	rlv := left.RealValue()
	rrv := right.RealValue()
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

func (left Value) Equal(right Value) bool {
	c, err := left.Compare(right)
	if err != nil {
		return false
	}
	return c == 0
}

func (left Value) Mod(right Value) (val Value, err error) {
	var lv, rv int64
	if lv, err = left.Int(); err != nil {
		return
	}
	if rv, err = right.Int(); err != nil {
		return
	}
	val = IntValue(lv % rv)
	return
}

func (left Value) And(right Value) Value {
	return BoolValue(left.Bool() && right.Bool())
}

func (left Value) Or(right Value) Value {
	return BoolValue(left.Bool() || right.Bool())
}

func (val Value) Bool() (ret bool) {
	val = val.RealValue()
	if b, ok := val.Value.(Booler); ok {
		return b.Bool()
	}
	return false
}

func (val Value) lookup(k []byte) Value {
	lookup := func() Value {
		lookuper, ok := val.Value.(lookuper)
		if ok {
			return lookuper.lookup(k)
		}
		return Value{Type: ValueNull}
	}
	i, r := splitKeyAndRest(k)
	if bytes.Equal(i, []byte{'_', 'p'}) {
		if val.Type == ValueObject && !val.Value.(Object).Has(i) {
			return lookup()
		}
		if val.p == nil {
			return Value{Type: ValueNull}
		}
		return val.p.lookup(r)
	}
	return lookup()
}
