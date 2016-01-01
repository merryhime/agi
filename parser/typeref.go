package parser

import "github.com/MerryMage/agi/lexer"

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
func (o NamedTypeRef) End() lexer.Position     { return o.Name.End() }
func (o NamedTypeRef) _astNode()               {}
func (o NamedTypeRef) _typeRef()               {}
func (o NamedTypeRef) _interfaceTypeRefField() {}

func extractIdent(o TypeRef) Identifier {
	i, ok := o.(NamedTypeRef)
	if !ok || i.Package != nil {
		panic("This isn't an identifier")
	}
	return i.Name
}
func extractIdentPtr(o TypeRef) *Identifier {
	i := extractIdent(o)
	return &i
}

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
	Fields []StructTypeRefField
	end    lexer.Position
}

func (o StructTypeRef) Begin() lexer.Position { return o.begin }
func (o StructTypeRef) End() lexer.Position   { return o.end }
func (o StructTypeRef) _astNode()             {}
func (o StructTypeRef) _typeRef()             {}

type StructTypeRefField struct {
	Names *[]Identifier // If not present, anonymous
	Type  TypeRef
	Tag   *lexer.Token // If not present, no tag
}

func (o StructTypeRefField) Begin() lexer.Position {
	if o.Names == nil {
		return o.Type.Begin()
	} else {
		return (*o.Names)[0].Begin()
	}
}
func (o StructTypeRefField) End() lexer.Position { return o.Type.End() }
func (o StructTypeRefField) _astNode()           {}

type PointerTypeRef struct {
	begin    lexer.Position
	BaseType TypeRef
}

func (o PointerTypeRef) Begin() lexer.Position { return o.begin }
func (o PointerTypeRef) End() lexer.Position   { return o.BaseType.End() }
func (o PointerTypeRef) _astNode()             {}
func (o PointerTypeRef) _typeRef()             {}

type FunctionTypeRef struct {
	begin     lexer.Position
	Signature FunctionSignature
}

func (o FunctionTypeRef) Begin() lexer.Position { return o.begin }
func (o FunctionTypeRef) End() lexer.Position   { return o.Signature.End() }
func (o FunctionTypeRef) _astNode()             {}
func (o FunctionTypeRef) _typeRef()             {}

type InterfaceTypeRef struct {
	begin  lexer.Position
	Fields []ASTNode // either InterfaceMethodSpec or NamedTypeRef
	end    lexer.Position
}

func (o InterfaceTypeRef) Begin() lexer.Position { return o.begin }
func (o InterfaceTypeRef) End() lexer.Position   { return o.end }
func (o InterfaceTypeRef) _astNode()             {}
func (o InterfaceTypeRef) _typeRef()             {}

type InterfaceTypeRefField interface {
	ASTNode
	_interfaceTypeRefField()
}

type InterfaceMethodSpec struct {
	MethodName Identifier
	Signature  FunctionSignature
}

func (o InterfaceMethodSpec) Begin() lexer.Position   { return o.MethodName.Begin() }
func (o InterfaceMethodSpec) End() lexer.Position     { return o.Signature.End() }
func (o InterfaceMethodSpec) _astNode()               {}
func (o InterfaceMethodSpec) _interfaceTypeRefField() {}

type MapTypeRef struct {
	begin     lexer.Position
	KeyType   TypeRef
	ValueType TypeRef
}

func (o MapTypeRef) Begin() lexer.Position { return o.begin }
func (o MapTypeRef) End() lexer.Position   { return o.ValueType.End() }
func (o MapTypeRef) _astNode()             {}
func (o MapTypeRef) _typeRef()             {}

type ChanDir int

const (
	ChanSend     ChanDir = 1
	ChanRecv             = 2
	ChanSendRecv         = ChanSend | ChanRecv
)

type ChanTypeRef struct {
	begin lexer.Position
	Dir   ChanDir
	Inner TypeRef
}

func (o ChanTypeRef) IsSend() bool { return o.Dir|ChanSend == ChanSend }
func (o ChanTypeRef) IsRecv() bool { return o.Dir|ChanRecv == ChanRecv }

func (o ChanTypeRef) Begin() lexer.Position { return o.begin }
func (o ChanTypeRef) End() lexer.Position   { return o.Inner.End() }
func (o ChanTypeRef) _astNode()             {}
func (o ChanTypeRef) _typeRef()             {}

