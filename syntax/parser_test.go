package syntax

import (
	"path/filepath"
	"testing"

	"github.com/andrewpillar/req/token"
)

func checkNodes(t *testing.T, expected, actual *Node) {
	if expected.Op != actual.Op {
		t.Errorf("%s - unexpected node Op, expected=%q, got=%q\n", actual.Pos, expected.Op, actual.Op)
	}

	if expected.Value != actual.Value {
		t.Errorf("%s - unexpected node Value, expected=%q, got=%q\n", actual.Pos, expected.Value, actual.Value)
	}

	if expected.Body != nil {
		if actual.Body == nil {
			t.Errorf("%s - expected Body for node\n", actual.Pos)
			return
		}
		checkNodes(t, expected.Body, actual.Body)
	}

	if expected.List != nil {
		if actual.List == nil {
			t.Errorf("%s - expected List for node\n", actual.Pos)
			return
		}
		checkNodes(t, expected.List, actual.List)
	}

	if expected.Next != nil {
		if actual.Next == nil {
			t.Errorf("%s - expected Next for node\n", actual.Pos)
			return
		}
		checkNodes(t, expected.Next, actual.Next)
	}

	if expected.Left != nil {
		if actual.Left == nil {
			t.Errorf("%s - expected Left for node\n", actual.Pos)
			return
		}
		checkNodes(t, expected.Left, actual.Left)
	}

	if expected.Right != nil {
		if actual.Right == nil {
			t.Errorf("%s - expected Right for node\n", actual.Pos)
			return
		}
		checkNodes(t, expected.Right, actual.Right)
	}
}

func Test_Parser(t *testing.T) {
	nn, err := ParseFile(filepath.Join("testdata", "gh.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	expected := []*Node{
		{
			Op: OVAR,
			Left: &Node{
				Op:    ONAME,
				Value: "Stdout",
			},
			Right: &Node{
				Op: OOPEN,
				Left: &Node{
					Op: OLIST,
					List: &Node{
						Op:    OLIT,
						Type:  token.String,
						Value: "/dev/stdout",
					},
				},
			},
		},
		{
			Op: OVAR,
			Left: &Node{
				Op:    ONAME,
				Value: "Stderr",
			},
			Right: &Node{
				Op: OOPEN,
				Left: &Node{
					Op: OLIST,
					List: &Node{
						Op:    OLIT,
						Type:  token.String,
						Value: "/dev/stderr",
					},
				},
			},
		},
		{
			Op: OVAR,
			Left: &Node{
				Op:    ONAME,
				Value: "Endpoint",
			},
			Right: &Node{
				Op:    OLIT,
				Type:  token.String,
				Value: "https://api.github.com",
			},
		},
		{
			Op: OVAR,
			Left: &Node{
				Op:    ONAME,
				Value: "Token",
			},
			Right: &Node{
				Op: OENV,
				Left: &Node{
					Op: OLIST,
					List: &Node{
						Op:    OLIT,
						Type:  token.String,
						Value: "GH_TOKEN",
					},
				},
			},
		},
		{
			Op: OVAR,
			Left: &Node{
				Op:    ONAME,
				Value: "Resp",
			},
			Right: &Node{
				Op:    OMETHOD,
				Value: "GET",
				Left: &Node{
					Op: OLIST,
					List: &Node{
						Op: OOBJ,
						Body: &Node{
							Op: OKEY,
							Left: &Node{
								Op:    ONAME,
								Value: "Authorization",
							},
							Right: &Node{
								Op:    OLIT,
								Type:  token.String,
								Value: "Bearer ${Token}",
							},
							Next: &Node{
								Op: OKEY,
								Left: &Node{
									Op:    ONAME,
									Value: "Content-Type",
								},
								Right: &Node{
									Op:    OLIT,
									Type:  token.String,
									Value: "application/json; charset=utf-8",
								},
							},
						},
					},
				},
				Right: &Node{
					Op:    OLIT,
					Type:  token.String,
					Value: "${Endpoint}/user",
				},
			},
		},
		{
			Op: OWRITE,
			Left: &Node{
				Op: OLIST,
				List: &Node{
					Op: OREFDOT,
					Left: &Node{
						Op: OREF,
						Left: &Node{
							Op:    ONAME,
							Value: "Resp",
						},
					},
					Right: &Node{
						Op:    ONAME,
						Value: "Body",
					},
				},
			},
			Right: &Node{
				Op: OMATCH,
				Left: &Node{
					Op: OREFDOT,
					Left: &Node{
						Op: OREF,
						Left: &Node{
							Op:    ONAME,
							Value: "Resp",
						},
					},
					Right: &Node{
						Op:    ONAME,
						Value: "StatusCode",
					},
				},
				Body: &Node{
					Op: OCASE,
					Left: &Node{
						Op:    OLIT,
						Type:  token.Int,
						Value: "200",
					},
					Right: &Node{
						Op: OYIELD,
						Left: &Node{
							Op: OREF,
							Left: &Node{
								Op:    ONAME,
								Value: "Stdout",
							},
						},
					},
					Next: &Node{
						Op: OCASE,
						Left: &Node{
							Op:    ONAME,
							Value: "_",
						},
						Right: &Node{
							Op: OBLOCK,
							Body: &Node{
								Op: OYIELD,
								Left: &Node{
									Op: OREF,
									Left: &Node{
										Op:    ONAME,
										Value: "Stderr",
									},
								},
								Next: &Node{
									Op: OEXIT,
									Left: &Node{
										Op:    OLIT,
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
		checkNodes(t, expected[i], n)
	}
}
