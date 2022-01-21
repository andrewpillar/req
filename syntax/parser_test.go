package syntax

import (
	"path/filepath"
	"testing"
)

func checkAssignStmt(t *testing.T, expected, actual *AssignStmt) {
	checkNode(t, expected.Left, actual.Left)
	checkNode(t, expected.Right, actual.Right)
}

func checkExprList(t *testing.T, expected, actual *ExprList) {
	if len(expected.Nodes) != len(actual.Nodes) {
		t.Errorf("%s - unexpected ExprList length, expected=%d, got=%d\n", actual.Pos(), len(expected.Nodes), len(actual.Nodes))
		return
	}

	for i, n := range actual.Nodes {
		checkNode(t, expected.Nodes[i], n)
	}
}

func checkRef(t *testing.T, expected, actual *Ref) {
	if expected.Left != nil {
		if actual.Left == nil {
			t.Errorf("%s - expected Left for Ref\n", actual.Pos())
			return
		}
		checkNode(t, expected.Left, actual.Left)
	}
}

func checkDotExpr(t *testing.T, expected, actual *DotExpr) {
	if expected.Left != nil {
		if actual.Left == nil {
			t.Errorf("%s - expected Left for DotExpr\n", actual.Pos())
			return
		}
		checkNode(t, expected.Left, actual.Left)
	}

	if expected.Right != nil {
		if actual.Right == nil {
			t.Errorf("%s - expected Right for DotExpr\n", actual.Pos())
			return
		}
		checkNode(t, expected.Right, actual.Right)
	}
}

func checkIndExpr(t *testing.T, expected, actual *IndExpr) {
	if expected.Left != nil {
		if actual.Left == nil {
			t.Errorf("%s - expected Left for IndExpr\n", actual.Pos())
			return
		}
		checkNode(t, expected.Left, actual.Left)
	}

	if expected.Right != nil {
		if actual.Right == nil {
			t.Errorf("%s - expected Right for IndExpr\n", actual.Pos())
			return
		}
		checkNode(t, expected.Right, actual.Right)
	}
}

func checkLit(t *testing.T, expected, actual *Lit) {
	if expected.Type != actual.Type {
		t.Errorf("%s - unexpected Lit.Type, expected=%q, got=%q\n", actual.Pos(), expected.Type, actual.Type)
	}

	if expected.Value != actual.Value {
		t.Errorf("%s - unexpected Lit.Value, expected=%q, got=%q\n", actual.Pos(), expected.Value, actual.Value)
	}
}

func checkName(t *testing.T, expected, actual *Name) {
	if expected.Value != actual.Value {
		t.Errorf("%s - unexpected Name.Value, expected=%q, got=%q\n", actual.Pos(), expected.Value, actual.Value)
	}
}

func checkArray(t *testing.T, expected, actual *Array) {
	if len(expected.Items) != len(actual.Items) {
		t.Errorf("%s - unexpected Array length, expeced=%d, got=%d\n", actual.Pos(), len(expected.Items), len(actual.Items))
		return
	}

	for i, n := range actual.Items {
		checkNode(t, expected.Items[i], n)
	}
}

func checkObject(t *testing.T, expected, actual *Object) {
	if len(expected.Pairs) != len(actual.Pairs) {
		t.Errorf("%s - unexpected Object.Pairs length, expeced=%d, got=%d\n", actual.Pos(), len(expected.Pairs), len(actual.Pairs))
		return
	}

	for i, n := range actual.Pairs {
		checkNode(t, expected.Pairs[i], n)
	}
}

func checkKeyExpr(t *testing.T, expected, actual *KeyExpr) {
	if expected.Key != nil {
		if actual.Key == nil {
			t.Errorf("%s - expected Key for KeyExpr\n", actual.Pos())
			return
		}
		checkNode(t, expected.Key, actual.Key)
	}

	if expected.Value != nil {
		if actual.Value == nil {
			t.Errorf("%s - expected Value for KeyExpr\n", actual.Pos())
			return
		}
		checkNode(t, expected.Value, actual.Value)
	}
}

