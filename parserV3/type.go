package parserV3

import "github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"

type TypeExpr interface {
}

func ParseTypeExpr(scanner scan.Scanner) TypeExpr {
	scanner.ConsumeTokenInfo()
	return nil
}
