package djson

import (
	"bytes"
	"testing"
)

func TestTranslator(t *testing.T) {
	translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
	input := `name = "test";
version = "v0.0.1";
{
  "name": name,
  "version": version
};
`
	ib := bytes.NewBuffer([]byte(input))
	ob := bytes.Buffer{}
	_, err := translator.Translate(ib, &ob)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(ob.String())
}
