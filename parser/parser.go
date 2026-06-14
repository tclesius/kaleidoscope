package parser

import (
	"fmt"
	"strconv"
	"tclesius/kaleidoscope/lexer"
)

type Expr any

type NumberExpr struct {
	Val float64
}

type VariableExpr struct {
	Name string
}

type BinaryExpr struct {
	Op  rune
	Lhs Expr
	Rhs Expr
}

type CallExpr struct {
	Callee string
	Args   []Expr
}

type Prototype struct {
	// Represents the "Prototype" of a function,
	Name string
	Args []string
}

type Function struct {
	// Represents the function definition itself
	Proto Prototype
	Body  Expr
}

////

type Parser struct {
	lexer      *lexer.Lexer
	curr       lexer.Token
	precedence map[string]int
}

func NewParser(lexer *lexer.Lexer) *Parser {
	p := &Parser{
		lexer: lexer,
		precedence: map[string]int{
			"<": 10,
			//">": 10,
			"+": 20,
			"-": 20,
			"*": 30,
			//"/": 30,
		},
	}

	p.advance()
	return p
}
func (p *Parser) peekPrecedence() int {
	val, ok := p.precedence[p.Peek().Value]
	if !ok {
		return -1
	}
	return val
}

func (p *Parser) advance() {
	p.curr = p.lexer.NextToken()
}

func (p *Parser) Advance() {
	p.advance()
}

func (p *Parser) consume() lexer.Token {
	tok := p.curr
	p.advance()
	return tok
}

func (p *Parser) Peek() lexer.Token { // TODO: make private
	return p.curr
}

func (p *Parser) parseNumberExpr() (Expr, error) {
	/// numberexpr ::= number
	result := p.consume()

	s, error := strconv.ParseFloat(result.Value, 64)
	if error != nil {
		return nil, error
	}
	return NumberExpr{
		Val: s,
	}, nil
}

func (p *Parser) parseParenExpr() (Expr, error) {
	//// parenexpr ::= '(' expression ')'
	p.consume()                          // (
	result, error := p.parseExpression() // TODO: handle missing result

	if error != nil {
		return nil, error
	}

	if p.Peek().Value != ")" {
		return nil, fmt.Errorf("Error: Expected ')'")
	}
	p.consume() // )
	return result, nil
}

func (p *Parser) parseIdentifierExpr() (Expr, error) {
	/// identifierexpr
	///   ::= identifier
	///   ::= identifier '(' expression* ')'
	name := p.consume() // ident name

	if p.Peek().Value != "(" {
		// Simple variable ref.
		return VariableExpr{Name: name.Value}, nil
	}

	// Call
	p.consume() // (
	var args []Expr
	for p.Peek().Value != ")" {
		arg, error := p.parseExpression()

		if error != nil {
			return nil, error
		}

		args = append(args, arg)
		if p.Peek().Value == ")" {
			break
		}
		if p.Peek().Value != "," {
			return nil, fmt.Errorf("Error: Expected ')' or ',' in argument list")
		}
		p.consume() // ,
	}
	p.consume() // )

	return CallExpr{
		Callee: name.Value,
		Args:   args,
	}, nil
}

func (p *Parser) parsePrimary() (Expr, error) {
	/// primary
	///   ::= identifierexpr
	///   ::= numberexpr
	///   ::= parenexpr
	switch p.Peek().Type {
	default:
		return nil, fmt.Errorf("Error: Unknown token when expecting an expression")
	case lexer.IDENTIFIER:
		return p.parseIdentifierExpr()
	case lexer.NUMBER:
		return p.parseNumberExpr()
	case '(':
		return p.parseParenExpr()
	}
}

func (p *Parser) parseExpression() (Expr, error) {
	/// expression
	///   ::= primary binoprhs
	lhs, error := p.parsePrimary()
	if error != nil {
		return nil, error
	}

	rhs, error := p.parseBinOpRhs(0, lhs)
	if error != nil {
		return nil, error
	}
	return rhs, nil
}

func (p *Parser) parseBinOpRhs(minPrecedence int, lhs Expr) (Expr, error) {
	/// minPrecendence is the minimal operator precedence that the function is allowed to consume
	/// binoprhs
	///   ::= ('+' primary)*
	for {
		prec := p.peekPrecedence()
		if prec < minPrecedence {
			return lhs, nil
		}
		binOp := p.consume()
		rhs, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		nextPrec := p.peekPrecedence()
		if prec < nextPrec {
			var err error
			rhs, err = p.parseBinOpRhs(prec+1, rhs)
			if err != nil {
				return nil, err
			}
		}
		lhs = BinaryExpr{
			Op:  rune(binOp.Type),
			Lhs: lhs,
			Rhs: rhs,
		}
	}
}

func (p *Parser) parsePrototype() (*Prototype, error) {
	/// prototype
	///   ::= id '(' id* ')'
	if p.Peek().Type != lexer.IDENTIFIER {
		return nil, fmt.Errorf("Error: Expected function name in prototype")
	}

	fnName := p.consume().Value

	if p.Peek().Value != "(" {
		return nil, fmt.Errorf("Error: Expected '(' in prototype")
	}
	p.consume() // (

	// Read the list of argument names.
	var argNames []string
	for p.Peek().Type == lexer.IDENTIFIER {
		argNames = append(argNames, p.consume().Value)
	}

	if p.Peek().Value != ")" {
		return nil, fmt.Errorf("Error: Expected ')' in prototype")
	}
	// success.
	p.consume() // ')'
	return &Prototype{
		Name: fnName,
		Args: argNames,
	}, nil
}

func (p *Parser) parseDefinition() (Expr, error) {
	/// definition ::= 'def' prototype expression
	p.consume() // 'def'
	proto, err := p.parsePrototype()
	if err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return Function{
		Proto: *proto,
		Body:  expr,
	}, nil
}

func (p *Parser) parseExtern() (Expr, error) {
	/// external ::= 'extern' prototype
	p.consume() // 'extern'
	return p.parsePrototype()
}

func (p *Parser) parseTopLevelExpr() (Expr, error) {
	/// toplevelexpr ::= expression

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	// Make an anonymous proto.
	proto := Prototype{
		Name: "__anon_expr",
	}
	return Function{
		Proto: proto,
		Body:  expr,
	}, nil
}

func (p *Parser) ParseDefinition() (Expr, error) {
	return p.parseDefinition()
}

func (p *Parser) ParseExtern() (Expr, error) {
	return p.parseExtern()
}

func (p *Parser) ParseTopLevelExpr() (Expr, error) {
	return p.parseTopLevelExpr()
}
