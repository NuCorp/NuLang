package scan

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"math/rand"
	"testing"

	"github.com/DarkMiMolle/GTL/array"
)

func TestCodeLiterals(t *testing.T) {
	run := func(code, expected string, tokenList ...tokens.Token) func(t *testing.T) {
		return func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Errorf("initial value %v; expected: %v; got panic:", code, expected)
					t.Error(err)
				}
			}()
			scanner := Code(code)
			scanCode := scanner.LookUp(-1)
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
		got := Code(code).LookUp(1).String()
		expected := "18 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 2", func(t *testing.T) {
		code := "31"
		got := Code(code).LookUp(1).String()
		expected := "31 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 3", func(t *testing.T) {
		code := "42"
		got := Code(code).LookUp(1).String()
		expected := "42 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 4", func(t *testing.T) {
		code := "23"
		got := Code(code).LookUp(1).String()
		expected := "23 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("binary format", func(t *testing.T) {
		code := "0b010"
		got := Code(code).LookUp(1).String()
		expected := "2 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("octal format", func(t *testing.T) {
		code := "0o70"
		got := Code(code).LookUp(1).String()
		expected := "56 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("hex format", func(t *testing.T) {
		code := "0x0A0"
		got := Code(code).LookUp(1).String()
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

func TestCodeOperators(t *testing.T) {
	run := func(code string, expectedTokens ...tokens.Token) func(t2 *testing.T) {
		return func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()
			//expectedTokens = append(expectedTokens)
			scanner := Code(code)
			got := scanner.LookUp(-1)
			if len(got.TokenList()) != len(expectedTokens) {
				t.Fatalf("expected %v element but got %v\nexpectedTokens: %v\ngot: %v", len(expectedTokens), len(got), expectedTokens, got.TokenList())
			}
			for idx, expectedToken := range expectedTokens {
				if expectedToken != got[idx].Token() {
					t.Errorf("wrong token at %v (n°%v) expected: %v but got %v", got[idx].FromPos(), idx+1, expectedToken, got[idx].Token())
				}
			}
		}
	}

	t.Run("arithmetic operators", run("+ - * / \\ %", tokens.PLUS, tokens.MINUS, tokens.TIME, tokens.DIV, tokens.FRAC_DIV, tokens.MOD))
	t.Run("boolean operators", run("&& || !", tokens.AND, tokens.OR, tokens.NOT))
	t.Run("logical operators", run("& | ~ << >>", tokens.LAND, tokens.LOR, tokens.XOR, tokens.LSHIFT, tokens.RSHIFT))
	t.Run("unary operators", run("++ -- ?", tokens.PLUS_PLUS, tokens.MINUS_MINUS, tokens.ASK))
	t.Run("optional operator", run("? ??", tokens.ASK, tokens.ASKOR))
	t.Run("comparison operators", run("== > >= <= < !=", tokens.EQ, tokens.GT, tokens.GE, tokens.LE, tokens.LT, tokens.NEQ))
	t.Run("assignment operators", run("= += -= *= /= %= &&= ||= &= |= ~= :=",
		tokens.ASSIGN, tokens.PLUS_ASSIGN, tokens.MINUS_ASSIGN, tokens.TIME_ASSIGN, tokens.DIV_ASSIGN, tokens.MOD_ASSIGN,
		tokens.AND_ASSIGN, tokens.OR_ASSIGN, tokens.LAND_ASSIGN, tokens.LOR_ASSIGN, tokens.XOR_ASSIGN, tokens.DEFINE))
	t.Run("punctuations", run(". ... : ; { } ( ) [ ] -> <- => ,",
		tokens.DOT, tokens.ELLIPSIS, tokens.COLON, tokens.SEMI, tokens.OBRAC, tokens.CBRAC, tokens.OPAREN, tokens.CPAREN,
		tokens.OBRAK, tokens.CBRAK, tokens.RARROW, tokens.LARROW, tokens.IMPL, tokens.COMA))

	t.Run("period error", run("..+", tokens.ERR, tokens.PLUS))
}

func TestCodeText(t *testing.T) {
	run := func(code string, expectedTokens ...tokens.Token) func(t2 *testing.T) {
		return func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()

			scanner := Code(code)
			got := scanner.LookUp(-1)
			if len(got.TokenList()) != len(expectedTokens) {
				t.Fatalf("expected %v element but got %v\nexpectedTokens: %v\ngot: %v", len(expectedTokens), len(got.TokenList()), expectedTokens, got.TokenList())
			}
			for idx, expectedToken := range expectedTokens {
				if expectedToken != got[idx].Token() {
					t.Errorf("wrong token at %v (n°%v) expected: %v but got %v", got[idx].FromPos(), idx+1, expectedToken, got[idx].Token())
				}
			}
		}
	}

	t.Run("one keyword", run("package", tokens.PKG))
	t.Run("some keyword", run("package import var func", tokens.PKG, tokens.IMPORT, tokens.VAR, tokens.FUNC))
	t.Run("keyword with ident", run("var a int import Std as standard", tokens.VAR, tokens.IDENT, tokens.IDENT,
		tokens.IMPORT, tokens.IDENT, tokens.AS, tokens.IDENT))
	t.Run("no ident ident and underscored ident", run("_ _yo", tokens.NO_IDENT, tokens.IDENT))

	var keywordTokens []tokens.Token
	tokens.ForEach(func(tok tokens.Token) {
		if !tok.IsKeyword() {
			return
		}
		keywordTokens = append(keywordTokens, tok)
	})

	t.Run("random keywords", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			idx := rand.Int() % len(keywordTokens)
			token := keywordTokens[idx]
			t.Logf("[%v/10] testing token: %v", i+1, token)
			run(token.String(), token)
		}
	})
}

func TestCode(t *testing.T) {
	run := func(code string, expectedTokens ...tokens.Token) func(t2 *testing.T) {
		return func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Error(err)
				}
			}()

			scanner := Code(code)
			got := scanner.LookUp(-1)
			if len(got.TokenList()) != len(expectedTokens) {
				t.Fatalf("expected %v element but got %v\nexpectedTokens: %v\ngot: %v", len(expectedTokens), len(got.TokenList()), expectedTokens, got.TokenList())
			}
			for idx, expectedToken := range expectedTokens {
				if expectedToken != got[idx].Token() {
					t.Errorf("wrong token at %v (n°%v) expected: %v but got %v", got[idx].FromPos(), idx+1, expectedToken, got[idx].Token())
				}
			}
		}
	}
	t.Run("with dot", run("a.b", tokens.IDENT, tokens.DOT, tokens.IDENT))
}

func MockScannerWith(toks ...tokens.Token) Scanner {
	scanner := &CodeScanner{}
	var i = 0
	scanner.commonScanner = commonScanner{
		Scanner: scanner,
		tokens: array.Map(toks, func(tok tokens.Token) TokenInfo {
			var value any = tok.String()
			if tok.IsLiteral() {
				switch tok {
				case tokens.STR:
					value = "42"
				case tokens.INT:
					value = 42
				case tokens.FLOAT:
					value = 42.0
				case tokens.CHAR:
					value = '*'
				case tokens.FRACTION:
					value = Fraction{42, 1}
				}
			}
			defer func() {
				i += len(tok.String())
			}()
			return TokenInfo{
				rawValue: tok.String(),
				token:    tok,
				from:     TokenPos{col: i},
				to:       TokenPos{col: i + len(tok.String()) - 1},
				value:    value,
				errorRef: 0,
			}
		}),
		current: 0,
		ended:   false,
	}

	return scanner
}
