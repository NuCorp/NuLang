package parser

import (
	"github.com/LicorneSharing/GTL/optional"
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

/*
different type of init:
- TYPE{ opt(repeat(NL)) EXPR opt(, repeat{, opt(repeat(NL))}(IDENT|*IDENT opt(: EXPR))) opt(repeat(NL)) }
- TYPE:IDENT{ARGS_BINDING}
- TYPE:{ opt(repeat(NL)) METHOD_DEF opt(repeat(NL)) }
- TYPE opt((ARGS_DEF) opt(TYPE)) => EXPR

METHOD_DEF:
- opt(const|set) IDENT ( ARGS_DEF ) opt(TYPE) METHOD_SCOPE
- opt(const|set) IDENT ( ARGS_DEF ) opt(TYPE) => EXPR

TYPE { --> 1
TYPE? --> 1
TYPE! --> 1
TYPE:IDENT --> 2
TYPE:{ --> 3
TYPE ( --> 4
TYPE => --> 4
*/

type simpleInit struct {
	expr           ParserOf[ast.Expr]
	simpleInitArgs listOf[bracesSurrounding, simpleInitArg]
}

type knownErrorContinuer[F, T any] struct {
	sharedScanner scan.SharedScanner
	errorMsg      string
	errorValue    T
}

func (e knownErrorContinuer[F, T]) ContinueParsing(_ F, s scan.Scanner, errors *Errors) T {
	if !e.sharedScanner.IsLinkedTo(s) {
		panic("should not use different scanner inside a same parsing session")
	}

	e.sharedScanner.ReSync()
	errors.Set(s.CurrentPos(), e.errorMsg)

	return e.errorValue
}

func (i initExpr) selectInit(s scan.SharedScanner) Continuer[ast.Type, ast.InitExpr] {
	type continuerInitAs[T any] = ContinuerCast[ast.Type, T, ast.InitExpr]

	switch s.ConsumeToken() {
	case tokens.COLON: // TYPE :
		if s.ConsumeToken() == tokens.IDENT { // TYPE : IDENT
			return nil // namedInit
		}

		// if it is not `TYPE : IDENT` it shall be: TYPE : { --> 3
		fallthrough
	case tokens.OPAREN, tokens.ARROW: // `TYPE (` or `TYPE =>` --> 4
		return continuerInitAs[ast.InterfaceInitExpr]{
			FromContinuer: i.interfaceInit,
		}
	case tokens.OBRAC, tokens.NOT, tokens.ASK: // TYPE { --> 1
		return nil // simple init
	default:
		return nil // no init matching
	}
}

type simpleInitArgParser struct {
	expr  ParserOf[ast.Expr]
	named ParserOf[ast.NamedArgBinding]

	asSelf bool
}

type simpleInitArg struct {
	fromAs        ast.Expr
	simpleInitArg *ast.SimpleInitArg
}

func (si *simpleInitArgParser) Parse(s scan.Scanner, errors *Errors) simpleInitArg {
	if !si.asSelf && s.CurrentToken() != tokens.STAR {
		si.asSelf = true
		return simpleInitArg{
			fromAs: si.expr.Parse(s, errors),
		}
	}

	if s.CurrentToken() == tokens.STAR {
		var (
			destructured bool

			arg = si.named.Parse(s, errors)
		)

		if s.CurrentToken() == tokens.ELLIPSIS {
			destructured = true
			s.ConsumeTokenInfo()
		}

		return simpleInitArg{
			simpleInitArg: &ast.SimpleInitArg{
				Name:         arg.Name,
				Value:        arg.Expr,
				Destructured: destructured,
			},
		}
	}

	const errorMsg = "should be IDENT (for bool value as true), !IDENT (bool value as false) or *IDENT (to set the argument by name)"

	if !s.CurrentToken().IsOneOf(tokens.IDENT, tokens.NOT) {
		errors.Set(s.CurrentPos(), errorMsg)

		skipTo(s, tokens.COMMA, tokens.CBRAC, tokens.SEMI)
		return simpleInitArg{}
	}

	arg := ast.SimpleInitArg{
		Bool: optional.Some(s.CurrentToken() != tokens.NOT),
	}

	if tok := s.ConsumeTokenInfo(); tok.Token() == tokens.IDENT {
		arg.Name = ast.DotIdent{tok.Value().(string)}
	} else if tok := s.ConsumeTokenInfo(); tok.Token() == tokens.IDENT {
		arg.Name = ast.DotIdent{tok.Value().(string)}
	} else {
		errors.Set(s.CurrentPos(), errorMsg)
		skipTo(s, tokens.COMMA, tokens.CBRAC, tokens.SEMI)
	}

	return simpleInitArg{simpleInitArg: &arg}
}

/*
| {}
| {EXPR}
| {EXPR, Left}

Left:
| ø
| NameArgBinding
| NameArgBindingDestructured
| IDENT
*/

func (i simpleInit) ContinueParsing(from ast.Type, s scan.Scanner, errors *Errors) ast.SimpleInitExpr {
	assert(s.CurrentToken().IsOneOf(tokens.OBRAC, tokens.NOT, tokens.ASK))

	init := ast.SimpleInitExpr{Type: from}

	switch s.CurrentToken() {
	case tokens.NOT:
		init.MayThrow = ast.MustThrow
		s.CurrentTokenInfo()
	case tokens.ASK:
		init.MayThrow = ast.MayThrow
		s.ConsumeTokenInfo()
	default:
	}

	if s.CurrentToken() != tokens.OBRAC {
		errors.Set(s.CurrentPos(), "expected '{' to start an init expression")
		return init
	}

	initArgs := i.simpleInitArgs.Parse(s, errors)

	if len(initArgs) == 0 {
		return init
	}

	if initArgs[0].fromAs != nil {
		init.FromAs.Set(initArgs[0].fromAs)
		initArgs = initArgs[1:]
	}

	if len(initArgs) > 0 {
		init.Args = make(map[string]ast.SimpleInitArg, len(initArgs))
	}

	for _, initArg := range initArgs {
		init.Args[initArg.simpleInitArg.Name.Pack()] = *initArg.simpleInitArg
	}

	return init

	s.ConsumeTokenInfo()

	if s.CurrentToken() != tokens.STAR {
		init.FromAs.Set(i.expr.Parse(s, errors))
	}

	if s.CurrentToken() == tokens.CBRAC {
		s.ConsumeTokenInfo()
		return init
	}

	switch {
	case init.FromAs.HasValue() && s.CurrentToken() != tokens.COMMA:
		errors.Set(s.CurrentPos(), "expected ',' to continue the init expression")
		skipTo(s, tokens.COMMA)
		return init
	case init.FromAs.HasValue():
		s.ConsumeTokenInfo()
	default:
	}

	// TODO: it can be a bool argument too !
	if s.CurrentToken() != tokens.STAR {
		errors.Set(s.CurrentPos(), "init argument must be named")
		skipTo(s, tokens.CBRAC)
		return init
	}

	s.ConsumeTokenInfo()
	ident := s.ConsumeTokenInfo()

	if ident.Token() != tokens.IDENT {
		errors.Set(s.CurrentPos(), "init argument must be named")
		skipTo(s, tokens.CBRAC)
		return init
	}

	if _, exists := init.Args[ident.Value().(string)]; exists {
		errors.Set(s.CurrentPos(), "init argument already exists")
		// continue
	}

	if s.CurrentToken() != tokens.COLON {

	}

	return ast.SimpleInitExpr{}
}

type initExpr struct {
	interfaceInit Continuer[ast.Type, ast.InterfaceInitExpr]
	simpleInit    Continuer[ast.Type, ast.SimpleInitExpr]
	expr          ParserOf[ast.Expr]
}

func (i initExpr) isInterfaceInit(scanner scan.SharedScanner) bool {
	if scanner.ConsumeToken().IsOneOf(tokens.CONST, tokens.SET) {
		// TYPE { const Method ...
		// or
		// Type { set Method ...
		return true
	}

	if scanner.ConsumeToken() != tokens.IDENT {
		return false
	}

	// TYPE { Method ...

	if scanner.CurrentToken() == tokens.OPAREN {
		// Type { Method() ...
		return true
	}

	if scanner.ConsumeToken() != tokens.ASK {
		return false
	}

	// Type { Method? ...

	return scanner.CurrentToken() == tokens.OPAREN // Type { Method?() ...
}

/*func (i initExpr) ContinueParsing(from ast.Type, s scan.Scanner, errors *Errors) ast.InitExpr {
	assert(
		s.CurrentToken().IsOneOf(tokens.COLON, tokens.OBRAC, tokens.ARROW),
		"expected `:`, `{` or =>, but got %v", s.CurrentToken(),
	)

	if from.AsType() == "type:interface" {
		return i.interfaceInit.ContinueParsing(from, s, errors)
	}

	if from.AsType() == "type:named" && s.CurrentToken() == tokens.ARROW {
		return i.interfaceInit.ContinueParsing(from, s, errors)
	}

	if s.CurrentToken() == tokens.OBRAC && from.AsType() == "type:named" && i.isInterfaceInit(s.Clone()) {
		return i.interfaceInit.ContinueParsing(from, s, errors)
	}

	// classic init
	assert(s.CurrentToken().IsOneOf(tokens.COLON, tokens.OBRAC))

	var (
		init       = ast.SimpleInitExpr{Type: from}
		tmpScanner = s.Clone()
	)

	if tmpScanner.ConsumeToken() == tokens.COLON {
		if tmpScanner.ConsumeToken() != tokens.IDENT {
			errors.Set(tmpScanner.CurrentPos(), "expected an identifier after `TYPE:` in order to make a named init expr")
			skipToEOI(s)
			return init
		}

		// init.Named.Set(tmpScanner.ConsumeTokenInfo().Value().(string))
		tmpScanner.ReSync()
	}

	if s.CurrentToken() != tokens.OBRAC {
		errors.Set(s.CurrentPos(), "expected `{` to start init expr")
		skipToEOI(s)
		return init
	}

	s.ConsumeTokenInfo()

	// if init.Named? then we can make an innerArgumentParsing
	// if not, we must see if the token is `*` => innerArgumentParsing
	// finally if it is an expr, we parse the expr and then check that token is `*` => innerArgumentParsing

	return init
}
*/
