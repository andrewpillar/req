package syntax

import (
	"path/filepath"
	"testing"

	"github.com/andrewpillar/req/token"
)

func checkVarDecl(t *testing.T, expected, actual *VarDecl) {
	if expected.Name != nil {
		if actual.Name == nil {
			t.Errorf("%s - expected Name for VarDecl\n", actual.Pos())
			return
		}
	}

	checkName(t, expected.Name, actual.Name)

	if expected.Value != nil {
		if actual.Value == nil {
			t.Errorf("%s - expected Value for VarDecl\n", actual.Pos())
			return
		}
	}
	checkNode(t, expected.Value, actual.Value)
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

	if len(expected.Cases) != len(actual.Cases) {
		t.Errorf("%s - unexpected MatchStmt.Cases length, expected=%d, got=%d\n", actual.Pos(), len(expected.Cases), len(actual.Cases))
		return
	}

	for i, n := range expected.Cases {
		checkNode(t, n.Value, actual.Cases[i].Value)
		checkNode(t, n.Then, actual.Cases[i].Then)
	}
}

func checkYieldStmt(t *testing.T, expected, actual *YieldStmt) {
	if expected.Value != nil {
		if actual.Value == nil {
			t.Errorf("%s - expected Value for YieldStmt\n", actual.Pos())
			return
		}
		checkNode(t, expected.Value, actual.Value)
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
	case *VarDecl:
		decl, ok := actual.(*VarDecl)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkVarDecl(t, v, decl)
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
	case *YieldStmt:
		yield, ok := actual.(*YieldStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkYieldStmt(t, v, yield)
	case *IfStmt:
		if_, ok := actual.(*IfStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkIfStmt(t, v, if_)
	default:
		t.Errorf("%s - unknown node type=%T\n", actual.Pos(), v)
	}
}

func Test_Parser(t *testing.T) {
	nn, err := ParseFile(filepath.Join("testdata", "gh.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	expected := []Node{
		&VarDecl{
			Name:  &Name{Value: "Stdout"},
			Value: &CommandStmt{
				Name: &Name{Value: "open"},
				Args: []Node{
					&Lit{
						Type:  token.String,
						Value: "/dev/stdout",
					},
				},
			},
		},
		&VarDecl{
			Name:  &Name{Value: "Stderr"},
			Value: &CommandStmt{
				Name: &Name{Value: "open"},
				Args: []Node{
					&Lit{
						Type:  token.String,
						Value: "/dev/stderr",
					},
				},
			},
		},
		&VarDecl{
			Name:  &Name{Value: "Endpoint"},
			Value: &Lit{
				Type:  token.String,
				Value: "https://api.github.com",
			},
		},
		&VarDecl{
			Name:  &Name{Value: "Token"},
			Value: &CommandStmt{
				Name: &Name{Value: "env"},
				Args: []Node{
					&Lit{
						Type:  token.String,
						Value: "GH_TOKEN",
					},
				},
			},
		},
		&VarDecl{
			Name:  &Name{Value: "Resp"},
			Value: &ChainExpr{
				Commands: []*CommandStmt{
					&CommandStmt{
						Name: &Name{Value: "GET"},
						Args: []Node{
							&Lit{
								Type: token.String,
								Value: "{$Endpoint}/user",
							},
							&Object{
								Pairs: []*KeyExpr{
									&KeyExpr{
										Key: &Name{Value: "Authorization"},
										Value: &Lit{
											Type: token.String,
											Value: "Bearer {$Token}",
										},
									},
									&KeyExpr{
										Key: &Name{Value: "Content-Type"},
										Value: &Lit{
											Type: token.String,
											Value: "application/json; charset=utf-8",
										},
									},
								},
							},
						},
					},
					&CommandStmt{
						Name: &Name{Value: "send"},
					},
				},
			},
		},
		&CommandStmt{
			Name: &Name{Value: "print"},
			Args: []Node{
				&Ref{
					Left: &DotExpr{
						Left:  &Name{Value: "Resp"},
						Right: &Name{Value: "Body"},
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
						&CaseStmt{
							Value: &Lit{
								Type:  token.Int,
								Value: "200",
							},
							Then: &YieldStmt{
								Value: &Ref{
									Left: &Name{Value: "Stdout"},
								},
							},
						},
						&CaseStmt{
							Value: &Name{Value: "_"},
							Then: &BlockStmt{
								Nodes: []Node{
									&YieldStmt{
										Value: &Ref{
											Left: &Name{Value: "Stderr"},
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
