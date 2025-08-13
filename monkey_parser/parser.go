package monkey_parser

import (
	"fmt"
	ast "myMonkey/monkey_ast"
	lexer "myMonkey/monkey_lexer"
	token "myMonkey/monkey_token"
	"strconv"
)

type (
	Precedence int
	nudFn      func() ast.Expression
	ledFn      func(ast.Expression) ast.Expression
	Parser     struct {
		l                   *lexer.Lexer
		errors              []string
		curToken, peekToken token.Token
		nudFns              map[token.TokenType]nudFn
		ledFns              map[token.TokenType]ledFn
	}
)

const (
	_ Precedence = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	SHIFT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]Precedence{
	token.EQ:       EQUALS,
	token.NEQ:      EQUALS,
	token.LT:       LESSGREATER,
	token.LE:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.GE:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.MULTIPLY: PRODUCT,
	token.DIVIDE:   PRODUCT,
	token.LSHIFT:   SHIFT,
	token.RSHIFT:   SHIFT,
	token.LPAREN:   CALL,
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}, nudFns: map[token.TokenType]nudFn{}, ledFns: map[token.TokenType]ledFn{}}
	p.registerNud(token.IDENTIFIER, p.parseIdent)
	p.registerNud(token.NUMBER, p.parseDecimal)
	p.registerNud(token.REVERSE, p.parsePrefix)
	p.registerNud(token.MINUS, p.parsePrefix)
	p.registerNud(token.BUMPPLUS, p.parsePrefix)
	p.registerNud(token.BUMPMINUS, p.parsePrefix)
	p.registerNud(token.TRUE, p.parseBoolean)
	p.registerNud(token.FALSE, p.parseBoolean)
	p.registerNud(token.LPAREN, p.parseGroup)
	p.registerNud(token.IF, p.parseIf)
	p.registerNud(token.FUNCTION, p.parseFn)
	p.registerLed(token.EQ, p.parseInfix)
	p.registerLed(token.NEQ, p.parseInfix)
	p.registerLed(token.LT, p.parseInfix)
	p.registerLed(token.LE, p.parseInfix)
	p.registerLed(token.GT, p.parseInfix)
	p.registerLed(token.GE, p.parseInfix)
	p.registerLed(token.PLUS, p.parseInfix)
	p.registerLed(token.MINUS, p.parseInfix)
	p.registerLed(token.MULTIPLY, p.parseInfix)
	p.registerLed(token.DIVIDE, p.parseInfix)
	p.registerLed(token.LSHIFT, p.parseInfix)
	p.registerLed(token.RSHIFT, p.parseInfix)
	p.registerLed(token.LPAREN, p.parseCall)
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) registerNud(tokenType token.TokenType, fn nudFn) { p.nudFns[tokenType] = fn }

func (p *Parser) registerLed(tokenType token.TokenType, fn ledFn) { p.ledFns[tokenType] = fn }

func (p *Parser) curTokenIs(t token.TokenType) bool { return p.curToken.Type == t }

func (p *Parser) nextTokenIs(t token.TokenType) bool { return p.peekToken.Type == t }

func (p *Parser) curPrecedence() Precedence {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) nextPrecedence() Precedence {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) expectNext(t token.TokenType) bool {
	if p.nextTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.nextError(t)
		return false
	}
}

func (p *Parser) nextError(t token.TokenType) {
	msg := fmt.Sprintf("Expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noNudError(tt token.TokenType) {
	msg := fmt.Sprintf("No null denotation function for %s found", tt)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.DEFINE:
		return p.parseDefinition()
	case token.RETURN:
		return p.parseReturn()
	default:
		return p.parseExpression()
	}
}

func (p *Parser) parseDefinition() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectNext(token.IDENTIFIER) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectNext(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.prattParser(LOWEST)
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturn() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.prattParser(LOWEST)
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.prattParser(LOWEST)
	if p.nextTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) prattParser(pre Precedence) ast.Expression {
	nud := p.nudFns[p.curToken.Type]
	if nud == nil {
		p.noNudError(p.curToken.Type)
		return nil
	}
	leftExp := nud()
	for !p.nextTokenIs(token.SEMICOLON) && pre < p.nextPrecedence() {
		led := p.ledFns[p.peekToken.Type]
		if led == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = led(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIdent() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseDecimal() ast.Expression {
	dl := &ast.DecimalLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("Could not parse %s as decimal", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	dl.Value = value
	return dl
}

func (p *Parser) parsePrefix() ast.Expression {
	exp := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	exp.Right = p.prattParser(PREFIX)
	return exp
}

func (p *Parser) parseInfix(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{Token: p.curToken, Operator: p.curToken.Literal, Left: left}
	pre := p.curPrecedence()
	p.nextToken()
	exp.Right = p.prattParser(pre)
	return exp
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroup() ast.Expression {
	p.nextToken()
	exp := p.prattParser(LOWEST)
	if !p.expectNext(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseBlock() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseIf() ast.Expression {
	exp := &ast.ConditionExpression{Token: p.curToken}
	if !p.expectNext(token.LPAREN) {
		return nil
	}
	p.nextToken()
	exp.Condition = p.prattParser(LOWEST)
	if !p.expectNext(token.RPAREN) {
		return nil
	}
	if !p.expectNext(token.LBRACE) {
		return nil
	}
	exp.True = p.parseBlock()
	if p.nextTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectNext(token.LBRACE) {
			return nil
		}
		exp.False = p.parseBlock()
	}
	return exp
}

func (p *Parser) parseFn() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectNext(token.LPAREN) {
		return nil
	}
	fn.Parameters = p.parseFnParams()
	if !p.expectNext(token.LBRACE) {
		return nil
	}
	fn.Body = p.parseBlock()
	return fn
}

func (p *Parser) parseFnParams() []*ast.Identifier {
	idents := []*ast.Identifier{}
	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		return idents
	}
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	idents = append(idents, ident)
	for p.nextTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		idents = append(idents, ident)
	}
	if !p.expectNext(token.RPAREN) {
		return nil
	}
	return idents
}

func (p *Parser) parseCall(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: fn}
	exp.Arguments = p.parseCallArgs()
	return exp
}

func (p *Parser) parseCallArgs() []ast.Expression {
	args := []ast.Expression{}
	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		return args
	}
	args = append(args, p.prattParser(LOWEST))
	for p.nextTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.prattParser(LOWEST))
	}
	if !p.expectNext(token.RPAREN) {
		return nil
	}
	return args
}
