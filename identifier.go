package expression

import "regexp"

type IIdentifier interface {
	Name() string
	Value() interface{}
}

const identifierNamePattern = `[a-zA-Z]([a-zA-Z0-9_]*[a-zA-Z0-9])?`

var identifierRegex = regexp.MustCompile("^" + identifierNamePattern + "$")

func IsValidIdentifier(s string) bool {
	return identifierRegex.MatchString(s)
}
