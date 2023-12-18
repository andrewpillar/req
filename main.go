package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/andrewpillar/req/eval"
	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/version"

	"golang.org/x/term"
)

func repl(term *term.Terminal) {
	fmt.Fprintln(term, "req", version.Build)

	in := make(chan string)

	go func() {
		for {
			line, err := term.ReadLine()

			if err != nil {
				close(in)
				return
			}
			in <- line
		}
	}()

	e := eval.New(term)

	var c eval.Context

	for line := range in {
		if line == "" {
			continue
		}

		nn, err := syntax.ParseExpr(line)

		if err != nil {
			fmt.Fprintln(term, err)
			continue
		}

		for _, n := range nn {
			val, err := e.Eval(&c, n)

			if err != nil {
				if evalerr, ok := err.(eval.Error); ok {
					fmt.Fprintln(term, evalerr.Err)
					continue
				}
				fmt.Fprintln(term, err)
				continue
			}

			if val != nil {
				fmt.Fprintln(term, val)
			}
		}
	}
}

func main() {
	argv0 := os.Args[0]

	var showVersion bool

	fs := flag.NewFlagSet(argv0, flag.ExitOnError)
	fs.BoolVar(&showVersion, "version", false, "show version and exit")
	_ = fs.Parse(os.Args[1:])

	if showVersion {
		fmt.Println(version.Build)
		return
	}

	args := fs.Args()

	if len(args) == 0 {
		fd := int(os.Stdin.Fd())

		state, err := term.MakeRaw(fd)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
			os.Exit(1)
		}

		t := term.NewTerminal(os.Stdout, "> ")

		repl(t)
		_ = term.Restore(fd, state)
		fmt.Println()
		return
	}

	nn, err := syntax.ParseFile(args[0], func(pos syntax.Pos, msg string) {
		fmt.Fprintf(os.Stderr, "%s - %s\n", pos, msg)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
		os.Exit(1)
	}

	e := eval.New(os.Stdout)

	if err := e.Run(nn); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
		os.Exit(1)
	}
}
