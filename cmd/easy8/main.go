package main

import (
	"os"

	"easy8-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
