package tokens

type Token int

const (
	NoInit = Token(iota)
	ERR    = Token(iota)
	IDENT
	literalStart
	INT      // int
	FLOAT    // float
	FRACTION // fraction
	CHAR     // char
	STR      // string
	literalEnd
	keywordStart
	TRUE  // true,
	FALSE // false,
	PKG   // package,
	keywordEnd

	operatorStart
	PLUS
	MINUS
	TIME
	DIV
	FRACDIV

	operatorEnd
	punctuationStart

	SEMI //;

	punctuationEnd
)

func (t Token) String() string {
	if str, found := tokenStr[t]; found {
		return str
	}
	return "❏"
}
func (t Token) IsLiteral() bool     { return literalStart < t && t < literalEnd }
func (t Token) IsKeyword() bool     { return keywordStart < t && t < keywordEnd }
func (t Token) IsOperator() bool    { return operatorStart < t && t < operatorEnd }
func (t Token) IsPunctuation() bool { return punctuationStart < t && t < punctuationEnd }

var tokenStr = map[Token]string{
	ERR: "TokenError",

	IDENT: "Identifier",

	// literals
	INT:      "INT",
	FLOAT:    "FLOAT",
	FRACTION: "FRACTION",
	CHAR:     "CHAR",
	STR:      "STR",

	// keywords
	TRUE:  "true",
	FALSE: "false",

	PKG: "package",

	// operators
	PLUS:    "+",
	MINUS:   "-",
	TIME:    "*",
	DIV:     "/",
	FRACDIV: "\\",

	// punctuations
	SEMI: ";",
}
