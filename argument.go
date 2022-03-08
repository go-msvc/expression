package expression

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

type IArgument interface {
	Eval(ctx IContext) (interface{}, error)
	String() string
}

type argument struct {
	//one of the following
	literal    ILiteral
	identifier IIdentifier
}

func ParseArgument(s string) (IArgument, error) {
	s = strings.TrimSpace(s)
	if isQuoted(s, '\'') || isQuoted(s, '"') {
		l := len(s)
		return &literal{value: s[1 : l-1]}, nil
	}

	if i, err := strconv.Atoi(s); err == nil {
		return &literal{value: i}, nil
	}

	if i, err := strconv.ParseBool(s); err == nil {
		return &literal{value: i}, nil
	}

	if i, err := strconv.ParseFloat(s, 64); err == nil {
		return &literal{value: i}, nil
	}

	if IsValidIdentifier(s) {
		return &identifier{name: s}, nil
	}

	return nil, fmt.Errorf("unknown type of argument(%s)", s)
}

//parseArgument from start of s and return remaining expession
func parseArgument(s string) (string, IArgument, error) {
	if len(s) == 0 {
		return s, nil, fmt.Errorf("no argument")
	}

	var l int
	if s[0] == '\'' || s[0] == '"' {
		//read quoted argument
		// var sc scanner.Scanner
		// sc.Init(strings.NewReader(s))
		// token := sc.Scan()
		// if token == scanner.EOF {
		// 	return s, nil, fmt.Errorf("failed to read quoted argument from %s", s)
		// }
		// l = len(sc.TokenText())

		l = quotedLength(s)
	} else {
		//unquoted: get next separator: '(', ')', any operator, or white space
		l = 1
		for l < len(s) {
			if unicode.IsSpace(rune(s[l])) || s[l] == '(' || s[l] == ')' {
				break
			}
			if _, oper := ParseOperator(s[l:]); oper != nil {
				break
			}
			l++
		}
	}

	arg, err := ParseArgument(s[:l])
	if err != nil {
		return s, nil, errors.Wrapf(err, "expected valid argument (l=%d) at %s", l, s)
	}

	return s[l:], arg, nil
}

type ILiteral interface {
	Value() interface{}
}

func isQuoted(s string, c byte) bool {
	l := len(s)
	if l >= 2 && s[0] == c && s[l-1] == c {
		return true
	}
	return false
}

type literal struct {
	value interface{}
}

func (l literal) String() string {
	if s, ok := l.value.(string); ok {
		return fmt.Sprintf("\"%s\"", s)
	}
	return fmt.Sprintf("%v", l.value)
}

func (l literal) Eval(ctx IContext) (interface{}, error) {
	return l.value, nil
}

type identifier struct {
	name string
}

func (l identifier) String() string {
	return l.name
}

func (l identifier) Eval(ctx IContext) (interface{}, error) {
	i := ctx.Get(l.name)
	if i == nil {
		return nil, fmt.Errorf("identifier(%s) not found", l.name)
	}
	return i, nil
}

func quotedLength(s string) int {
	if s[0] != '\'' && s[0] != '"' {
		return 0
	}
	c := s[0]
	l := 1
	for l < len(s) {
		if s[l] == c {
			l++
			break //end of string
		}
		l++ //skip over non-quote char
	}
	return l
}
