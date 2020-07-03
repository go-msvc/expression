package expression

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type IOperator interface {
	String() string
	Eval(ctx IContext, t1, t2 IArgument) (interface{}, error)
}

//operator implements IOperator
type operator struct {
	oper string
}

func (o operator) String() string {
	return o.oper
}

var (
	operMutex sync.Mutex
	operators = []IOperator{} //sorted from long to short
)

func MustAddOperator(oper IOperator) {
	if err := AddOperator(oper); err != nil {
		panic(err)
	}
}

func AddOperator(oper IOperator) error {
	operName := strings.TrimSpace(oper.String())
	if operName == "" || operName != oper.String() {
		return fmt.Errorf("invalid operator(%s)", operName)
	}

	operMutex.Lock()
	defer operMutex.Unlock()
	for _, o := range operators {
		if o.String() == operName {
			return fmt.Errorf("duplicate operator(%s)", operName)
		}
	}

	//insert sorted from long to short
	insertIndex := sort.Search(len(operators), func(i int) bool {
		return len(operators[i].String()) < len(operName)
	})
	operators = append(operators, nil)
	copy(operators[insertIndex+1:], operators[insertIndex:])
	operators[insertIndex] = oper

	//log.Debugf("Added oper[%d](%s)", insertIndex, operName)
	return nil
}

//parse operator at beginning of string, return remaining string and oper
func ParseOperator(s string) (string, IOperator) {
	operMutex.Lock()
	defer operMutex.Unlock()
	for _, o := range operators {
		oper := o.String()
		l := len(oper)
		if len(s) >= l && s[0:l] == oper {
			return strings.TrimSpace(s[l:]), o
		}
	}
	return s, nil //no operator found
}

//Split to get terms before and after an operator, and trim space
//<term><oper><term>
func SplitOnOperator(s string) ([]string, IOperator, error) {
	operMutex.Lock()
	defer operMutex.Unlock()
	if len(operators) == 0 {
		return []string{}, nil, fmt.Errorf("no operators registered")
	}
	for _, o := range operators {
		oper := o.String()
		operIndex := strings.Index(s, oper)
		if operIndex < 0 {
			continue
		}
		if operIndex == 0 {
			if oper == "-" { //exception: "-" before a number is part of the number
				continue
			}
			return []string{}, o, fmt.Errorf("expr(%s) has no term before %s", s, oper)
		}
		operLen := len(oper)
		if len(s) <= operIndex+operLen {
			return []string{}, o, fmt.Errorf("expr(%s) has no term after %s", s, oper)
		}

		term1 := strings.TrimSpace(s[0:operIndex])
		term2 := strings.TrimSpace(s[operIndex+operLen:])
		return []string{term1, term2}, o, nil
	}

	//not found
	list := ""
	for _, o := range operators {
		list += "|" + o.String()
	}
	if len(list) > 0 {
		list = list[1:]
	}
	return []string{}, nil, fmt.Errorf("expr(%s) no operator, expecting(%s)", s, list)
}

func init() {
	MustAddOperator(operNumeric{name: "+", nfnc: func(n1, n2 float64) float64 { return n1 + n2 }})
	MustAddOperator(operNumeric{name: "-", nfnc: func(n1, n2 float64) float64 { return n1 - n2 }})
	MustAddOperator(operNumeric{name: "*", nfnc: func(n1, n2 float64) float64 { return n1 * n2 }})
	MustAddOperator(operNumeric{name: "/", nfnc: func(n1, n2 float64) float64 {
		if n2 == 0 {
			return 0
		}
		return n1 / n2
	}})

	MustAddOperator(operCompare{name: "==", sfnc: func(s1, s2 string) bool { return s1 == s2 }, nfnc: func(s1, s2 float64) bool { return s1 == s2 }})
	MustAddOperator(operCompare{name: "<", sfnc: func(s1, s2 string) bool { return s1 < s2 }, nfnc: func(s1, s2 float64) bool { return s1 < s2 }})
	MustAddOperator(operCompare{name: "<=", sfnc: func(s1, s2 string) bool { return s1 <= s2 }, nfnc: func(s1, s2 float64) bool { return s1 <= s2 }})
	MustAddOperator(operCompare{name: ">", sfnc: func(s1, s2 string) bool { return s1 > s2 }, nfnc: func(s1, s2 float64) bool { return s1 > s2 }})
	MustAddOperator(operCompare{name: ">=", sfnc: func(s1, s2 string) bool { return s1 >= s2 }, nfnc: func(s1, s2 float64) bool { return s1 >= s2 }})

	MustAddOperator(operBool{name: "&&", bfnc: func(b1, b2 bool) bool { return b1 && b2 }})
	MustAddOperator(operBool{name: "||", bfnc: func(b1, b2 bool) bool { return b1 || b2 }})

	MustAddOperator(operString{name: "~=", sfnc: func(s1, s2 string) bool {
		r, err := regexp.Compile(s2)
		if err != nil {
			log.Errorf("cannot compile regex(%s): %v", s2, err)
			return false
		}
		return r.MatchString(s1)
	}})
}

