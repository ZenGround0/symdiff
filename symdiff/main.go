package main

import (
	"bufio"
	"os"
	. "github.com/zenground0/symdiff"
	"fmt"
)


func main() {
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
}
