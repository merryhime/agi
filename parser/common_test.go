package parser

import "github.com/MerryMage/agi/lexer"
import "bytes"
import t "testing"

func getParser(s string) *Parser {
	b := bytes.NewBufferString(s)
	l := lexer.MakeLexer(b, "<test>")
	p := Parser{}
	p.l = &l
	p.nextToken()
	return &p
}

func shouldPanic(t *t.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fail()
		}
	}()

	f()
}

func assert(t *t.T, b bool) {
	if !b {
		t.Fail()
	}
}

func TestIdentifier(t *t.T) {
	assert(t, getParser("a").parseIdentifier().Name == "a")
	assert(t, getParser("_x9").parseIdentifier().Name == "_x9")
	assert(t, getParser("ThisVariableIsExported").parseIdentifier().Name == "ThisVariableIsExported")
	assert(t, getParser("αβ").parseIdentifier().Name == "αβ")
	shouldPanic(t, func() { getParser("func").parseIdentifier() })
	shouldPanic(t, func() { getParser("32").parseIdentifier() })
}
