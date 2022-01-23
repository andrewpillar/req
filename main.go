package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/andrewpillar/req/eval"
	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/version"
)

func files(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	fnames := make([]string, 0, len(ents))

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}

		if fname := ent.Name(); strings.HasSuffix(fname, ".req") {
			fnames = append(fnames, filepath.Join(dir, fname))
		}
	}

	sort.Strings(fnames)
	return fnames, nil
}

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

func errh(errs chan error) func(syntax.Pos, string) {
	return func(pos syntax.Pos, msg string) {
		errs <- errors.New(pos.String() + " - " + msg)
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

	fnames := make([]string, 0, len(args))

	for _, arg := range args {
		info, err := os.Stat(arg)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
			os.Exit(1)
		}

		if info.IsDir() {
			paths, err := files(info.Name())

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
				os.Exit(1)
			}

			fnames = append(fnames, paths...)
			continue
		}
		fnames = append(fnames, arg)
	}

	sems := make(chan struct{}, runtime.GOMAXPROCS(0)+10)
	errs := make(chan error)

	var wg sync.WaitGroup
	wg.Add(len(fnames))

	for _, fname := range fnames {
		go func(fname string) {
			sems <- struct{}{}
			defer func() {
				wg.Done()
				<-sems
			}()

			nn, err := syntax.ParseFile(fname, errh(errs))

			if err != nil {
				errs <- err
				return
			}

			e := eval.New()

			if err := e.Run(nn); err != nil {
				errs <- err
				return
			}
		}(fname)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	var errc int

	errmax := 50

	for err := range errs {
		if errc < errmax {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		errc++
	}

	if errc > 0 {
		if errc > errmax {
			fmt.Fprintf(os.Stderr, "%s: too many errors\n", argv0)
		}
		os.Exit(1)
	}
}
