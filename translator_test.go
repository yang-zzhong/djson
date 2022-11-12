package djson

import (
	"bytes"
	"testing"
)

func TestTranslator(t *testing.T) {
	translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
	input := []byte("\"hello\"")
	ib := bytes.NewBuffer(input)
	ob := bytes.Buffer{}
	_, err := translator.Translate(ib, &ob)
	if err != nil {
		t.Fatal(err)
	}
}
