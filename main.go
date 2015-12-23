package main

import "os"
import "io"
import "unicode"
import "fmt"
import "bufio"

type Position struct {
	Filename string
	Line     int
	Column   int
}

type TokenType int

func (tt TokenType) String() string {
	return tokenToStringMap[tt]
}

// Reference: https://golang.org/ref/spec#Lexical_elements
const (
	EndOfFile TokenType = iota

	// Comments
	commentMin

	LineComment
	BlockComment

	commentMax
	// Identifiers

	Identifier

	// Keywords
	keywordMin

	BreakKeyword
	CaseKeyword
	ChanKeyword
	ConstKeyword
	ContinueKeyword
	DefaultKeyword
	DeferKeyword
	ElseKeyword
	FallthroughKeyword
	ForKeyword
	FuncKeyword
	GoKeyword
	GotoKeyword
	IfKeyword
	ImportKeyword
	InterfaceKeyword
	MapKeyword
	PackageKeyword
	RangeKeyword
	ReturnKeyword
	SelectKeyword
	StructKeyword
	SwitchKeyword
	TypeKeyword
	VarKeyword

	keywordMax
	// Operators
	opMin

	AddOp // +
	SubOp // -
	MulOp // *
	DivOp // /
	ModOp // %

	BitAndOp   // &
	BitOrrOp   // |
	BitXorOp   // ^
	ShlOp      // <<
	ShrOp      // >>
	BitClearOp // &^

	assignopMin
	DefineOp         // :=
	AssignOp         // =
	AddAssignOp      // +=
	SubAssignOp      // -=
	MulAssignOp      // *=
	DivAssignOp      // /=
	ModAssignOp      // %=
	BitAndAssignOp   // &=
	BitOrrAssignOp   // |=
	BitXorAssignOp   // ^=
	ShlAssignOp      // <<=
	ShrAssignOp      // >>=
	BitClearAssignOp // &^=
	assignOpMax

	LogicNotOp // !
	LogicAndOp // &&
	LogicOrrOp // ||
	LtOp       // <
	GtOp       // >
	EqOp       // ==
	NeqOp      // !=
	LteOp      // <=
	GteOp      // >=

	ChanOpOp    // <-
	IncrementOp // ++
	DecrementOp // --

	EllipsisOp // ...

	opMax
	// Delimiters
	delimMin

	EndOfLine
	ElidedSemicolon
	Semicolon

	LParen
	LBracket
	LBrace
	RParen
	RBracket
	RBrace

	Comma
	Dot
	Colon

	delimMax
	// Literals
	literalMin

	DecimalIntegerLiteral
	OctalIntegerLiteral
	HexIntegerLiteral

	FloatLiteral

	ImaginaryLiteral

	RuneLiteral

	InterpretedStringLiteral
	RawStringLiteral

	literalMax
)

var keywordMap = map[string]TokenType{
	"break":       BreakKeyword,
	"case":        CaseKeyword,
	"chan":        ChanKeyword,
	"const":       ConstKeyword,
	"continue":    ContinueKeyword,
	"default":     DefaultKeyword,
	"defer":       DeferKeyword,
	"else":        ElseKeyword,
	"fallthrough": FallthroughKeyword,
	"for":         ForKeyword,
	"func":        FuncKeyword,
	"go":          GoKeyword,
	"goto":        GotoKeyword,
	"if":          IfKeyword,
	"import":      ImportKeyword,
	"interface":   InterfaceKeyword,
	"map":         MapKeyword,
	"package":     PackageKeyword,
	"range":       RangeKeyword,
	"return":      ReturnKeyword,
	"select":      SelectKeyword,
	"struct":      StructKeyword,
	"switch":      SwitchKeyword,
	"type":        TypeKeyword,
	"var":         VarKeyword,
}

