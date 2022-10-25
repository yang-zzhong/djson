package djson

import "bytes"

type object []pair

type pair struct {
	key []byte
	val value
}

func (obj *object) set(key []byte, val value) {
	for i, p := range *obj {
		if !bytes.Equal(p.key, key) {
			continue
		}
		(*obj)[i] = pair{key: key, val: val}
		return
	}
	*obj = append(*obj, pair{key: key, val: val})
}

func (obj *object) get(key []byte) value {
	for _, p := range *obj {
		if !bytes.Equal(p.key, key) {
			continue
		}
		return p.val
	}
	return value{typ: valueNull}
}

func (obj *object) del(key []byte) {
	for i := range *obj {
		if !bytes.Equal((*obj)[i].key, key) {
			continue
		}
		*obj = append((*obj)[0:i], (*obj)[i+1:]...)
	}
}
