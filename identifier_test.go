package djson

import "testing"

func TestIdentifier_Value(t *testing.T) {
	vs := NewContext(
		Variable{
			Name: []byte("hello"),
			Value: Value{
				Type: ValueArray,
				Value: NewArray(
					Value{Type: ValueInt, Value: int64(1)},
				),
			},
		})
	// hello.0
	id := NewIdentifier([]byte{'0'}, vs)
	id.SetParent(Value{
		Type: ValueIdentifier, Value: NewIdentifier([]byte("hello"), vs),
	})
	val := id.Value()
	if !(val.Type == ValueInt && val.Value.(int64) == 1) {
		t.Fatal("value of array error")
	}
}

func TestIdentifier_Assign(t *testing.T) {
	vs := NewContext(
		Variable{
			Name: []byte("hello"),
			Value: Value{
				Type: ValueArray,
				Value: NewArray(
					Value{Type: ValueInt, Value: int64(1)},
				),
			},
		})
	// hello.0 = 2
	id := NewIdentifier([]byte{'0'}, vs)
	id.SetParent(Value{
		Type: ValueIdentifier, Value: NewIdentifier([]byte("hello"), vs),
	})
	err := id.Assign(Value{Type: ValueInt, Value: int64(2)})
	if err != nil {
		t.Fatal(err)
	}
	val := id.Value()
	if !(val.Type == ValueInt && val.Value.(int64) == 2) {
		t.Fatal("value of array error")
	}
}
