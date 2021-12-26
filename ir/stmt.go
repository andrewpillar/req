package ir

type MatchStmt struct {
	node

	Cond   Node
	Jmptab map[int]Node
}

type IfStmt struct {
	node

	Cond Node
	Then Node
	Else Node
}
