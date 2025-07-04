package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/NuCorp/NuLang/container"
)

type Fraction struct {
	Num   int64
	Denum uint
}

func GCD(n1, n2 uint) uint {
	var (
		prevQ = n1
		Q     = n2
	)

	for Q != 0 {
		prevQ, Q = Q, prevQ%Q
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

func MakeFractionFromString(str string) (Fraction, error) {
	var (
		intSplit = strings.Split(str, ".")

		res Fraction
	)

	switch len(intSplit) {
	case 1, 2:
		intPart, err := strconv.ParseInt(intSplit[0], 10, 64)

		if err != nil {
			return Fraction{}, fmt.Errorf("invalid fraction: %w", err)
		}

		res = MakeFraction(intPart, 1)

		if len(intSplit) == 1 {
			return res, nil
		}
	default:
		return Fraction{}, fmt.Errorf("invalid fraction format")
	}

	var (
		repeatSplit = strings.Split(intSplit[1], "(")
		floatOffset int
	)

	switch len(repeatSplit) {
	case 1, 2:
		floatPart, err := strconv.ParseFloat("0."+repeatSplit[0], 64)

		if err != nil {
			return Fraction{}, fmt.Errorf("invalid fraction: %w", err)
		}

		floatOffset = len(repeatSplit[0])

		power := math.Pow(10, float64(floatOffset))

		res = MakeFraction(int64((float64(res.Num)+floatPart)*power), uint(power))

		if len(repeatSplit) == 1 {
			return res, nil
		}
	default:
		return Fraction{}, fmt.Errorf("invalid fraction format")
	}

	var (
		repeatStr   = strings.TrimSuffix(repeatSplit[1], ")")
		repeat, err = strconv.ParseInt(repeatStr, 10, 64)
	)

	if err != nil {
		return Fraction{}, fmt.Errorf("invalid fraction: %w", err)
	}

	var (
		log10Repeat = len([]rune(repeatStr))

		repeatFrac = MakeFraction(repeat, uint(math.Pow(10, float64(log10Repeat))-1))
	)

	return res.Add(
		repeatFrac.Mult(MakeFraction(1, uint(math.Pow10(floatOffset)))),
	), nil
}

func (frac Fraction) String() string {
	sign := ""
	if frac.Num < 0 {
		sign = "-"
		frac.Num *= -1
	}
	type pair = [2]uint
	rests := []pair{}
	num := uint(frac.Num)

	fix := fmt.Sprint(num / frac.Denum)
	fixFloat := ""
	repeat := ""

	reachingFloatingPoint := false
	for {
		quotient := num / frac.Denum
		rest := num % frac.Denum
		if rest == 0 {
			break
		}
		couple := pair{quotient, rest}
		if container.Contains(rests, couple) {
			passed := false
			for _, pair := range rests {
				if pair == couple {
					passed = true
				}
				if passed {
					repeat += fmt.Sprint(pair[0])
				} else {
					fixFloat += fmt.Sprint(pair[0])
				}
			}
			break
		}
		rests = append(rests, couple)
		num = rest
		if num < frac.Denum {
			num *= 10
			if !reachingFloatingPoint {
				rests = []pair{}
				reachingFloatingPoint = true
			}
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