func checkBlockStmt(t *testing.T, expected, actual *BlockStmt) {
	if len(expected.Nodes) != len(actual.Nodes) {
		t.Errorf("%s - unexpected BlockStmt length, expected=%d, got=%d\n", actual.Pos(), len(expected.Nodes), len(actual.Nodes))
		return
	}

	for i, n := range actual.Nodes {
		checkNode(t, expected.Nodes[i], n)
	}
}

func checkCommandStmt(t *testing.T, expected, actual *CommandStmt) {
	if expected.Name.Value != actual.Name.Value {
		t.Errorf("%s - unexpected CommandStmt.Name, expected=%q, got=%q\n", actual.Pos(), expected.Name.Value, actual.Name.Value)
		return
	}

	if len(expected.Args) != len(actual.Args) {
		t.Errorf("%s - unexpected CommandStmt.Args length, expected=%d, got=%d\n", actual.Pos(), len(expected.Args), len(actual.Args))
		return
	}

	for i, n := range actual.Args {
		checkNode(t, expected.Args[i], n)
	}
}

func checkChainExpr(t *testing.T, expected, actual *ChainExpr) {
	if len(expected.Commands) != len(actual.Commands) {
		t.Errorf("%s - unexpected ChainExpr.Commands length, expected=%d, got=%d\n", actual.Pos(), len(expected.Commands), len(actual.Commands))
		return
	}

	for i, n := range actual.Commands {
		checkNode(t, expected.Commands[i], n)
	}
}

