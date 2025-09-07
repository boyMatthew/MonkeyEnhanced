package monkey_code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)
		if len(instruction) != len(test.expected) {
			t.Errorf("Instruction has wrong length. want=%d, got=%d", len(instruction), len(test.expected))
		}
		for i, b := range test.expected {
			if instruction[i] != b {
				t.Errorf("Wrong byte at position %d. want=%d, got=%d", i, b, instruction[i])
			}
		}
	}
}
