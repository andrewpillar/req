package syntax

import "github.com/andrewpillar/req/token"

type Op uint

//go:generate stringer -type Op -trimprefix O
const (
	OVAR    Op = iota + 1 // Left = Right
	OREF                  // $Left
	OREFDOT               // $Left.Right
	OREFIND               // $Left[Right]

	OLIT  // Value
	ONAME // Value

	OARR   // [ List ]
	OOBJ   // { Body }
	OKEY   // Left: Right
	OBLOCK // { Body }
	OLIST  // List

	OMETHOD // Value Left -> Right

	OMATCH // match Left { Body }
	OCASE  // Left -> { Right }
	OYIELD // yield Left

	OIF // if Left { Body } else { Next }

	OOPEN  // open Left
	OENV   // env Left
	OEXIT  // exit Left
	OWRITE // write Left -> Right
)

type Node struct {
	Pos   token.Pos
	Op    Op
	Type  token.Type
	Value string
	Body  *Node
	List  *Node
	Next  *Node
	Left  *Node
	Right *Node
}

func (n *Node) insertNext(n2 *Node) {
	if n.Next == nil {
		n.Next = n2
		return
	}
	n.Next.insertNext(n2)
}

func (n *Node) InsertBody(n2 *Node) {
	if n.Body == nil {
		n.Body = n2
		return
	}
	n.Body.insertNext(n2)
}

func (n *Node) InsertList(n2 *Node) {
	if n.List == nil {
		n.List = n2
		return
	}
	n.List.insertNext(n2)
}
