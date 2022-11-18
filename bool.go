package djson

import "fmt"

type Booler interface {
	Bool() bool
}

type Bool interface {
	TypeConverter
	Comparable
}

type booln bool

func NewBool(b bool) Bool {
	return booln(b)
}

func (b booln) Bool() bool {
	return bool(b)
}

func (b booln) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b booln) Int() (int64, error) {
	if b {
		return 1, nil
	}
	return 0, nil
}

func (b booln) Float() (float64, error) {
	if b {
		return 1, nil
	}
	return 0, nil
}

func (b booln) Bytes() []byte {
	if b {
		return []byte{'t', 'r', 'u', 'e'}
	}
	return []byte{'f', 'a', 'l', 's', 'e'}
}

func (b booln) Copy() booln {
	return b
}

func (b booln) Compare(val Value) (ret int, err error) {
	if val.Type != ValueBool {
		err = fmt.Errorf("can't compare bool with [%s]", val.TypeName())
		return
	}
	bi, _ := b.Int()
	vi, _ := val.Value.(Inter).Int()
	ret = int(bi - vi)
	return
}
