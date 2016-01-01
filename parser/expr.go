package parser

import "github.com/MerryMage/agi/lexer"

type Expr struct{}

type Block struct{}

func (o Block) End() lexer.Position { panic("unimplemented") }

func (p *Parser) parseBlock() Block { panic("unimplemented") }

func (p *Parser) parseExpr() Expr { p.nextToken(); return Expr{} }
