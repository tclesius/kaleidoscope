package main

import (
	"fmt"
	"tclesius/kaleidoscope/codegen"
	"tclesius/kaleidoscope/lexer"
	"tclesius/kaleidoscope/linker"
	"tclesius/kaleidoscope/parser"
	"time"

	"tinygo.org/x/go-llvm"
)

func main() {
	var err error
	var start time.Time
	var elapsed time.Duration

	start = time.Now()
	tokens := lexer.Lex("def main() if 0 then 3 else 0")
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
	module, err := codegen.Gen(program)
	if err != nil {
		panic(err)
	}
	elapsed = time.Since(start)
	fmt.Printf("Codegen took: %s\n", &elapsed)
	// fmt.Printf("----------\n")
	// fmt.Print(code.String())

	targetTriple := llvm.DefaultTargetTriple()

	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()

	target, err := llvm.GetTargetFromTriple(targetTriple)
	if err != nil {
		panic(err)
	}

	cpu := "generic"
	features := ""

	targetMachine := target.CreateTargetMachine(
		targetTriple,
		cpu,
		features,
		llvm.CodeGenLevelDefault,
		llvm.RelocDefault,
		llvm.CodeModelDefault,
	)
	targetData := targetMachine.CreateTargetData()
	defer targetData.Dispose()

	module.SetDataLayout(targetData.String())
	module.SetTarget(targetTriple)

	buf, err := targetMachine.EmitToMemoryBuffer(module, llvm.ObjectFile)
	if err != nil {
		panic(err)
	}
	defer buf.Dispose()

	start = time.Now()
	err = linker.LinkObject(buf.Bytes(), "output")
	if err != nil {
		panic(err)
	}
	elapsed = time.Since(start)
	fmt.Printf("Linker took: %s\n", &elapsed)

}
