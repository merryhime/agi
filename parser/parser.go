package parser

import "github.com/MerryMage/agi/lexer"

////////////////////////////////////////////////////////////////////////////////
// ASTNode

type ASTNode interface {
	Begin() lexer.Position
	End() lexer.Position
	_astNode()
}

////////////////////////////////////////////////////////////////////////////////
// Parser

type Parser struct {
	l     *lexer.Lexer
	t     lexer.Token // Current Token
	peekt lexer.Token // One Token Lookahead
}

////////////////////////////////////////////////////////////////////////////////
// Parse File

type File struct {
	PackageName string
	Imports     []Import
	Decls       []Decl
}

type Import struct {
	PackageNickname string
	ImportPath      string
}

/*
	SourceFile       = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .
*/
func (p *Parser) ParseFile() File {
	var f File

	/*
		PackageClause  = "package" PackageName .
		PackageName    = identifier .
	*/
	p.expect("a Go file must start with a 'package' declaration.", lexer.PackageKeyword)
	p.expect("expected a package name after 'package'", lexer.Identifier)
	f.PackageName = p.t.Payload.(string)
	p.expect("a package name is a single identifier", lexer.Semicolon)

	if p.maybe(lexer.EndOfFile) {
		return f
	}

	/*
		ImportDecl       = "import" ( ImportSpec | "(" { ImportSpec ";" } ")" ) .
		ImportSpec       = [ "." | PackageName ] ImportPath .
		ImportPath       = string_lit .
	*/
	for p.maybe(lexer.ImportKeyword) {
		parseImportSpec := func() {
			var i Import

			if p.maybe(lexer.Dot) {
				i.PackageNickname = "."
			} else if p.maybe(lexer.Identifier) {
				i.PackageNickname = p.t.Payload.(string)
			}

			p.expect("Malformed import statement", lexer.RawStringLiteral, lexer.InterpretedStringLiteral)
			i.ImportPath = p.t.Payload.(string)
			p.expect("One import statment per line please", lexer.Semicolon)

			f.Imports = append(f.Imports, i)
		}

		if p.maybe(lexer.LParen) {
			for p.maybe(lexer.RParen) {
				parseImportSpec()
			}
		} else {
			parseImportSpec()
		}
	}

	for !p.maybe(lexer.EndOfFile) {
		for p.maybe(lexer.Semicolon) {
			// empty
		}
		d := p.ParseTopLevel()
		f.Decls = append(f.Decls, d)
	}

	return f
}

////////////////////////////////////////////////////////////////////////////////
// Parse Top Level

/*
	Declaration   = ConstDecl | TypeDecl | VarDecl .
	TopLevelDecl  = Declaration | FunctionDecl | MethodDecl .
*/
func (p *Parser) ParseTopLevel() Decl {
	if p.peek(lexer.FuncKeyword) {
		return p.parseFuncOrMethodDecl()
	} else if p.peek(lexer.ConstKeyword) {
		panic("unimplemented")
		// return p.parseConstDecl()
	} else if p.peek(lexer.TypeKeyword) {
		panic("unimplemented")
		// return p.parseTypeDecl()
	} else if p.peek(lexer.VarKeyword) {
		panic("unimplemented")
		// return p.parseVarDecl()
	} else {
		panic("Did not expect *this* weirdness at toplevel")
	}
}
