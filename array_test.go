package djson

import "testing"

func TestArray_set(t *testing.T) {
	// arr.set(k == 0 => 4)
	arr := newArray(
		value{typ: valueInt, value: int64(1)},
		value{typ: valueInt, value: int64(2)},
		value{typ: valueInt, value: int64(3)},
	)
	scanner := newTokenScanner(newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'i'}},
		{Type: TokenEqual},
		{Type: TokenNumber, Raw: []byte{'0'}},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenParenthesesClose},
	}))
	val, err := setArray(value{typ: valueArray, value: arr}, scanner, newVariables())
	if err != nil {
		t.Fatal(err)
	}
	if val.typ != valueNull {
		t.Fatal("returned error")
	}
}
