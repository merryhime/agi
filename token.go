package main

type Token struct {
	Position   Position
	Type       TokenType
	SourceCode string
	Payload    interface{}
}

type Position struct {
	Filename string
	Line     int
	Column   int
}

type TokenType int

func (tt TokenType) String() string {
	return tokenToStringMap[tt]
}

func (t Token) IsComment() bool {
	return commentMin < t.Type && t.Type < commentMax
}

func (t Token) IsKeyword() bool {
	return keywordMin < t.Type && t.Type < keywordMax
}

func (t Token) IsOp() bool {
	return opMin < t.Type && t.Type < opMax
}

func (t Token) IsAssignOp() bool {
	return assignopMin < t.Type && t.Type < assignopMax
}

func (t Token) IsDelim() bool {
	return delimMin < t.Type && t.Type < delimMax
}

func (t Token) IsLiteral() bool {
	return literalMin < t.Type && t.Type < literalMax
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
	assignopMax

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
	BinaryIntegerLiteral

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
	BreakKeyword:       "break",
	CaseKeyword:        "case",
	ChanKeyword:        "chan",
	ConstKeyword:       "const",
	ContinueKeyword:    "continue",
	DefaultKeyword:     "default",
	DeferKeyword:       "defer",
	ElseKeyword:        "else",
	FallthroughKeyword: "fallthrough",
	ForKeyword:         "for",
	FuncKeyword:        "func",
	GoKeyword:          "go",
	GotoKeyword:        "goto",
	IfKeyword:          "if",
	ImportKeyword:      "import",
	InterfaceKeyword:   "interface",
	MapKeyword:         "map",
	PackageKeyword:     "package",
	RangeKeyword:       "range",
	ReturnKeyword:      "return",
	SelectKeyword:      "select",
	StructKeyword:      "struct",
	SwitchKeyword:      "switch",
	TypeKeyword:        "type",
	VarKeyword:         "var",
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
	EndOfLine:          "\\n",
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
	BinaryIntegerLiteral:     "BinaryIntegerLiteral",
	FloatLiteral:             "FloatLiteral",
	ImaginaryLiteral:         "ImaginaryLiteral",
	RuneLiteral:              "RuneLiteral",
	InterpretedStringLiteral: "InterpretedStringLiteral",
	RawStringLiteral:         "RawStringLiteral",
}
