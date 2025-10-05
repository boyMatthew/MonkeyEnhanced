package monkey_compiler

import (
	ast "myMonkey/monkey_ast"
	code "myMonkey/monkey_code"
	object "myMonkey/monkey_object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    make([]object.Object, 0),
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch n := node.(type) {
	case *ast.Program:
		for _, s := range n.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(n.Expression)
		if err != nil {
			return err
		}
	case *ast.InfixExpression:
		err := c.Compile(n.Left)
		if err != nil {
			return err
		}
		err = c.Compile(n.Right)
		if err != nil {
			return err
		}
	case *ast.DecimalLiteral:
		decimal := &object.Decimal{Value: n.Value}
		c.emit(code.OpConstant, c.addConstant(decimal))
	}
	return nil
}

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	return c.addIns(ins)
}

func (c *Compiler) addIns(ins []byte) int {
	posNewIns := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewIns
}
