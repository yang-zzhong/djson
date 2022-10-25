package djson

type valueType int

const (
	valueNull = valueType(iota)
	valueObject
	valueArray
	valueString
	valueFloat
	valueInt
	valueBool
)

type value struct {
	value interface{}
	typ   valueType
}

type variable struct {
	name  []byte
	value value
}

type variables []variable

type p struct {
	key []byte
	idx int
}

func (val value) each(handle func(p p, val value)) {
	switch val.typ {
	case valueObject:
		for _, v := range val.value.(object) {
			handle(p{key: v.key}, v.val)
		}
	case valueInt:
		for i, v := range val.value.(array) {
			handle(p{idx: i}, v)
		}
	default:
		handle(p{}, val)
	}
}
