package ast

import "strings"

type DotIdent []string

func (d DotIdent) Pack() string {
	return strings.Join(d, ".")
}

func (d DotIdent) First() string {
	if d[0] == "" {
		return "self"
	}

	return d[0]
}

func (d DotIdent) Last() string {
	return d[len(d)-1]
}
