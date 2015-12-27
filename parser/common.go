package parser

import "github.com/MerryMage/agi/lexer"

////////////////////////////////////////////////////////////////////////////////
// Token handling

func (p *Parser) nextToken() {
	p.t = p.peekt
again:
	p.peekt = p.l.NextToken()
	if p.peekt.IsComment() {
		goto again
	} else if p.peekt.Type == lexer.ElidedSemicolon {
		p.peekt.Type = lexer.Semicolon
	} else if p.peekt.Type == lexer.EndOfLine {
		goto again
	}
}

// Will an expect(tt) succeed?
func (p *Parser) peek(tt lexer.TokenType) bool {
	if p.peekt.Type == tt {
		return true
	}
	return false
}

// If peek(tt), p.t is advanced to that token. (In other words, p.t.Type == tt)
// Otherwise, don't do anything.
func (p *Parser) maybe(tt lexer.TokenType) bool {
	if p.peekt.Type == tt {
		p.nextToken()
		return true
	}
	return false
}

// If peek(tt), return false, advance token to that token.
// If not, return true, advance token to that failing token.
// Intended for error handling more complex than panicing.
func (p *Parser) isnot(tt lexer.TokenType) bool {
	p.nextToken()
	if p.t.Type == tt {
		return false
	}
	return true
}

// If peek(tt), advance token to that token.
// Otherwise, panic.
func (p *Parser) expect(panicstr string, tts ...lexer.TokenType) {
	for _, tt := range tts {
		if tt == p.peekt.Type {
			p.nextToken()
			return
		}
	}
	panic(panicstr)
}

func (p *Parser) expectSemicolon(panicstr string) {
	if p.peekt.Type == lexer.RParen || p.peekt.Type == lexer.RBrace {
		// Semicolon is optional before ) or }
		p.t = p.peekt // Duplicate the token! (Pretend it's a semicolon the first time round)
		p.t.Type = lexer.Semicolon
		return
	}
	// Otherwise expect a semicolon like a normal person would
	p.expect(panicstr, lexer.Semicolon)
}

////////////////////////////////////////////////////////////////////////////////
// Identifier

type Identifier struct {
	pos  lexer.Position
	Name string
}

func (i Identifier) Begin() lexer.Position { return i.pos }
func (i Identifier) End() lexer.Position   { return i.pos.Move(len(i.Name)) }
func (i Identifier) _astNode()             {}

func (p Parser) parseIdentifier() Identifier {
	p.expect("ICE", lexer.Identifier)
	return Identifier{p.t.Position, p.t.Payload.(string)}
}

////////////////////////////////////////////////////////////////////////////////
// Type References
//   Named TypeRefs to distingush the syntatical names in the source code from
//   the Type representations produced during semantic anaylsis.

type TypeRef interface {
	ASTNode
	_typeRef()
}

// We are referring to a named type. Either:
// Identifier | (PackageName "." Identifier)
type NamedTypeRef struct {
	Package *Identifier // If this is non-nil, this is a qualified type name
	Name    Identifier
}

func (o NamedTypeRef) Begin() lexer.Position {
	if o.Package != nil {
		return o.Package.Begin()
	} else {
		return o.Name.Begin()
	}
}
func (o NamedTypeRef) End() lexer.Position { return o.Name.End() }
func (o NamedTypeRef) _astNode()           {}
func (o NamedTypeRef) _typeRef()           {}

// "[" Expr "]" ElemType
type ArrayTypeRef struct {
	begin    lexer.Position
	Length   Expr
	ElemType TypeRef
}

func (o ArrayTypeRef) Begin() lexer.Position { return o.begin }
func (o ArrayTypeRef) End() lexer.Position   { return o.ElemType.End() }
func (o ArrayTypeRef) _astNode()             {}
func (o ArrayTypeRef) _typeRef()             {}

// "[" "]" ElemType
type SliceTypeRef struct {
	begin    lexer.Position
	ElemType TypeRef
}

func (o SliceTypeRef) Begin() lexer.Position { return o.begin }
func (o SliceTypeRef) End() lexer.Position   { return o.ElemType.End() }
func (o SliceTypeRef) _astNode()             {}
func (o SliceTypeRef) _typeRef()             {}

// "[" "..." "]" ElemType
type ArrayEllipsesTypeRef struct {
	begin    lexer.Position
	ElemType TypeRef
}

