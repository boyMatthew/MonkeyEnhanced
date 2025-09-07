package monkey_code

import (
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
