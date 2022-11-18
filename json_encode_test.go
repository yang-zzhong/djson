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
	NewJsonEncoder("").Encode(StringValue([]byte("hello")...), &buf)
	if buf.String() != "\"hello\"" {
		t.Fatal("string error")
	}
}

func TestJsonEncoderBool(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder().Encode(BoolValue(true), &buf)
	if buf.String() != "true" {
		t.Fatal("bool error")
	}
}

func TestJsonEncoderInt(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder().Encode(IntValue(int64(123)), &buf)
	if buf.String() != "123" {
		t.Fatal("int error")
	}
}

func TestJsonEncoderFloat(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder().Encode(FloatValue(float64(1.23)), &buf)
	if buf.String() != "1.23" {
		t.Fatal("float error")
	}
}

func TestJsonEncoderArray(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder("  ").Encode(Value{Type: ValueArray, Value: NewArray(
		IntValue(int64(123)),
		FloatValue(float64(1.23)),
		StringValue([]byte("1.23")...),
		Value{Type: ValueNull},
	)}, &buf)
	t.Logf("%s\n", buf.String())
}

func TestJsonEncoderObject(t *testing.T) {
	var buf bytes.Buffer
	NewJsonEncoder("  ").Encode(Value{Type: ValueObject, Value: NewObject(
		&pair{key: []byte{'0'}, val: IntValue(int64(123))},
		&pair{key: []byte{'1'}, val: FloatValue(float64(1.23))},
		&pair{key: []byte{'2'}, val: StringValue([]byte("1.23")...)},
		&pair{key: []byte{'3'}, val: Value{Type: ValueNull}},
	)}, &buf)
	t.Logf("%s\n", buf.String())
}
