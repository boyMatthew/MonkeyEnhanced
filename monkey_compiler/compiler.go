package monkey_compiler

import (
	ast "myMonkey/monkey_ast"
	code "myMonkey/monkey_code"
	object "myMonkey/monkey_object"
)

type Complier struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Complier {
	return &Complier{
		instructions: code.Instructions{},
		constants:    make([]object.Object, 0),
	}
}

func (c *Complier) Compile(node ast.Node) error {
	return nil
}

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Complier) ByteCode() *ByteCode {
	return &ByteCode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
