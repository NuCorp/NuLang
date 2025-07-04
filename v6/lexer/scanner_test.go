package lexer

import (
	"testing"

	"github.com/LicorneSharing/GTL/slices"
	"github.com/stretchr/testify/assert"

	"github.com/NuCorp/NuLang/v6/lexer/token"
)

func Test_common_operator(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		expectToken token.Token
	}{
		{"ADD", "+", token.ADD},
		{"SUB", "-", token.SUB},
		{"MUL", "*", token.MUL},
		{"DIV", "/", token.DIV},
		{"MOD", "%", token.MOD},
		{"DIVF", "\\", token.DIVF},
		{"POW", "**", token.POW},
		{"EQ", "==", token.EQ},
		{"NEQ", "!=", token.NEQ},
		{"LT", "<", token.LT},
		{"GT", ">", token.GT},
		{"LE", "<=", token.LE},
		{"GE", ">=", token.GE},
		{"AND", "&&", token.AND},
		{"OR", "||", token.OR},
		{"XOR", "^", token.XOR},
		{"LAND", "&", token.LAND},
		{"LOR", "|", token.LOR},
		{"ASK", "?", token.ASK},
		{"ASKOR", "??", token.ASKOR},
		{"NOT", "!", token.NOT},
		{"ASSIGN", "=", token.ASSIGN},
		{"PLUSPLUS", "++", token.PLUSPLUS},
		{"MINUSMINUS", "--", token.MINUSMINUS},
		{"ADD_ASSIGN", "+=", token.ADD_ASSIGN},
		{"SUB_ASSIGN", "-=", token.SUB_ASSIGN},
		{"MUL_ASSIGN", "*=", token.MUL_ASSIGN},
		{"DIV_ASSIGN", "/=", token.DIV_ASSIGN},
		{"MOD_ASSIGN", "%=", token.MOD_ASSIGN},
		{"AND_ASSIGN", "&&=", token.AND_ASSIGN},
		{"OR_ASSIGN", "||=", token.OR_ASSIGN},
		{"XOR_ASSIGN", "^=", token.XOR_ASSIGN},
		{"LAND_ASSIGN", "&=", token.LAND_ASSIGN},
		{"LOR_ASSIGN", "|=", token.LOR_ASSIGN},
		{"DEF", ":=", token.DEF},
		{"DOT", ".", token.DOT},
		{"COMMA", ",", token.COMMA},
		{"COLON", ":", token.COLON},
		{"ELLIPSIS", "...", token.ELLIPSIS},
		{"ARROW", "=>", token.ARROW},
		{"ARROW_LEFT", "->", token.ARROW_LEFT},
		{"ARROW_RIGHT", "<-", token.ARROW_RIGHT},
		{"DOLLAR", "$", token.DOLLAR},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			sc := CodeScanner(tt.input).(*scanner)
			sc.operator()
			assert.Len(t, sc.tokens, 1)
			assert.Equal(
				t, tt.expectToken, sc.tokens[0].Token,
				"got %v, expected %v", sc.tokens[0].Token, tt.expectToken,
			)
		})
	}
}

func Test_common_identifier(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		expectToken token.Token
	}{
		// Mots-clés (keywords)
		{"PKG", "package", token.PKG},
		{"IMPORT", "import", token.IMPORT},
		{"FUNC", "func", token.FUNC},
		{"VAR", "var", token.VAR},
		{"CONST", "const", token.CONST},
		{"TYPE", "type", token.TYPE},
		{"STRUCT", "struct", token.STRUCT},
		{"INTERFACE", "interface", token.INTERFACE},
		{"ENUM", "enum", token.ENUM},
		{"EXTENSION", "extension", token.EXTENSION},
		{"GET", "get", token.GET},
		{"SET", "set", token.SET},
		{"INIT", "init", token.INIT},
		{"AS", "as", token.AS},
		{"EXPLICIT", "explicit", token.EXPLICIT},
		{"IMPLICIT", "implicit", token.IMPLICIT},
		{"IS", "is", token.IS},
		{"OPERATOR", "operator", token.OPERATOR},
		{"IF", "if", token.IF},
		{"ELSE", "else", token.ELSE},
		{"WITH", "with", token.WITH},
		{"FOR", "for", token.FOR},
		{"WHILE", "while", token.WHILE},
		{"DO", "do", token.DO},
		{"THEN", "then", token.THEN},
		{"TRY", "try", token.TRY},
		{"CATCH", "catch", token.CATCH},
		{"THROW", "throw", token.THROW},
		{"RETURN", "return", token.RETURN},
		{"CONTINUE", "continue", token.CONTINUE},
		{"BREAK", "break", token.BREAK},
		{"DEFER", "defer", token.DEFER},
		{"RUN", "run", token.RUN},
		{"CHAN", "chan", token.CHAN},
		{"AWAIT", "await", token.AWAIT},
		{"SELECT", "select", token.SELECT},
		{"CASE", "case", token.CASE},
		{"DEFAULT", "default", token.DEFAULT},

		// Identifiants simples
		{"IDENT_simple", "foo", token.IDENT},
		{"IDENT_underscore", "_bar", token.IDENT},
		{"IDENT_mixed", "foo123", token.IDENT},
		{"IDENT_caps", "FOO", token.IDENT},
		{"IDENT_with_digit", "a1b2c3", token.IDENT},
		{"IDENT_looks_like_keyword", "if123", token.IDENT},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			sc := CodeScanner(tt.input).(*scanner)
			sc.identifier()
			assert.Len(t, sc.tokens, 1)
			assert.Equal(
				t, tt.expectToken, sc.tokens[0].Token,
				"got %v, expected %v", sc.tokens[0].Token, tt.expectToken,
			)
		})
	}
}

