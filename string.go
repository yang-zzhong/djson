package djson

import (
	"bytes"
	"errors"
	"regexp"
)

type String interface {
	Literal() []byte
	Copy() String
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

func (s *str) Literal() []byte {
	return s.bytes
}

func (s *str) Copy() String {
	return NewString(s.bytes...)
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
	sub := stmt.value.Value.(String).Literal()
	s := val.Value.(String).Literal()
	ret = Value{Type: ValueInt, Value: int64(bytes.Index(s, sub))}
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
	s := val.Value.(String).Literal()
	arr := stmt.value.Value.(Array)
	a1 := arr.Get(0)
	var start int
	if a1.Type == ValueInt {
		start = int(a1.Value.(int64))
	} else if a1.Type == ValueNull {
		start = 0
	} else {
		err = errors.New("string sub only accept a [start, end] as the range")
		return
	}
	a2 := arr.Get(1)
	var end int
	if a2.Type == ValueInt {
		end = int(a2.Value.(int64)) + 1
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
	if reg, err = regexp.Compile(string(stmt.value.Value.(String).Literal())); err != nil {
		return
	}
	ret = Value{Type: ValueBool, Value: reg.Match(val.Value.(String).Literal())}
	return
}
