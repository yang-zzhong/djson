package djson

import "fmt"

type Booler interface {
	Bool() bool
}

type Bool bool

func (b Bool) Bool() bool {
	return bool(b)
}

func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b Bool) Int() (int64, error) {
	if b {
		return 1, nil
	}
	return 0, nil
}

func (b Bool) Float() (float64, error) {
	if b {
		return 1, nil
	}
	return 0, nil
}

func (b Bool) Bytes() []byte {
	if b {
		return []byte{'t', 'r', 'u', 'e'}
	}
	return []byte{'f', 'a', 'l', 's', 'e'}
}

func (b Bool) Copy() Bool {
	return b
}

func (b Bool) Compare(val Value) (ret int, err error) {
	if val.Type != ValueBool {
		err = fmt.Errorf("can't compare bool with [%s]", val.TypeName())
		return
	}
	bi, _ := b.Int()
	vi, _ := val.Value.(Inter).Int()
	ret = int(bi - vi)
	return
}
