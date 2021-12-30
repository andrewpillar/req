package syntax

import "github.com/andrewpillar/req/token"

type Node interface {
	Pos() token.Pos
}

type node struct {
	pos token.Pos
}

func (n node) Pos() token.Pos {	return n.pos }

type VarDecl struct {
	node

	Ident *Ident
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

type Lit struct {
	node

	Type  token.Type
	Value string
}

type Ident struct {
	node

	Name string
}

type Array struct {
	node

	Items []Node
}

type Object struct {
	node

	objtab map[string]Node
	Body   []Node
}

type KeyExpr struct {
	node

	Key   *Ident
	Value Node
}

type BlockStmt struct {
	node

	Nodes []Node
}

type ActionStmt struct {
	node

	Name string
	Args []Node
	Dest Node
}

type MatchStmt struct {
	node

	Cond   Node
	Jmptab map[uint32]Node
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
