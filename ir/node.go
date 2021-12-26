package ir

import (
	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/token"
)

type node struct {
	pos token.Pos
}

func (n node) Pos() token.Pos { return n.pos }

type Node interface {
	Pos() token.Pos
}

func rewrite(n *Node) (Node, error) {
	var ir Node

	switch n.Op {
	case syntax.OVAR:
		ir = NewVarDecl(n)
	default:
		panic("cannot rewrite " + n.Op.String())
	}
	return ir, nil
}

func Rewrite(nn []*syntax.Node) ([]Node, error) {
	ir := make([]Node, 0, len(nn))

	for _, n := range nn {
		n2, err := rewrite(n)

		if err != nil {
			panic(err)
		}
		ir = appennd(ir, n2)
	}
	return ir, nil
}
