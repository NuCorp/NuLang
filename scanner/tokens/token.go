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
	arithmeticOperatorStart
	PLUS
	MINUS
	TIME
	DIV
	FRACDIV
	MOD

	PLUSPLUS   // "++",
	MINUSMINUS // "--",
	ASK        // "?",
	ASKOR      // "??",
	arithmeticOperatorEnd
	booleanOperatorStart
	AND // "&&",
	OR  // "||",
	NOT // "!"
	booleanOperatorEnd
	logicalOperatorStart
	LAND // "&",
	LOR  // "|",
	XOR  // "~"
	logicalOperatorEnd

	EQ  // "=="
	NEQ // "!="
	GT  // ">"
	LT  // "<"
	GE  // ">="
	LE  // "<="

	ASSIGN // "="

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
func (t Token) IsLiteral() bool  { return literalStart < t && t < literalEnd }
func (t Token) IsKeyword() bool  { return keywordStart < t && t < keywordEnd }
func (t Token) IsOperator() bool { return operatorStart < t && t < operatorEnd }
func (t Token) IsLogicalOperator() bool {
	return logicalOperatorStart < t && t < logicalOperatorEnd || t == NOT
}
func (t Token) IsPunctuation() bool { return punctuationStart < t && t < punctuationEnd || t == NOT }

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
	PLUS:       "+",
	MINUS:      "-",
	TIME:       "*",
	DIV:        "/",
	FRACDIV:    "\\",
	MOD:        "%",
	PLUSPLUS:   "++",
	MINUSMINUS: "--",

	ASK:   "?",
	ASKOR: "??",

	AND: "&&",
	OR:  "||",
	NOT: "!",

	LAND: "&",
	LOR:  "|",
	XOR:  "~",

	EQ:  "==",
	NEQ: "!=",
	GT:  ">",
	LT:  "<",
	GE:  ">=",
	LE:  "<=",

	ASSIGN: "=",

	// punctuations
	SEMI: ";",
}
