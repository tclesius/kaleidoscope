package main

import (
	"fmt"
	"tclesius/kaleidoscope/codegen"
	"tclesius/kaleidoscope/lexer"
	"tclesius/kaleidoscope/parser"
	"time"
)

func main() {
	var start time.Time
	var elapsed time.Duration

	start = time.Now()
	tokens := lexer.Lex("4+5")
	elapsed = time.Since(start)
	fmt.Printf("Lexer took: %s\n", &elapsed)

	start = time.Now()
	program, err := parser.Parse(tokens)
	if err != nil {
		panic(err)
	}
	elapsed = time.Since(start)
	fmt.Printf("Parser took: %s\n", &elapsed)

	start = time.Now()
	code, err := codegen.Gen(program)
	if err != nil {
		panic(err)
	}
	elapsed = time.Since(start)
	fmt.Printf("Codegen took: %s\n", &elapsed)
	fmt.Printf("----------\n")
	fmt.Print(code.String())
}
