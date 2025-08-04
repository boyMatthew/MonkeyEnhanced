package monkey_lexer

import token "myMonkey/monkey_token"

type Lexer struct {
	src          string
	pos, readPos int
	ch           byte
}

func NewLexer(src string) *Lexer {
	l := &Lexer{src: src}
	l.readCh()
	return l
}

func (l *Lexer) readCh() {
	if l.readPos >= len(l.src) {
		l.ch = 0
	} else {
		l.ch = l.src[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) readTwoCh() string {
	ch := l.ch
	l.readCh()
	return string(ch) + string(l.ch)
}

func (l *Lexer) peekCh() byte {
	if l.readPos >= len(l.src) {
		return 0
	} else {
		return l.src[l.readPos]
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	var True = true
	l.skipEmpty()
	switch l.ch {
	case '=':
		if l.peekCh() == '=' {
			tok = token.Token{Type: token.EQ, Literal: l.readTwoCh()}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		if l.peekCh() == '+' {
			tok = token.Token{Type: token.BUMPPLUS, Literal: l.readTwoCh()}
		} else {
			tok = newToken(token.PLUS, l.ch)
		}
	case '-':
		if l.peekCh() == '-' {
			tok = token.Token{Type: token.BUMPMINUS, Literal: l.readTwoCh()}
		} else {
			tok = newToken(token.MINUS, l.ch)
		}
	case '*':
		tok = newToken(token.MULTIPLY, l.ch)
	case '/':
		tok = newToken(token.DIVIDE, l.ch)
	case '!':
		if l.peekCh() == '=' {
			tok = token.Token{Type: token.NEQ, Literal: l.readTwoCh()}
		} else {
			tok = newToken(token.REVERSE, l.ch)
		}
	case '<':
		if l.peekCh() == '=' {
			tok = token.Token{Type: token.LE, Literal: l.readTwoCh()}
		} else if l.peekCh() == '<' {
			tok = token.Token{Type: token.LSHIFT, Literal: l.readTwoCh()}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekCh() == '=' {
			tok = token.Token{Type: token.GE, Literal: l.readTwoCh()}
		} else if l.peekCh() == '>' {
			tok = token.Token{Type: token.RSHIFT, Literal: l.readTwoCh()}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			ident := l.readIdent()
			tok = token.Token{Type: token.LookupKeyword(ident), Literal: ident}
			return tok
		} else if isNumber(l.ch, &True) {
			return token.Token{Type: token.NUMBER, Literal: l.readNum()}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readCh()
	return tok
}

func newToken(tType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tType, Literal: string(ch)}
}

func (l *Lexer) readIdent() string {
	pos := l.pos
	for isLetter(l.ch) {
		l.readCh()
	}
	return l.src[pos:l.pos]
}

func (l *Lexer) readNum() string {
	pos := l.pos
	l.readCh()
	for hasFloatingPoint := false; isNumber(l.ch, &hasFloatingPoint); l.readCh() {
	}
	return l.src[pos:l.pos]
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_' || ch == '$'
}

func isNumber(ch byte, hasFloatingPoint *bool) bool {
	if ch == '.' && !*hasFloatingPoint {
		*hasFloatingPoint = true
		return true
	}
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipEmpty() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readCh()
	}
}
