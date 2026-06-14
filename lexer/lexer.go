package lexer

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type TokenType int

const (
	EOF TokenType = -1
	// commands
	DEF    = -2
	EXTERN = -3
	// primary
	IDENTIFIER = -4
	NUMBER     = -5
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	r    *bufio.Reader
	curr rune
	pos  int
	err  error
}

func NewLexer(input io.Reader) *Lexer {
	l := &Lexer{
		r:   bufio.NewReader(input),
		pos: -1,
	}

	l.advance()
	return l
}

func (l *Lexer) eof() bool {
	return l.err == io.EOF
}

func (l *Lexer) advance() {
	ch, _, err := l.r.ReadRune()
	l.curr = ch
	l.err = err

	if err == nil {
		l.pos++
	}
}

func (l *Lexer) consume() rune {
	ch := l.curr
	l.advance()
	return ch
}

func (l *Lexer) peek() rune {
	return l.curr
}

////

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.peek()) {
		l.consume()
	}
}

func (l *Lexer) skipComment() {
	l.consume() // #

	for !l.eof() && l.peek() != '\n' && l.peek() != '\r' {
		// Comment until EOL
		l.consume()
	}
}

func (l *Lexer) lexIdentifier() Token {
	var ident strings.Builder

	for unicode.IsLetter(l.peek()) {
		ident.WriteRune(l.consume())
	}

	if ident.String() == "def" {
		return Token{DEF, ident.String()}
	}
	if ident.String() == "extern" {
		return Token{EXTERN, ident.String()}
	}

	return Token{IDENTIFIER, ident.String()}
}

func (l *Lexer) lexNumeric() Token {
	var numeric strings.Builder

	for unicode.IsDigit(l.peek()) || l.peek() == '.' {
		numeric.WriteRune(l.consume())
	}

	return Token{NUMBER, numeric.String()}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if unicode.IsLetter(l.peek()) {
		return l.lexIdentifier()
	}

	if unicode.IsDigit(l.peek()) || l.peek() == '.' {
		return l.lexNumeric()
	}

	if l.peek() == '#' {
		l.skipComment()
	}

	if l.eof() {
		return Token{EOF, ""}
	}

	ch := l.consume()
	return Token{TokenType(ch), string(ch)}
}
