package expression

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jansemmelink/log"
	"github.com/pkg/errors"
)

type ICompound interface {
	Eval(ctx IContext) (interface{}, error)
	String() string
}

func NewCompound(s string) (ICompound, error) {
	e := Compound{}
	if err := e.Parse(s); err != nil {
		return e, err
	}
	return e, nil
}

type Compound struct {
	terms []term
}

func (c Compound) String() string {
	s := ""
	for _, t := range c.terms {
		if t.oper != nil {
			s += t.oper.String()
		}
		s += t.arg.String()
	}
	return s
}

//parses a compound expression breaks the expression into terms and operators
func (e *Compound) Parse(s string) error {
	log.Debugf("parsing compound[[ %s ]]", s)

	rem, err := e.parse(0, s) //0=top level
	if err != nil {
		return errors.Wrapf(err, "failed to parse expr[%s]", s)
	}
	if rem != "" {
		return errors.Wrapf(err, "text remain after expression: %s", rem)
	}
	return nil
}

type term struct {
	oper IOperator
	arg  IArgument
}

//recursive parsing functions
func (e *Compound) parse(level int, s string) (string, error) {
	//split on brackets
	//a             -> [1]{ {oper:nil, arg:a}                                   }
	//a + 2         -> [2]{ {oper:nil, arg:a}, {oper:+, arg:2}                  }
	//(a + 2)       -> [1]{ {oper:nil, arg:a+2} }
	//++a           -> [1]{ {oper:++,  arg:a} }
	//a + 2 + 3     -> [3]{ {oper:nil, arg:a}, {oper:+, arg:2} {oper:+, term:3} }
	//a + (2 + 3)	-> [2]{ {oper:nil, arg:a}, {oper:+, arg:"(2+3)"}            }

	//so parse sequentual from start
	//first term may be just arg, rest required oper then term
	//when get bracket, start scope to look recursively for closing brackets in nested values
	e.terms = []term{}
	rem := s
	for {
		rem = strings.TrimSpace(rem)
		log.Debugf("rem: \"%s\"", rem)
		if len(rem) == 0 {
			break
		}
		if rem[0] == ')' && level > 0 {
			//end of this expression
			log.Debugf("end of nested expression")
			return rem, nil
		}

		//oper is optional on first term only
		afterOper, oper := ParseOperator(rem)
		if oper == nil && len(e.terms) > 0 {
			return rem, fmt.Errorf("expect operator before \"...%s\"", rem)
		}
		if oper != nil {
			log.Debugf("oper: %s", oper)
			rem = afterOper
		}

		term := term{
			oper: oper,
			arg:  nil,
		}

		if rem[0] == '(' {
			arg := Compound{}
			var err error
			rem, err = arg.parse(
				level+1,
				rem[1:])
			if err != nil {
				return rem, errors.Wrapf(err, "failed to parse: %s", rem)
			}
			if len(rem) == 0 || rem[0] != ')' {
				return rem, fmt.Errorf("expecting ')' before %s", rem)
			}
			log.Debugf("Parsed (arg): %T %v, rem: %s", arg, arg, rem)
			term.arg = arg
			rem = rem[1:]
		} else {
			afterArg, arg, err := parseArgument(rem)
			if err != nil {
				return rem, errors.Wrapf(err, "invalid argument: %s", rem)
			}
			term.arg = arg
			rem = afterArg
			log.Debugf("Parsed arg: %T %v, rem: %s", arg, arg, rem)
		}

		e.terms = append(e.terms, term)
		log.Debugf("Now %d terms", len(e.terms))

	}

	log.Debugf("\"%s\" -> %d terms", s, len(e.terms))
	for i, t := range e.terms {
		if t.oper == nil {
			log.Debugf("  [%d] %5.5s %s", i, "", t.arg)
		} else {
			log.Debugf("  [%d] %5.5s %s", i, t.oper, t.arg)
		}
	}
	return rem, nil
}

func (c Compound) Eval(ctx IContext) (interface{}, error) {
	var val interface{}
	for _, t := range c.terms {
		if t.oper == nil {
			//first argument without operation (i.e. not "++1" etc)
			var err error
			val, err = t.arg.Eval(ctx)
			if err != nil {
				return val, errors.Wrapf(err, "cannot evaluate argument")
			}
		} else {
			//eval: value = <current value> <term.oper> <term.arg>
			newValue, err := t.oper.Eval(ctx, literal{value: val}, t.arg)
			if err != nil {
				return val, errors.Wrapf(err, "failed operation")
			}
			val = newValue
		}
	}
	return val, nil
}

func (c *Compound) UnmarshalJSON(jsonValue []byte) error {
	s := ""
	if err := json.Unmarshal(jsonValue, &s); err != nil {
		return errors.Wrapf(err, "failed to decode expression string")
	}
	if isQuoted(s, '"') {
		s = s[1 : len(s)-1]
	}
	if err := c.Parse(s); err != nil {
		return errors.Wrapf(err, "invalid expression from %s", string(jsonValue))
	}
	return nil
}

func (c Compound) MarshalJSON() ([]byte, error) {
	s := c.String()
	return json.Marshal(s)
}
