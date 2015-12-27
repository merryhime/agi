package parser

import "github.com/MerryMage/agi/lexer"

type Decl interface {
	ASTNode
	_decl()
}

////////////////////////////////////////////////////////////////////////////////
// Parse Function or Method Decl

type FuncOrMethodDecl struct {
	begin        lexer.Position
	Receiver     *ParameterDeclList // If this exists, it's a method. Otherwise, it's a function.
	FunctionName Identifier
	Signature    FunctionSignature
	Body         *Block
}

func (o FuncOrMethodDecl) Begin() lexer.Position { return o.begin }
func (o FuncOrMethodDecl) End() lexer.Position {
	if o.Body != nil {
		return o.Body.End()
	} else {
		return o.Signature.End()
	}
}
func (o FuncOrMethodDecl) _astNode() {}
func (o FuncOrMethodDecl) _decl()    {}

/*
	FunctionDecl = "func" FunctionName ( Function | Signature ) .
	FunctionName = identifier .
	Function     = Signature FunctionBody .
	FunctionBody = Block .
*/
/*
	MethodDecl   = "func" Receiver MethodName ( Function | Signature ) .
	Receiver     = Parameters .
*/

func (p *Parser) parseFuncOrMethodDecl() FuncOrMethodDecl {
	var d FuncOrMethodDecl

	p.expect("ICE", lexer.FuncKeyword)
	d.begin = p.t.Position

	if p.maybe(lexer.LParen) {
		// A method
		dl := p.parseParameterDeclList(true)
		d.Receiver = &dl
		if len(d.Receiver.Decls) > 1 {
			panic("More than one receiver")
		} else if len(d.Receiver.Decls) == 0 {
			panic("I'm not sure what you're up to here")
		}
	} else {
		// A function
		d.Receiver = nil
	}

	p.expect("a function name was expected here", lexer.Identifier)
	d.FunctionName = p.parseIdentifier()

	d.Signature = p.parseFunctionSignature(true)

	if p.maybe(lexer.LBrace) {
		// A function with a body
		b := p.parseBlock()
		d.Body = &b
	} else if p.maybe(lexer.Semicolon) {
		// An external function reference
		d.Body = nil
	} else {
		panic("Expected a function body here")
	}

	return d
}
