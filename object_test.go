package djson

import "testing"

func TestObject_set(t *testing.T) {
	// obj.set(k == "0" => 4)
	obj := newObject(
		&pair{key: []byte{'0'}, val: Value{Type: ValueInt, Value: int64(1)}},
		&pair{key: []byte{'1'}, val: Value{Type: ValueInt, Value: int64(2)}},
		&pair{key: []byte{'2'}, val: Value{Type: ValueInt, Value: int64(3)}},
	)
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'k'}},
		{Type: TokenEqual},
		{Type: TokenString, Raw: []byte{'0'}},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenParenthesesClose},
	}))
	val, err := setObject(Value{Type: ValueArray, Value: obj}, scanner, newVariables())
	if err != nil {
		t.Fatal(err)
	}
	if val.Type != ValueNull {
		t.Fatal("returned error")
	}
	val = obj.get([]byte("0"))
	if !(val.Type == ValueInt && val.Value.(int64) == 4) {
		t.Fatal("set error")
	}
}

func TestObject_del(t *testing.T) {
	// obj.del(k == "0")
	obj := newObject(
		&pair{key: []byte{'0'}, val: Value{Type: ValueInt, Value: int64(1)}},
		&pair{key: []byte{'1'}, val: Value{Type: ValueInt, Value: int64(2)}},
		&pair{key: []byte{'2'}, val: Value{Type: ValueInt, Value: int64(3)}},
	)
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'k'}},
		{Type: TokenEqual},
		{Type: TokenString, Raw: []byte{'0'}},
		{Type: TokenParenthesesClose},
	}))
	_, err := delObject(Value{Type: ValueArray, Value: obj}, scanner, newVariables())
	if err != nil {
		t.Fatal(err)
	}
	if len(obj.pairs) != 2 {
		t.Fatal("del object error")
	}
}
