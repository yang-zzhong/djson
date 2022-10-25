package djson

import (
	"bytes"
	"strconv"
)

type lookuper interface {
	lookup([]byte) value
}

func path(p string) []byte {
	return []byte(p)
}

func splitKeyAndRest(ik []byte) (k []byte, rest []byte) {
	dot := bytes.Index(ik, []byte{'.'})
	if dot < 0 {
		k = ik
		return
	}
	k = ik[0:dot]
	rest = ik[dot+1:]
	return
}

func (vs variables) lookup(k []byte) value {
	i, r := splitKeyAndRest(k)
	for _, v := range vs {
		if !bytes.Equal(v.name, i) {
			continue
		}
		if len(r) == 0 {
			return v.value
		}
		return v.value.lookup(r)
	}
	return value{typ: valueNull}
}

func (val value) lookup(k []byte) value {
	lookuper, ok := val.value.(lookuper)
	if ok {
		return lookuper.lookup(k)
	}
	return value{typ: valueNull}
}

func (obj object) lookup(k []byte) value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		val := obj.get(i)
		if val.typ == valueNull || len(r) == 0 {
			return val
		}
		return val.lookup(r)
	}
	var arr array
	for _, p := range obj {
		if len(r) == 0 {
			arr = append(arr, p.val)
			continue
		}
		item := p.val.lookup(r)
		if item.typ != valueNull {
			arr = append(arr, item)
		}
	}
	return value{typ: valueArray, value: arr}
}

func (arr array) lookup(k []byte) value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		idx, err := strconv.Atoi(string(i))
		if err != nil {
			return value{typ: valueNull}
		}
		if idx > len(arr) {
			return value{typ: valueNull}
		}
		if len(r) == 0 {
			return arr[idx]
		}
		return arr[idx].lookup(r)
	}
	if len(r) == 0 {
		return value{typ: valueArray, value: arr}
	}
	var ret array
	for _, item := range arr {
		v := item.lookup(r)
		if v.typ != valueNull {
			ret = append(ret, v)
		}
	}
	return value{typ: valueArray, value: ret}
}
