package parser

import "github.com/MerryMage/agi/lexer"
import "bytes"
import t "testing"

func dump(t *t.T, o interface{}) {
	t.Logf("%#v\n", o)
}

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
		t.FailNow()
	}
}

func TestIdentifier(t *t.T) {
	assert(t, getParser("a").parseIdentifier().Name == "a")
	assert(t, getParser("_x9").parseIdentifier().Name == "_x9")
	assert(t, getParser("ThisVariableIsExported").parseIdentifier().Name == "ThisVariableIsExported")
	assert(t, getParser("αβ").parseIdentifier().Name == "αβ")
	shouldPanic(t, func() { getParser("func").parseIdentifier() })
	shouldPanic(t, func() { getParser("32").parseIdentifier() })

	p := getParser("α β _")
	assert(t, p.parseIdentifier().Name == "α")
	dump(t, p)
	assert(t, p.parseIdentifier().Name == "β")
	dump(t, p)
	assert(t, p.parseIdentifier().Name == "_")
	dump(t, p)
}

func TestTypeRef(t *t.T) {
	// assert(t, getParser("").parseTypeRef().(). == "")

	assert(t, getParser("a.b").parseTypeRef().(NamedTypeRef).Package.Name == "a")
	assert(t, getParser("a.b").parseTypeRef().(NamedTypeRef).Name.Name == "b")

	assert(t, getParser("[32]byte").parseTypeRef().(ArrayTypeRef).ElemType.(NamedTypeRef).Package == nil)
	assert(t, getParser("[32]byte").parseTypeRef().(ArrayTypeRef).ElemType.(NamedTypeRef).Name.Name == "byte")
	// assert(t, getParser("[32]byte").parseTypeRef().(ArrayTypeRef).Length)

	assert(t, getParser("[2][2][2]float32").parseTypeRef().(ArrayTypeRef).ElemType.(ArrayTypeRef).ElemType.(ArrayTypeRef).ElemType.(NamedTypeRef).Name.Name == "float32")

	assert(t, getParser("[]int").parseTypeRef().(SliceTypeRef).ElemType.(NamedTypeRef).Name.Name == "int")

	assert(t, len(getParser("struct{}").parseTypeRef().(StructTypeRef).Fields) == 0)

	s := getParser("struct { x, y int; u float32; F func() ; _ float32 }").parseTypeRef().(StructTypeRef)
	assert(t, len(s.Fields) == 4)
	assert(t, len(*s.Fields[0].Names) == 2)
	assert(t, len(s.Fields[2].Type.(FunctionTypeRef).Signature.Args.Decls) == 0)
	assert(t, s.Fields[2].Type.(FunctionTypeRef).Signature.Return == nil)

	assert(t, getParser("<-chan int").parseTypeRef().(ChanTypeRef).Dir == ChanRecv)
	shouldPanic(t, func() { getParser("<-chan<- int").parseTypeRef() })

	assert(t, getParser("<-chan<-chan int").parseTypeRef().(ChanTypeRef).Inner.(ChanTypeRef).Dir == ChanRecv)
	assert(t, getParser("chan<-chan<- int").parseTypeRef().(ChanTypeRef).Inner.(ChanTypeRef).Dir == ChanSend)

	i := getParser(`
interface {
	Read(b Buffer) bool
	Write(b Buffer) bool
	Close()
}
	`).parseTypeRef().(InterfaceTypeRef)
	assert(t, len(i.Fields) == 3)
	assert(t, i.Fields[0].(InterfaceMethodSpec).MethodName.Name == "Read")
	assert(t, i.Fields[1].(InterfaceMethodSpec).MethodName.Name == "Write")
	assert(t, i.Fields[2].(InterfaceMethodSpec).MethodName.Name == "Close")
	assert(t, len(i.Fields[0].(InterfaceMethodSpec).Signature.Args.Decls) == 1)
	assert(t, len(i.Fields[1].(InterfaceMethodSpec).Signature.Args.Decls) == 1)
	assert(t, len(i.Fields[2].(InterfaceMethodSpec).Signature.Args.Decls) == 0)
	assert(t, i.Fields[0].(InterfaceMethodSpec).Signature.Args.Decls[0].Type.(NamedTypeRef).Name.Name == "Buffer")
	assert(t, i.Fields[1].(InterfaceMethodSpec).Signature.Args.Decls[0].Name.Name == "b")
	assert(t, i.Fields[1].(InterfaceMethodSpec).Signature.Return.Decls[0].Type.(NamedTypeRef).Name.Name == "bool")
	assert(t, i.Fields[2].(InterfaceMethodSpec).Signature.Return == nil)
}
