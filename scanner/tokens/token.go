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
	SEMI //;
	operatorEnd
)

func (t Token) String() string {
	if str, found := tokenStr[t]; found {
		return str
	}
	return "❏"
}
func (t Token) IsLiteral() bool  { return literalStart < t && t < literalEnd }
func (t Token) IsKeyword() bool  { return keywordStart < t && t < keywordEnd }
func (t Token) IsOperator() bool { return operatorStart < t && t < operatorEnd }

var tokenStr = map[Token]string{
	ERR: "TokenError",

	IDENT: "Identifier",

	INT:      "INT",
	FLOAT:    "FLOAT",
	FRACTION: "FRACTION",
	CHAR:     "CHAR",
	STR:      "STR",

	TRUE:  "true",
	FALSE: "false",

	PKG: "package",

	SEMI: ";",
}