type operCompare struct {
	name string
	sfnc func(v1, v2 string) bool
	nfnc func(v1, v2 float64) bool
}

func (o operCompare) String() string { return o.name }

func (o operCompare) Eval(ctx IContext, t1, t2 IArgument) (interface{}, error) {
	v1, err := t1.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 1st term")
	}
	v2, err := t2.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 2nd term")
	}

	if s1, ok := v1.(string); ok {
		if s2, ok := v2.(string); ok {
			return o.sfnc(s1, s2), nil
		}
	}

	n1, n2, ok := getNumbers(v1, v2)
	if !ok {
		return nil, fmt.Errorf("cannot compare %T %s %T", v1, o.name, v2)
	}
	return o.nfnc(n1, n2), nil
}

type operBool struct {
	name string
	bfnc func(b1, b2 bool) bool
}

func (o operBool) String() string { return o.name }

func (o operBool) Eval(ctx IContext, t1, t2 IArgument) (interface{}, error) {
	v1, err := t1.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 1st term")
	}
	v2, err := t2.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 2nd term")
	}
	if b1, ok := v1.(bool); ok {
		if b2, ok := v2.(bool); ok {
			return o.bfnc(b1, b2), nil
		}
	}
	return nil, fmt.Errorf("cannot compare %T %s %T", v1, o.name, v2)
}

type operString struct {
	name string
	sfnc func(v1, v2 string) bool
}

func (o operString) String() string { return o.name }

func (o operString) Eval(ctx IContext, t1, t2 IArgument) (interface{}, error) {
	v1, err := t1.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 1st term")
	}
	v2, err := t2.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 2nd term")
	}
	if s1, ok := v1.(string); ok {
		if s2, ok := v2.(string); ok {
			return o.sfnc(s1, s2), nil
		}
	}
	return nil, fmt.Errorf("cannot evaluate %T %s %T", v1, o.name, v2)
}

//todo:
//bool operators:
//	not
//	in set [a,t,5]
//non-bool operators - +-/*...
//	integer mod & div

type operNumeric struct {
	name string
	nfnc func(v1, v2 float64) float64
}

func (o operNumeric) String() string { return o.name }

func (o operNumeric) Eval(ctx IContext, t1, t2 IArgument) (interface{}, error) {
	v1, err := t1.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 1st term")
	}
	v2, err := t2.Eval(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to eval 2nd term")
	}

	n1, n2, ok := getNumbers(v1, v2)
	if !ok {
		return nil, fmt.Errorf("cannot evaluate %T %s %T", v1, o.name, v2)
	}
	return o.nfnc(n1, n2), nil
}

func getNumbers(v1, v2 interface{}) (float64, float64, bool) {
	n1, ok1 := getNumber(v1)
	n2, ok2 := getNumber(v2)
	if ok1 && ok2 {
		return n1, n2, true
	}
	return n1, n2, false
}

func getNumber(v interface{}) (float64, bool) {
	var n float64
	var ok bool
	if n, ok = v.(float64); !ok {
		var i int
		if i, ok = v.(int); ok {
			n = float64(i)
		} else {
			var b bool
			if b, ok = v.(bool); ok {
				if b {
					n = 1
				} else {
					n = 0
				}
			}
		}

	}
	return n, ok
}
