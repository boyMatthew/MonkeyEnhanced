package monkey_evaluator

import (
	lexer "myMonkey/monkey_lexer"
	object "myMonkey/monkey_object"
	parser "myMonkey/monkey_parser"
	"testing"
)

func TestEvalDecimalExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5", 5.0},
		{"10", 10.0},
		{"9.2", 9.2},
		{"-5", -5.0},
		{"-10", -10.0},
		{"-9.2", -9.2},
		{"++5", 6.0},
		{"--10", 9.0},
		{"5+5+5+5-10", 10.0},
		{"2*2*2*2*2", 32.0},
		{"-50+100+-50", 0.0},
		{"5*2+10", 20.0},
		{"5+2*10", 25.0},
		{"20+2*-10", 0.0},
		{"50/2*2+10", 60.0},
		{"2*(5+10)", 30.0},
		{"3*3*3+10", 37.0},
		{"3*(3*3)+10", 37.0},
		{"(5+10*2+15/3)*2+-10", 50},
		{"2<<5", 64.0},
		{"64>>5", 2.0},
	}

	for _, test := range tests {
		evaluated := testEval(test.input)
		testDecimalObj(t, evaluated, test.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"1<2", true},
		{"1>2", false},
		{"1>1", false},
		{"1<1", false},
		{"1==1", true},
		{"1!=1", false},
		{"1==2", false},
		{"1!=2", true},
		{"1>=1", true},
		{"1<=2", true},
	}
	for _, test := range tests {
		evaluated := testEval(test.input)
		testBooleanObj(t, evaluated, test.expected)
	}
}

func TestConditionExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if(true){10}", 10.0},
		{"if(false){10}", nil},
		{"if(1){10}", 10.0},
		{"if(1<2){10}", 10.0},
		{"if(1>2){10}", nil},
		{"if(1<2){10}else{20}", 10.0},
		{"if(1>2){10}else{20}", 20.0},
	}
	for _, test := range tests {
		evaluated := testEval(test.input)
		decimal, ok := test.expected.(float64)
		if ok {
			testDecimalObj(t, evaluated, decimal)
		} else {
			testNullObj(t, evaluated)
		}
	}
}

func TestReturnValue(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"ret 10;", 10.0},
		{"ret 10;9;", 10.0},
		{"ret 2*5;9;", 10.0},
		{"9;ret 2*5;9;", 10.0},
		{"if(10>1){if(10>1){ret 10;}ret 1;}", 10.0},
	}

	for _, test := range tests {
		evaluated := testEval(test.input)
		testDecimalObj(t, evaluated, test.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	pro := p.Parse()
	return Eval(pro)
}

func testDecimalObj(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Decimal)
	if !ok {
		t.Errorf("object is not Decimal. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
		return false
	}
	return true
}

func testBooleanObj(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}
	return true
}

func testNullObj(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
