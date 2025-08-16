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

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlock(node, env)
	case *ast.DecimalLiteral:
		return &object.Decimal{Value: node.Value}
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}
	case *ast.Boolean:
		return convertBoolean(node.Value)
	case *ast.Identifier:
		return evalIdent(node, env)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefix(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfix(node.Operator, left, right)
	case *ast.ConditionExpression:
		return evalCondition(node, env)
	case *ast.AssignExpression:
		name := node.Name.Value
		if !env.Exist(name) {
			return newError("identifier not found: %s", name)
		}
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(name, val)
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExps(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunc(function, args)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	}
	return nil
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlock(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range block.Statements {
		result = Eval(stmt, env)
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

func evalCondition(ce *ast.ConditionExpression, env *object.Environment) object.Object {
	condition := Eval(ce.Condition, env)
	if isError(condition) {
		return condition
	} else if isTrue(condition) {
		return Eval(ce.True, env)
	} else if ce.False != nil {
		return Eval(ce.False, env)
	} else {
		return NULL
	}
}

func evalIdent(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s", node.Value)
	}
	return val
}

func evalExps(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, exp := range exps {
		evaluated := Eval(exp, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunc(fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}
	env := extendFuncEnv(function, args)
	evaluated := Eval(function.Body, env)
	return getReturnValue(evaluated)
}

func extendFuncEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}
	return env
}

func getReturnValue(obj object.Object) object.Object {
	if retValue, ok := obj.(*object.ReturnValue); ok {
		return retValue.Value
	}
	return obj
}
