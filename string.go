package djson

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Stringer interface {
	String() string
}

type Bytesable interface {
	Bytes() []byte
}

type String interface {
	TypeConverter
	Comparable
	Arithmacable
	Copy() String
	Concat([]byte)
	Replace([]byte, []byte)
}

type str struct {
	bytes []byte
	*callableImp
}

var _ String = &str{}

func NewString(bs ...byte) *str {
	s := &str{bytes: bs, callableImp: newCallable("string")}
	s.register("index", indexString)
	s.register("match", matchString)
	s.register("sub", subString)
	return s
}

func (s *str) Bytes() []byte {
	return s.bytes
}

func (s *str) Copy() String {
	return NewString(s.bytes...)
}

func (s *str) String() string {
	return string(s.bytes)
}

func (s *str) Int() (int64, error) {
	return strconv.ParseInt(string(s.bytes), 10, 64)
}

func (s *str) Float() (float64, error) {
	return strconv.ParseFloat(string(s.bytes), 64)
}

func (s *str) Bool() bool {
	return bytes.EqualFold(s.bytes, []byte{'t', 'r', 'u', 'e'})
}

func (s *str) Add(val Value) (ret Value, err error) {
	if l, ok := val.Value.(Bytesable); ok {
		r := s.Copy()
		r.Concat(l.Bytes())
		ret = Value{Type: ValueString, Value: r}
		return
	} else if l, ok := val.Value.(Stringer); ok {
		r := s.Copy()
		r.Concat([]byte(l.String()))
		ret = Value{Type: ValueString, Value: r}
		return
	}
	err = fmt.Errorf("string can't + a [%s]", val.TypeName())
	return
}

func (s *str) Minus(val Value) (ret Value, err error) {
	if l, ok := val.Value.(Bytesable); ok {
		r := s.Copy()
		r.Replace(l.Bytes(), []byte{})
		ret = Value{Type: ValueString, Value: r}
		return
	} else if l, ok := val.Value.(Stringer); ok {
		r := s.Copy()
		r.Replace([]byte(l.String()), []byte{})
		ret = Value{Type: ValueString, Value: r}
		return
	}
	err = fmt.Errorf("string can't - a [%s]", val.TypeName())
	return
}

func (s *str) Multiply(val Value) (ret Value, err error) {
	err = fmt.Errorf("string can't * a [%s]", val.TypeName())
	return
}

func (s *str) Devide(val Value) (ret Value, err error) {
	err = fmt.Errorf("string can't / a [%s]", val.TypeName())
	return
}

func (s *str) And(val Value) (ret Value, err error) {
	ret = Value{Type: ValueBool, Value: s.Bool() && val.Bool()}
	return
}

func (s *str) Or(val Value) (ret Value, err error) {
	ret = Value{Type: ValueBool, Value: s.Bool() || val.Bool()}
	return
}

func (s *str) Compare(val Value) (ret int, err error) {
	if val.Type != ValueString {
		err = fmt.Errorf("can't compare string with [%s]", val.TypeName())
		return
	}
	ret = bytes.Compare(s.Bytes(), val.Value.(String).Bytes())
	return
}

func (s *str) Replace(search []byte, r []byte) {
	s.bytes = bytes.ReplaceAll(s.bytes, search, r)
}

func (s *str) Concat(ss []byte) {
	s.bytes = append(s.bytes, ss...)
}

func indexString(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(1)
	stmt := NewStmt(scanner, vars)
	if err = stmt.Execute(); err != nil {
		return
	}
	if stmt.value.Type != ValueString {
		err = errors.New("string match only accept a string as the regexp")
		return
	}
	sub := stmt.value.Value.(String).Bytes()
	s := val.Value.(String).Bytes()
	ret = Value{Type: ValueInt, Value: NewInt(int64(bytes.Index(s, sub)))}
	return
}

func subString(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(1)
	stmt := NewStmt(scanner, vars)
	if err = stmt.Execute(); err != nil {
		return
	}
	if !(stmt.value.Type == ValueArray && stmt.value.Value.(Array).Total() == 2) {
		err = errors.New("string sub only accept a [start, end] as the range")
		return
	}
	s := val.Value.(String).Bytes()
	arr := stmt.value.Value.(Array)
	a1 := arr.Get(0)
	var start int
	if a1.Type == ValueInt {
		v, _ := a1.Value.(Inter).Int()
		start = int(v)
	} else if a1.Type == ValueNull {
		start = 0
	} else {
		err = errors.New("string sub only accept a [start, end] as the range")
		return
	}
	a2 := arr.Get(1)
	var end int
	if a2.Type == ValueInt {
		v, _ := a2.Value.(Inter).Int()
		end = int(v) + 1
	} else if a2.Type == ValueNull {
		end = len(s)
	} else {
		err = errors.New("string sub only accept a [start, end] as the range")
		return
	}
	if end > len(s) {
		end = len(s)
	}
	ret = Value{Type: ValueString, Value: NewString(s[start:end]...)}
	return
}

func matchString(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	scanner.PushEnds(TokenParenthesesClose)
	defer scanner.PopEnds(1)
	stmt := NewStmt(scanner, vars)
	if err = stmt.Execute(); err != nil {
		return
	}
	if stmt.value.Type != ValueString {
		err = errors.New("string match only accept a string as the regexp")
		return
	}
	var reg *regexp.Regexp
	if reg, err = regexp.Compile(string(stmt.value.Value.(String).Bytes())); err != nil {
		return
	}
	ret = Value{Type: ValueBool, Value: reg.Match(val.Value.(String).Bytes())}
	return
}
