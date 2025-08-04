package monkey_token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"func":  FUNCTION,
	"def":   DEFINE,
	"true":  TRUE,
	"false": FALSE,
	"if":    IF,
	"else":  ELSE,
	"ret":   RETURN,
}

func LookupKeyword(keyword string) TokenType {
	if tok, ok := keywords[keyword]; ok {
		return tok
	}
	return IDENTIFIER
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENTIFIER = "IDENT"
	NUMBER     = "DECIMAL"

	ASSIGN    = "="
	PLUS      = "+"
	MINUS     = "-"
	REVERSE   = "!"
	MULTIPLY  = "*"
	DIVIDE    = "/"
	LSHIFT    = "<<"
	RSHIFT    = ">>"
	BUMPPLUS  = "++"
	BUMPMINUS = "--"

	LT  = "<"
	GT  = ">"
	EQ  = "=="
	NEQ = "!="
	LE  = "<="
	GE  = ">="

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACKET  = "["
	RBRACKET  = "]"
	LBRACE    = "{"
	RBRACE    = "}"

	FUNCTION = "FUNC"
	DEFINE   = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)
