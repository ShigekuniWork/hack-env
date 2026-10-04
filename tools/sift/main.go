// Command sift filters lines using command-line conditions.
package main

import (
	"fmt"
	"os"
	"tools/sift/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
