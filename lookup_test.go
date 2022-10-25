package djson

import "testing"

func TestLookup(t *testing.T) {
	v := value{typ: valueObject, value: object{
		{
			key: []byte("val1"),
			val: value{
				typ: valueArray,
				value: array{
					value{typ: valueInt, value: int64(1)},
					value{typ: valueInt, value: int64(2)},
					value{typ: valueInt, value: int64(3)},
				},
			},
		},
	}}
	vs := variables{{name: []byte("var1"), value: v}}
	v = vs.lookup(path("var1.val1.0"))
	if v.value.(int64) != 1 {
		t.Fatal("find failed")
	}
	vi := vs.lookup(path("var1.val1.*"))
	if !(vi.typ == valueArray && len(vi.value.(array)) == 3) {
		t.Fatal("* find failed")
	}
	arr := vi.value.(array)
	if arr[0].value.(int64) != 1 || arr[1].value.(int64) != 2 || arr[2].value.(int64) != 3 {
		t.Fatal("* find value failed")
	}
}
