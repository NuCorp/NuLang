package token

import (
	"maps"

	"github.com/LicorneSharing/GTL/iter"
)

type Token int

const (
	// Special tokens
	special_start Token = iota - 1
	Illegal
	EOF
	EOI     // ; or \n
	COMMENT // // or /* ... */
	special_end

	// Identifiers and literals
	literal_start
	IDENT
	INT
	FLOAT
	FRAC // fraction number either: INT \ INT or INT.INT(INT) (or .INT(INT) or INT.(INT) or .(INT))
	STRING
	CHAR
	literal_end

	// Keywords
	keyword_start
	PKG
	IMPORT

	FUNC
	VAR
	CONST
	TYPE
	STRUCT
	INTERFACE
	ENUM
	EXTENSION
	GET
	SET
	INIT
	AS
	EXPLICIT
	IMPLICIT
	IS
	OPERATOR // to be developed in the future

	IF
	ELSE
	WITH
	FOR
	WHILE
	DO
	THEN
	TRY
	CATCH
	THROW
	RETURN
	CONTINUE
	BREAK
	DEFER
	RUN
	CHAN
	AWAIT // may be not needed
	SELECT
	CASE
	DEFAULT
	keyword_end

	// Operators and punctuation
	op_punctu_start
	ADD  // +
	SUB  // -
	MUL  // *
	DIV  // /
	MOD  // %
	DIVF // \
	POW  // **

	EQ  // ==
	NEQ // !=
	LT  // <
	GT  // >
	LE  // <=
	GE  // >=
	AND // &&
	OR  // ||
	XOR // ^

	LAND // & (bitwise AND)
	LOR  // | (bitwise OR)

	ASK   // ?
	ASKOR // ??

	NOT    // !
	ASSIGN // =

	PLUSPLUS   // ++
	MINUSMINUS // --
	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	DIV_ASSIGN // /=
	MOD_ASSIGN // %=

	AND_ASSIGN // &&=
	OR_ASSIGN  // ||=
	XOR_ASSIGN // ^=

	LAND_ASSIGN // &=
	LOR_ASSIGN  // |=

	DEF // `:=` (short variable declaration, like in Go)

	DOT         // .
	COMMA       // ,
	COLON       // :
	ELLIPSIS    // ...
	ARROW       // =>
	ARROW_LEFT  // ->
	ARROW_RIGHT // <-
	DOLLAR      // $

	OBRAC  // {
	CBRAC  // }
	OBRAK  // [
	CBRAK  // ]
	OPAREN // (
	CPAREN // )

	op_punctu_end
)

const (
	BANG = NOT  // !
	STAR = MUL  // *
	ADDR = LAND // &
	PIPE = LOR  // |
)

func (t Token) IsOneOf(toks ...Token) bool {
	for _, tok := range toks {
		if tok == t {
			return true
		}
	}

	return false
}

func (t Token) String() string {
	if str, ok := Strings[t]; ok {
		return str
	}
	return "Unknown"
}

// IsKeyword checks if the token is a keyword.
func (t Token) IsKeyword() bool {
	return t >= keyword_start && t < keyword_end
}

// IsLiteral checks if the token is a literal.
func (t Token) IsLiteral() bool {
	return t >= literal_start && t < literal_end
}

// IsOperator checks if the token is an operator or punctuation.
func (t Token) IsOperator() bool {
	return t >= op_punctu_start && t < op_punctu_end
}

// IsSpecial checks if the token is a special token (like Illegal, EOF, etc.).
func (t Token) IsSpecial() bool {
	return t >= special_start && t < special_end
}

var (
	Strings = map[Token]string{
		Illegal:     "Illegal",
		EOF:         "EOF",
		EOI:         "EOI",
		COMMENT:     "COMMENT",
		IDENT:       "IDENT",
		INT:         "INT",
		FLOAT:       "FLOAT",
		FRAC:        "FRAC",
		STRING:      "STRING",
		CHAR:        "CHAR",
		PKG:         "package",
		IMPORT:      "import",
		FUNC:        "func",
		VAR:         "var",
		CONST:       "const",
		TYPE:        "type",
		STRUCT:      "struct",
		INTERFACE:   "interface",
		ENUM:        "enum",
		EXTENSION:   "extension",
		GET:         "get",
		SET:         "set",
		INIT:        "init",
		AS:          "as",
		EXPLICIT:    "explicit",
		IMPLICIT:    "implicit",
		IS:          "is",
		OPERATOR:    "operator",
		IF:          "if",
		ELSE:        "else",
		WITH:        "with",
		FOR:         "for",
		WHILE:       "while",
		DO:          "do",
		THEN:        "then",
		TRY:         "try",
		CATCH:       "catch",
		THROW:       "throw",
		RETURN:      "return",
		CONTINUE:    "continue",
		BREAK:       "break",
		DEFER:       "defer",
		RUN:         "run",
		CHAN:        "chan",
		AWAIT:       "await",
		SELECT:      "select",
		CASE:        "case",
		DEFAULT:     "default",
		ADD:         "+",
		SUB:         "-",
		MUL:         "*",
		DIV:         "/",
		MOD:         "%",
		DIVF:        "\\",
		POW:         "**",
		EQ:          "==",
		NEQ:         "!=",
		LT:          "<",
		GT:          ">",
		LE:          "<=",
		GE:          ">=",
		AND:         "&&",
		OR:          "||",
		XOR:         "^",
		LAND:        "&",
		LOR:         "|",
		ASK:         "?",
		ASKOR:       "??",
		NOT:         "!",
		ASSIGN:      "=",
		PLUSPLUS:    "++",
		MINUSMINUS:  "--",
		ADD_ASSIGN:  "+=",
		SUB_ASSIGN:  "-=",
		MUL_ASSIGN:  "*=",
		DIV_ASSIGN:  "/=",
		MOD_ASSIGN:  "%=",
		AND_ASSIGN:  "&&=",
		OR_ASSIGN:   "||=",
		XOR_ASSIGN:  "^=",
		LAND_ASSIGN: "&=",
		LOR_ASSIGN:  "|=",
		OBRAC:       "{",
		CBRAC:       "}",
		OBRAK:       "[",
		CBRAK:       "]",
		OPAREN:      "(",
		CPAREN:      ")",
		DEF:         ":=",
		DOT:         ".",
		COMMA:       ",",
		COLON:       ":",
		ELLIPSIS:    "...",
		ARROW:       "=>",
		ARROW_LEFT:  "->",
		ARROW_RIGHT: "<-",
		DOLLAR:      "$",
	}
	fromString = func() map[string]Token {
		m := make(map[string]Token, len(Strings))

		for t, str := range Strings {
			m[str] = t
		}

		return m
	}
)

func FromString(s string) (Token, bool) {
	if t, ok := fromString()[s]; ok {
		return t, true
	}

	return Illegal, false
}

func Iter() iter.Seq2[Token, string] {
	return iter.Seq2[Token, string](maps.All(Strings))
}
