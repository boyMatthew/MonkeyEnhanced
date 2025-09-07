package monkey_compiler

import (
	"fmt"
	ast "myMonkey/monkey_ast"
	code "myMonkey/monkey_code"
	lexer "myMonkey/monkey_lexer"
	object "myMonkey/monkey_object"
	parser "myMonkey/monkey_parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func parse(input string) *ast.Program {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	return p.Parse()
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, i := range s {
		out = append(out, i...)
	}
	return out
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)
	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot=%q", concatted, actual)
	}
	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot=%q", i, concatted, actual)
		}
	}
	return nil
}

func testConstants(expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong constants length. want=%d got=%d", len(expected), len(actual))
	}
	for i, con := range expected {
		switch con := con.(type) {
		case float64:
			err := testDecimalObject(con, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testDecimalObject failed: %s", i, err)
			}
		}
	}
	return nil
}

func testDecimalObject(expected float64, actual object.Object) error {
	res, ok := actual.(*object.Decimal)
	if !ok {
		return fmt.Errorf("object is not Decimal. got=%T(%+v)", actual, actual)
	}
	if res.Value != expected {
		return fmt.Errorf("object has wrong value. got=%f, want=%f", res.Value, expected)
	}
	return nil
}

func TestDecimalArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1.0, 2.0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
			},
		},
	}
	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, test := range tests {
		pro := parse(test.input)
		compiler := New()
		err := compiler.Compile(pro)
		if err != nil {
			t.Errorf("Compile failed: %s", err)
		}
		bytecode := compiler.ByteCode()
		err = testInstructions(test.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}
		err = testConstants(test.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}
