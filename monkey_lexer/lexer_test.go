package monkey_lexer

import (
	token "myMonkey/monkey_token"
	"testing"
)

type aTest struct {
	expectedType    token.TokenType
	expectedLiteral string
}

func TestLexer(t *testing.T) {
	input := `
		def five = 5;
		def ten = 10;
		def add = func(x, y) {ret x + y;};
		def result = add(five, ten);
		!-/*5;
		5 < 10 > 5;
		5 <= 10 >= 5;
		5 << 10 >> 5;
		fuck[you];
		8.7 == -9.8;
		if (5!=10) {false;} else {true;}
		karma:a:bitch;
		for(def x = 0;| x < 10;| ++x) {}
		while(true) {}
	`
	tests := []aTest{
		{token.DEFINE, "def"},
		{token.IDENTIFIER, "five"},
		{token.ASSIGN, "="},
		{token.NUMBER, "5"},
		{token.SEMICOLON, ";"},
		{token.DEFINE, "def"},
		{token.IDENTIFIER, "ten"},
		{token.ASSIGN, "="},
		{token.NUMBER, "10"},
		{token.SEMICOLON, ";"},
		{token.DEFINE, "def"},
		{token.IDENTIFIER, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "func"},
		{token.LPAREN, "("},
		{token.IDENTIFIER, "x"},
		{token.COMMA, ","},
		{token.IDENTIFIER, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "ret"},
		{token.IDENTIFIER, "x"},
		{token.PLUS, "+"},
		{token.IDENTIFIER, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.DEFINE, "def"},
		{token.IDENTIFIER, "result"},
		{token.ASSIGN, "="},
		{token.IDENTIFIER, "add"},
		{token.LPAREN, "("},
		{token.IDENTIFIER, "five"},
		{token.COMMA, ","},
		{token.IDENTIFIER, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.REVERSE, "!"},
		{token.MINUS, "-"},
		{token.DIVIDE, "/"},
		{token.MULTIPLY, "*"},
		{token.NUMBER, "5"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "5"},
		{token.LT, "<"},
		{token.NUMBER, "10"},
		{token.GT, ">"},
		{token.NUMBER, "5"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "5"},
		{token.LE, "<="},
		{token.NUMBER, "10"},
		{token.GE, ">="},
		{token.NUMBER, "5"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "5"},
		{token.LSHIFT, "<<"},
		{token.NUMBER, "10"},
		{token.RSHIFT, ">>"},
		{token.NUMBER, "5"},
		{token.SEMICOLON, ";"},
		{token.IDENTIFIER, "fuck"},
		{token.LBRACKET, "["},
		{token.IDENTIFIER, "you"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "8.7"},
		{token.EQ, "=="},
		{token.MINUS, "-"},
		{token.NUMBER, "9.8"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.NUMBER, "5"},
		{token.NEQ, "!="},
		{token.NUMBER, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.IDENTIFIER, "karma"},
		{token.COLON, ":"},
		{token.IDENTIFIER, "a"},
		{token.COLON, ":"},
		{token.IDENTIFIER, "bitch"},
		{token.SEMICOLON, ";"},
		{token.LOOP, "for"},
		{token.LPAREN, "("},
		{token.DEFINE, "def"},
		{token.IDENTIFIER, "x"},
		{token.ASSIGN, "="},
		{token.NUMBER, "0"},
		{token.SEMICOLON, ";"},
		{token.VERTICAL, "|"},
		{token.IDENTIFIER, "x"},
		{token.LT, "<"},
		{token.NUMBER, "10"},
		{token.SEMICOLON, ";"},
		{token.VERTICAL, "|"},
		{token.BUMPPLUS, "++"},
		{token.IDENTIFIER, "x"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LOOP, "while"},
		{token.LPAREN, "("},
		{token.TRUE, "true"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}
	dealTesting(t, input, tests)
}

func TestLexerParsingFloating(t *testing.T) {
	input := "-127.1.1.1;"
	tests := []aTest{
		{token.MINUS, "-"},
		{token.NUMBER, "127.1"},
		{token.ILLEGAL, "."},
		{token.NUMBER, "1.1"},
	}
	dealTesting(t, input, tests)
}

func dealTesting(t *testing.T, input string, tests []aTest) {
	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q expectedLiteral=%q gotLiteral%q", i, tt.expectedType, tok.Type, tt.expectedLiteral, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
