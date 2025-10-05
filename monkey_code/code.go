package monkey_code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
)

type Definition struct {
	Name     string
	OpWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
}

func Lookup(op Opcode) (*Definition, error) {
	def, ok := definitions[op]
	if !ok {
		return nil, fmt.Errorf("Unknown opcode: %v", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}
	iLen := 1
	for _, w := range def.OpWidths {
		iLen += w
	}
	instruction := make([]byte, iLen)
	instruction[0] = byte(op)
	offset := 1
	for i, o := range operands {
		wid := def.OpWidths[i]
		switch wid {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += wid
	}
	return instruction
}

func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(Opcode(ins[i]))
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtIns(def, operands))
		i += read + 1
	}
	return out.String()
}

func (ins Instructions) fmtIns(def *Definition, operands []int) string {
	operandCount := len(def.OpWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}
	switch operandCount {
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}
	return fmt.Sprintf("ERROR: unhandled operand count for %s\n", def.Name)
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OpWidths))
	offset := 0
	for i, w := range def.OpWidths {
		switch w {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += w
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
