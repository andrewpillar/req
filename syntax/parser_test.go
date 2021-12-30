package syntax

import (
	"path/filepath"
	"testing"

	"github.com/andrewpillar/req/token"
)

func checkVarDecl(t *testing.T, expected, actual *VarDecl) {
	if expected.Ident != nil {
		if actual.Ident == nil {
			t.Errorf("%s - expected Ident for VarDecl\n", actual.Pos())
			return
		}
	}

	checkIdent(t, expected.Ident, actual.Ident)

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

func checkIdent(t *testing.T, expected, actual *Ident) {
	if expected.Name != actual.Name {
		t.Errorf("%s - unexpected Ident.Name, expected=%q, got=%q\n", actual.Pos(), expected.Name, actual.Name)
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
	if len(expected.Body) != len(actual.Body) {
		t.Errorf("%s - unexpected Object length, expeced=%d, got=%d\n", actual.Pos(), len(expected.Body), len(actual.Body))
		return
	}

	for i, n := range actual.Body {
		checkNode(t, expected.Body[i], n)
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
		t.Errorf("%s - unexpected BlockStmt length, expected=%q, got=%d\n", actual.Pos(), len(expected.Nodes), len(actual.Nodes))
		return
	}

	for i, n := range actual.Nodes {
		checkNode(t, expected.Nodes[i], n)
	}
}

func checkActionStmt(t *testing.T, expected, actual *ActionStmt) {
	if expected.Name != actual.Name {
		t.Errorf("%s - unexpected Action.Name, expected=%q, got=%q\n", actual.Pos(), expected.Name, actual.Name)
		return
	}

	if len(expected.Args) != len(actual.Args) {
		t.Errorf("%s - unexpected Action.Args length, expected=%d, got=%d\n", actual.Pos(), len(expected.Args), len(actual.Args))
		return
	}

	for i, n := range actual.Args {
		checkNode(t, expected.Args[i], n)
	}

	if expected.Dest != nil {
		if actual.Dest == nil {
			t.Errorf("%s - expected Dest for Action\n", actual.Pos())
			return
		}
		checkNode(t, expected.Dest, actual.Dest)
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

	if len(expected.Jmptab) != len(actual.Jmptab) {
		t.Errorf("%s - unexpected MatchStmt.Jmptab length, expected=%d, got=%d\n", actual.Pos(), len(expected.Jmptab), len(actual.Jmptab))
		return
	}

	for k, n := range expected.Jmptab {
		if _, ok := actual.Jmptab[k]; !ok {
			t.Errorf("%s - could not find key %d in Jmptab\n", n.Pos(), k)
			continue
		}
		checkNode(t, n, actual.Jmptab[k])
	}

//	for k, n := range actual.Jmptab {
//		if _, ok := expected.Jmptab[k]; !ok {
//			t.Errorf("%s - could not find key %d in Jmptab\n", n.Pos(), k)
//			continue
//		}
//		checkNode(t, expected.Jmptab[k], n)
//	}
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
	case *Ident:
		ident, ok := actual.(*Ident)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkIdent(t, v, ident)
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
	case *ActionStmt:
		action, ok := actual.(*ActionStmt)

		if !ok {
			t.Errorf("%s - unexpected node type, expected=%T, got=%T\n", actual.Pos(), v, actual)
			return
		}
		checkActionStmt(t, v, action)
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
			Ident: &Ident{Name: "Stdout"},
			Value: &ActionStmt{
				Name: "open",
				Args: []Node{
					&Lit{
						Type:  token.String,
						Value: "/dev/stdout",
					},
				},
			},
		},
		&VarDecl{
			Ident: &Ident{Name: "Stderr"},
			Value: &ActionStmt{
				Name: "open",
				Args: []Node{
					&Lit{
						Type:  token.String,
						Value: "/dev/stderr",
					},
				},
			},
		},
		&VarDecl{
			Ident: &Ident{Name: "Endpoint"},
			Value: &Lit{
				Type:  token.String,
				Value: "https://api.github.com",
			},
		},
		&VarDecl{
			Ident: &Ident{Name: "Token"},
			Value: &ActionStmt{
				Name: "env",
				Args: []Node{
					&Lit{
						Type:  token.String,
						Value: "GH_TOKEN",
					},
				},
			},
		},
		&VarDecl{
			Ident: &Ident{Name: "Resp"},
			Value: &ActionStmt{
				Name: "GET",
				Args: []Node{
					&Object{
						Body: []Node{
							&KeyExpr{
								Key: &Ident{Name: "Authorization"},
								Value: &Lit{
									Type: token.String,
									Value: "Bearer ${Token}",
								},
							},
							&KeyExpr{
								Key: &Ident{Name: "Content-Type"},
								Value: &Lit{
									Type: token.String,
									Value: "application/json; charset=utf-8",
								},
							},
						},
					},
				},
				Dest: &Lit{
					Type: token.String,
					Value: "${Endpoint}/user",
				},
			},
		},
		&ActionStmt{
			Name: "write",
			Args: []Node{
				&Ref{
					Left: &DotExpr{
						Left:  &Ident{Name: "Resp"},
						Right: &Ident{Name: "Body"},
					},
				},
			},
			Dest: &MatchStmt{
				Cond: &Ref{
					Left: &DotExpr{
						Left:  &Ident{Name: "Resp"},
						Right: &Ident{Name: "StatusCode"},
					},
				},
				Jmptab: map[uint32]Node{
					1859371669: &YieldStmt{
						Value: &Ref{Left: &Ident{Name: "Stdout"}},
					},
					84696384: &BlockStmt{
						Nodes: []Node{
							&YieldStmt{
								Value: &Ref{Left: &Ident{Name: "Stderr"}},
							},
							&ActionStmt{
								Name: "exit",
								Args: []Node{
									&Lit{
										Type:  token.Int,
										Value: "1",
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