var tokenToStringMap = map[TokenType]string{
	EndOfFile:          "EOF",
	LineComment:        "// ...",
	BlockComment:       "/* ... */",
	Identifier:         "identifier",
	BreakKeyword:       "Break",
	CaseKeyword:        "Case",
	ChanKeyword:        "Chan",
	ConstKeyword:       "Const",
	ContinueKeyword:    "Continue",
	DefaultKeyword:     "Default",
	DeferKeyword:       "Defer",
	ElseKeyword:        "Else",
	FallthroughKeyword: "Fallthrough",
	ForKeyword:         "For",
	FuncKeyword:        "Func",
	GoKeyword:          "Go",
	GotoKeyword:        "Goto",
	IfKeyword:          "If",
	ImportKeyword:      "Import",
	InterfaceKeyword:   "Interface",
	MapKeyword:         "Map",
	PackageKeyword:     "Package",
	RangeKeyword:       "Range",
	ReturnKeyword:      "Return",
	SelectKeyword:      "Select",
	StructKeyword:      "Struct",
	SwitchKeyword:      "Switch",
	TypeKeyword:        "Type",
	VarKeyword:         "Var",
	AddOp:              "+",
	SubOp:              "-",
	MulOp:              "*",
	DivOp:              "/",
	ModOp:              "%",
	BitAndOp:           "&",
	BitOrrOp:           "|",
	BitXorOp:           "^",
	ShlOp:              "<<",
	ShrOp:              ">>",
	BitClearOp:         "&^",
	DefineOp:           ":=",
	AssignOp:           "=",
	AddAssignOp:        "+=",
	SubAssignOp:        "-=",
	MulAssignOp:        "*=",
	DivAssignOp:        "/=",
	ModAssignOp:        "%=",
	BitAndAssignOp:     "&=",
	BitOrrAssignOp:     "|=",
	BitXorAssignOp:     "^=",
	ShlAssignOp:        "<<=",
	ShrAssignOp:        ">>=",
	BitClearAssignOp:   "&^=",
	LogicNotOp:         "!",
	LogicAndOp:         "&&",
	LogicOrrOp:         "||",
	LtOp:               "<",
	GtOp:               ">",
	EqOp:               "==",
	NeqOp:              "!=",
	LteOp:              "<=",
	GteOp:              ">=",
	ChanOpOp:           "<-",
	IncrementOp:        "++",
	DecrementOp:        "--",
	EllipsisOp:         "...",
	EndOfLine:          "EOL",
	ElidedSemicolon:    "[;]",
	Semicolon:          ";",
	LParen:             "(",
	LBracket:           "[",
	LBrace:             "{",
	RParen:             ")",
	RBracket:           "]",
	RBrace:             "}",
	Comma:              ",",
	Dot:                ".",
	Colon:              ":",
	DecimalIntegerLiteral:    "DecimalIntegerLiteral",
	OctalIntegerLiteral:      "OctalIntegerLiteral",
	HexIntegerLiteral:        "HexIntegerLiteral",
	FloatLiteral:             "FloatLiteral",
	ImaginaryLiteral:         "ImaginaryLiteral",
	RuneLiteral:              "RuneLiteral",
	InterpretedStringLiteral: "InterpretedStringLiteral",
	RawStringLiteral:         "RawStringLiteral",
}

type Token struct {
	Position Position
	Type     TokenType
	Payload  string
}

func MakeLexer(f io.ByteReader, fname string) Lexer {
	l := Lexer{}
	l.f = f
	l.pos.Filename = fname
	l.nextch()
	return l
}

type Lexer struct {
	// character stream
	f   io.ByteReader
	pos Position
	ch  rune

	// token state
	canElideSemicolon bool
	t                 Token // current token
}

func (l *Lexer) nextbyte() uint8 {
	b, err := l.f.ReadByte()
	if err != nil {
		return 0
	}
	return b
}

// Get next unicode codepoint (decodes UTF-8)
// Reference: https://golang.org/ref/spec#Source_code_representation
func (l *Lexer) nextch() {
	l.t.Payload += string(l.ch)

	b := l.nextbyte()

	var ret rune

	// Unicode continuation byte : 0b10xxxxxx
	nextContByte := func() {
		ret <<= 6
		b = l.nextbyte()
		if b&0xC0 != 0x80 {
			panic("Invalid UTF-8")
		}
		ret |= rune(b & 0x3F)
	}

	switch {
	case b&0x80 == 0: // One byte sequence      : 0b0xxxxxxx
		ret = rune(b)
	case b&0xE0 == 0xC0: // Two byte sequence   : 0b110xxxxx
		ret = rune(b & 0x1F)
		nextContByte()
	case b&0xF0 == 0xE0: // Three byte sequence : 0b1110xxxx
		ret = rune(b & 0x0F)
		nextContByte()
		nextContByte()
	case b&0xF8 == 0xF0: // Four byte sequence  : 0b11110xxx
		ret = rune(b & 0x07)
		nextContByte()
		nextContByte()
		nextContByte()
	default:
		panic("Invalid UTF-8")
	}

	if l.ch == 0x000A {
		l.pos.Line++
		l.pos.Column = 1
		l.ch = ret
	} else {
		l.pos.Column++
		l.ch = ret
	}
}

