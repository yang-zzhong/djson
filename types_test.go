package djson

import (
	"testing"
)

func TestInt_arithmatic(t *testing.T) {
	// add
	a := Int(1)
	val, err := a.Add(IntValue(2))
	if err != nil {
		t.Fatal(err)
	}
	if val.MustInt() != 3 {
		t.Fatal("1 + 2 failed")
	}
	// minus
	a = Int(2)
	if val, err = a.Minus(IntValue(1)); err != nil {
		t.Fatal(err)
	}
	if val.MustInt() != 1 {
		t.Fatal("2 - 1 failed")
	}
	// multiply
	a = Int(2)
	if val, err = a.Multiply(IntValue(2)); err != nil {
		t.Fatal(err)
	}
	if val.MustInt() != 4 {
		t.Fatal("2 * 2 failed")
	}
	// devide
	a = Int(4)
	if val, err = a.Devide(IntValue(2)); err != nil {
		t.Fatal(err)
	}
	if val.MustInt() != 2 {
		t.Fatal("4 / 2 failed")
	}
}

func TestFloat_arithmatic(t *testing.T) {
	// add
	a := Float(1)
	val, err := a.Add(FloatValue(2))
	if err != nil {
		t.Fatal(err)
	}
	if val.MustFloat() != 3 {
		t.Fatal("1 + 2 failed")
	}
	// minus
	a = Float(2)
	if val, err = a.Minus(FloatValue(1)); err != nil {
		t.Fatal(err)
	}
	if val.MustFloat() != 1 {
		t.Fatal("2 - 1 failed")
	}
	// multiply
	a = Float(2)
	if val, err = a.Multiply(FloatValue(2)); err != nil {
		t.Fatal(err)
	}
	if val.MustFloat() != 4 {
		t.Fatal("2 * 2 failed")
	}
	// devide
	a = Float(4)
	if val, err = a.Devide(FloatValue(2)); err != nil {
		t.Fatal(err)
	}
	if val.MustFloat() != 2 {
		t.Fatal("4 / 2 failed")
	}
}

func TestString_arithmatic(t *testing.T) {
	// add
	a := NewString('h', 'e', 'l', 'l', 'o')
	val, err := a.Add(StringValue('w', 'o', 'r', 'l', 'd'))
	if err != nil {
		t.Fatal(err)
	}
	if val.String() != "helloworld" {
		t.Fatal("hello + world failed")
	}
	// minus
	a = NewString('h', 'e', 'l', 'l', 'o')
	if val, err = a.Minus(StringValue('e', 'l')); err != nil {
		t.Fatal(err)
	}
	if val.String() != "hlo" {
		t.Fatal("hello - el failed")
	}
	// multiply
	a = NewString('h')
	if val, err = a.Multiply(StringValue('e')); err == nil {
		t.Fatal("h * e should error")
	}
	// devide
	a = NewString('h')
	if val, err = a.Devide(StringValue('e')); err == nil {
		t.Fatal("h / e should error")
	}
}
