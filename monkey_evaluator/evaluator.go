package monkey_evaluator

import (
	ast "myMonkey/monkey_ast"
	object "myMonkey/monkey_object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.DecimalLiteral:
		return &object.Decimal{Value: node.Value}
	case *ast.Boolean:
		return convertBoolean(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefix(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfix(node.Operator, left, right)
	}
	return nil
}

func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt)
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
		return evalReverse(right)
	case "-":
		return evalMinus(right)
	case "++":
		return evalBumpPlus(right)
	case "--":
		return evalBumpMinus(right)
	default:
		return NULL
	}
}

func evalReverse(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinus(right object.Object) object.Object {
	if right.Type() != object.DECIMAL_OBJ {
		return NULL
	}
	value := right.(*object.Decimal).Value
	return &object.Decimal{Value: -value}
}

func evalBumpPlus(right object.Object) object.Object {
	if right.Type() != object.DECIMAL_OBJ {
		return NULL
	}
	value := right.(*object.Decimal).Value
	return &object.Decimal{Value: value + 1}
}

func evalBumpMinus(right object.Object) object.Object {
	if right.Type() != object.DECIMAL_OBJ {
		return NULL
	}
	value := right.(*object.Decimal).Value
	return &object.Decimal{Value: value - 1}
}

func evalInfix(op string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.DECIMAL_OBJ && right.Type() == object.DECIMAL_OBJ:
		return evalDecimalInfix(op, left.(*object.Decimal).Value, right.(*object.Decimal).Value)
	case op == "==":
		return convertBoolean(left == right)
	case op == "!=":
		return convertBoolean(left != right)
	default:
		return NULL
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
		return NULL
	}
}
