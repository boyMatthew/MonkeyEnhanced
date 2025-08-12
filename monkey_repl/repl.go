package monkey_repl

import (
	"bufio"
	"fmt"
	"io"
	lexer "myMonkey/monkey_lexer"
	parser "myMonkey/monkey_parser"
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
		p := parser.NewParser(l)
		pro := p.Parse()
		_, err = fmt.Fprintf(write, WRITEPROMPT, i)
		if err != nil {
			return fmt.Errorf("%d: write WRITEPROMPT failed: %v", i, err)
		}
		if len(p.Errors()) != 0 {
			printParserErrors(write, p.Errors())
			continue
		}
		io.WriteString(write, pro.String())
		io.WriteString(write, "\n")
	}
	return nil
}

func printParserErrors(write io.Writer, errors []string) {
	io.WriteString(write, "Whoops! We've encountered some errors!\nParser errors:\n")
	for _, msg := range errors {
		io.WriteString(write, "\t"+msg+"\n")
	}
}
