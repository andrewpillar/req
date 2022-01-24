package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/andrewpillar/req/eval"
	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/version"
)

func repl(ctx context.Context, w io.Writer, r io.Reader) {
	sc := bufio.NewScanner(r)

	e := eval.New()

	var c eval.Context

	fmt.Println("req", version.Build)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("> ")

			if !sc.Scan() {
				return
			}

			if err := sc.Err(); err != nil {
				fmt.Fprintln(w, "ERR", err)
				continue
			}

			line := sc.Text()

			if line == "" {
				continue
			}

			nn, err := syntax.ParseExpr(line)

			if err != nil {
				fmt.Fprintln(w, err)
				continue
			}

			for _, n := range nn {
				val, err := e.Eval(&c, n)

				if err != nil {
					if evalerr, ok := err.(eval.Error); ok {
						fmt.Fprintln(w, evalerr.Err)
						continue
					}
					fmt.Fprintln(w, err)
					continue
				}

				if val != nil {
					fmt.Fprintln(w, val.String())
				}
			}
		}
	}
}

func main() {
	argv0 := os.Args[0]

	var showVersion bool

	fs := flag.NewFlagSet(argv0, flag.ExitOnError)
	fs.BoolVar(&showVersion, "version", false, "show version and exit")
	fs.Parse(os.Args[1:])

	if showVersion {
		fmt.Println(version.Build)
		return
	}

	args := fs.Args()

	if len(args) == 0 {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := make(chan os.Signal, 1)

		signal.Notify(ch, os.Interrupt)

		go repl(ctx, os.Stdout, os.Stdin)

		<-ch
		cancel()
		return
	}

	nn, err := syntax.ParseFile(args[0], func(pos syntax.Pos, msg string) {
		fmt.Fprintf(os.Stderr, "%s - %s\n", pos, msg)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
		os.Exit(1)
	}

	e := eval.New()

	if err := e.Run(nn); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
		os.Exit(1)
	}
}