func Test_common_string(t *testing.T) {
	testcases := []struct {
		name         string
		input        string
		expectToken  token.Token
		expectValue  string
		expectInterp bool
	}{
		{"simple", `"abc"`, token.STRING, "abc", false},
		{"escaped_quote", `"a\"b"`, token.STRING, `a"b`, false},
		{"escaped_newline", `"a\nb"`, token.STRING, "a\nb", false},
		{"interpolation", `"a\{b}"`, token.STRING, "a\\{b}", true},
		{"unterminated", `"abc`, token.Illegal, "", false},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			sc := CodeScanner(tt.input).(*scanner)
			sc.string()
			assert.Len(t, sc.tokens, 1)
			assert.Equal(t, tt.expectToken, sc.tokens[0].Token)
			if tt.expectToken == token.STRING {
				assert.Equal(t, tt.expectValue, sc.tokens[0].RawValue)
				assert.Equal(t, tt.expectInterp, sc.tokens[0].WithInterpolation)
			}
		})
	}
}

func Test_common_number(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		expectToken token.Token
		expectRaw   string
	}{
		{"int", "123", token.INT, "123"},
		{"float", "123.45", token.FLOAT, "123.45"},
		{"frac", "1.2(34)", token.FRAC, "1.2(34)"},
		{"hex", "0x1A", token.INT, "0x1A"},
		{"octal", "0o77", token.INT, "0o77"},
		{"binary", "0b101", token.INT, "0b101"},
		{"illegal", "12a", token.Illegal, "12"},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			sc := CodeScanner(tt.input).(*scanner)
			sc.number()
			assert.Len(t, sc.tokens, 1)
			assert.Equal(
				t, tt.expectToken, sc.tokens[0].Token,
				"got %v, expected %v", sc.tokens[0].Token, tt.expectToken,
			)
			assert.Equal(t, tt.expectRaw, sc.tokens[0].Raw)
		})
	}
}

func Test_common_eoi(t *testing.T) {
	testcases := []struct {
		name         string
		input        string
		expectTokens []token.Token
	}{
		{"semicolon", "a;", []token.Token{token.IDENT, token.EOI}},
		{"newline", "a\n", []token.Token{token.IDENT, token.EOI}},
		{"no_eoi", "+\n", []token.Token{token.ADD}},
		{"remove spaces", "\n\t a\n\t;", []token.Token{token.IDENT, token.EOI}},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			sc := CodeScanner(tt.input).(*scanner)

			for sc.Scan() {
			}

			gotTokens := slices.Map(sc.tokens, TokenInfo.GetToken)

			assert.Equal(
				t, tt.expectTokens, gotTokens,
				"expected tokens %v, got %v", tt.expectTokens, gotTokens,
			)
		})
	}
}

func Test_common_char(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		expectToken token.Token
		expectValue string
	}{
		{"simple", `'a'`, token.CHAR, "a"},
		{"escaped_quote", `'\''`, token.CHAR, "'"},
		{"unicode_code", `'\{65}'`, token.CHAR, "A"},
		{"illegal_unclosed", `'a`, token.Illegal, ""},
		{"illegal_escape", `'\\x'`, token.Illegal, ""},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			sc := CodeScanner(tt.input).(*scanner)
			sc.char()
			assert.Len(t, sc.tokens, 1)
			assert.Equal(
				t, tt.expectToken, sc.tokens[0].Token,
				"got %v, expected %v", sc.tokens[0].Token, tt.expectToken,
			)
			assert.Equal(t, tt.expectValue, sc.tokens[0].RawValue)

		})
	}
}
