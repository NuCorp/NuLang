package tokens

type Token int

const EOF = Token(-1)

const (
	NoInit = Token(iota)
	ERR    = Token(iota)
	IDENT
	NO_IDENT // "_"
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

	PKG    // package,
	IMPORT // import

	TYPE      // type
	STRUCT    // struct
	INTERFACE // interface
	ENUM      // enum
	EXTENSION // extension
	CAST      // cast
	AS        // as
	IS        // is
	EXPLICIT  // explicit
	IMPLICIT  // implicit
	DELETE    // delete
	NEW       // new
	INIT      // init
	OPERATOR  // operator
	WITH      // with
	ALIAS     // alias
	TYPEOF

	VAR   // var
	CONST // const
	FUNC  // func
	DEFER // defer
	RUN   // run
	CHAN  // chan

	IF       // if
	THEN     // then
	ELSE     // else
	FOR      // for
	IN       // in
	TO       // to
	WHILE    // while
	DO       // do
	CASE     // case
	BREAK    // break
	CONTINUE // continue
	DEFAULT  // default

	TRY    // try
	CATCH  // catch
	THROW  // throw
	NIL    // nil
	RETURN // return

	keywordEnd

	operatorStart
	PLUS_PLUS   // "++",
	MINUS_MINUS // "--",

	PLUS
	MINUS
	TIME
	DIV
	FRAC_DIV
	MOD

	ASK   // "?",
	ASKOR // "??",

	NOT // "!"

	LAND   // "&",
	LOR    // "|",
	XOR    // "~"
	LSHIFT // "<<",
	RSHIFT // ">>",

	AND // "&&",
	OR  // "||",
	EQ  // "=="
	NEQ // "!="
	GT  // ">"
	LT  // "<"
	GE  // ">="
	LE  // "<="

	ASSIGN       // "="
	DEFINE       // ":="
	PLUS_ASSIGN  // "+="
	MINUS_ASSIGN // "-="
	TIME_ASSIGN  // "*="
	DIV_ASSIGN   // "/="
	MOD_ASSIGN   // "%="
	LAND_ASSIGN  // "&="
	LOR_ASSIGN   // "|="
	XOR_ASSIGN   // "~="
	AND_ASSIGN   // "&&="
	OR_ASSIGN    // "||="

	operatorEnd

	punctuationStart

	SEMI   //;
	COLON  // ":"
	COMA   // ",",
	DOT    // ".",
	PERIOD // "...",

	IMPL   // "=>",
	RARROW // "->",
	LARROW // "<-",

	OBRAC  // "{",
	OBRAK  // "[",
	OPAREN // "(",
	CBRAC  // "}",
	CBRAK  // "]",
	CPAREN // ")",
	punctuationEnd

	NL // new line
)

func (t Token) String() string {
	if str, found := tokenStr[t]; found {
		return str
	}
	return "❏"
}
func (t Token) IsLiteral() bool     { return literalStart < t && t < literalEnd || t == TRUE || t == FALSE }
func (t Token) IsKeyword() bool     { return keywordStart < t && t < keywordEnd }
func (t Token) IsOperator() bool    { return operatorStart < t && t < operatorEnd }
func (t Token) IsPunctuation() bool { return punctuationStart < t && t < punctuationEnd }
func (t Token) IsAssignation() bool {
	return t >= ASSIGN && t <= OR_ASSIGN
}
func (t Token) IsEoI() bool { return t == NL || t == SEMI }
func EoI() []Token          { return []Token{NL, SEMI} }

func ForEach(forFunction func(token Token)) {
	for _, token := range strKeyword {
		forFunction(token)
	}
}

var strKeyword = func() map[string]Token {
	ret := map[string]Token{}
	for tok := keywordStart + 1; tok < keywordEnd; tok++ {
		ret[tok.String()] = tok
	}
	return ret
}()

func GetKeywordForText(text string) Token {
	tok, found := strKeyword[text]
	if !found {
		tok = IDENT
	}
	return tok
}

var tokenStr = map[Token]string{
	ERR: "TokenError",

	IDENT:    "Identifier",
	NO_IDENT: "_",

	// literals
	INT:      "INT",
	FLOAT:    "FLOAT",
	FRACTION: "FRACTION",
	CHAR:     "CHAR",
	STR:      "STR",

	// keywords
	TRUE:  "true",
	FALSE: "false",

	PKG:    "package",
	IMPORT: "import",

	TYPE:      "type",
	STRUCT:    "struct",
	INTERFACE: "interface",
	ENUM:      "enum",
	EXTENSION: "extension",
	CAST:      "cast",
	AS:        "as",
	IS:        "is",
	EXPLICIT:  "explicit",
	IMPLICIT:  "implicit",
	DELETE:    "delete",
	NEW:       "new",
	INIT:      "init",
	OPERATOR:  "operator",
	WITH:      "with",
	ALIAS:     "alias",
	TYPEOF:    "typeof",

	VAR:   "var",
	CONST: "const",
	FUNC:  "func",
	DEFER: "defer",
	RUN:   "run",
	CHAN:  "chan",

	IF:       "if",
	THEN:     "then",
	ELSE:     "else",
	FOR:      "for",
	IN:       "in",
	TO:       "to",
	WHILE:    "while",
	DO:       "do",
	CASE:     "case",
	BREAK:    "break",
	CONTINUE: "continue",
	DEFAULT:  "default",

	TRY:    "try",
	CATCH:  "catch",
	THROW:  "throw",
	NIL:    "nil",
	RETURN: "return",

	// operators
	PLUS:        "+",
	MINUS:       "-",
	TIME:        "*",
	DIV:         "/",
	FRAC_DIV:    "\\",
	MOD:         "%",
	PLUS_PLUS:   "++",
	MINUS_MINUS: "--",

	ASK:   "?",
	ASKOR: "??",

	AND: "&&",
	OR:  "||",
	NOT: "!",

	LAND:   "&",
	LOR:    "|",
	XOR:    "~",
	LSHIFT: "<<",
	RSHIFT: ">>",

	EQ:  "==",
	NEQ: "!=",
	GT:  ">",
	LT:  "<",
	GE:  ">=",
	LE:  "<=",

	ASSIGN:       "=",
	DEFINE:       ":=",
	PLUS_ASSIGN:  "+=",
	MINUS_ASSIGN: "-=",
	TIME_ASSIGN:  "*=",
	DIV_ASSIGN:   "/=",
	MOD_ASSIGN:   "%=",
	LAND_ASSIGN:  "&=",
	LOR_ASSIGN:   "|=",
	XOR_ASSIGN:   "~=",
	AND_ASSIGN:   "&&=",
	OR_ASSIGN:    "||=",

	// punctuations
	SEMI:   ";",
	COLON:  ":",
	COMA:   ",",
	DOT:    ".",
	PERIOD: "...",

	IMPL:   "=>",
	RARROW: "->",
	LARROW: "<-",

	OBRAC:  "{",
	OBRAK:  "[",
	OPAREN: "(",
	CBRAC:  "}",
	CBRAK:  "]",
	CPAREN: ")",

	NL: "\n",
}