func (o ArrayEllipsesTypeRef) Begin() lexer.Position { return o.begin }
func (o ArrayEllipsesTypeRef) End() lexer.Position   { return o.ElemType.End() }
func (o ArrayEllipsesTypeRef) _astNode()             {}
func (o ArrayEllipsesTypeRef) _typeRef()             {}

type StructTypeRef struct {
	begin  lexer.Position
	Name   Identifier
	Fields StructTypeRefFields
}

func (o StructTypeRef) Begin() lexer.Position { return o.begin }
func (o StructTypeRef) End() lexer.Position   { return o.Fields.End() }
func (o StructTypeRef) _astNode()             {}
func (o StructTypeRef) _typeRef()             {}

type StructTypeRefFields struct {
	begin  lexer.Position
	Fields []StructTypeRefField
	end    lexer.Position
}

func (o StructTypeRefFields) End() lexer.Position { return o.end }

type StructTypeRefField struct {
	Name *Identifier // If not present, anonymous
	Type TypeRef
	Tag  *lexer.Token // If not present, no tag
}

type PointerTypeRef struct {
	begin    lexer.Position
	BaseType TypeRef
}

type FunctionTypeRef struct {
	begin     lexer.Position
	Signature FunctionSignature
}

func (p *Parser) parseTypeRef() TypeRef {
	r, _ := p.parseTypeRefOrIdent()
	return r
}
func (p *Parser) parseTypeRefOrIdent() (TypeRef, bool) {
	panic("unimplemented")
}

////////////////////////////////////////////////////////////////////////////////
// FunctionSignature
//   This is an annoying part of the Golang grammar. It's workable but annoying.

type FunctionSignature struct {
	Args   ParameterDeclList
	Return *ParameterDeclList
}

func (o FunctionSignature) Begin() lexer.Position { return o.Args.Begin() }
func (o FunctionSignature) End() lexer.Position {
	if o.Return != nil {
		return o.Return.End()
	} else {
		return o.Args.End()
	}
}
func (o FunctionSignature) _astNode() {}

func (p *Parser) parseFunctionSignature(mustNotElideParamNames bool) FunctionSignature {
	panic("unimplemented")
}

type ParameterDeclList struct {
	begin lexer.Position
	Decls []ParameterDecl
	end   lexer.Position
}

func (o ParameterDeclList) Begin() lexer.Position { return o.begin }
func (o ParameterDeclList) End() lexer.Position   { return o.end }
func (o ParameterDeclList) _astNode()             {}

type ParameterDecl struct {
	Name          *Identifier
	TypeWasElided bool
	Type          TypeRef
}

func (o ParameterDecl) Begin() lexer.Position { return o.Name.Begin() }
func (o ParameterDecl) End() lexer.Position   { return o.Type.End() }
func (o ParameterDecl) _astNode()             {}

func (p *Parser) parseParameterDeclList(mustNotElideParamNames bool) ParameterDeclList {
	extractIdent := func(t TypeRef) *Identifier {
		i, ok := t.(NamedTypeRef)
		if !ok || i.Package != nil {
			panic("This isn't an identifier")
		}
		return &(i.Name)
	}

	hasTwo := false

	var dl ParameterDeclList

	p.expect("ICE", lexer.LParen)
	dl.begin = p.t.Position
	for {
		var d ParameterDecl

		t1, maybeIdent := p.parseTypeRefOrIdent()
		if maybeIdent && !p.peek(lexer.Comma) {
			hasTwo = true
			t2 := p.parseTypeRef()
			d.Type = t2
			d.Name = extractIdent(t1)
		} else {
			d.Type = t1
		}

		dl.Decls = append(dl.Decls, d)

		if !p.maybe(lexer.Comma) {
			break
		}
	}
	p.expect("Unexpected thing in parameter list", lexer.RParen)
	dl.end = p.t.Position.Move(1)

	if len(dl.Decls) == 0 {
		return dl
	}

	if mustNotElideParamNames && !hasTwo {
		panic("must not elide param names in this context")
	}

	if hasTwo {
		var t TypeRef = dl.Decls[len(dl.Decls)-1].Type
		for i := int(len(dl.Decls)) - 2; i >= 0; i-- {
			if dl.Decls[i].Name == nil {
				dl.Decls[i].Name = extractIdent(dl.Decls[i].Type)
				dl.Decls[i].TypeWasElided = true
				dl.Decls[i].Type = t
			} else {
				t = dl.Decls[i].Type
			}
		}
	}

	return dl
}
