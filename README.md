# Kaleidoscope 

This is a simple toy language build soely to learn LLVM.

- [Introduction and the Lexer](https://llvm.org/docs/tutorial/MyFirstLanguageFrontend/LangImpl01.html)
- [Implementing a Parser and AST](https://llvm.org/docs/tutorial/MyFirstLanguageFrontend/LangImpl02.html)
- [Code generation to LLVM IR](https://llvm.org/docs/tutorial/MyFirstLanguageFrontend/LangImpl03.html)
- [Extending the Language: Control Flow](https://llvm.org/docs/tutorial/MyFirstLanguageFrontend/LangImpl05.html)

## Dependencies
- LLVM 20
- Clang
- Go 1.26

## Examples

`fibonacci`
```sh
go run . -s examples/fib.k -o build/fib.o
clang build/fib.o examples/lib.c -o build/fib
./build/fib
```

`putchar ffi`
```sh
go run . -s examples/std.k -o build/std.o
clang build/std.o -o build/std
./build/std
```
