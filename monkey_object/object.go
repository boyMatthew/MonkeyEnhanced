package monkey_object

import "fmt"

type ObjectType string

const (
	DECIMAL_OBJ ObjectType = "DECIMAL"
	BOOLEAN_OBJ            = "BOOLEAN"
	NULL_OBJ               = "NULL"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Decimal struct {
	Value float64
}

func (d *Decimal) Inspect() string  { return fmt.Sprintf("%f", d.Value) }
func (d *Decimal) Type() ObjectType { return DECIMAL_OBJ }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL_OBJ }
