package expression

import (
	"fmt"

	logger "github.com/go-msvc/log"
	"github.com/pkg/errors"
)

var log = logger.ForThisPackage().WithLevel(logger.ErrorLevel)

type IExpression interface {
	Eval(ctx IContext) (interface{}, error)
}

func New(s string) (IExpression, error) {
	e := Expression{}
	if err := e.Parse(s, nil); err != nil {
		return e, err
	}
	return e, nil
}

//Expression
type Expression struct {
	valid bool
	term1 IArgument
	oper  IOperator
	term2 IArgument
}

//parses a simple <term1><oper><term2> expression
func (e *Expression) Parse(s string, ctx IContext) error {
	terms, oper, err := SplitOnOperator(s)
	if err != nil {
		return errors.Wrapf(err, "invalid expression")
	}

	log.Debugf("expr(%s) -> term1(%s) oper(%s) term2(%s)", s, terms[0], oper, terms[1])

	//parse the terms
	e.term1, err = ParseArgument(terms[0])
	if err != nil {
		return errors.Wrapf(err, "invalid 1st term(%s)", terms[0])
	}
	e.term2, err = ParseArgument(terms[1])
	if err != nil {
		return errors.Wrapf(err, "invalid 2nd term(%s)", terms[1])
	}
	e.oper = oper
	e.valid = true
	return nil
}

func (e Expression) Eval(ctx IContext) (interface{}, error) {
	if !e.valid || e.oper == nil || e.term1 == nil || e.term2 == nil {
		return nil, fmt.Errorf("cannot eval expression(%v,%v,%v,%v)", e.valid, e.oper, e.term1, e.term2)
	}
	return e.oper.Eval(ctx, e.term1, e.term2)
}
