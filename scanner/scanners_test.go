package scanner

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"testing"
)

func TestScanCodeLiterals(t *testing.T) {
	run := func(code, expected string, tokenList ...tokens.Token) func(t *testing.T) {
		return func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Errorf("initial value %v; expected: %v; got panic:", code, expected)
					t.Error(err)
				}
			}()
			scanCode := ScanCode(code)
			got := scanCode.String()
			if got != expected {
				t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
			}
			if tokenList != nil {
				for i, token := range scanCode.TokenList() {
					if i >= len(tokenList) {
						t.Errorf("Got more element that it should be\nexpected: %v\n got: %v", tokenList, scanCode.TokenList())
					} else if tokenList[i] != token {
						t.Errorf("invalid token\ngot: %v\nexpected: %v\ndiff at %v: %v -> %v", scanCode.TokenList(), tokenList, i, token, tokenList[i])
					}
				}
			}
		}
	}

	t.Run("simple integer 1", func(t *testing.T) {
		code := "18"
		got := ScanCode(code).String()
		expected := "18 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 2", func(t *testing.T) {
		code := "31"
		got := ScanCode(code).String()
		expected := "31 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 3", func(t *testing.T) {
		code := "42"
		got := ScanCode(code).String()
		expected := "42 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 4", func(t *testing.T) {
		code := "23"
		got := ScanCode(code).String()
		expected := "23 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("binary format", func(t *testing.T) {
		code := "0b010"
		got := ScanCode(code).String()
		expected := "2 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("octal format", func(t *testing.T) {
		code := "0o70"
		got := ScanCode(code).String()
		expected := "56 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("hex format", func(t *testing.T) {
		code := "0x0A0"
		got := ScanCode(code).String()
		expected := "160 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})

	// floating point
	t.Run("simple float 1", run("18.02", "18.02 ", tokens.FLOAT))
	t.Run("simple float 2 starts with 0", run("0.3108", "0.3108 ", tokens.FLOAT))
	t.Run("float at 0", run("0.0", "0 ", tokens.FLOAT))

	// fraction
	t.Run("simple fraction 1", run("1.0(3)", "1.0(3) ", tokens.FRACTION))
	t.Run("simple fraction 2", run("1.(3)", "1.(3) ", tokens.FRACTION))
	t.Run("simple fraction 3", run("0.01(53)", "0.01(53) ", tokens.FRACTION))

	t.Run("simple char 1", run("'a'", "'a' ", tokens.CHAR))
	t.Run("simple char 2", run("'*'", "'*' ", tokens.CHAR))
	t.Run("simple char 3", run("'0'", "'0' ", tokens.CHAR))
	t.Run("simple escape char", run(`'\n'`, "'\n' ", tokens.CHAR))
	t.Run("value escape char", run(`'\0'`, "'\000' ", tokens.CHAR))
	t.Run("complex escape char", run(`'\u{0x1f984}'`, "'\U0001f984' ", tokens.CHAR))

	t.Run("simple string", run(`"Hello there !"`, `"Hello there !" `, tokens.STR))
	t.Run("string with escape", run(`"\nGeneral Kenobi !?"`, `"\nGeneral Kenobi !?" `, tokens.STR))
	t.Run("large string", run(`"""	"ok"	"""`, `"\t\"ok\"\t" `, tokens.STR))
	t.Run("large string with quote", run(`"""\""""`, `"\"" `, tokens.STR))
	// t.Run("large string with special escape", run(`"""\	ok\ """`, `"ok" `, tokens.STR)) // Nu 1.0
}

func TestScanCodeOperators(t *testing.T) {
	run := func(code string, expectedTokens ...tokens.Token) func(t2 *testing.T) {
		return func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()

			got := ScanCode(code)
			if len(got) != len(expectedTokens) {
				t.Fatalf("expected %v element but got %v\nexpectedTokens: %v\ngot: %v", len(expectedTokens), len(got), expectedTokens, got)
			}
			for idx, expectedToken := range expectedTokens {
				if expectedToken != got[idx].Token() {
					t.Errorf("wrong token at %v (n°%v) expected: %v but got %v", got[idx].FromPos(), idx+1, expectedToken, got[idx].Token())
				}
			}
		}
	}

	t.Run("arithmetic operators", run("+ - * / \\ %", tokens.PLUS, tokens.MINUS, tokens.TIME, tokens.DIV, tokens.FRACDIV, tokens.MOD))
	t.Run("boolean operators", run("&& || !", tokens.AND, tokens.OR, tokens.NOT))
	t.Run("logical operators", run("& | ~", tokens.LAND, tokens.LOR, tokens.XOR))
	t.Run("unary operators", run("++ -- ?", tokens.PLUSPLUS, tokens.MINUSMINUS, tokens.ASK))
	t.Run("optional operator", run("? ??", tokens.ASK, tokens.ASKOR))
	t.Run("comparison operators", run("== > >= <= < !=", tokens.EQ, tokens.GT, tokens.GE, tokens.LE, tokens.LT, tokens.NEQ))
}
