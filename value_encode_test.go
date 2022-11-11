package djson

import (
	"bytes"
	"testing"
)

func TestEncodeJSONNull(t *testing.T) {
	var buf bytes.Buffer
	encodeJSON(value{typ: valueNull}, &buf)
	if buf.String() != "null" {
		t.Fatal("null error")
	}
}

func TestEncodeJSONString(t *testing.T) {
	var buf bytes.Buffer
	encodeJSON(value{typ: valueString, value: []byte("hello")}, &buf)
	if buf.String() != "\"hello\"" {
		t.Fatal("string error")
	}
}

func TestEncodeJSONBool(t *testing.T) {
	var buf bytes.Buffer
	encodeJSON(value{typ: valueBool, value: []byte("true")}, &buf)
	if buf.String() != "true" {
		t.Fatal("bool error")
	}
}

func TestEncodeJSONInt(t *testing.T) {
	var buf bytes.Buffer
	encodeJSON(value{typ: valueInt, value: int64(123)}, &buf)
	if buf.String() != "123" {
		t.Fatal("int error")
	}
}

func TestEncodeJSONFloat(t *testing.T) {
	var buf bytes.Buffer
	encodeJSON(value{typ: valueFloat, value: float64(1.23)}, &buf)
	if buf.String() != "1.23" {
		t.Fatal("float error")
	}
}
