package value

import "github.com/andrewpillar/req/syntax"

// Name is the value for an identifier, such as a variable declaration or
// key.
type Name struct {
	Value string
}

// ToName attempts to type assert the given value to a name.
func ToName(v Value) (Name, error) {
	n, ok := v.(Name)

	if !ok {
		return Name{}, typeError(v.valueType(), nameType)
	}
	return n, nil
}

func (n Name) String() string {
	return n.Value
}

func (n Name) Sprint() string {
	return n.String()
}

func (n Name) valueType() valueType {
	return nameType
}

func (n Name) cmp(op syntax.Op, _ Value) (Value, error) {
	return nil, opError(op, nameType)
}
