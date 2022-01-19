package syntax

type Node interface {
	// Pos returns the position the node can be found out in the original
	// source.
	Pos() Pos

	// Err reports an error that occurred at the node's position.
	Err(msg string) error
}

type node struct {
	pos Pos
}

func (n node) Pos() Pos { return n.pos }

func (n node) Err(msg string) error { return n.pos.Err(msg) }

// VarDecl for variable declarataions.
// Name = Value
type VarDecl struct {
	node

	Name  *Name
	Value Node
}

// Ref for referring to previous variable declarations.
// $Left
type Ref struct {
	node

	Left Node
}

// DotExpr for selecting fields on an entity. The left most node of this
// expression will always be a Ref.
// $Left.Right
type DotExpr struct {
	node

	Left  Node
	Right Node
}

// IndExpr for accessing items in an indexable type such as an array or an
// object. The left most node of this expression will always be a Ref.
// $Left[Right]
type IndExpr struct {
	node

	Left  Node
	Right Node
}

// ChainExpr for using the output of one command as the first argument to a
// subsequent command.
// Commands[0] -> Commands[1] -> ...
type ChainExpr struct {
	node

	Commands []*CommandStmt
}

// Lit represents either a String, Int, or Bool literal.
// Value
type Lit struct {
	node

	Type  LitType
	Value string
}

// Name represents an identifier, such as a variable name, object key, or
// reference name.
// Value
type Name struct {
	node

	Value string
}

// Array represents an array of operands. Arrays defined by the user will
// contain only a single type, either being a literal, or an array or object.
// [Items[0], Items[1], ...]
type Array struct {
	node

	Items []Node
}

// Object is the type for key-value pairs.
// {Pairs[0], Pairs[1], ...}
type Object struct {
	node

	Pairs []*KeyExpr
}

// KeyExpr for defining a key-value pair within an object. The key of this pair
// is a name and not a string literal.
// Key: Value
type KeyExpr struct {
	node

	Key   *Name
	Value Node
}

// BlockStmt for a list of top-level statements, each separated by a semi.
// {Nodes[0]; Nodes[1]; ...}
type BlockStmt struct {
	node

	Nodes []Node
}

// CommandStmt for the invocation of a command. The arguments passed to the
// command are space separated.
// Name Args[0] Args[1] ...
type CommandStmt struct {
	node

	Name *Name
	Args []Node
}

// CaseStmt for use within a match statement. If the condition of the match
// statement matches the value, the Then block is executed.
// Value -> Then
// Value -> { Then }
type CaseStmt struct {
	node

	Value Node
	Then  Node
}

// MatchStmt for checking the condition against different literals. The default
// condition for a match statement is defined with the name _.
// match Cond {
//     Cases[0];
//     Cases[1];
//     Default;
// }
type MatchStmt struct {
	node

	Cond    Node
	Cases   []*CaseStmt
	Default Node
}

// Operation represents a binary expression.
// Left Op Right
type Operation struct {
	node

	Op    Op
	Left  Node
	Right Node
}

// IfStmt for checking the condition and executing the given Then node if that
// condition evaluates to a truthy value.
// if Cond { Then } else { Else }
type IfStmt struct {
	node

	Cond Node
	Then Node
	Else Node
}

// BranchStmt for break or continue statements within a for-loop.
type BranchStmt struct {
	node

	Tok token
}

// ForStmt for executing a block code multiple times.
// for Cond { Body }
// for Init; Cond; Post { Body }
type ForStmt struct {
	node

	Init Node
	Cond Node
	Post Node
	Body *BlockStmt
}
