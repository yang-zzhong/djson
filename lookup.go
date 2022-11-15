package djson

import (
	"bytes"
	"strconv"
)

type lookuper interface {
	lookup([]byte) Value
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

func (vs ctx) lookup(k []byte) Value {
	i, r := splitKeyAndRest(k)
	for _, v := range vs {
		if !bytes.Equal(v.Name, i) {
			continue
		}
		if len(r) == 0 {
			return v.Value
		}
		return v.Value.lookup(r)
	}
	return Value{Type: ValueNull}
}

func (val Value) lookup(k []byte) Value {
	lookup := func() Value {
		lookuper, ok := val.Value.(lookuper)
		if ok {
			return lookuper.lookup(k)
		}
		return Value{Type: ValueNull}
	}
	i, r := splitKeyAndRest(k)
	if bytes.Equal(i, []byte{'_', 'p'}) {
		if val.Type == ValueObject && !val.Value.(Object).Has(i) {
			return lookup()
		}
		if val.p == nil {
			return Value{Type: ValueNull}
		}
		return val.p.lookup(r)
	}
	return lookup()
}

func (obj *object) lookup(k []byte) Value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		val := obj.Get(i)
		if val.Type == ValueNull || len(r) == 0 {
			return val
		}
		return val.lookup(r)
	}
	arr := NewArray()
	for _, p := range obj.pairs {
		if len(r) == 0 {
			arr.items = append(arr.items, p.val)
			continue
		}
		item := p.val.lookup(r)
		if item.Type != ValueNull {
			arr.items = append(arr.items, item)
		}
	}
	return Value{Type: ValueArray, Value: arr}
}

func (arr *array) lookup(k []byte) Value {
	i, r := splitKeyAndRest(k)
	if !bytes.Equal(i, []byte{'*'}) {
		idx, err := strconv.Atoi(string(i))
		if err != nil {
			return Value{Type: ValueNull}
		}
		if idx > len(arr.items) {
			return Value{Type: ValueNull}
		}
		if len(r) == 0 {
			return arr.items[idx]
		}
		return arr.items[idx].lookup(r)
	}
	if len(r) == 0 {
		return Value{Type: ValueArray, Value: arr}
	}
	ret := NewArray()
	for _, item := range arr.items {
		v := item.lookup(r)
		if v.Type != ValueNull {
			ret.items = append(ret.items, v)
		}
	}
	return Value{Type: ValueArray, Value: ret}
}