// Reference: https://golang.org/ref/spec#Letters_and_digits
func isletter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}
func isdigit(r rune) bool {
	return unicode.IsDigit(r)
}
func isdecimaldigit(r rune) bool {
	return '0' <= r && r <= '9'
}
func isoctaldigit(r rune) bool {
	return '0' <= r && r <= '7'
}
func ishexdigit(r rune) bool {
	return isdecimaldigit(r) || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F')
}

// Reference: https://golang.org/ref/spec#Tokens
func iswhitespace(r rune) bool {
	switch r {
	case 0x0020, 0x0009, 0x000D:
		return true
	default:
		return false
	}
}
func (l *Lexer) skipwhitespace() {
	for iswhitespace(l.ch) {
		l.nextch()
	}
}

func (l *Lexer) NextToken() Token {
	// Deal with whitespace and elided semicolons
	l.skipwhitespace()
	if l.ch == '\n' {
		l.t = Token{}
		l.t.Position = l.pos
		l.nextch() // skip over it

		if l.canElideSemicolon {
			l.t.Type = ElidedSemicolon
			l.canElideSemicolon = false
		} else {
			l.t.Type = EndOfLine
		}

		return l.t
	} else if l.ch == 0 {
		l.t = Token{}
		l.t.Position = l.pos

		if l.canElideSemicolon {
			l.t.Type = ElidedSemicolon
			l.canElideSemicolon = false
		} else {
			l.t.Type = EndOfFile
		}

		return l.t
	}

	// Deal with other things
	l.t = Token{}
	l.t.Position = l.pos
	l.canElideSemicolon = false

	ch := l.ch
	l.nextch()

	switch ch {
	case '"':
		l.lextranslatedstr()
	case '\'':
		l.lexchar()
	case '`':
		l.lexrawstr()
	case ':':
		l.t.Type = Colon
		if l.maybech('=') {
			l.t.Type = DefineOp
		}
	case '.':
		if isdigit(l.ch) {
			l.lexnumerical(ch)
		} else {
			l.t.Type = Dot
			if l.maybech('.') {
				if l.maybech('.') {
					l.t.Type = EllipsisOp
				} else {
					panic("Not sure why you have \"..\" in your code dude")
				}
			}
		}
	case ',':
		l.t.Type = Comma
	case ';':
		l.t.Type = Semicolon
	case '(':
		l.t.Type = LParen
	case '[':
		l.t.Type = LBracket
	case '{':
		l.t.Type = LBrace
	case ')':
		l.t.Type = RParen
	case ']':
		l.t.Type = RBracket
	case '}':
		l.t.Type = RBrace
	case '+':
		l.t.Type = AddOp
		if l.maybech('=') {
			l.t.Type = AddAssignOp
		} else if l.maybech('+') {
			l.t.Type = IncrementOp
		}
	case '-':
		l.t.Type = SubOp
		if l.maybech('=') {
			l.t.Type = SubAssignOp
		} else if l.maybech('-') {
			l.t.Type = DecrementOp
		}
	case '*':
		l.t.Type = MulOp
		if l.maybech('=') {
			l.t.Type = MulAssignOp
		}
	case '/':
		l.t.Type = DivOp
		if l.maybech('=') {
			l.t.Type = DivAssignOp
		} else {
			if l.ch == '/' || l.ch == '*' {
				l.lexcomment()
			}
		}
	case '%':
		l.t.Type = ModOp
		if l.maybech('=') {
			l.t.Type = ModAssignOp
		}
	case '^':
		l.t.Type = BitXorOp
		if l.maybech('=') {
			l.t.Type = BitXorAssignOp
		}
	case '<':
		l.t.Type = LtOp
		if l.maybech('-') {
			l.t.Type = ChanOpOp
		} else if l.maybech('=') {
			l.t.Type = LteOp
		} else if l.maybech('<') {
			l.t.Type = ShlOp
			if l.maybech('=') {
				l.t.Type = ShlAssignOp
			}
		}
	case '>':
		l.t.Type = GtOp
		if l.maybech('=') {
			l.t.Type = GteOp
		} else if l.maybech('>') {
			l.t.Type = ShrOp
			if l.maybech('=') {
				l.t.Type = ShrAssignOp
			}
		}
	case '=':
		l.t.Type = AssignOp
		if l.maybech('=') {
			l.t.Type = EqOp
		}
	case '!':
		l.t.Type = LogicNotOp
		if l.maybech('=') {
			l.t.Type = NeqOp
		}
	case '&':
		l.t.Type = BitAndOp
		if l.maybech('^') {
			l.t.Type = BitClearOp
			if l.maybech('=') {
				l.t.Type = BitClearAssignOp
			}
		} else if l.maybech('=') {
			l.t.Type = BitAndAssignOp
		} else if l.maybech('&') {
			l.t.Type = LogicAndOp
		}
	case '|':
		l.t.Type = BitOrrOp
		if l.maybech('=') {
			l.t.Type = BitOrrAssignOp
		} else if l.maybech('|') {
			l.t.Type = LogicOrrOp
		}
	default:
		switch {
		case isletter(ch):
			l.lexidentifierorkeyword()
		case isdigit(ch):
			l.lexnumerical(ch)
		default:
			fmt.Printf("%x %c\n", ch, ch)
			panic("Illegal character")
		}
	}

	switch l.t.Type {
	case Identifier,
		HexIntegerLiteral, OctalIntegerLiteral, DecimalIntegerLiteral,
		FloatLiteral,
		ImaginaryLiteral,
		RuneLiteral,
		RawStringLiteral, InterpretedStringLiteral,
		BreakKeyword, ContinueKeyword, FallthroughKeyword, ReturnKeyword,
		IncrementOp, DecrementOp,
		RParen, RBracket, RBrace:
		l.canElideSemicolon = true
	}

	return l.t
}

