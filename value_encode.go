package djson

import (
	"fmt"
	"io"
	"regexp"
)

func encodeJSONIndent(val value, w io.Writer, tab []byte, privs ...[]byte) (totalWrites int, err error) {
	var writes int
	write := func(b []byte) bool {
		writes, err = w.Write(b)
		if err != nil {
			return false
		}
		totalWrites += writes
		return true
	}
	priv := []byte{}
	if len(privs) > 0 {
		priv = privs[0]
	}
	switch val.typ {
	case valueNull:
		write([]byte{'n', 'u', 'l', 'l'})
		return
	case valueString:
		if write([]byte{'"'}) && write(val.value.([]byte)) && write([]byte{'"'}) {
			return
		}
		return
	case valueFloat:
		float := fmt.Sprintf("%f", val.value)
		reg := regexp.MustCompile("0*$")
		n := reg.ReplaceAllString(float, "")
		write([]byte(n))
	case valueInt:
		write([]byte(fmt.Sprintf("%d", val.value)))
	case valueBool:
		write(val.value.([]byte))
	case valueObject:
		writes, err = encodeObjectJSON(val.value.(*object), w, tab, append(priv, tab...))
		if err != nil {
			return
		}
		totalWrites += writes
	case valueArray:
		writes, err = encodeArrayJSON(val.value.(*array), w, tab, append(priv, tab...))
		if err != nil {
			return
		}
		totalWrites += writes
	}
	return
}

func encodeJSON(val value, w io.Writer) (totalWrites int, err error) {
	return encodeJSONIndent(val, w, []byte{})
}

func encodeObjectJSON(obj *object, w io.Writer, tab []byte, priv []byte) (totalWrites int, err error) {
	var writes int
	write := func(b []byte) bool {
		writes, err = w.Write(b)
		if err != nil {
			return false
		}
		totalWrites += writes
		return true
	}
	if len(priv) > 0 && !write(priv) {
		return
	}
	if !write([]byte{'{'}) {
		return
	}
	defer func() {
		if len(priv) > 0 && !write(priv) {
			return
		}
		write([]byte{'}'})
	}()
	indent := append(priv, tab...)
	for i, p := range obj.pairs {
		if len(indent) > 0 && !(write([]byte{'\n'}) && write(indent)) {
			return
		}
		if !(write([]byte{'"'}) && write(p.key) && write([]byte{'"', ':'})) {
			return
		}
		if writes, err = encodeJSON(p.val, w); err != nil {
			return
		}
		totalWrites += writes
		if i < len(obj.pairs)-1 && !write([]byte{','}) {
			return
		}
	}
	return
}

func encodeArrayJSON(arr *array, w io.Writer, tab []byte, priv []byte) (totalWrites int, err error) {
	var writes int
	write := func(b []byte) bool {
		writes, err = w.Write(b)
		if err != nil {
			return false
		}
		totalWrites += writes
		return true
	}
	if len(priv) > 0 && !write(priv) {
		return
	}
	if !write([]byte{'['}) {
		return
	}
	defer func() {
		if len(priv) > 0 && !write(priv) {
			return
		}
		write([]byte{']'})
	}()
	indent := append(priv, tab...)
	for i, item := range arr.items {
		if len(indent) > 0 && !(write([]byte{'\n'}) && write(indent)) {
			return
		}
		if !write([]byte(fmt.Sprintf("%d", i))) {
			return
		}
		if writes, err = encodeJSON(item, w); err != nil {
			return
		}
		totalWrites += writes
		if i < len(arr.items)-1 && !write([]byte{','}) {
			return
		}
	}
	return
}
