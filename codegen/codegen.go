package codegen

import (
	"fmt"
	"tclesius/kaleidoscope/parser"

	"tinygo.org/x/go-llvm"
)

type CodeGen struct {
	ctx     llvm.Context
	builder llvm.Builder
	module  llvm.Module
	values  map[string]llvm.Value
	parser  parser.Parser
}

func (cg *CodeGen) codegenExpr(expr parser.Expr) (llvm.Value, error) {
	switch e := expr.(type) {
	case parser.NumberExpr:
		return cg.codegenNumberExpr(e), nil
	case parser.VariableExpr:
		return cg.codegenVariableExpr(e)
	case parser.BinaryExpr:
		return cg.codegenBinaryExpr(e)
	case parser.CallExpr:
		return cg.codegenCallExpr(e)
	default:
		return llvm.Value{}, fmt.Errorf("Unknown expr type")
	}
}

func (cg *CodeGen) codegenNumberExpr(expr parser.NumberExpr) llvm.Value {
	return llvm.ConstFloat(cg.ctx.DoubleType(), expr.Val)
}

func (cg *CodeGen) codegenVariableExpr(expr parser.VariableExpr) (llvm.Value, error) {
	val, ok := cg.values[expr.Name]
	if !ok {
		return llvm.Value{}, fmt.Errorf("Unknown variable `%q`", expr.Name)
	}
	return val, nil
}

func (cg *CodeGen) codegenBinaryExpr(e parser.BinaryExpr) (llvm.Value, error) {
	l, err := cg.codegenExpr(e.Lhs)
	if err != nil {
		return llvm.Value{}, err
	}

	r, err := cg.codegenExpr(e.Rhs)
	if err != nil {
		return llvm.Value{}, err
	}

	switch e.Op {
	case '+':
		return cg.builder.CreateFAdd(l, r, "addtmp"), nil
	case '-':
		return cg.builder.CreateFSub(l, r, "subtmp"), nil
	case '*':
		return cg.builder.CreateFMul(l, r, "multmp"), nil
	case '<':
		cmp := cg.builder.CreateFCmp(llvm.FloatULT, l, r, "cmptmp")
		// Convert bool 0/1 to double 0.0 or 1.0
		return cg.builder.CreateUIToFP(cmp, cg.ctx.DoubleType(), "booltmp"), nil
	default:
		return llvm.Value{}, fmt.Errorf("Invalid binary operator `%q`", e.Op)
	}
}

func (cg *CodeGen) codegenCallExpr(e parser.CallExpr) (llvm.Value, error) {
	callee := cg.module.NamedFunction(e.Callee)
	if callee.IsNil() {
		return llvm.Value{}, fmt.Errorf("Unknown function `%q` referenced", e.Callee)
	}
	if len(e.Args) != callee.ParamsCount() {
		return llvm.Value{}, fmt.Errorf("Incorrect number of arguments passed to `%q`", e.Callee)
	}
	var args []llvm.Value
	for _, arg := range e.Args {
		val, err := cg.codegenExpr(arg)
		if err != nil {
			return llvm.Value{}, err
		}
		args = append(args, val)
	}
	return cg.builder.CreateCall(callee.Type().ElementType(), callee, args, "calltmp"), nil
}

func (cg *CodeGen) codegenPrototype(proto parser.Prototype) (llvm.Value, error) {
	// Make the function type:  double(double,double) etc.
	doubles := make([]llvm.Type, len(proto.Args))
	for i := range doubles {
		doubles[i] = cg.ctx.DoubleType()
	}

	fnType := llvm.FunctionType(cg.ctx.DoubleType(), doubles, false)
	fn := llvm.AddFunction(cg.module, proto.Name, fnType)
	fn.SetLinkage(llvm.ExternalLinkage) // function may be defined outside the current module

	for i, arg := range fn.Params() {
		arg.SetName(proto.Args[i])
	}

	return fn, nil
}

func (cg *CodeGen) codegenFunction(f parser.Function) (llvm.Value, error) {
	// check for an existing function from a previous 'extern' declaration.
	fn := cg.module.NamedFunction(f.Proto.Name)

	if fn.IsNil() {
		var err error
		fn, err = cg.codegenPrototype(f.Proto)
		if err != nil {
			return llvm.Value{}, err
		}
	}

	if fn.BasicBlocksCount() != 0 {
		// we shouldnt be in here
		return llvm.Value{}, fmt.Errorf("Function `%q` cannot be redefined", f.Proto.Name)
	}

	// check if proto args fit function args
	if fn.ParamsCount() != len(f.Proto.Args) {
		return llvm.Value{}, fmt.Errorf(
			"function %q declared with %d args, redefined with %d",
			f.Proto.Name,
			fn.ParamsCount(),
			len(f.Proto.Args),
		)
	}

	for i, arg := range fn.Params() {
		if arg.Name() != f.Proto.Args[i] {
			return llvm.Value{}, fmt.Errorf(
				"function %q argument %d was declared as %q, redefined as %q",
				f.Proto.Name,
				i,
				arg.Name(),
				f.Proto.Args[i],
			)
		}
	}

	// Create a new basic block to start insertion into.
	basicBlock := cg.ctx.AddBasicBlock(fn, "entry")
	cg.builder.SetInsertPointAtEnd(basicBlock)

	// Clear the NamedValues map
	for k := range cg.values {
		delete(cg.values, k)
	}
	// Add the function arguments to the NamedValues map
	for _, arg := range fn.Params() {
		cg.values[arg.Name()] = arg
	}

	body, err := cg.codegenExpr(f.Body)
	if err != nil {
		fn.EraseFromParentAsFunction()
		return llvm.Value{}, err
	}
	cg.builder.CreateRet(body)

	// IR consistency checks
	if err := llvm.VerifyFunction(fn, llvm.PrintMessageAction); err != nil {
		return llvm.Value{}, err
	}

	return fn, nil
}

func Gen(program []parser.Expr) (llvm.Module, error) {
	ctx := llvm.NewContext()
	cg := &CodeGen{
		ctx:     ctx,
		builder: ctx.NewBuilder(),
		module:  ctx.NewModule("main"),
		values:  make(map[string]llvm.Value),
	}
	for _, expr := range program {
		var err error

		switch e := expr.(type) {
		case parser.Function:
			_, err = cg.codegenFunction(e)
		case parser.Prototype:
			_, err = cg.codegenPrototype(e)
		default:
			return llvm.Module{}, fmt.Errorf("unknown top-level expression type %T", expr)
		}

		if err != nil {
			return llvm.Module{}, err
		}
	}
	return cg.module, nil
}
