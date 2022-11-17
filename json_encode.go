package djson

import (
	"fmt"
	"io"
	"regexp"
)

type jsonEncoder struct {
	indent string
}

func NewJsonEncoder(indent ...string) *jsonEncoder {
	id := ""
	if len(indent) > 0 {
		id = indent[0]
	}
	return &jsonEncoder{id}
}

func (jt jsonEncoder) Encode(val Value, w io.Writer) (int, error) {
	return jt.encodeJSONIndent(val, w, []byte(jt.indent), []byte{})
}

func (jt jsonEncoder) encodeJSONIndent(val Value, w io.Writer, tab []byte, priv []byte) (totalWrites int, err error) {
	var writes int
	write := func(b []byte) bool {
		writes, err = w.Write(b)
		if err != nil {
			return false
		}
		totalWrites += writes
		return true
	}
	val = val.realValue()
	switch val.Type {
	case ValueNull:
		write([]byte{'n', 'u', 'l', 'l'})
		return
	case ValueString:
		if write([]byte{'"'}) && write(val.Value.(String).Literal()) && write([]byte{'"'}) {
			return
		}
		return
	case ValueFloat:
		float := fmt.Sprintf("%f", val.Value)
		reg := regexp.MustCompile("0*$")
		n := reg.ReplaceAllString(float, "")
		write([]byte(n))
	case ValueInt:
		write([]byte(fmt.Sprintf("%d", val.Value)))
	case ValueBool:
		if val.Value.(bool) {
			write([]byte{'t', 'r', 'u', 'e'})
		} else {
			write([]byte{'f', 'a', 'l', 's', 'e'})
		}
	case ValueObject:
		writes, err = jt.encodeObjectJSON(val.Value.(*object), w, tab, priv)
		if err != nil {
			return
		}
		totalWrites += writes
	case ValueArray:
		writes, err = jt.encodeArrayJSON(val.Value.(*array), w, tab, priv)
		if err != nil {
			return
		}
		totalWrites += writes
	}
	return
}

func (jt jsonEncoder) encodeObjectJSON(obj *object, w io.Writer, tab []byte, priv []byte) (totalWrites int, err error) {
	var writes int
	write := func(b []byte) bool {
		writes, err = w.Write(b)
		if err != nil {
			return false
		}
		totalWrites += writes
		return true
	}
	if !write([]byte{'{'}) {
		return
	}
	indent := append(priv, tab...)
	for i, p := range obj.pairs {
		if len(indent) > 0 && !(write([]byte{'\n'}) && write(indent)) {
			return
		}
		if !(write([]byte{'"'}) && write(p.key) && write([]byte{'"', ':'})) {
			return
		}
		if writes, err = jt.encodeJSONIndent(p.val, w, tab, indent); err != nil {
			return
		}
		totalWrites += writes
		if i < len(obj.pairs)-1 && !write([]byte{','}) {
			return
		}
	}
	if len(obj.pairs) > 0 && write([]byte{'\n'}) && !write(priv) {
		return
	}
	write([]byte{'}'})
	return
}

func (jt jsonEncoder) encodeArrayJSON(arr *array, w io.Writer, tab []byte, priv []byte) (totalWrites int, err error) {
	var writes int
	write := func(b []byte) bool {
		writes, err = w.Write(b)
		if err != nil {
			return false
		}
		totalWrites += writes
		return true
	}
	if !write([]byte{'['}) {
		return
	}
	defer func() {
		if len(arr.items) > 0 && !write([]byte{'\n'}) {
			return
		}
		_ = write(priv) && write([]byte{']'})
	}()
	indent := append(priv, tab...)
	for i, item := range arr.items {
		if len(indent) > 0 && !(write([]byte{'\n'}) && write(indent)) {
			return
		}
		if writes, err = jt.encodeJSONIndent(item, w, tab, indent); err != nil {
			return
		}
		totalWrites += writes
		if i < len(arr.items)-1 && !write([]byte{','}) {
			return
		}
	}
	return
}
