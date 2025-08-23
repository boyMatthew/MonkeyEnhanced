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
	INDEX
	ASSIGN
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
	token.ASSIGN:   ASSIGN,
	token.LBRACKET: INDEX,
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}, nudFns: map[token.TokenType]nudFn{}, ledFns: map[token.TokenType]ledFn{}}
	p.registerNuds(p.parseIdent, token.IDENTIFIER)
	p.registerNuds(p.parseDecimal, token.NUMBER)
	p.registerNuds(p.parseString, token.STRING)
	p.registerNuds(p.parseGroup, token.LPAREN)
	p.registerNuds(p.parseIf, token.IF)
	p.registerNuds(p.parseFn, token.FUNCTION)
	p.registerNuds(p.parseBoolean, token.TRUE, token.FALSE)
	p.registerNuds(p.parsePrefix, token.REVERSE, token.MINUS, token.BUMPPLUS, token.BUMPMINUS)
	p.registerNuds(p.parseHash, token.LBRACE)
	p.registerNuds(p.parseArray, token.LBRACKET)
	p.registerLeds(p.parseCall, token.LPAREN)
	p.registerLeds(p.parseAssign, token.ASSIGN)
	p.registerLeds(p.parseInfix, token.EQ, token.NEQ, token.LT, token.LE, token.GT, token.GE, token.PLUS, token.MINUS, token.MULTIPLY, token.DIVIDE, token.LSHIFT, token.RSHIFT)
	p.registerLeds(p.parseIndex, token.LBRACKET)
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

func (p *Parser) registerNuds(fn nudFn, tokenTypes ...token.TokenType) {
	for _, tokenType := range tokenTypes {
		p.nudFns[tokenType] = fn
	}
}

func (p *Parser) registerLeds(fn ledFn, tokenTypes ...token.TokenType) {
	for _, tokenType := range tokenTypes {
		p.ledFns[tokenType] = fn
	}
}

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
	case token.LOOP:
		return p.parseLoop()
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

func (p *Parser) parseLoop() *ast.LoopStatement {
	loop := &ast.LoopStatement{Token: p.curToken}
	if !p.expectNext(token.LPAREN) {
		return nil
	}
	p.nextToken()
	if loop.Token.Literal == "for" {
		loop.Initial = p.parseDefinition()
		p.nextToken()
		loop.Condition = p.parseExpression()
		p.nextToken()
		loop.AfterBlock = p.parseExpression()
	} else {
		loop.Condition = p.parseExpression()
	}
	if !p.expectNext(token.RPAREN) {
		return nil
	}
	if !p.expectNext(token.LBRACE) {
		return nil
	}
	loop.Body = p.parseBlock()
	return loop
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
	exp.Arguments = p.parseExpList(token.RPAREN)
	return exp
}

func (p *Parser) parseArray() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Value = p.parseExpList(token.RBRACKET)
	return array
}

func (p *Parser) parseIndex(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.prattParser(LOWEST)
	if !p.expectNext(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseExpList(end token.TokenType) []ast.Expression {
	args := []ast.Expression{}
	p.nextToken()
	if p.curTokenIs(end) {
		return args
	}
	args = append(args, p.prattParser(LOWEST))
	for p.nextTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.prattParser(LOWEST))
	}
	if !p.expectNext(end) {
		return nil
	}
	return args
}

func (p *Parser) parseAssign(left ast.Expression) ast.Expression {
	name, ok := left.(*ast.Identifier)
	if !ok {
		msg := fmt.Sprintf("Expected an identifier, got %s", left.String())
		p.errors = append(p.errors, msg)
		return nil
	}
	assign := &ast.AssignExpression{Token: p.curToken, Name: name}
	p.nextToken()
	assign.Value = p.prattParser(LOWEST)
	return assign
}

func (p *Parser) parseString() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseHash() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.nextTokenIs(token.RBRACE) {
		p.nextToken()
		k := p.prattParser(LOWEST)
		if !p.expectNext(token.COLON) {
			return nil
		}
		p.nextToken()
		v := p.prattParser(LOWEST)
		hash.Pairs[k] = v
		if !p.nextTokenIs(token.RBRACE) && !p.expectNext(token.COMMA) {
			return nil
		}
	}
	if !p.expectNext(token.RBRACE) {
		return nil
	}
	return hash
}
