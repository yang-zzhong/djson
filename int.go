package djson

import (
	"fmt"
	"strconv"
)

type Inter interface {
	Int() (int64, error)
}

type Int interface {
	TypeConverter
	Comparable
	Arithmacable
}

type intl int64

func NewInt(i int64) Int {
	return intl(i)
}

func (i intl) Bool() bool {
	return i != 0
}

func (i intl) String() string {
	return strconv.Itoa(int(i))
}

func (i intl) Bytes() []byte {
	return []byte(i.String())
}

func (i intl) Int() (int64, error) {
	return int64(i), nil
}

func (i intl) Float() (float64, error) {
	return float64(i), nil
}

func (i intl) Compare(val Value) (int, error) {
	if val.Type != ValueInt {
		return 0, fmt.Errorf("can't compare int with [%s]", val.TypeName())
	}
	r, _ := val.Value.(Inter).Int()
	return int(int64(i) - r), nil
}

func (i intl) Add(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = Value{Type: ValueFloat, Value: NewFloat(ri + rr)}
	default:
		inter, ok := val.Value.(Inter)
		if !ok {
			err = fmt.Errorf("int can't + a [%s]", val.TypeName())
			return
		}
		var rr int64
		rr, err = inter.Int()
		if err != nil {
			strer, ok := val.Value.(Stringer)
			if !ok {
				err = fmt.Errorf("int can't + a [%s] with unvisible value", val.TypeName())
				return
			}
			err = fmt.Errorf("int can't + a [%s] with value %s", val.TypeName(), strer.String())
			return
		}
		ret = Value{Type: ValueInt, Value: NewInt(int64(i) + rr)}
	}
	return
}

func (i intl) Minus(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = Value{Type: ValueFloat, Value: NewFloat(ri - rr)}
	default:
		inter, ok := val.Value.(Inter)
		if !ok {
			err = fmt.Errorf("int can't - a [%s]", val.TypeName())
			return
		}
		var rr int64
		rr, err = inter.Int()
		if err != nil {
			strer, ok := val.Value.(Stringer)
			if !ok {
				err = fmt.Errorf("int can't - a [%s] with unvisible value", val.TypeName())
				return
			}
			err = fmt.Errorf("int can't - a [%s] with value %s", val.TypeName(), strer.String())
			return
		}
		ret = Value{Type: ValueInt, Value: NewInt(int64(i) - rr)}
	}
	return
}

func (i intl) Multiply(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = Value{Type: ValueFloat, Value: NewFloat(ri * rr)}
	default:
		inter, ok := val.Value.(Inter)
		if !ok {
			err = fmt.Errorf("int can't - a [%s]", val.TypeName())
			return
		}
		var rr int64
		rr, err = inter.Int()
		if err != nil {
			strer, ok := val.Value.(Stringer)
			if !ok {
				err = fmt.Errorf("int can't * a [%s] with unvisible value", val.TypeName())
				return
			}
			err = fmt.Errorf("int can't * a [%s] with value %s", val.TypeName(), strer.String())
			return
		}
		ret = Value{Type: ValueInt, Value: NewInt(int64(i) * rr)}
	}
	return
}

func (i intl) Devide(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = Value{Type: ValueFloat, Value: NewFloat(ri / rr)}
	default:
		inter, ok := val.Value.(Inter)
		if !ok {
			err = fmt.Errorf("int can't - a [%s]", val.TypeName())
			return
		}
		var rr int64
		rr, err = inter.Int()
		if err != nil {
			strer, ok := val.Value.(Stringer)
			if !ok {
				err = fmt.Errorf("int can't / a [%s] with unvisible value", val.TypeName())
				return
			}
			err = fmt.Errorf("int can't / a [%s] with value %s", val.TypeName(), strer.String())
			return
		}
		ret = Value{Type: ValueInt, Value: NewInt(int64(i) / rr)}
	}
	return
}
