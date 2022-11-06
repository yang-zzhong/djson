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
	lookup := func() value {
		lookuper, ok := val.value.(lookuper)
		if ok {
			return lookuper.lookup(k)
		}
		return value{typ: valueNull}
	}
	i, r := splitKeyAndRest(k)
	if bytes.Equal(i, []byte{'_', 'p'}) {
		if val.typ == valueObject && !val.value.(*object).has(i) {
			return lookup()
		}
		if val.p == nil {
			return value{typ: valueNull}
		}
		return val.p.lookup(r)
	}
	return lookup()
}

func (obj *object) lookup(k []byte) value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		val := obj.get(i)
		if val.typ == valueNull || len(r) == 0 {
			return val
		}
		return val.lookup(r)
	}
	arr := newArray()
	for _, p := range obj.pairs {
		if len(r) == 0 {
			arr.items = append(arr.items, p.val)
			continue
		}
		item := p.val.lookup(r)
		if item.typ != valueNull {
			arr.items = append(arr.items, item)
		}
	}
	return value{typ: valueArray, value: arr}
}

func (arr *array) lookup(k []byte) value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		idx, err := strconv.Atoi(string(i))
		if err != nil {
			return value{typ: valueNull}
		}
		if idx > len(arr.items) {
			return value{typ: valueNull}
		}
		if len(r) == 0 {
			return arr.items[idx]
		}
		return arr.items[idx].lookup(r)
	}
	if len(r) == 0 {
		return value{typ: valueArray, value: arr}
	}
	ret := newArray()
	for _, item := range arr.items {
		v := item.lookup(r)
		if v.typ != valueNull {
			ret.items = append(ret.items, v)
		}
	}
	return value{typ: valueArray, value: ret}
}
