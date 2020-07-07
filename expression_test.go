package expression_test

import (
	"encoding/json"
	"testing"

	"github.com/go-msvc/expression"
	logger "github.com/go-msvc/log"
	"github.com/pkg/errors"
)

func TestCompareNumbers(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []entry{
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
	testList(t, list)
}

func TestCompareStrings(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []entry{
		{" 'ja\"n' ==\"Jan\"", false},
		{"  'jan'  =='jan'", true},
		{"  'jan'  ==\"jan\"", true},
		{"  \"jan\"  ==\"jan\"", true},
		{"  'jan'  <= 'jan'", true},
		{"  'jan'  >=\"jan\"", true},
		{"  'jan'  <= 'Jan'", false},
		{"  'jan'  >='Jan'", true},
	}
	testList(t, list)
}

func TestBool(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []entry{
		{"false || false", false},
		{"true || false", true},
		{"false || true", true},
		{"true || true", true},
		{"false && false", false},
		{"true && false", false},
		{"false && true", false},
		{"true && true", true},
	}
	testList(t, list)
}

func TestRegex(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []entry{
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
	testList(t, list)
}

func TestCompound(t *testing.T) {
	logger.Top().SetLevel(logger.ErrorLevel)
	list := []entry{
		{"(1==2)", false},
		{"(2==2)", true},
		{"1+2==3", true},
		{"(1+2)==3", true},
		{"1+(2==3)", float64(1)}, //false=0
		{"1+(3==3)", float64(2)}, //true=1
		{"(1+5)*(4-7)", float64(-18)},
		{"((1+2)*(3+4))*(-7+4)", float64(-63)}, //double brackets
	}
	testList(t, list)
}

func TestCtx(t *testing.T) {
	c := expression.NewContext()
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("s", "27821234567")
	list := []entry{
		{"a", 1},
		{"b", 2},
		{"a+b", float64(3)},
		{"b-a", float64(1)},
		{"a>b", false},
		{"b>a", true},
		{"s=='2'", false},
		{"s=='27821234567'", true},
		{"s==\"27821234567\"", true},
	}
	testListCtx(t, list, c)
}

type entry struct {
	expr     string
	expValue interface{}
}

func testList(t *testing.T, list []entry) {
	testListCtx(t, list, nil)
}

func testListCtx(t *testing.T, list []entry, ctx expression.IContext) {
	if ctx == nil {
		ctx = expression.NewContext()
	}
	for _, l := range list {
		e, err := expression.NewCompound(l.expr)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to create expression"))
		}
		val, err := e.Eval(ctx)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to eval expr(%s)", l.expr))
		}
		if val != l.expValue {
			t.Fatalf("%s -> %T(%v) != %T(%v)", l.expr, val, val, l.expValue, l.expValue)
		}
		//t.Logf("OK: %s -> %T(%v)", l.expr, val, val)

		//print expr to string
		es := e.String()

		//parse printed into new expression
		en, err := expression.NewCompound(l.expr)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to parse printed expr(%s) original(%s)", es, l.expr))
		}
		val, err = en.Eval(ctx)
		if err != nil {
			t.Fatal(errors.Wrapf(err, "failed to eval printed expr(%s) original(%s)", es, l.expr))
		}
		if val != l.expValue {
			t.Fatalf("original(%s) printed(%s) -> %T(%v) != %T(%v)", l.expr, es, val, val, l.expValue, l.expValue)
		}
		t.Logf("OK: %30.30s -> %30.30s -> %T(%v)", l.expr, es, val, val)
	}
}

//todo: test with identifiers and context

func TestJSONMarshal(t *testing.T) {
	type rule struct {
		Expr expression.Compound `json:"expr"`
	}
	r1 := rule{}
	jsonRule := `{"expr":"value==\"123\""}`
	if err := json.Unmarshal([]byte(jsonRule), &r1); err != nil {
		t.Fatal(err)
	}
	t.Logf("r1: %+v", r1)

	jsonNew := []byte{}
	var err error
	jsonNew, err = json.Marshal(r1)
	if err != nil {
		t.Fatal(err)
	}
	if len(jsonRule) != len(jsonNew) {
		t.Fatal("Different")
	}
	t.Logf("encoded: %s", string(jsonNew))

}
