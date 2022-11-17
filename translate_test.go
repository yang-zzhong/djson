package djson

import (
	"bytes"
	"testing"
)

func TestTranslator(t *testing.T) {
	translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
	input := `name = ((5 + 2) * 3 == 21) || false => "test";
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
	if ob.String() != "{\n    \"name\":\"test\",\n    \"version\":\"v0.0.1\"\n}" {
		t.Fatal("not equal")
	}
	t.Logf(ob.String())
}

func BenchmarkTranslator(b *testing.B) {
	input := `name = ((5 + 2) * 3 == 21) || false => "test";
    version = "v0.0.1";
    {
      "name": name,
      "version": version
    };
    `
	for i := 0; i < b.N; i++ {
		translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
		ib := bytes.NewBuffer([]byte(input))
		ob := bytes.Buffer{}
		_, err := translator.Translate(ib, &ob)
		if err != nil {
			b.Fatal(err)
		}
	}
}
