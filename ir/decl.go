package ir

import (
	"github.com/andrewpillar/req/syntax"
)

type VarDecl struct {
	node

	Left  Node
	Right Node
}

func NewVarDecl(n *syntax.Node) *VarDecl {
	return &VarDecl{
		node:  node{pos: n.Pos},
		Left:  rewrite(n.Left),
		Right: rewrite(n.Right),
	}
}
