package djson

import (
	"fmt"
)

type Floater interface {
	Float() (float64, error)
}

type Float float64

func (i Float) Bool() bool {
	return i != 0
}

func (i Float) String() string {
	return fmt.Sprintf("%f", i)
}

func (i Float) Bytes() []byte {
	return []byte(i.String())
}

func (i Float) Int() (int64, error) {
	return int64(i), nil
}

func (i Float) Float() (float64, error) {
	return float64(i), nil
}

func (i Float) Compare(val Value) (int, error) {
	if val.Type != ValueFloat {
		return 0, fmt.Errorf("can't compare float with [%s]", val.TypeName())
	}
	r, _ := val.Value.(Floater).Float()
	return int(float64(i) - r), nil
}

func (i Float) Add(val Value) (ret Value, err error) {
	floater, ok := val.Value.(Floater)
	if !ok {
		err = fmt.Errorf("float can't + a [%s]", val.TypeName())
	}
	rr, err := floater.Float()
	if err != nil {
		strer, ok := val.Value.(Stringer)
		if !ok {
			err = fmt.Errorf("float can't + a [%s] with unvisible value", val.TypeName())
			return
		}
		err = fmt.Errorf("int can't + a [%s] with value %s", val.TypeName(), strer.String())
		return
	}
	ret = Value{Type: ValueFloat, Value: Float(float64(i) + rr)}
	return
}

func (i Float) Minus(val Value) (ret Value, err error) {
	floater, ok := val.Value.(Floater)
	if !ok {
		err = fmt.Errorf("float can't - a [%s]", val.TypeName())
		return
	}
	rr, err := floater.Float()
	if err != nil {
		strer, ok := val.Value.(Stringer)
		if !ok {
			err = fmt.Errorf("float can't - a [%s] with unvisible value", val.TypeName())
			return
		}
		err = fmt.Errorf("int can't - a [%s] with value %s", val.TypeName(), strer.String())
		return
	}
	ret = Value{Type: ValueFloat, Value: Float(float64(i) - rr)}
	return
}

func (i Float) Multiply(val Value) (ret Value, err error) {
	floater, ok := val.Value.(Floater)
	if !ok {
		err = fmt.Errorf("float can't * a [%s]", val.TypeName())
		return
	}
	rr, err := floater.Float()
	if err != nil {
		strer, ok := val.Value.(Stringer)
		if !ok {
			err = fmt.Errorf("float can't * a [%s] with unvisible value", val.TypeName())
			return
		}
		err = fmt.Errorf("int can't * a [%s] with value %s", val.TypeName(), strer.String())
	}
	ret = Value{Type: ValueFloat, Value: Float(float64(i) * rr)}
	return
}

func (i Float) Devide(val Value) (ret Value, err error) {
	floater, ok := val.Value.(Floater)
	if !ok {
		err = fmt.Errorf("float can't / a [%s]", val.TypeName())
		return
	}
	rr, err := floater.Float()
	if err != nil {
		strer, ok := val.Value.(Stringer)
		if !ok {
			err = fmt.Errorf("float can't / a [%s] with unvisible value", val.TypeName())
			return
		}
		err = fmt.Errorf("int can't / a [%s] with value %s", val.TypeName(), strer.String())
	}
	ret = Value{Type: ValueFloat, Value: Float(float64(i) / rr)}
	return
}
