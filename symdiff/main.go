package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	. "github.com/zenground0/symdiff"
)

func main() {

	cmds := []*cli.Command{
		replCmd,
		ddxCmd,
		simplifyCmd,
	}
	app := &cli.App{
		Name:     "symdiff",
		Usage:    "s-expression based symbolic differentiation",
		Version:  "0",
		Commands: cmds,
	}

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		fmt.Printf("error running symdiff: %s\n", err)
		os.Exit(1)
		return
	}

}

var replCmd = &cli.Command{
	Name:  "repl",
	Usage: "d/dx, simplify, print loop",
	Action: func(cctx *cli.Context) error {
		fmt.Printf("\nd/dx, simplify, print\n")
		bio := bufio.NewReader(os.Stdin)
		// rough repl
		for {
			// prompt
			fmt.Printf("\nd/dx ")
			// await user input
			input, err := bio.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading user input %s\n", err)
				continue
			}
			// parse
			var sexp SExp
			if err := sexp.Parse(input); err != nil {
				fmt.Printf("Error parsing user input as sexp: %s\n", err)
				continue
			}
			var poly PolyExp
			if err := poly.Parse(sexp); err != nil {
				fmt.Printf("Error parsing user input as polynomial: %s\n", err)
				continue
			}

			// differentiate in x
			d, err := Differentiate(Symbol("x"), poly)
			if err != nil {
				fmt.Printf("Error taking derivative: %s\n", err)
				continue
			}
			// simplify
			s, err := Simplify(*d)
			if err != nil {
				fmt.Printf("Error simplifying expression %s: %s", d.ToSExp().String(), err)
				continue
			}

			// return value
			prettyString, err := RainbowParens(s.ToSExp().String(), Rainbow)
			if err != nil {
				fmt.Printf("Error formatting output: %s\n", err)
			}
			fmt.Printf("%s\n", prettyString)
		}
	},
}

var ddxCmd = &cli.Command{
	Name:        "d/dx",
	Description: "Take derivative in bound variable x",
	Usage:       "d/dx <poly expr>",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 1 {
			return fmt.Errorf("invalid arguments to d/dx")
		}
		var sexp SExp
		if err := sexp.Parse(cctx.Args().First()); err != nil {
			return fmt.Errorf("error parsing user input as sexp: %s", err)
		}
		var poly PolyExp
		if err := poly.Parse(sexp); err != nil {
			return fmt.Errorf("error parsing user input as polynomial: %s", err)
		}

		d, err := Differentiate(Symbol("x"), poly)
		if err != nil {
			return fmt.Errorf("error taking derivative: %s", err)
		}
		prettyString, err := RainbowParens(d.ToSExp().String(), Rainbow)
		if err != nil {
			fmt.Printf("Error formatting output: %s", err)
		}
		fmt.Printf("%s\n", prettyString)
		return nil
	},
}

var simplifyCmd = &cli.Command{
	Name:        "simplify",
	Description: "Run polynomial simplification",
	Usage:       "simplify <poly expr>",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 1 {
			return fmt.Errorf("invalid arguments to simplify")
		}
		var sexp SExp
		if err := sexp.Parse(cctx.Args().First()); err != nil {
			return fmt.Errorf("error parsing user input as sexp: %s", err)
		}
		var poly PolyExp
		if err := poly.Parse(sexp); err != nil {
			return fmt.Errorf("error parsing user input as polynomial: %s", err)
		}
		// simplify
		s, err := Simplify(poly)
		if err != nil {
			return fmt.Errorf("error simplifying expression %s: %s", poly.ToSExp().String(), err)
		}

		prettyString, err := RainbowParens(s.ToSExp().String(), Rainbow)
		if err != nil {
			fmt.Printf("Error formatting output: %s", err)
		}
		fmt.Printf("%s\n", prettyString)
		return nil
	},
}
