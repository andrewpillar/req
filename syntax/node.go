package syntax

import "github.com/andrewpillar/req/token"

type Node interface {
	Pos() token.Pos

	Err(msg string) error
}

type node struct {
	pos token.Pos
}

func (n node) Pos() token.Pos { return n.pos }

func (n node) Err(msg string) error {
	return n.pos.Err(msg)
}

type VarDecl struct {
	node

	Name  *Name
	Value Node
}

type Ref struct {
	node

	Left Node
}

type DotExpr struct {
	node

	Left  Node
	Right Node
}

type IndExpr struct {
	node

	Left  Node
	Right Node
}

type ChainExpr struct {
	node

	Commands []*CommandStmt
}

type Lit struct {
	node

	Type  token.Type
	Value string
}

type Name struct {
	node

	Value string
}

type Array struct {
	node

	Items []Node
}

type Object struct {
	node

	Pairs []*KeyExpr
}

type KeyExpr struct {
	node

	Key   *Name
	Value Node
}

type BlockStmt struct {
	node

	Nodes []Node
}

type CommandStmt struct {
	node

	Name *Name
	Args []Node
}

type CaseStmt struct {
	node

	Value Node
	Then  Node
}

type MatchStmt struct {
	node

	Cond    Node
	Cases   []*CaseStmt
	Default Node
}

type YieldStmt struct {
	node

	Value Node
}

type IfStmt struct {
	node

	Cond Node
	Then Node
	Else Node
}
