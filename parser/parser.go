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
	tokens     []lexer.Token
	curr       int
	precedence map[string]int
}

func (p *Parser) peekPrecedence() int {
	val, ok := p.precedence[p.peek().Value]
	if !ok {
		return -1
	}
	return val
}

func (p *Parser) advance() {
	if p.curr < len(p.tokens)-1 {
		p.curr += 1
	}
}

func (p *Parser) consume() lexer.Token {
	tok := p.peek()
	p.advance()
	return tok
}

func (p *Parser) peek() lexer.Token {
	if p.curr >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[p.curr]
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

	if p.peek().Value != ")" {
		return nil, fmt.Errorf("Expected ')'")
	}
	p.consume() // )
	return result, nil
}

func (p *Parser) parseIdentifierExpr() (Expr, error) {
	/// identifierexpr
	///   ::= identifier
	///   ::= identifier '(' expression* ')'
	name := p.consume() // ident name

	if p.peek().Value != "(" {
		// Simple variable ref.
		return VariableExpr{Name: name.Value}, nil
	}

	// Call
	p.consume() // (
	var args []Expr
	for p.peek().Value != ")" {
		arg, error := p.parseExpression()

		if error != nil {
			return nil, error
		}

		args = append(args, arg)
		if p.peek().Value == ")" {
			break
		}
		if p.peek().Value != "," {
			return nil, fmt.Errorf("Expected ')' or ',' in argument list")
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
	switch p.peek().Type {
	default:
		return nil, fmt.Errorf("Unknown token when expecting an expression")
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
	if p.peek().Type != lexer.IDENTIFIER {
		return nil, fmt.Errorf("Expected function name in prototype")
	}

	fnName := p.consume().Value

	if p.peek().Value != "(" {
		return nil, fmt.Errorf("Expected '(' in prototype")
	}
	p.consume() // (

	// Read the list of argument names.
	var argNames []string
	for p.peek().Type == lexer.IDENTIFIER {
		argNames = append(argNames, p.consume().Value)
	}

	if p.peek().Value != ")" {
		return nil, fmt.Errorf("Expected ')' in prototype")
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
	proto, err := p.parsePrototype()
	if err != nil {
		return nil, err
	}
	return *proto, nil
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

func Parse(t []lexer.Token) ([]Expr, error) {
	p := &Parser{
		tokens: t,
		precedence: map[string]int{
			"<": 10,
			//">": 10,
			"+": 20,
			"-": 20,
			"*": 30,
			//"/": 30,
		},
	}
	var program []Expr
	for p.peek().Type != lexer.EOF {
		switch p.peek().Type {
		case lexer.TokenType(';'):
			p.advance()
		case lexer.DEF:
			expr, err := p.parseDefinition()
			if err != nil {
				return nil, err
			}
			program = append(program, expr)
		case lexer.EXTERN:
			expr, err := p.parseExtern()
			if err != nil {
				return nil, err
			}
			program = append(program, expr)
		default:
			expr, err := p.parseTopLevelExpr()
			if err != nil {
				return nil, err
			}
			program = append(program, expr)
		}
	}

	return program, nil
}
