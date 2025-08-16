package monkey_repl

import (
	"bufio"
	"fmt"
	"io"
	evaluator "myMonkey/monkey_evaluator"
	lexer "myMonkey/monkey_lexer"
	object "myMonkey/monkey_object"
	parser "myMonkey/monkey_parser"
)

const (
	READPROMPT  = "In[%d] :"
	WRITEPROMPT = "Out[%d] :"
)

func Start(read io.Reader, write io.Writer) error {
	scanner := bufio.NewScanner(read)
	env := object.NewEnvironment()

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
		if len(p.Errors()) != 0 {
			_, err = fmt.Fprintf(write, WRITEPROMPT, i)
			if err != nil {
				return fmt.Errorf("%d: write WRITEPROMPT failed: %v", i, err)
			}
			printParserErrors(write, p.Errors())
			continue
		}
		evaluated := evaluator.Eval(pro, env)
		if evaluated != nil {
			_, err = fmt.Fprintf(write, WRITEPROMPT, i)
			if err != nil {
				return fmt.Errorf("%d: write WRITEPROMPT failed: %v", i, err)
			}
			io.WriteString(write, evaluated.Inspect())
			io.WriteString(write, "\n")
		}
	}
	return nil
}

func printParserErrors(write io.Writer, errors []string) {
	io.WriteString(write, "Whoops! We've encountered some errors!\nParser errors:\n")
	for _, msg := range errors {
		io.WriteString(write, "\t"+msg+"\n")
	}
}
