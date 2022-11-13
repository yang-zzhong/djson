package djson

import (
	"bytes"
	"testing"
)

func TestJsonEncoderNull(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder("").Encode(Value{Type: ValueNull}, &buf)
	if buf.String() != "null" {
		t.Fatal("null error")
	}
}

func TestJsonEncoderString(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder("").Encode(Value{Type: ValueString, Value: []byte("hello")}, &buf)
	if buf.String() != "\"hello\"" {
		t.Fatal("string error")
	}
}

func TestJsonEncoderBool(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder().Encode(Value{Type: ValueBool, Value: []byte("true")}, &buf)
	if buf.String() != "true" {
		t.Fatal("bool error")
	}
}

func TestJsonEncoderInt(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder().Encode(Value{Type: ValueInt, Value: int64(123)}, &buf)
	if buf.String() != "123" {
		t.Fatal("int error")
	}
}

func TestJsonEncoderFloat(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder().Encode(Value{Type: ValueFloat, Value: float64(1.23)}, &buf)
	if buf.String() != "1.23" {
		t.Fatal("float error")
	}
}

func TestJsonEncoderArray(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder("  ").Encode(Value{Type: ValueArray, Value: NewArray(
		Value{Type: ValueInt, Value: int64(123)},
		Value{Type: ValueFloat, Value: float64(1.23)},
		Value{Type: ValueString, Value: []byte("1.23")},
		Value{Type: ValueNull},
	)}, &buf)
	t.Logf("%s\n", buf.String())
}

func TestJsonEncoderObject(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder("  ").Encode(Value{Type: ValueObject, Value: NewObject(
		&pair{key: []byte{'0'}, val: Value{Type: ValueInt, Value: int64(123)}},
		&pair{key: []byte{'1'}, val: Value{Type: ValueFloat, Value: float64(1.23)}},
		&pair{key: []byte{'2'}, val: Value{Type: ValueString, Value: []byte("1.23")}},
		&pair{key: []byte{'3'}, val: Value{Type: ValueNull}},
	)}, &buf)
	t.Logf("%s\n", buf.String())
}