// Parser
func (p *Parser) parseTypeRefOrIdent() (TypeRef, bool) {
	r := p.parseTypeRef()
	if ntr, ok := r.(NamedTypeRef); ok {
		if ntr.Package == nil {
			return r, true
		}
	}
	return r, false
}
func (p *Parser) parseTypeRef() TypeRef {
	r := p.maybeParseTypeRef()
	if r == nil {
		panic("Expected TypeRef (Or Ident)")
	}
	return r
}
func (p *Parser) maybeParseTypeRef() TypeRef {
	begin := p.peekt.Position
	switch {
	case p.maybe(lexer.LParen): // <(> TypeRef <)>
		tr := p.parseTypeRef()
		p.expect("Expected close bracket to match this one", lexer.RParen)
		return tr
	case p.peek(lexer.Identifier): // Ident | Ident <.> Ident
		first := p.parseIdentifier()
		if p.maybe(lexer.Dot) {
			second := p.parseIdentifier()
			return NamedTypeRef{Package: &first, Name: second}
		} else {
			return NamedTypeRef{Name: first}
		}
	case p.maybe(lexer.LBracket): // <[> <]> TypeRef | <[> <...> <]> TypeRef | <[> Expr <]> TypeRef
		switch {
		case p.maybe(lexer.RBracket): // <[> <]> TypeRef
			inner := p.parseTypeRef()
			return SliceTypeRef{begin: begin, ElemType: inner}
		case p.maybe(lexer.EllipsisOp): // <[> <...> <]> TypeRef
			p.expect("[...] expected", lexer.RBracket)
			inner := p.parseTypeRef()
			return ArrayEllipsesTypeRef{begin: begin, ElemType: inner}
		default: // <[> Expr <]> TypeRef
			expr := p.parseExpr()
			p.expect("<[> <expr> <]> expected", lexer.RBracket)
			inner := p.parseTypeRef()
			return ArrayTypeRef{begin: begin, Length: expr, ElemType: inner}
		}
	case p.maybe(lexer.StructKeyword): // <struct> <{> { FieldDecl <;> } <}>
		p.expect("{ must occur after a struct keyword", lexer.LBrace)
		ret := StructTypeRef{begin: begin}
		ret.begin = p.t.Position
		for !p.peek(lexer.RBrace) {
			// FieldDecl = (IdentifierList TypeRef | TypeRef) [Tag]
			field := StructTypeRefField{}

			first, first_ident := p.parseTypeRefOrIdent()

			if first_ident && !(p.peek(lexer.Semicolon) || p.peek(lexer.RBrace) || p.peek(lexer.RawStringLiteral) || p.peek(lexer.InterpretedStringLiteral)) {
				// FieldDecl = IdentifierList TypeRef [Tag]
				names := []Identifier{extractIdent(first)}
				for p.maybe(lexer.Comma) {
					names = append(names, p.parseIdentifier())
				}
				field.Names = &names
				field.Type = p.parseTypeRef()
			} else {
				// FieldDecl = TypeRef [Tag]
				field.Type = first
			}

			if p.maybe(lexer.RawStringLiteral) || p.maybe(lexer.InterpretedStringLiteral) {
				field.Tag = &(p.t)
			}

			ret.Fields = append(ret.Fields, field)

			p.expectSemicolon("no semicolon?")
		}
		p.expect("expected }", lexer.RBrace)
		ret.end = p.t.Position
		return ret
	case p.maybe(lexer.MulOp): // <*> TypeRef
		inner := p.parseTypeRef()
		return PointerTypeRef{begin: begin, BaseType: inner}
	case p.maybe(lexer.FuncKeyword): // <func> FunctionSignature
		sig := p.parseFunctionSignature(false)
		return FunctionTypeRef{begin: begin, Signature: sig}
	case p.maybe(lexer.InterfaceKeyword): // <interface> <{> { MethodSpec <;> } <}>
		p.expect("{ expected", lexer.LBrace)
		ret := InterfaceTypeRef{begin: begin}
		for !p.peek(lexer.RBrace) {
			var field InterfaceTypeRefField

			// MethodSpec = MethodName FunctionSignature | TypeName
			name := p.parseIdentifier()
			if p.maybe(lexer.Dot) { // MethodSpec = TypeName = QualifiedName = Identifier <.> Identifier
				name2 := p.parseIdentifier()
				field = NamedTypeRef{Package: &name, Name: name2}
			} else if p.peek(lexer.LParen) { // MethodSpec = MethodName FunctionSignature
				sig := p.parseFunctionSignature(false)
				field = InterfaceMethodSpec{MethodName: name, Signature: sig}
			} else { // MethodSpec = TypeName = Identifier
				field = NamedTypeRef{Name: name}
			}

			ret.Fields = append(ret.Fields, field)

			p.expectSemicolon("expect ; at end of methodspec")
		}
		p.expect("expected }", lexer.RBrace)
		ret.end = p.t.Position
		return ret
	case p.maybe(lexer.MapKeyword): // <map> <[> TypeRef <]> TypeRef
		p.expect("[ expected", lexer.LBracket)
		first := p.parseTypeRef()
		p.expect("] expected", lexer.RBracket)
		second := p.parseTypeRef()
		return MapTypeRef{begin, first, second}
	case p.maybe(lexer.ChanKeyword): // <chan> [<<->] TypeRef
		if p.maybe(lexer.ChanOpOp) {
			inner := p.parseTypeRef()
			return ChanTypeRef{begin, ChanSend, inner}
		} else {
			inner := p.parseTypeRef()
			return ChanTypeRef{begin, ChanSendRecv, inner}
		}
	case p.maybe(lexer.ChanOpOp): // <<-> <chan> TypeRef
		p.expect("expected <-chan here", lexer.ChanKeyword)
		inner := p.parseTypeRef()
		return ChanTypeRef{begin, ChanRecv, inner}
	}
	return nil
}