func checkMatchStmt(t *testing.T, expected, actual *MatchStmt) {
	if expected.Cond != nil {
		if actual.Cond == nil {
			t.Errorf("%s - expected Cond for MatchStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Cond, actual.Cond)
	}

	if expected.Default != nil {
		if expected.Default == nil {
			t.Errorf("%s - expected Default for MatchStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Default, actual.Default)
	}

	if len(expected.Cases) != len(actual.Cases) {
		t.Errorf("%s - unexpected MatchStmt.Cases length, expected=%d, got=%d\n", actual.Pos(), len(expected.Cases), len(actual.Cases))
		return
	}

	for i, n := range expected.Cases {
		checkNode(t, n.Value, actual.Cases[i].Value)
		checkNode(t, n.Then, actual.Cases[i].Then)
	}
}

func checkOperation(t *testing.T, expected, actual *Operation) {
	if expected.Op != actual.Op {
		t.Errorf("%s - unexpected Operation.Op, expected=%q, got=%q\n", actual.Pos(), expected.Op, actual.Op)
		return
	}

	if expected.Left != nil {
		if actual.Left == nil {
			t.Errorf("%s - expected Left of Operation\n", actual.Pos())
			return
		}
		checkNode(t, expected.Left, actual.Left)
	}

	if expected.Right != nil {
		if actual.Right == nil {
			t.Errorf("%s - expected Right of Operation\n", actual.Pos())
			return
		}
		checkNode(t, expected.Right, actual.Right)
	}
}

func checkForStmt(t *testing.T, expected, actual *ForStmt) {
	if expected.Init != nil {
		if actual.Init == nil {
			t.Errorf("%s - expected Init for ForStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Init, actual.Init)
	}

	if expected.Cond != nil {
		if actual.Cond == nil {
			t.Errorf("%s - expected Cond for ForStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Cond, actual.Cond)
	}

	if expected.Post != nil {
		if actual.Post == nil {
			t.Errorf("%s - expected Post for ForStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Post, actual.Post)
	}

	if expected.Body != nil {
		if actual.Body == nil {
			t.Errorf("%s - expected Body for ForStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Body, actual.Body)
	}
}

func checkIfStmt(t *testing.T, expected, actual *IfStmt) {
	if expected.Cond != nil {
		if actual.Cond == nil {
			t.Errorf("%s - expected Cond for IfStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Cond, actual.Cond)
	}

	if expected.Then != nil {
		if actual.Then == nil {
			t.Errorf("%s - expected Then for IfStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Then, actual.Then)
	}

	if expected.Else != nil {
		if actual.Else == nil {
			t.Errorf("%s - expected Else for IfStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Else, actual.Else)
	}
}

func checkNode(t *testing.T, expected, actual Node) {
	switch v := expected.(type) {
	case *ExprList:
		list, ok := actual.(*ExprList)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkExprList(t, v, list)
	case *AssignStmt:
		decl, ok := actual.(*AssignStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkAssignStmt(t, v, decl)
	case *Ref:
		ref, ok := actual.(*Ref)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkRef(t, v, ref)
	case *DotExpr:
		dot, ok := actual.(*DotExpr)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkDotExpr(t, v, dot)
	case *IndExpr:
		ind, ok := actual.(*IndExpr)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkIndExpr(t, v, ind)
	case *Lit:
		lit, ok := actual.(*Lit)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkLit(t, v, lit)
	case *Name:
		name, ok := actual.(*Name)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkName(t, v, name)
	case *Array:
		arr, ok := actual.(*Array)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkArray(t, v, arr)
	case *Object:
		obj, ok := actual.(*Object)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkObject(t, v, obj)
	case *KeyExpr:
		key, ok := actual.(*KeyExpr)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkKeyExpr(t, v, key)
	case *BlockStmt:
		block, ok := actual.(*BlockStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkBlockStmt(t, v, block)
	case *CommandStmt:
		cmd, ok := actual.(*CommandStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkCommandStmt(t, v, cmd)
	case *ChainExpr:
		chain, ok := actual.(*ChainExpr)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkChainExpr(t, v, chain)
	case *MatchStmt:
		match, ok := actual.(*MatchStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkMatchStmt(t, v, match)
	case *Operation:
		op, ok := actual.(*Operation)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkOperation(t, v, op)
	case *IfStmt:
		if_, ok := actual.(*IfStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkIfStmt(t, v, if_)
	case *ForStmt:
		for_, ok := actual.(*ForStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkForStmt(t, v, for_)
	default:
		t.Errorf("%s - unknown node type=%T\n", actual.Pos(), v)
	}
}

func Test_ParseRef(t *testing.T) {
	tests := []struct {
		expr     string
		expected *Ref
	}{
		{
			"$Resp.Body",
			&Ref{
				Left: &DotExpr{
					Left:  &Name{Value: "Resp"},
					Right: &Name{Value: "Body"},
				},
			},
		},
		{
			`$Resp.Header["Content-Type"]`,
			&Ref{
				Left: &IndExpr{
					Left: &DotExpr{
						Left:  &Name{Value: "Resp"},
						Right: &Name{Value: "Header"},
					},
					Right: &Lit{Type: StringLit, Value: "Content-Type"},
				},
			},
		},
		{
			"$Resp.Status.Code",
			&Ref{
				Left: &DotExpr{
					Left: &DotExpr{
						Left:  &Name{Value: "Resp"},
						Right: &Name{Value: "Status"},
					},
					Right: &Name{Value: "Code"},
				},
			},
		},
		{
			`$Hash["Array"][0]`,
			&Ref{
				Left: &IndExpr{
					Left: &IndExpr{
						Left:  &Name{Value: "Hash"},
						Right: &Lit{Type: StringLit, Value: "Array"},
					},
					Right: &Lit{Type: IntLit, Value: "0"},
				},
			},
		},
	}

	for i, test := range tests {
		n, err := ParseRef(test.expr)

		if err != nil {
			t.Errorf("tests[%d] - %s\n", i, err)
			continue
		}

		ref, ok := n.(*Ref)

		if !ok {
			t.Errorf("tests[%d] - unexpected node type, expected=%T, got=%T\n", i, &Ref{}, n)
			continue
		}

		checkRef(t, test.expected, ref)
	}
}

func Test_ParseArray(t *testing.T) {
	nn, err := ParseFile(filepath.Join("testdata", "array.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	expected := []Node{
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "ObjArray"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Array{
						Items: []Node{
							&Object{
								Pairs: []*KeyExpr{
									{
										Key: &Name{Value: "N"},
										Value: &Lit{
											Type:  IntLit,
											Value: "10",
										},
									},
								},
							},
							&Object{
								Pairs: []*KeyExpr{
									{
										Key: &Name{Value: "S"},
										Value: &Lit{
											Type:  StringLit,
											Value: "S",
										},
									},
								},
							},
							&Object{
								Pairs: []*KeyExpr{
									{
										Key: &Name{Value: "T"},
										Value: &Lit{
											Type:  BoolLit,
											Value: "true",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if len(nn) != len(expected) {
		t.Fatalf("node count mismatch, expected=%d, got=%d\n", len(expected), len(nn))
	}

	for i, n := range nn {
		checkNode(t, expected[i], n)
	}
}

func Test_ParseFor(t *testing.T) {
	nn, err := ParseFile(filepath.Join("testdata", "for.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	expected := []Node{
		&ForStmt{},
		&ForStmt{
			Cond: &Lit{
				Type:  BoolLit,
				Value: "true",
			},
		},
		&ForStmt{
			Cond: &Operation{
				Op: AndOp,
				Left: &Lit{
					Type:  BoolLit,
					Value: "true",
				},
				Right: &Lit{
					Type:  BoolLit,
					Value: "true",
				},
			},
		},
		&ForStmt{
			Init: &AssignStmt{
				Left: &ExprList{
					Nodes: []Node{
						&Name{Value: "Line"},
					},
				},
				Right: &ExprList{
					Nodes: []Node{&CommandStmt{
						Name: &Name{Value: "read"},
						Args: []Node{
							&Ref{
								Left: &Name{Value: "F"},
							},
						},
					},
					},
				},
			},
			Cond: &Operation{
				Op: NeqOp,
				Left: &Ref{
					Left: &Name{Value: "Line"},
				},
				Right: &Lit{
					Type: StringLit,
				},
			},
			Post: &AssignStmt{
				Left: &ExprList{
					Nodes: []Node{
						&Name{Value: "Line"},
					},
				},
				Right: &ExprList{
					Nodes: []Node{&CommandStmt{
						Name: &Name{Value: "read"},
						Args: []Node{
							&Ref{
								Left: &Name{Value: "F"},
							},
						},
					},
					},
				},
			},
		},
		&ForStmt{
			Init: &AssignStmt{
				Left: &ExprList{
					Nodes: []Node{
						&Name{Value: "Line"},
					},
				},
				Right: &ExprList{
					Nodes: []Node{
						&CommandStmt{
							Name: &Name{Value: "read"},
							Args: []Node{
								&Ref{
									Left: &Name{Value: "F"},
								},
							},
						},
					},
				},
			},
			Cond: &Operation{
				Op: AndOp,
				Left: &Operation{
					Op: NeqOp,
					Left: &Ref{
						Left: &Name{Value: "Line"},
					},
					Right: &Lit{
						Type: StringLit,
					},
				},
				Right: &Lit{
					Type:  BoolLit,
					Value: "true",
				},
			},
			Post: &AssignStmt{
				Left: &ExprList{
					Nodes: []Node{
						&Name{Value: "Line"},
					},
				},
				Right: &ExprList{
					Nodes: []Node{
						&CommandStmt{
							Name: &Name{Value: "read"},
							Args: []Node{
								&Ref{
									Left: &Name{Value: "F"},
								},
							},
						},
					},
				},
			},
		},
	}

	if len(nn) != len(expected) {
		t.Fatalf("node count mismatch, expected=%d, got=%d\n", len(expected), len(nn))
	}

	for i, n := range nn {
		checkNode(t, expected[i], n)
	}
}

func Test_ParseAssign(t *testing.T) {
	nn, err := ParseFile(filepath.Join("testdata", "assign.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	expected := []Node{
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "Arr"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Array{},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&IndExpr{
						Left: &Name{Value: "Arr"},
						Right: &Array{},
					},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Lit{
						Type:  IntLit,
						Value: "1",
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&IndExpr{
						Left: &Name{Value: "Arr"},
						Right: &Lit{
							Type:  IntLit,
							Value: "0",
						},
					},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Lit{
						Type:  IntLit,
						Value: "2",
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "S"},
					&Name{Value: "I"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Lit{
						Type:  StringLit,
						Value: "string",
					},
					&Lit{
						Type:  IntLit,
						Value: "10",
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "Obj"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Object{
						Pairs: []*KeyExpr{
							{
								Key:   &Name{Value: "Arr"},
								Value: &Array{},
							},
						},
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&IndExpr{
						Left: &IndExpr{
							Left: &Name{Value: "Obj"},
							Right: &Lit{
								Type:  StringLit,
								Value: "Arr",
							},
						},
						Right: &Array{},
					},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Lit{
						Type:  IntLit,
						Value: "1",
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&IndExpr{
						Left: &IndExpr{
							Left:  &Name{Value: "Obj"},
							Right: &Lit{Type: StringLit, Value: "Arr"},
						},
						Right: &Lit{Type: IntLit, Value: "0"},
					},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Lit{Type: IntLit, Value: "2"},
				},
			},
		},
	}

	if len(nn) != len(expected) {
		t.Fatalf("node count mismatch, expected=%d, got=%d\n", len(expected), len(nn))
	}

	for i, n := range nn {
		checkNode(t, expected[i], n)
	}
}

func Test_ParseIf(t *testing.T) {
	nn, err := ParseFile(filepath.Join("testdata", "if.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	expected := []Node{
		&IfStmt{
			Cond: &Lit{
				Type:  BoolLit,
				Value: "true",
			},
		},
		&IfStmt{
			Cond: &Operation{
				Op: OrOp,
				Left: &Operation{
					Op: EqOp,
					Left: &Lit{
						Type:  IntLit,
						Value: "10",
					},
					Right: &Lit{
						Type:  IntLit,
						Value: "11",
					},
				},
				Right: &Lit{
					Type:  BoolLit,
					Value: "true",
				},
			},
		},
		&IfStmt{
			Cond: &Operation{
				Op: AndOp,
				Left: &Operation{
					Op: EqOp,
					Left: &Lit{
						Type:  IntLit,
						Value: "10",
					},
					Right: &Lit{
						Type:  IntLit,
						Value: "10",
					},
				},
				Right: &Operation{
					Op: EqOp,
					Left: &Lit{
						Type:  IntLit,
						Value: "11",
					},
					Right: &Lit{
						Type:  IntLit,
						Value: "11",
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "StatusCode"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Lit{
						Type:  IntLit,
						Value: "204",
					},
				},
			},
		},
		&IfStmt{
			Cond: &Operation{
				Op: AndOp,
				Left: &Operation{
					Op: GeqOp,
					Left: &Ref{
						Left: &Name{Value: "StatusCode"},
					},
					Right: &Lit{
						Type:  IntLit,
						Value: "200",
					},
				},
				Right: &Operation{
					Op: LtOp,
					Left: &Ref{
						Left: &Name{Value: "StatusCode"},
					},
					Right: &Lit{
						Type:  IntLit,
						Value: "300",
					},
				},
			},
			Else: &IfStmt{
				Cond: &Operation{
					Op: AndOp,
					Left: &Operation{
						Op: GeqOp,
						Left: &Ref{
							Left: &Name{Value: "StatusCode"},
						},
						Right: &Lit{
							Type:  IntLit,
							Value: "400",
						},
					},
					Right: &Operation{
						Op: LtOp,
						Left: &Ref{
							Left: &Name{Value: "StatusCode"},
						},
						Right: &Lit{
							Type:  IntLit,
							Value: "500",
						},
					},
				},
				Else: &BlockStmt{},
			},
		},
	}

	for i, n := range nn {
		checkNode(t, expected[i], n)
	}
}

func Test_Parser(t *testing.T) {
	nn, err := ParseFile(filepath.Join("testdata", "gh.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	expected := []Node{
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "Stdout"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&CommandStmt{
						Name: &Name{Value: "open"},
						Args: []Node{
							&Lit{
								Type:  StringLit,
								Value: "/dev/stdout",
							},
						},
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "Stderr"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&CommandStmt{
						Name: &Name{Value: "open"},
						Args: []Node{
							&Lit{
								Type:  StringLit,
								Value: "/dev/stderr",
							},
						},
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "Endpoint"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&Lit{
						Type:  StringLit,
						Value: "https://api.github.com",
					},
				},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "Token"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&CommandStmt{
						Name: &Name{Value: "env"},
						Args: []Node{
							&Lit{
								Type:  StringLit,
								Value: "GH_TOKEN",
							},
						},
					},
				},
			},
		},
		&IfStmt{
			Cond: &Operation{
				Op: EqOp,
				Left: &Ref{
					Left: &Name{Value: "Token"},
				},
				Right: &Lit{Type: StringLit},
			},
		},
		&AssignStmt{
			Left: &ExprList{
				Nodes: []Node{
					&Name{Value: "Resp"},
				},
			},
			Right: &ExprList{
				Nodes: []Node{
					&ChainExpr{
						Commands: []*CommandStmt{
							{
								Name: &Name{Value: "GET"},
								Args: []Node{
									&Lit{
										Type:  StringLit,
										Value: "{$Endpoint}/user",
									},
									&Object{
										Pairs: []*KeyExpr{
											{
												Key: &Name{Value: "Authorization"},
												Value: &Lit{
													Type:  StringLit,
													Value: "Bearer {$Token}",
												},
											},
											{
												Key: &Name{Value: "Content-Type"},
												Value: &Lit{
													Type:  StringLit,
													Value: "application/json; charset=utf-8",
												},
											},
										},
									},
								},
							},
							{
								Name: &Name{Value: "send"},
							},
						},
					},
				},
			},
		},
		&MatchStmt{
			Cond: &Ref{
				Left: &DotExpr{
					Left:  &Name{Value: "Resp"},
					Right: &Name{Value: "StatusCode"},
				},
			},
			Cases: []*CaseStmt{
				{
					Value: &Lit{
						Type:  IntLit,
						Value: "200",
					},
					Then: &BlockStmt{
						Nodes: []Node{
							&AssignStmt{
								Left: &ExprList{
									Nodes: []Node{
										&Name{Value: "User"},
									},
								},
								Right: &ExprList{
									Nodes: []Node{
										&CommandStmt{
											Name: &Name{Value: "decode"},
											Args: []Node{
												&Name{Value: "json"},
												&Ref{
													Left: &DotExpr{
														Left:  &Name{Value: "Resp"},
														Right: &Name{Value: "Body"},
													},
												},
											},
										},
									},
								},
							},
							&CommandStmt{
								Name: &Name{Value: "print"},
								Args: []Node{
									&Lit{
										Type:  StringLit,
										Value: `Hello {$User["login"]}`,
									},
								},
							},
						},
					},
				},
			},
			Default: &BlockStmt{
				Nodes: []Node{
					&CommandStmt{
						Name: &Name{Value: "print"},
						Args: []Node{
							&Ref{
								Left: &DotExpr{
									Left:  &Name{Value: "Resp"},
									Right: &Name{Value: "Body"},
								},
							},
							&Ref{
								Left: &Name{Value: "Stderr"},
							},
						},
					},
				},
			},
		},
	}

	if len(nn) != len(expected) {
		t.Fatalf("node count mismatch, expected=%d, got=%d\n", len(expected), len(nn))
	}

	for i, n := range nn {
		checkNode(t, expected[i], n)
	}
}
