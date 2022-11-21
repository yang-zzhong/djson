package djson

import (
	"bytes"
	"io"
	"os"
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
	if ob.String() != "{\n  \"name\":\"test\",\n  \"version\":\"v0.0.1\"\n}" {
		t.Fatal("not equal")
	}
	t.Logf(ob.String())
}

func TestTranslator_full(t *testing.T) {
	f, err := os.Open("./testdata/full.djson")
	if err != nil {
		t.Fatal(err)
	}
	translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
	ob := bytes.Buffer{}
	_, err = translator.Translate(f, &ob)
	if err != nil {
		t.Fatal(err)
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

func TestTranslator_assign(t *testing.T) {
	data := `
header = {
  "type": "h3",
  "innerText": ""
};
header.map(k == "innerText" => "基本信息");
`
	ib := bytes.NewBuffer([]byte(data))
	ob := bytes.Buffer{}
	translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
	_, err := translator.Translate(ib, &ob)
	if err != nil {
		t.Fatal(err)
	}
	if ob.String() != "{\n  \"type\":\"h3\",\n  \"innerText\":\"基本信息\"\n}" {
		t.Fatal("not equal")
	}
	t.Logf(ob.String())
}

func BenchmarkTranslator_full(b *testing.B) {
	f, err := os.Open("./testdata/full.djson")
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		f.Seek(0, io.SeekStart)
		translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
		ob := bytes.Buffer{}
		_, err := translator.Translate(f, &ob)
		if err != nil {
			b.Fatal(err)
		}
	}
}
