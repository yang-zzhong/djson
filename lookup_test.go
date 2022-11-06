package djson

import "testing"

func TestLookup(t *testing.T) {
	v := value{typ: valueObject, value: &object{
		pairs: []*pair{
			{
				key: []byte("val1"),
				val: value{
					typ: valueArray,
					value: &array{
						items: []value{
							{typ: valueInt, value: int64(1)},
							{typ: valueInt, value: int64(2)},
							{typ: valueInt, value: int64(3)},
						},
					},
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
	if !(vi.typ == valueArray && len(vi.value.(*array).items) == 3) {
		t.Fatal("* find failed")
	}
	arr := vi.value.(*array)
	if arr.items[0].value.(int64) != 1 || arr.items[1].value.(int64) != 2 || arr.items[2].value.(int64) != 3 {
		t.Fatal("* find value failed")
	}
}
