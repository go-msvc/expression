package expression_test

import (
	"testing"

	"github.com/go-msvc/expression"
	logger "github.com/go-msvc/log"
	"github.com/pkg/errors"
)

func TestCompareNumbers(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []struct {
		expr     string
		expValue interface{}
	}{
		{"  1 == 2 ", false},
		{"  2 == 2 ", true},
		{"5>6", false},
		{"5< 6", true},
		{"12.3 == 2", false},
		{"  34.0 == 34 ", true},
		{"-5<10", true},
		{"-5>10", false},
		{"10>=-5", true},
		{"10<=-5", false},
	}
	for _, l := range list {
		e, err := expression.New(l.expr)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to create expression"))
		}
		ctx := expression.NewContext()
		val, err := e.Eval(ctx)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to eval expr(%s)", l.expr))
		}
		if val != l.expValue {
			t.Fatalf("%s -> %T(%v) != %T(%v)", l.expr, val, val, l.expValue, l.expValue)
		}
		t.Logf("OK: %s -> %T(%v)", l.expr, val, val)
	}
}

func TestCompareStrings(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []struct {
		expr     string
		expValue interface{}
	}{
		{" 'jan' =='Jan'", false},
		{"  'jan'  =='jan'", true},
		{"  'jan'  <= 'jan'", true},
		{"  'jan'  >='jan'", true},
		{"  'jan'  <= 'Jan'", false},
		{"  'jan'  >='Jan'", true},
	}
	for _, l := range list {
		e, err := expression.New(l.expr)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to create expression"))
		}
		ctx := expression.NewContext()
		val, err := e.Eval(ctx)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to eval expr(%s)", l.expr))
		}
		if val != l.expValue {
			t.Fatalf("%s -> %T(%v) != %T(%v)", l.expr, val, val, l.expValue, l.expValue)
		}
		t.Logf("OK: %s -> %T(%v)", l.expr, val, val)
	}
}

func TestBool(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []struct {
		expr     string
		expValue interface{}
	}{
		{"false || false", false},
		{"true || false", true},
		{"false || true", true},
		{"true || true", true},
		{"false && false", false},
		{"true && false", false},
		{"false && true", false},
		{"true && true", true},
	}
	for _, l := range list {
		e, err := expression.New(l.expr)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to create expression"))
		}
		ctx := expression.NewContext()
		val, err := e.Eval(ctx)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to eval expr(%s)", l.expr))
		}
		if val != l.expValue {
			t.Fatalf("%s -> %T(%v) != %T(%v)", l.expr, val, val, l.expValue, l.expValue)
		}
		t.Logf("OK: %s -> %T(%v)", l.expr, val, val)
	}
}

func TestRegex(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []struct {
		expr     string
		expValue interface{}
	}{
		{"'abc' ~= '[a-z]'", true},
		{"'ABC' ~= '[a-z]'", false},
		{"'ABC' ~= '[A-Z]'", true},
		{"'ABC' ~= '^[A-Z]$'", false},
		{"'ABC' ~= '^[A-Z]'", true},
		{"'ABC' ~= '[A-Z]$'", true},
		{"'ABC1' ~= '[A-Z]'", true},
		{"'ABC1' ~= '^[A-Z]$'", false},
		{"'ABC1' ~= '^[A-Z]'", true},
		{"'ABC1' ~= '[A-Z]$'", false},
	}
	for _, l := range list {
		e, err := expression.New(l.expr)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to create expression"))
		}
		ctx := expression.NewContext()
		val, err := e.Eval(ctx)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to eval expr(%s)", l.expr))
		}
		if val != l.expValue {
			t.Fatalf("%s -> %T(%v) != %T(%v)", l.expr, val, val, l.expValue, l.expValue)
		}
		t.Logf("OK: %s -> %T(%v)", l.expr, val, val)
	}
}

func TestCompount(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []struct {
		expr     string
		expValue interface{}
	}{
		{"(1==2)", false},
		{"(2==2)", true},
		{"1+2==3", true},
		{"(1+2)==3", true},
		{"1+(2==3)", float64(1)}, //false=0
		{"1+(3==3)", float64(2)}, //true=1
		{"(1+5)*(4-7)", float64(-18)},
	}
	for _, l := range list {
		e, err := expression.NewCompound(l.expr)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to create expression"))
		}
		ctx := expression.NewContext()
		val, err := e.Eval(ctx)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to eval expr(%s)", l.expr))
		}
		if val != l.expValue {
			t.Fatalf("%s -> %T(%v) != %T(%v)", l.expr, val, val, l.expValue, l.expValue)
		}
		t.Logf("OK: %s -> %T(%v)", l.expr, val, val)
	}
}
