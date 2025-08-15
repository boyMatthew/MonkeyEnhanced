package monkey_evaluator

import (
	"fmt"
	ast "myMonkey/monkey_ast"
	object "myMonkey/monkey_object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.BlockStatement:
		return evalBlock(node)
	case *ast.DecimalLiteral:
		return &object.Decimal{Value: node.Value}
	case *ast.Boolean:
		return convertBoolean(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalPrefix(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		if isError(left) {
			return left
		}
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalInfix(node.Operator, left, right)
	case *ast.ConditionExpression:
		return evalCondition(node)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	}
	return nil
}

func evalProgram(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlock(block *ast.BlockStatement) object.Object {
	var result object.Object
	for _, stmt := range block.Statements {
		result = Eval(stmt)
		if result != nil && (result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ) {
			return result
		}
	}
	return result
}

func convertBoolean(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefix(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return convertBoolean(!isTrue(right))
	case "-":
		return evalMinus(right)
	case "++":
		return evalBumpPlus(right)
	case "--":
		return evalBumpMinus(right)
	default:
		return newError("unknown operator: %s%s", op, right.Type())
	}
}

func isTrue(obj object.Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	default:
		return true
	}
}

func evalMinus(right object.Object) object.Object {
	if right.Type() != object.DECIMAL_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Decimal).Value
	return &object.Decimal{Value: -value}
}

func evalBumpPlus(right object.Object) object.Object {
	if right.Type() != object.DECIMAL_OBJ {
		return newError("unknown operator: ++%s", right.Type())
	}
	value := right.(*object.Decimal).Value
	return &object.Decimal{Value: value + 1}
}

func evalBumpMinus(right object.Object) object.Object {
	if right.Type() != object.DECIMAL_OBJ {
		return newError("unknown operator: --%s", right.Type())
	}
	value := right.(*object.Decimal).Value
	return &object.Decimal{Value: value - 1}
}

func evalInfix(op string, left, right object.Object) object.Object {
	switch {
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	case left.Type() == object.DECIMAL_OBJ && right.Type() == object.DECIMAL_OBJ:
		return evalDecimalInfix(op, left.(*object.Decimal).Value, right.(*object.Decimal).Value)
	case op == "==":
		return convertBoolean(left == right)
	case op == "!=":
		return convertBoolean(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalDecimalInfix(op string, left, right float64) object.Object {
	switch op {
	case "+":
		return &object.Decimal{Value: left + right}
	case "-":
		return &object.Decimal{Value: left - right}
	case "*":
		return &object.Decimal{Value: left * right}
	case "/":
		return &object.Decimal{Value: left / right}
	case "<<":
		return &object.Decimal{Value: float64(int64(left) << int64(right))}
	case ">>":
		return &object.Decimal{Value: float64(int64(left) >> int64(right))}
	case "<":
		return convertBoolean(left < right)
	case ">":
		return convertBoolean(left > right)
	case "==":
		return convertBoolean(left == right)
	case "!=":
		return convertBoolean(left != right)
	case ">=":
		return convertBoolean(left >= right)
	case "<=":
		return convertBoolean(left <= right)
	default:
		return newError("unknown operator: %f %s %f", left, op, right)
	}
}

func evalCondition(ce *ast.ConditionExpression) object.Object {
	condition := Eval(ce.Condition)
	if isError(condition) {
		return condition
	} else if isTrue(condition) {
		return Eval(ce.True)
	} else if ce.False != nil {
		return Eval(ce.False)
	} else {
		return NULL
	}
}
