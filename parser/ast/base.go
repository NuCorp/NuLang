package ast

import "strings"

type DotIdent []string

func (d DotIdent) Pack() string {
	return strings.Join(d, ".")
}
