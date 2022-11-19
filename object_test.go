package djson

import "testing"

func TestObject_set(t *testing.T) {
	// obj.set(k == "0" => 4)
	obj := NewObject(
		&pair{key: []byte{'0'}, val: Value{Type: ValueInt, Value: Int(int64(1))}},
		&pair{key: []byte{'1'}, val: Value{Type: ValueInt, Value: Int(int64(2))}},
		&pair{key: []byte{'2'}, val: Value{Type: ValueInt, Value: Int(int64(3))}},
	)
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'k'}},
		{Type: TokenEqual},
		{Type: TokenString, Raw: []byte{'0'}},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenParenthesesClose},
	}))
	val, err := setObject(Value{Type: ValueArray, Value: obj}, scanner, NewContext())
	if err != nil {
		t.Fatal(err)
	}
	if val.Type != ValueObject {
		t.Fatal("returned error")
	}
	val = val.Value.(Object).Get([]byte("0"))
	if val.Type != ValueInt {
		t.Fatal("set error")
	}
	if v, _ := val.Value.(Inter).Int(); v != 4 {
		t.Fatal("set error")
	}
}

func TestObject_del(t *testing.T) {
	// obj.del(k == "0")
	obj := NewObject(
		&pair{key: []byte{'0'}, val: Value{Type: ValueInt, Value: Int(int64(1))}},
		&pair{key: []byte{'1'}, val: Value{Type: ValueInt, Value: Int(int64(2))}},
		&pair{key: []byte{'2'}, val: Value{Type: ValueInt, Value: Int(int64(3))}},
	)
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'k'}},
		{Type: TokenEqual},
		{Type: TokenString, Raw: []byte{'0'}},
		{Type: TokenParenthesesClose},
	}))
	_, err := delObject(Value{Type: ValueArray, Value: obj}, scanner, NewContext())
	if err != nil {
		t.Fatal(err)
	}
	if obj.Total() != 2 {
		t.Fatal("del object error")
	}
}
