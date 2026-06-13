package main

import (
	"fmt"
	"strings"
	lexer "tclesius/kaleidoscope/src"
)

func main() {
	l := lexer.NewLexer(strings.NewReader("def foo(x) x + 1 # great work"))

	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF {
			break
		}

		fmt.Printf("%v %q\n", tok.Type, tok.Value)
	}
}
