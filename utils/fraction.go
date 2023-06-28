package utils

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
)

type Fraction struct {
	Num   int64
	Denum uint
}

func GCD(n1, n2 uint) uint {
	prevQ, Q := n2, n1%n2
	for Q != 0 {
		tmp := Q
		Q = prevQ % Q
		prevQ = tmp
	}
	return prevQ
}
func (frac *Fraction) reduce() Fraction {
	sign := false
	if frac.Num < 0 {
		sign = true
		frac.Num *= -1
	}
	gcd := GCD(uint(frac.Num), frac.Denum)
	frac.Num = frac.Num / int64(gcd)
	frac.Denum = frac.Denum / gcd
	if sign {
		frac.Num *= -1
	}
	return *frac
}

func MakeFraction(num int64, denum uint) Fraction {
	return (&Fraction{Num: num, Denum: denum}).reduce()
}

func (frac Fraction) String() string {
	sign := ""
	if frac.Num < 0 {
		sign = "-"
		frac.Num *= -1
	}
	rests := container.OrderedMap[uint, uint]{}
	num := uint(frac.Num)

	fix := fmt.Sprint(num / frac.Denum)
	fixFloat := ""
	repeat := ""
	for {
		quotient := num / frac.Denum
		rest := num % frac.Denum
		if rest == 0 {
			break
		}
		if _, found := rests.TryGet(rest); found {
			passed := false
			for _, key := range rests.Keys() {
				value := rests.Get(key)
				if key == rest {
					passed = true
				}
				if passed {
					repeat += fmt.Sprint(value)
				} else {
					fixFloat += fmt.Sprint(value)
				}
			}
			break
		}
		rests.Set(rest, quotient)
		num = rest
		if num < frac.Denum {
			num *= 10
		}
	}

	return fmt.Sprintf("%v%v.%v(%v)", sign, fix, fixFloat, repeat)
}

func (frac Fraction) Add(f2 Fraction) Fraction {
	return (&Fraction{Num: frac.Num*int64(f2.Denum) + f2.Num*int64(frac.Denum), Denum: frac.Denum * f2.Denum}).reduce()
}

func (frac *Fraction) AddEq(f2 Fraction) {
	*frac = frac.Add(f2)
}

func (frac Fraction) Mult(f2 Fraction) Fraction {
	return (&Fraction{Num: frac.Num * f2.Num, Denum: frac.Denum * f2.Denum}).reduce()
}
func (frac *Fraction) MultEq(f2 Fraction) {
	*frac = frac.Mult(f2)
}

func (frac Fraction) Div(f2 Fraction) Fraction {
	return (&Fraction{Num: frac.Num * int64(f2.Denum), Denum: frac.Denum * uint(f2.Num)}).reduce()
}
func (frac *Fraction) DivEq(f2 Fraction) {
	*frac = frac.Div(f2)
}

func (frac Fraction) Minus(f2 Fraction) Fraction {
	return frac.Add(f2.Neg())
}
func (frac *Fraction) MinusEq(f2 Fraction) {
	*frac = frac.Minus(f2)
}

func (frac Fraction) Neg() Fraction {
	frac.Num *= -1
	return frac
}