func (l *Lexer) maybech(r rune) bool {
	if l.ch == r {
		l.nextch()
		return true
	}
	return false
}

func (l *Lexer) lexidentifierorkeyword() {
	for isletter(l.ch) || isdigit(l.ch) {
		l.nextch()
	}
	if keywordMap[l.t.Payload] != 0 {
		l.t.Type = keywordMap[l.t.Payload]
	} else {
		l.t.Type = Identifier
	}
}

func (l *Lexer) lexnumerical(ch rune) {
	switch ch {
	case '0':
		l.nextch()
	case '.':
		l.nextch()
	default:
		l.nextch()
	}
	l.t.Type = DecimalIntegerLiteral
}

func (l *Lexer) lextranslatedstr() {
	l.t.Type = InterpretedStringLiteral
	for l.ch != '"' {
		if l.ch == '\\' {
			l.nextch()
		}
		l.nextch()
	}
	l.nextch()
}
func (l *Lexer) lexchar() {
	l.t.Type = RuneLiteral
	for l.ch != '\'' {
		if l.ch == '\\' {
			l.nextch()
		}
		l.nextch()
	}
	l.nextch()
}
func (l *Lexer) lexrawstr() {
	l.t.Type = RawStringLiteral
	for l.ch != '`' {
		l.nextch()
	}
	l.nextch()
}
func (l *Lexer) lexcomment() {
	if l.ch == '/' {
		l.t.Type = LineComment
		for l.ch != '\n' {
			l.nextch()
		}
		// !! Do not consume newline
	} else {
		l.t.Type = BlockComment
		l.nextch()
		for {
			for l.ch != '*' {
				l.nextch()
			}
			l.nextch()
			if l.ch == '/' {
				break
			}
		}
		l.nextch()
	}
}

func main() {
	f, err := os.Open("./main.go")
	if err != nil {
		panic("Couldn't open main.go")
	}
	l := MakeLexer(bufio.NewReader(f), "main.go")
	for {
		t := l.NextToken()
		fmt.Printf("%v\n", t)
		if t.Type == EndOfFile {
			break
		}
	}
}
