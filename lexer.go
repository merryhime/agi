package main

import "io"
import "unicode"
import "fmt"
import "math/big"

type Lexer struct {
	// character stream
	f   io.ByteReader
	pos Position
	ch  rune

	// token state
	canElideSemicolon bool
	t                 Token // current token
}

func MakeLexer(f io.ByteReader, fname string) Lexer {
	l := Lexer{}
	l.f = f
	l.pos.Filename = fname
	l.nextch()
	return l
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
	l.t.SourceCode += string(l.ch)

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
func hexdigitvalue(ch rune) rune {
	switch {
	case '0' <= ch && ch <= '9':
		return ch - '0'
	case 'a' <= ch && ch <= 'f':
		return ch - 'a' + 10
	case 'A' <= ch && ch <= 'F':
		return ch - 'A' + 10
	default:
		panic("ICE")
	}
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

func (l *Lexer) maybech(r rune) bool {
	if l.ch == r {
		l.nextch()
		return true
	}
	return false
}

func (l *Lexer) expectch(r rune, panicstr string) {
	if l.ch != r {
		panic(panicstr)
	}
	l.nextch()
}

func (l *Lexer) lexidentifierorkeyword() {
	for isletter(l.ch) || isdigit(l.ch) {
		l.nextch()
	}
	if keywordMap[l.t.SourceCode] != 0 {
		l.t.Type = keywordMap[l.t.SourceCode]
	} else {
		l.t.Type = Identifier
	}
}

/* Reference: https://golang.org/ref/spec#Integer_literals
int_lit     = decimal_lit | octal_lit | hex_lit .
decimal_lit = ( "1" … "9" ) { decimal_digit } .
octal_lit   = "0" { octal_digit } .
hex_lit     = "0" ( "x" | "X" ) hex_digit { hex_digit } .
*/
func (l *Lexer) lexnumerical(ch rune) {
	switch ch {
	case '0':
		if l.maybech('x') || l.maybech('X') {
			// Hex
			l.t.Type = HexIntegerLiteral
			for ishexdigit(l.ch) {
				l.nextch()
			}
		} else if l.maybech('b') || l.maybech('B') {
			// Binary
			l.t.Type = BinaryIntegerLiteral
			for l.ch == '0' || l.ch == '1' {
				l.nextch()
			}
		} else {
			// Octal
			l.t.Type = OctalIntegerLiteral
			for isoctaldigit(l.ch) {
				l.nextch()
			}
		}
		value, success := (&big.Int{}).SetString(l.t.SourceCode, 0)
		if value == nil || !success {
			panic("ICE: Could not parse verified integer literal")
		}
		l.t.Payload = value
	case '.':
		l.lexfloat(ch)
	default:
		for isdecimaldigit(l.ch) {
			l.nextch()
		}
		if l.maybech('.') {
			l.lexfloat('.')
		} else if l.maybech('e') || l.maybech('E') {
			l.lexfloat('e')
		} else {
			// Decimal
			l.t.Type = DecimalIntegerLiteral
			value, success := (&big.Int{}).SetString(l.t.SourceCode, 0)
			if value == nil || !success {
				panic("ICE: Could not parse verified integer literal")
			}
			l.t.Payload = value
		}
	}
}

/* Reference: https://golang.org/ref/spec#Floating-point_literals
float_lit = decimals "." [ decimals ] [ exponent ] |
            decimals exponent |
            "." decimals [ exponent ] .
decimals  = decimal_digit { decimal_digit } .
exponent  = ( "e" | "E" ) [ "+" | "-" ] decimals .
*/
func (l *Lexer) lexfloat(ch rune) {
	switch ch {
	case '.':
		if l.t.SourceCode == "" && !isdecimaldigit(l.ch) {
			panic("expecting at least one decimal digit")
		}
		for isdecimaldigit(l.ch) {
			l.nextch()
		}
		if !l.maybech('e') || !l.maybech('E') {
			break
		}
		fallthrough
	case 'e':
		if !l.maybech('-') {
			l.maybech('+')
		}
		if !isdecimaldigit(l.ch) {
			panic("expecting at least one decimal digit")
		}
		for isdecimaldigit(l.ch) {
			l.nextch()
		}
	}
	l.t.Type = FloatLiteral
	value, success := (&big.Rat{}).SetString(l.t.SourceCode)
	if value == nil || !success {
		panic("ICE: Could not parse verified float literal")
	}
	l.t.Payload = value
}

/* Reference: https://golang.org/ref/spec#Rune_literals
rune_lit         = "'" ( unicode_value | byte_value ) "'" .
unicode_value    = unicode_char | little_u_value | big_u_value | escaped_char .
byte_value       = octal_byte_value | hex_byte_value .
octal_byte_value = `\` octal_digit octal_digit octal_digit .
hex_byte_value   = `\` "x" hex_digit hex_digit .
little_u_value   = `\` "u" hex_digit hex_digit hex_digit hex_digit .
big_u_value      = `\` "U" hex_digit hex_digit hex_digit hex_digit
                           hex_digit hex_digit hex_digit hex_digit .
escaped_char     = `\` ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | `\` | "'" | `"` ) .

This func implements ( unicode_value | byte_value ).
*/
func (l *Lexer) lexsingletransch() rune {
	hexdigits := func(n int) rune {
		var value rune
		for i := 0; i < n; i++ {
			ch := l.ch
			l.nextch()
			if !ishexdigit(ch) {
				panic("Too few hex digits")
			}
			value <<= 4
			value |= hexdigitvalue(ch)
		}
		if value > 0x10FFFF || value < 0 {
			panic("invalid unicode value")
		} else if 0xD800 <= value && value <= 0xDFFF {
			panic("illegal: surrogate halves are not allowed")
		}
		return value
	}

	ch := l.ch
	l.nextch()
	switch ch {
	case '\\':
		ch := l.ch
		l.nextch()
		switch ch {
		case 'x':
			return hexdigits(2)
		case 'u':
			return hexdigits(4)
		case 'U':
			return hexdigits(8)
		case '0', '1', '2', '3', '4', '5', '6', '7':
			ch2 := l.ch
			l.nextch()
			ch3 := l.ch
			l.nextch()
			if !isoctaldigit(ch2) || !isoctaldigit(ch3) {
				panic("too few octal digits")
			}
			return (ch-'0')*64 + (ch2-'0')*8 + (ch3-'0')*1
		case 'a':
			return 0x0007
		case 'b':
			return 0x0008
		case 'f':
			return 0x000C
		case 'n':
			return 0x000A
		case 'r':
			return 0x000D
		case 't':
			return 0x0009
		case 'v':
			return 0x000B
		case '\\':
			return 0x005C
		case '\'':
			return 0x0027
		case '"':
			return 0x0022
		}
	case '\n':
		panic("Newline in the middle of a string/rune constant")
	default:
		return ch
	}

	panic("ICE: Unreachable")
}

func (l *Lexer) lexchar() {
	l.t.Type = RuneLiteral
	var value rune = l.lexsingletransch()
	l.expectch('\'', "Expected '; only single character rune literals are allowed")
	l.t.Payload = value
}

func (l *Lexer) lextranslatedstr() {
	l.t.Type = InterpretedStringLiteral
	var value string
	for l.ch != '"' {
		value += string(l.lexsingletransch())
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
		hasNewline := false
		l.t.Type = BlockComment
		l.nextch()
		for {
			for l.ch != '*' {
				if l.ch == '\n' {
					hasNewline = true
				}
				l.nextch()
			}
			l.nextch()
			if l.ch == '/' {
				break
			}
		}
		if hasNewline {
			l.ch = '\n'
		} else {
			l.nextch()
		}
	}
}
