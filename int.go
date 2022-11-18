package djson

import (
	"fmt"
	"strconv"
)

type Inter interface {
	Int() (int64, error)
}

type Int int64

func (i Int) Bool() bool {
	return i != 0
}

func (i Int) String() string {
	return strconv.Itoa(int(i))
}

func (i Int) Bytes() []byte {
	return []byte(i.String())
}

func (i Int) Int() (int64, error) {
	return int64(i), nil
}

func (i Int) Float() (float64, error) {
	return float64(i), nil
}

func (i Int) Compare(val Value) (int, error) {
	if val.Type != ValueInt {
		return 0, fmt.Errorf("can't compare int with [%s]", val.TypeName())
	}
	r, _ := val.Value.(Inter).Int()
	return int(int64(i) - r), nil
}

func (i Int) Add(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = FloatValue(ri + rr)
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
		ret = IntValue(int64(i) + rr)
	}
	return
}

func (i Int) Minus(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = FloatValue(ri - rr)
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
		ret = IntValue(int64(i) - rr)
	}
	return
}

func (i Int) Multiply(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = FloatValue(ri * rr)
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
		ret = IntValue(int64(i) * rr)
	}
	return
}

func (i Int) Devide(val Value) (ret Value, err error) {
	switch val.Type {
	case ValueFloat:
		ri, _ := i.Float()
		rr, _ := val.Value.(Floater).Float()
		ret = FloatValue(ri / rr)
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
		ret = IntValue(int64(i) / rr)
	}
	return
}
