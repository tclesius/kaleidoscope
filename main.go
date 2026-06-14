package main

import (
	"fmt"
	"os"
	"tclesius/kaleidoscope/lexer"
	"tclesius/kaleidoscope/parser"
)

func handleDefinition(p *parser.Parser) {
	if _, err := p.ParseDefinition(); err == nil {
		fmt.Fprintln(os.Stderr, "Parsed a function definition.")
	} else {
		fmt.Fprintln(os.Stderr, err)
		p.Advance()
	}
}

func handleExtern(p *parser.Parser) {
	if _, err := p.ParseExtern(); err == nil {
		fmt.Fprintln(os.Stderr, "Parsed an extern")
	} else {
		fmt.Fprintln(os.Stderr, err)
		p.Advance()
	}
}

func handleTopLevelExpression(p *parser.Parser) {
	if _, err := p.ParseTopLevelExpr(); err == nil {
		fmt.Fprintln(os.Stderr, "Parsed a top-level expr")
	} else {
		fmt.Fprintln(os.Stderr, err)
		p.Advance()
	}
}

func mainLoop(p *parser.Parser) {
	for {
		fmt.Fprint(os.Stderr, "ready> ")

		switch p.Peek().Type {
		case lexer.EOF:
			return
		case lexer.TokenType(';'):
			p.Advance()
		case lexer.DEF:
			handleDefinition(p)
		case lexer.EXTERN:
			handleExtern(p)
		default:
			handleTopLevelExpression(p)
		}
	}
}

func main() {
	l := lexer.NewLexer(os.Stdin)
	p := parser.NewParser(l)
	mainLoop(p)
}
