package monkey_object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"math"
	ast "myMonkey/monkey_ast"
	"strconv"
	"strings"
)

type ObjectType string
type BuiltinFn func(args ...Object) Object

const (
	DECIMAL_OBJ      ObjectType = "DECIMAL"
	BOOLEAN_OBJ                 = "BOOLEAN"
	NULL_OBJ                    = "NULL"
	RETURN_VALUE_OBJ            = "RETURN_VALUE"
	ERROR_OBJ                   = "ERROR"
	FUNCTION_OBJ                = "FUNCTION"
	STRING_OBJ                  = "STRING"
	BUILTIN_OBJ                 = "BUILTIN"
	ARRAY_OBJ                   = "ARRAY"
	HASH_OBJ                    = "HASH"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type HashAble interface {
	HashKey() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type HashPair struct {
	Key, Value Object
}

func boolToInt(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

type Decimal struct {
	Value float64
}

func (d *Decimal) Inspect() string  { return fmt.Sprintf("%f", d.Value) }
func (d *Decimal) Type() ObjectType { return DECIMAL_OBJ }
func (d *Decimal) HashKey() HashKey {
	var out bytes.Buffer
	iNT, frac := math.Modf(d.Value)
	fmt.Fprintf(&out, "%d%d%d%d", int(iNT), int(frac), countDecimalPlaces(d.Value), boolToInt(d.Value < 0))
	val, err := strconv.Atoi(out.String())
	if err != nil {
		panic(err)
	}
	return HashKey{Type: d.Type(), Value: uint64(val)}
}

func countDecimalPlaces(f float64) int {
	_, frac := math.Modf(f)
	fracStr := fmt.Sprintf("%g", frac)
	dotIndex := strings.IndexByte(fracStr, '.')
	if dotIndex == -1 {
		return 0
	}
	return len(fracStr) - dotIndex - 1
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) HashKey() HashKey { return HashKey{Type: b.Type(), Value: boolToInt(b.Value)} }

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL_OBJ }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object), outer: nil}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, value Object) Object {
	e.store[name] = value
	return value
}

func (e *Environment) Exist(name string) bool {
	_, ok := e.store[name]
	return ok
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("func(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type Builtin struct {
	Name string
	Fn   BuiltinFn
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return fmt.Sprintf("<builtin function: %s>", b.Name) }

type Array struct {
	Value []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	out.WriteString("[")
	elems := []string{}
	for _, e := range a.Value {
		elems = append(elems, e.Inspect())
	}
	out.WriteString(strings.Join(elems, ", "))
	out.WriteString("]")
	return out.String()
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, p := range h.Pairs {
		pStr := fmt.Sprintf("%s: %s", p.Key.Inspect(), p.Value.Inspect())
		pairs = append(pairs, pStr)
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
