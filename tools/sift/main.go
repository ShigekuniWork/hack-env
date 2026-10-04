// Command sift filters lines using command-line conditions.
package main

import (
	"os"
	"tools/sift/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		cli.PrintError(os.Args[1:], err, os.Stderr)
		os.Exit(1)
	}
}
