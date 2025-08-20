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

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Name: "len",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Decimal{Value: float64(len(arg.Value))}
			case *object.Array:
				return &object.Decimal{Value: float64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"truncate": &object.Builtin{
		Name: "truncate",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 && len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("first argument to `truncate` must be array, got %s", args[0].Type())
			}
			if args[1].Type() != object.DECIMAL_OBJ {
				return newError("second argument to `truncate` must be decimal, got %s", args[1].Type())
			}
			if len(args) == 3 && args[2].Type() != object.DECIMAL_OBJ {
				return newError("third argument to `truncate` must be decimal, got %s", args[2].Type())
			}
			arr := args[0].(*object.Array)
			length := len(arr.Value)
			if length <= 0 {
				return NULL
			}
			first := int(args[1].(*object.Decimal).Value)
			if len(args) == 2 {
				newArr := make([]object.Object, length-first)
				copy(newArr, arr.Value[first:length])
				return &object.Array{Value: newArr}
			} else {
				last := int(args[2].(*object.Decimal).Value)
				if last < first {
					first, last = last, first
				}
				newArr := make([]object.Object, last-first+1)
				copy(newArr, arr.Value[first:last])
				return &object.Array{Value: newArr}
			}
		},
	},
}

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
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		eles := evalExps(node.Value, env)
		if len(eles) == 1 && isError(eles[0]) {
			return eles[0]
		}
		return &object.Array{Value: eles}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndex(left, index)
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
	case left.Type() == object.STRING_OBJ && right.Type() == object.DECIMAL_OBJ:
		return evalStringMultiplication(op, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	case left.Type() == object.DECIMAL_OBJ && right.Type() == object.DECIMAL_OBJ:
		return evalDecimalInfix(op, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringConcentration(op, left, right)
	case op == "==":
		return convertBoolean(left == right)
	case op == "!=":
		return convertBoolean(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalDecimalInfix(op string, leftObj, rightObj object.Object) object.Object {
	left := leftObj.(*object.Decimal).Value
	right := rightObj.(*object.Decimal).Value
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
		return newError("unknown operator: %s %s %s", leftObj.Type(), op, rightObj.Type())
	}
}

func evalStringConcentration(op string, left, right object.Object) object.Object {
	if op != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalStringMultiplication(op string, left, right object.Object) object.Object {
	if op != "*" {
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := int(right.(*object.Decimal).Value)
	finalVal := ""
	for i := 0; i < rightVal; i++ {
		finalVal += leftVal
	}
	return &object.String{Value: finalVal}
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
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
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
	switch function := fn.(type) {
	case *object.Function:
		env := extendFuncEnv(function, args)
		evaluated := Eval(function.Body, env)
		return getReturnValue(evaluated)
	case *object.Builtin:
		return function.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
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

func evalIndex(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.DECIMAL_OBJ:
		return evalArrayIndex(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndex(array, index object.Object) object.Object {
	arrayObj := array.(*object.Array)
	idx := int64(index.(*object.Decimal).Value)
	maxLen := int64(len(arrayObj.Value) - 1)
	if idx > maxLen || idx < 0 {
		return NULL
	}
	return arrayObj.Value[idx]
}
