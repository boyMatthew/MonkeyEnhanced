package monkey_repl

import (
	"bufio"
	"fmt"
	"io"
	lexer "myMonkey/monkey_lexer"
	token "myMonkey/monkey_token"
)

const (
	READPROMPT  = "In[%d] :"
	WRITEPROMPT = "Out[%d] :"
)

func Start(read io.Reader, write io.Writer) error {
	scanner := bufio.NewScanner(read)

	for i := 1; true; i++ {
		_, err := fmt.Fprintf(write, READPROMPT, i)
		if err != nil {
			return fmt.Errorf("%d: write READPROMPT failed: %v", i, err)
		}
		scanned := scanner.Scan()
		if !scanned {
			return fmt.Errorf("%d: scan failed", i)
		}
		line := scanner.Text()
		l := lexer.NewLexer(line)
		_, err = fmt.Fprintf(write, WRITEPROMPT, i)
		if err != nil {
			return fmt.Errorf("%d: write WRITEPROMPT failed: %v", i, err)
		}
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			_, err := fmt.Fprintf(write, "%+v\n", tok)
			if err != nil {
				return fmt.Errorf("%d: write token failed: %v", i, err)
			}
		}
	}
	return nil
}
