package djson

import "testing"

func TestArray_set(t *testing.T) {
	// arr.set(k == 0 => 4)
	arr := NewArray(
		IntValue(int64(1)),
		IntValue(int64(2)),
		IntValue(int64(3)),
	)
	scanner := NewTokenScanner(NewLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'i'}},
		{Type: TokenEqual},
		{Type: TokenNumber, Raw: []byte{'0'}},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenParenthesesClose},
	}))
	val, err := setArray(Value{Type: ValueArray, Value: arr}, scanner, NewContext())
	if err != nil {
		t.Fatal(err)
	}
	if val.Type != ValueArray {
		t.Fatal("array set returned error")
	}
	val = val.Value.(Array).Get(0)
	if val.Type != ValueInt {
		t.Fatal("array set error")
	}
	if v, _ := val.Value.(Inter).Int(); v != 4 {
		t.Fatal("array set error")
	}

}

func TestArray_del(t *testing.T) {
	// arr.del(i == 0)
	arr := NewArray(
		IntValue(int64(1)),
		IntValue(int64(2)),
		IntValue(int64(3)),
	)
	scanner := NewTokenScanner(NewLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'i'}},
		{Type: TokenEqual},
		{Type: TokenNumber, Raw: []byte{'0'}},
		{Type: TokenParenthesesClose},
	}))
	newArr, err := delArray(Value{Type: ValueArray, Value: arr}, scanner, NewContext())
	if err != nil {
		t.Fatal(err)
	}
	if !(newArr.Type == ValueArray && newArr.Value.(Array).Total() == 2) {
		t.Fatal("del array error")
	}
}

func TestArray_get(t *testing.T) {
	// arr.filter(i > 1)
	arr := NewArray(
		IntValue(int64(1)),
		IntValue(int64(2)),
		IntValue(int64(3)),
	)
	scanner := NewTokenScanner(NewLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'i'}},
		{Type: TokenGreateThan},
		{Type: TokenNumber, Raw: []byte{'1'}},
		{Type: TokenParenthesesClose},
	}))
	val, err := filterArray(Value{Type: ValueArray, Value: arr}, scanner, NewContext())
	if err != nil {
		t.Fatal(err)
	}
	if val.Type != ValueArray {
		t.Fatal("returned type error")
	}
	if val.Value.(Array).Total() != 1 {
		t.Fatal("get error")
	}
}
