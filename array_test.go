package djson

import "testing"

func TestArray_set(t *testing.T) {
	// arr.set(k == 0 => 4)
	arr := newArray(
		Value{Type: ValueInt, Value: int64(1)},
		Value{Type: ValueInt, Value: int64(2)},
		Value{Type: ValueInt, Value: int64(3)},
	)
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'i'}},
		{Type: TokenEqual},
		{Type: TokenNumber, Raw: []byte{'0'}},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenParenthesesClose},
	}))
	val, err := setArray(Value{Type: ValueArray, Value: arr}, scanner, newVariables())
	if err != nil {
		t.Fatal(err)
	}
	if val.Type != ValueNull {
		t.Fatal("returned error")
	}
	val = arr.get(0)
	if !(val.Type == ValueInt && val.Value.(int64) == 4) {
		t.Fatal("array set error")
	}
}

func TestArray_del(t *testing.T) {
	// arr.del(i == 0)
	arr := newArray(
		Value{Type: ValueInt, Value: int64(1)},
		Value{Type: ValueInt, Value: int64(2)},
		Value{Type: ValueInt, Value: int64(3)},
	)
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'i'}},
		{Type: TokenEqual},
		{Type: TokenNumber, Raw: []byte{'0'}},
		{Type: TokenParenthesesClose},
	}))
	_, err := delArray(Value{Type: ValueArray, Value: arr}, scanner, newVariables())
	if err != nil {
		t.Fatal(err)
	}
	if len(arr.items) != 2 {
		t.Fatal("del array error")
	}
}
