package djson

import "testing"

func TestLookup(t *testing.T) {
	v := Value{Type: ValueObject, Value: &Object{
		pairs: []*pair{
			{
				key: []byte("val1"),
				val: Value{
					Type: ValueArray,
					Value: &array{
						items: []Value{
							{Type: ValueInt, Value: int64(1)},
							{Type: ValueInt, Value: int64(2)},
							{Type: ValueInt, Value: int64(3)},
						},
					},
				},
			},
		},
	}}
	vs := variables{{name: []byte("var1"), value: v}}
	v = vs.lookup(path("var1.val1.0"))
	if v.Value.(int64) != 1 {
		t.Fatal("find failed")
	}
	vi := vs.lookup(path("var1.val1.*"))
	if !(vi.Type == ValueArray && len(vi.Value.(*array).items) == 3) {
		t.Fatal("* find failed")
	}
	arr := vi.Value.(*array)
	if arr.items[0].Value.(int64) != 1 || arr.items[1].Value.(int64) != 2 || arr.items[2].Value.(int64) != 3 {
		t.Fatal("* find value failed")
	}
}
