package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils"
	"strings"
)

type ComputedString struct {
	str     string
	formats []CodeToken
}

func (c ComputedString) String() string {
	for idx, format := range c.formats {
		c.str = strings.ReplaceAll(c.str, fmt.Sprintf("{%v}", idx), fmt.Sprintf(`\{%v}`, format))
	}
	return strings.ReplaceAll(c.str, "{}", "{")
}

type Int = uint
type Float = float64
type String = string
type Char = rune
type Bool = bool
type Fraction = utils.Fraction
